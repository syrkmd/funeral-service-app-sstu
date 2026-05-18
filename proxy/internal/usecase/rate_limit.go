package usecase

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/pkg/ipmatch"
)

type RateLimitUseCase struct {
	repo          RateLimitStateRepository
	config        ConfigProvider
	logger        UseCaseLogger
	now           func() time.Time
	compiledRules atomic.Value
	keyLocks      [256]sync.Mutex
}

func NewRateLimitUseCase(repo RateLimitStateRepository, provider ConfigProvider, logger UseCaseLogger) *RateLimitUseCase {
	useCase := &RateLimitUseCase{
		repo:   repo,
		config: provider,
		logger: logger,
		now:    time.Now,
	}

	useCase.compiledRules.Store(compiledRateLimitConfig{})
	useCase.applyConfig(provider.Current())

	provider.Subscribe(func(cfg config.Config) {
		useCase.applyConfig(cfg)
	})

	return useCase
}

func (u *RateLimitUseCase) CheckRateLimit(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
	ip := strings.TrimSpace(rawIP)
	compiled := u.compiledRules.Load().(compiledRateLimitConfig)

	if !compiled.Enabled {
		return domain.RateLimitDecision{
			IP:      ip,
			Allowed: true,
			Reason:  "rate_limit_disabled",
		}, nil
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return domain.RateLimitDecision{}, fmt.Errorf("%w: %s", domain.ErrInvalidIP, rawIP)
	}
	addr = addr.Unmap()

	rules := u.resolveRules(addr, compiled)
	acquiredConnectionKeys := make([]string, 0, len(rules))

	for _, rule := range rules {
		allowed, limitType, currentValue, acquiredKey, err := u.consume(ctx, rule, normalizeTrafficBytes(uploadBytes), 0)
		if err != nil {
			_ = u.ReleaseConnections(ctx, acquiredConnectionKeys)
			return domain.RateLimitDecision{}, err
		}
		if acquiredKey != "" {
			acquiredConnectionKeys = append(acquiredConnectionKeys, acquiredKey)
		}
		if !allowed {
			_ = u.ReleaseConnections(ctx, acquiredConnectionKeys)
			return domain.RateLimitDecision{
				IP:             ip,
				Allowed:        false,
				Reason:         "rate_limit",
				RuleID:         rule.ID,
				RuleValue:      rule.Value,
				LimitType:      limitType,
				CurrentValue:   currentValue,
				ConnectionKeys: nil,
			}, nil
		}
	}

	return domain.RateLimitDecision{
		IP:             ip,
		Allowed:        true,
		Reason:         "rate_limit_ok",
		ConnectionKeys: acquiredConnectionKeys,
	}, nil
}

func (u *RateLimitUseCase) ReserveResponseBandwidth(ctx context.Context, rawIP string, downloadBytes int64) (domain.RateLimitDecision, error) {
	ip := strings.TrimSpace(rawIP)
	compiled := u.compiledRules.Load().(compiledRateLimitConfig)

	if !compiled.Enabled {
		return domain.RateLimitDecision{
			IP:      ip,
			Allowed: true,
			Reason:  "rate_limit_disabled",
		}, nil
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return domain.RateLimitDecision{}, fmt.Errorf("%w: %s", domain.ErrInvalidIP, rawIP)
	}
	addr = addr.Unmap()

	rules := u.resolveRules(addr, compiled)
	for _, rule := range rules {
		allowed, limitType, currentValue, err := u.consumeTraffic(ctx, rule, 0, normalizeTrafficBytes(downloadBytes))
		if err != nil {
			return domain.RateLimitDecision{}, err
		}
		if !allowed {
			return domain.RateLimitDecision{
				IP:           ip,
				Allowed:      false,
				Reason:       "rate_limit",
				RuleID:       rule.ID,
				RuleValue:    rule.Value,
				LimitType:    limitType,
				CurrentValue: currentValue,
			}, nil
		}
	}

	return domain.RateLimitDecision{
		IP:      ip,
		Allowed: true,
		Reason:  "rate_limit_ok",
	}, nil
}

func (u *RateLimitUseCase) AccountTraffic(ctx context.Context, rawIP string, uploadBytes int64, downloadBytes int64) error {
	uploadBytes = normalizeTrafficBytes(uploadBytes)
	downloadBytes = normalizeTrafficBytes(downloadBytes)
	if uploadBytes == 0 && downloadBytes == 0 {
		return nil
	}

	ip := strings.TrimSpace(rawIP)
	compiled := u.compiledRules.Load().(compiledRateLimitConfig)
	if !compiled.Enabled {
		return nil
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidIP, rawIP)
	}
	addr = addr.Unmap()

	rules := u.resolveRules(addr, compiled)
	for _, rule := range rules {
		if _, _, _, err := u.consumeTraffic(ctx, rule, uploadBytes, downloadBytes); err != nil {
			return err
		}
	}

	return nil
}

func (u *RateLimitUseCase) ListRules(_ context.Context) ([]domain.RateLimitRule, error) {
	compiled := u.compiledRules.Load().(compiledRateLimitConfig)
	rules := make([]domain.RateLimitRule, 0, 1+len(compiled.Subnets))

	if hasLimits(compiled.DefaultRule) {
		rules = append(rules, compiled.DefaultRule)
	}
	for _, subnetRule := range compiled.Subnets {
		rules = append(rules, subnetRule.Rule)
	}

	return rules, nil
}

func (u *RateLimitUseCase) ListBuckets(ctx context.Context) ([]domain.RateLimitBucketSnapshot, error) {
	return u.repo.SnapshotBuckets(ctx)
}

func (u *RateLimitUseCase) resolveRules(addr netip.Addr, compiled compiledRateLimitConfig) []domain.RateLimitRule {
	rules := make([]domain.RateLimitRule, 0, 1+len(compiled.Subnets))

	if hasLimits(compiled.DefaultRule) {
		ipRule := compiled.DefaultRule
		ipRule.Value = addr.String()
		rules = append(rules, ipRule)
	}

	for _, subnetRule := range compiled.Subnets {
		if !subnetRule.Matcher.Match(addr) {
			continue
		}
		rules = append(rules, subnetRule.Rule)
	}

	return rules
}

func (u *RateLimitUseCase) consume(ctx context.Context, rule domain.RateLimitRule, uploadBytes int64, downloadBytes int64) (bool, string, int, string, error) {
	if rule.CPS > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "cps"), float64(rule.CPS), float64(rule.CPS), time.Second)
		if err != nil || !allowed {
			return allowed, "cps", currentValue, "", err
		}
	}

	var acquiredKey string
	if rule.MaxConnections > 0 {
		key := bucketKey(rule, "connections")
		allowed, currentValue, err := u.repo.TryAcquireConnection(ctx, key, rule.MaxConnections)
		if err != nil || !allowed {
			return allowed, "connections", currentValue, "", err
		}
		acquiredKey = key
	}

	if rule.RPS > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rps"), float64(rule.RPS), float64(rule.RPS), time.Second)
		if err != nil || !allowed {
			if acquiredKey != "" {
				_ = u.repo.ReleaseConnection(ctx, acquiredKey)
			}
			return allowed, "rps", currentValue, "", err
		}
	}

	if rule.RPM > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rpm"), float64(rule.RPM), float64(rule.RPM), time.Minute)
		if err != nil || !allowed {
			if acquiredKey != "" {
				_ = u.repo.ReleaseConnection(ctx, acquiredKey)
			}
			return allowed, "rpm", currentValue, "", err
		}
	}

	if rule.RPH > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rph"), float64(rule.RPH), float64(rule.RPH), time.Hour)
		if err != nil || !allowed {
			if acquiredKey != "" {
				_ = u.repo.ReleaseConnection(ctx, acquiredKey)
			}
			return allowed, "rph", currentValue, "", err
		}
	}

	if rule.RPD > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rpd"), float64(rule.RPD), float64(rule.RPD), 24*time.Hour)
		if err != nil || !allowed {
			if acquiredKey != "" {
				_ = u.repo.ReleaseConnection(ctx, acquiredKey)
			}
			return allowed, "rpd", currentValue, "", err
		}
	}

	allowed, limitType, currentValue, err := u.consumeTraffic(ctx, rule, uploadBytes, downloadBytes)
	if err != nil || !allowed {
		if acquiredKey != "" {
			_ = u.repo.ReleaseConnection(ctx, acquiredKey)
		}
		return allowed, limitType, currentValue, "", err
	}

	return true, "", 0, acquiredKey, nil
}

func (u *RateLimitUseCase) consumeTraffic(ctx context.Context, rule domain.RateLimitRule, uploadBytes int64, downloadBytes int64) (bool, string, int, error) {
	if uploadBytes > 0 && rule.UploadBPS > 0 {
		allowed, currentValue, err := u.takeAmount(ctx, bucketKey(rule, "upload"), float64(rule.UploadBPS), float64(rule.UploadBPS), time.Second, float64(uploadBytes))
		if err != nil || !allowed {
			return allowed, "upload", currentValue, err
		}
	}

	if downloadBytes > 0 && rule.DownloadBPS > 0 {
		allowed, currentValue, err := u.takeAmount(ctx, bucketKey(rule, "download"), float64(rule.DownloadBPS), float64(rule.DownloadBPS), time.Second, float64(downloadBytes))
		if err != nil || !allowed {
			return allowed, "download", currentValue, err
		}
	}

	totalBytes := uploadBytes + downloadBytes
	if totalBytes > 0 && rule.TotalBytes > 0 && rule.TotalWindow > 0 {
		allowed, currentValue, err := u.takeAmount(ctx, bucketKey(rule, "total"), float64(rule.TotalBytes), float64(rule.TotalBytes), rule.TotalWindow, float64(totalBytes))
		if err != nil || !allowed {
			return allowed, "total", currentValue, err
		}
	}

	return true, "", 0, nil
}

func (u *RateLimitUseCase) take(ctx context.Context, key string, capacity float64, refillTokens float64, period time.Duration) (bool, int, error) {
	return u.takeAmount(ctx, key, capacity, refillTokens, period, 1)
}

func (u *RateLimitUseCase) takeAmount(ctx context.Context, key string, capacity float64, refillTokens float64, period time.Duration, amount float64) (bool, int, error) {
	if amount <= 0 {
		return true, 0, nil
	}

	lock := &u.keyLocks[lockIndex(key)]
	lock.Lock()
	defer lock.Unlock()

	now := u.now().UTC()
	state, exists, err := u.repo.GetBucket(ctx, key)
	if err != nil {
		return false, 0, err
	}

	if !exists {
		state = domain.BucketState{
			Tokens:     capacity,
			LastRefill: now,
		}
	}

	elapsed := now.Sub(state.LastRefill)
	if elapsed > 0 {
		state.Tokens = math.Min(capacity, state.Tokens+(elapsed.Seconds()/period.Seconds())*refillTokens)
		state.LastRefill = now
	}

	if state.Tokens < amount {
		if err := u.repo.SaveBucket(ctx, key, state); err != nil {
			return false, 0, err
		}
		return false, usedTokens(capacity, state.Tokens), nil
	}

	state.Tokens -= amount
	state.LastRefill = now

	if err := u.repo.SaveBucket(ctx, key, state); err != nil {
		return false, 0, err
	}

	return true, usedTokens(capacity, state.Tokens), nil
}

func (u *RateLimitUseCase) ReleaseConnections(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if err := u.repo.ReleaseConnection(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func bucketKey(rule domain.RateLimitRule, dimension string) string {
	return fmt.Sprintf("%s:%s:%s", rule.Scope, dimension, rule.Value)
}

type compiledSubnetRule struct {
	Rule    domain.RateLimitRule
	Matcher domain.IPMatcher
}

type compiledRateLimitConfig struct {
	Enabled     bool
	DefaultRule domain.RateLimitRule
	Subnets     []compiledSubnetRule
}

func (u *RateLimitUseCase) applyConfig(cfg config.Config) {
	compiled := compiledRateLimitConfig{
		Enabled: cfg.RateLimit.Enabled,
		DefaultRule: domain.RateLimitRule{
			ID:             "default-ip",
			Scope:          domain.RateLimitScopeIP,
			RPS:            cfg.RateLimit.RPS,
			RPM:            cfg.RateLimit.RPM,
			RPH:            cfg.RateLimit.RPH,
			RPD:            cfg.RateLimit.RPD,
			CPS:            cfg.RateLimit.CPS,
			MaxConnections: cfg.RateLimit.MaxConnections,
			UploadBPS:      cfg.RateLimit.UploadBPS,
			DownloadBPS:    cfg.RateLimit.DownloadBPS,
			TotalBytes:     cfg.RateLimit.TotalBytes,
			TotalWindow:    cfg.RateLimit.TotalWindow.Duration,
		},
		Subnets: make([]compiledSubnetRule, 0, len(cfg.RateLimit.Subnets)),
	}

	for idx, subnet := range cfg.RateLimit.Subnets {
		matcher, err := ipmatch.Parse(subnet.CIDR)
		if err != nil {
			if u.logger != nil {
				u.logger.Error("invalid subnet in rate limit config", err)
			}
			continue
		}

		ruleID := strings.TrimSpace(subnet.ID)
		if ruleID == "" {
			ruleID = fmt.Sprintf("subnet-%d", idx+1)
		}

		compiled.Subnets = append(compiled.Subnets, compiledSubnetRule{
			Rule: domain.RateLimitRule{
				ID:             ruleID,
				Scope:          domain.RateLimitScopeSubnet,
				Value:          subnet.CIDR,
				RPS:            subnet.RPS,
				RPM:            subnet.RPM,
				RPH:            subnet.RPH,
				RPD:            subnet.RPD,
				CPS:            subnet.CPS,
				MaxConnections: subnet.MaxConnections,
				UploadBPS:      subnet.UploadBPS,
				DownloadBPS:    subnet.DownloadBPS,
				TotalBytes:     subnet.TotalBytes,
				TotalWindow:    subnet.TotalWindow.Duration,
				Description:    subnet.Description,
			},
			Matcher: matcher,
		})
	}

	u.compiledRules.Store(compiled)
}

func usedTokens(capacity, tokens float64) int {
	value := int(math.Ceil(capacity - tokens))
	if value < 0 {
		return 0
	}
	return value
}

func hasLimits(rule domain.RateLimitRule) bool {
	return rule.RPS > 0 || rule.RPM > 0 || rule.RPH > 0 || rule.RPD > 0 || rule.CPS > 0 || rule.MaxConnections > 0 || rule.UploadBPS > 0 || rule.DownloadBPS > 0 || rule.TotalBytes > 0
}

func normalizeTrafficBytes(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func lockIndex(key string) uint32 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return hasher.Sum32() % uint32(len((RateLimitUseCase{}).keyLocks))
}
