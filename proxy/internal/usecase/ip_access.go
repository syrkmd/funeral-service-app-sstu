package usecase

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type AddRuleInput struct {
	ID          string
	Type        string
	Value       string
	Description string
}

type cachedDecision struct {
	version   uint64
	decision  domain.AccessDecision
	expiresAt time.Time
}

type decisionCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	cache *lru.Cache[string, cachedDecision]
}

func newDecisionCache(size int, ttl time.Duration) (*decisionCache, error) {
	cache, err := lru.New[string, cachedDecision](size)
	if err != nil {
		return nil, err
	}

	return &decisionCache{ttl: ttl, cache: cache}, nil
}

func (c *decisionCache) Get(key string) (cachedDecision, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.cache.Get(key)
	if !ok {
		return cachedDecision{}, false
	}
	if !value.expiresAt.IsZero() && time.Now().UTC().After(value.expiresAt) {
		c.cache.Remove(key)
		return cachedDecision{}, false
	}
	return value, true
}

func (c *decisionCache) Add(key string, value cachedDecision) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ttl > 0 {
		value.expiresAt = time.Now().UTC().Add(c.ttl)
	}
	c.cache.Add(key, value)
}

func (c *decisionCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

type IPAccessUseCase struct {
	repo         IPAccessRepository
	verifiedRepo VerifiedIPRepository
	cache        *decisionCache
	idSequence   atomic.Uint64
}

func NewIPAccessUseCase(repo IPAccessRepository, verifiedRepo VerifiedIPRepository, cacheSize int) *IPAccessUseCase {
	cache, err := newDecisionCache(cacheSize, 5*time.Minute)
	if err != nil {
		panic(err)
	}

	return &IPAccessUseCase{
		repo:         repo,
		verifiedRepo: verifiedRepo,
		cache:        cache,
	}
}

func (u *IPAccessUseCase) SetDecisionCacheTTL(ttl time.Duration) {
	u.cache.SetTTL(ttl)
}

func (u *IPAccessUseCase) ListRules(ctx context.Context) ([]domain.IPRule, error) {
	return u.repo.ListRules(ctx)
}

func (u *IPAccessUseCase) AddRule(ctx context.Context, input AddRuleInput) (domain.IPRule, error) {
	listType := domain.ListType(strings.ToLower(strings.TrimSpace(input.Type)))
	switch listType {
	case domain.ListTypeAllow, domain.ListTypeDeny, domain.ListTypeGray:
	default:
		return domain.IPRule{}, domain.ErrInvalidRuleType
	}

	value := strings.TrimSpace(input.Value)
	if value == "" {
		return domain.IPRule{}, domain.ErrEmptyRuleValue
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("runtime-%d", u.idSequence.Add(1))
	}

	return u.repo.AddRule(ctx, domain.IPRule{
		ID:          id,
		Type:        listType,
		Value:       value,
		Description: strings.TrimSpace(input.Description),
	})
}

func (u *IPAccessUseCase) DeleteRule(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.ErrRuleNotFound
	}
	return u.repo.DeleteRule(ctx, strings.TrimSpace(id))
}

func (u *IPAccessUseCase) CheckIP(ctx context.Context, rawIP string) (domain.AccessDecision, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(rawIP))
	if err != nil {
		return domain.AccessDecision{}, fmt.Errorf("%w: %s", domain.ErrInvalidIP, rawIP)
	}

	addr = addr.Unmap()

	snapshot, err := u.repo.GetSnapshot(ctx)
	if err != nil {
		return domain.AccessDecision{}, err
	}

	cacheKey := addr.String()
	if cached, ok := u.cache.Get(cacheKey); ok && cached.version == snapshot.Version {
		return cached.decision, nil
	}

	decision, cacheable, err := u.evaluateSnapshot(ctx, addr, snapshot)
	if err != nil {
		return domain.AccessDecision{}, err
	}
	if cacheable {
		u.cache.Add(cacheKey, cachedDecision{
			version:  snapshot.Version,
			decision: decision,
		})
	}

	return decision, nil
}

func (u *IPAccessUseCase) VerifyCaptcha(ctx context.Context, rawIP string, answer string) error {
	if strings.TrimSpace(answer) != "1234" {
		return domain.ErrInvalidCaptcha
	}

	addr, err := netip.ParseAddr(strings.TrimSpace(rawIP))
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidIP, rawIP)
	}

	return u.verifiedRepo.MarkVerified(ctx, addr.Unmap().String())
}

func (u *IPAccessUseCase) evaluateSnapshot(ctx context.Context, addr netip.Addr, snapshot domain.AccessSnapshot) (domain.AccessDecision, bool, error) {
	ip := addr.String()

	if rule, ok := snapshot.DenyLookup.Match(addr); ok {
		return domain.AccessDecision{
			IP:            ip,
			Allowed:       false,
			Decision:      "deny",
			Reason:        "matched denylist rule",
			MatchedRuleID: rule.Rule.ID,
			MatchedValue:  rule.Rule.Value,
		}, true, nil
	}

	if rule, ok := snapshot.AllowLookup.Match(addr); ok {
		return domain.AccessDecision{
			IP:            ip,
			Allowed:       true,
			Decision:      "allow",
			Reason:        "matched allowlist rule",
			MatchedRuleID: rule.Rule.ID,
			MatchedValue:  rule.Rule.Value,
		}, true, nil
	}

	if rule, ok := snapshot.GrayLookup.Match(addr); ok {
		verified, err := u.verifiedRepo.IsVerified(ctx, ip)
		if err != nil {
			return domain.AccessDecision{}, false, err
		}
		if verified {
			return domain.AccessDecision{
				IP:            ip,
				Allowed:       true,
				Decision:      "allow",
				Reason:        "matched graylist rule with active verification",
				MatchedRuleID: rule.Rule.ID,
				MatchedValue:  rule.Rule.Value,
			}, false, nil
		}

		return domain.AccessDecision{
			IP:                   ip,
			Allowed:              false,
			Decision:             "captcha_required",
			Reason:               "captcha verification required",
			VerificationRequired: true,
			MatchedRuleID:        rule.Rule.ID,
			MatchedValue:         rule.Rule.Value,
		}, false, nil
	}

	return defaultDecision(ip, snapshot.DefaultPolicy, "default policy", "", ""), true, nil
}

func defaultDecision(ip string, policy domain.DefaultPolicy, reason, ruleID, ruleValue string) domain.AccessDecision {
	allowed := policy == domain.DefaultPolicyAllow
	decision := "deny"
	if allowed {
		decision = "allow"
	}

	return domain.AccessDecision{
		IP:            ip,
		Allowed:       allowed,
		Decision:      decision,
		Reason:        reason,
		MatchedRuleID: ruleID,
		MatchedValue:  ruleValue,
	}
}
