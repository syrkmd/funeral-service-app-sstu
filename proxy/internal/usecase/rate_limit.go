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

	"proxy/internal/config"
	"proxy/internal/domain"
	"proxy/pkg/ipmatch"

	"github.com/rs/zerolog/log"
)

type RateLimitUseCase struct {
	repo          RateLimitStateRepository
	config        ConfigProvider
	now           func() time.Time
	compiledRules atomic.Value
	keyLocks      [256]sync.Mutex
}

func NewRateLimitUseCase(repo RateLimitStateRepository, provider ConfigProvider) *RateLimitUseCase {
	useCase := &RateLimitUseCase{
		repo:   repo,
		config: provider,
		now:    time.Now,
	}

	useCase.compiledRules.Store(compiledRateLimitConfig{})
	useCase.applyConfig(provider.Current())

	provider.Subscribe(func(cfg config.Config) {
		useCase.applyConfig(cfg)
	})

	return useCase
}

func (u *RateLimitUseCase) CheckRateLimit(ctx context.Context, rawIP string) (domain.RateLimitDecision, error) {
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
		allowed, limitType, currentValue, err := u.consume(ctx, rule)
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

func (u *RateLimitUseCase) consume(ctx context.Context, rule domain.RateLimitRule) (bool, string, int, error) {
	if rule.RPS > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rps"), float64(rule.RPS), float64(rule.RPS), time.Second)
		if err != nil || !allowed {
			return allowed, "rps", currentValue, err
		}
	}

	if rule.RPM > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rpm"), float64(rule.RPM), float64(rule.RPM), time.Minute)
		if err != nil || !allowed {
			return allowed, "rpm", currentValue, err
		}
	}

	if rule.RPH > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rph"), float64(rule.RPH), float64(rule.RPH), time.Hour)
		if err != nil || !allowed {
			return allowed, "rph", currentValue, err
		}
	}

	if rule.RPD > 0 {
		allowed, currentValue, err := u.take(ctx, bucketKey(rule, "rpd"), float64(rule.RPD), float64(rule.RPD), 24*time.Hour)
		if err != nil || !allowed {
			return allowed, "rpd", currentValue, err
		}
	}

	return true, "", 0, nil
}

func (u *RateLimitUseCase) take(ctx context.Context, key string, capacity float64, refillTokens float64, period time.Duration) (bool, int, error) {
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

	if state.Tokens < 1 {
		if err := u.repo.SaveBucket(ctx, key, state); err != nil {
			return false, 0, err
		}
		return false, usedTokens(capacity, state.Tokens), nil
	}

	state.Tokens--
	state.LastRefill = now

	if err := u.repo.SaveBucket(ctx, key, state); err != nil {
		return false, 0, err
	}

	return true, usedTokens(capacity, state.Tokens), nil
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
			ID:    "default-ip",
			Scope: domain.RateLimitScopeIP,
			RPS:   cfg.RateLimit.RPS,
			RPM:   cfg.RateLimit.RPM,
			RPH:   cfg.RateLimit.RPH,
			RPD:   cfg.RateLimit.RPD,
		},
		Subnets: make([]compiledSubnetRule, 0, len(cfg.RateLimit.Subnets)),
	}

	for idx, subnet := range cfg.RateLimit.Subnets {
		matcher, err := ipmatch.Parse(subnet.CIDR)
		if err != nil {
			log.Error().Err(err).Msg("invalid subnet in rate limit config")
			continue
		}

		ruleID := strings.TrimSpace(subnet.ID)
		if ruleID == "" {
			ruleID = fmt.Sprintf("subnet-%d", idx+1)
		}

		compiled.Subnets = append(compiled.Subnets, compiledSubnetRule{
			Rule: domain.RateLimitRule{
				ID:          ruleID,
				Scope:       domain.RateLimitScopeSubnet,
				Value:       subnet.CIDR,
				RPS:         subnet.RPS,
				RPM:         subnet.RPM,
				RPH:         subnet.RPH,
				RPD:         subnet.RPD,
				Description: subnet.Description,
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
	return rule.RPS > 0 || rule.RPM > 0 || rule.RPH > 0 || rule.RPD > 0
}

func lockIndex(key string) uint32 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return hasher.Sum32() % uint32(len((RateLimitUseCase{}).keyLocks))
}
