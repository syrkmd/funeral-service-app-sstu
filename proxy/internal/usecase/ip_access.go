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

	"proxy/internal/domain"
)

type AddRuleInput struct {
	ID          string
	Type        string
	Value       string
	Description string
}

type cachedDecision struct {
	version  uint64
	decision domain.AccessDecision
}

type decisionCache struct {
	mu    sync.Mutex
	cache *lru.Cache[string, cachedDecision]
}

func newDecisionCache(size int) (*decisionCache, error) {
	cache, err := lru.New[string, cachedDecision](size)
	if err != nil {
		return nil, err
	}

	return &decisionCache{cache: cache}, nil
}

func (c *decisionCache) Get(key string) (cachedDecision, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache.Get(key)
}

func (c *decisionCache) Add(key string, value cachedDecision) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Add(key, value)
}

type IPAccessUseCase struct {
	repo       IPAccessRepository
	cache      *decisionCache
	idSequence atomic.Uint64
}

func NewIPAccessUseCase(repo IPAccessRepository, cacheSize int) *IPAccessUseCase {
	cache, err := newDecisionCache(cacheSize)
	if err != nil {
		panic(err)
	}

	return &IPAccessUseCase{
		repo:  repo,
		cache: cache,
	}
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

	decision := evaluateSnapshot(addr, snapshot)
	u.cache.Add(cacheKey, cachedDecision{
		version:  snapshot.Version,
		decision: decision,
	})

	return decision, nil
}

func evaluateSnapshot(addr netip.Addr, snapshot domain.AccessSnapshot) domain.AccessDecision {
	ip := addr.String()

	for _, rule := range snapshot.DenyRules {
		if rule.Matcher.Match(addr) {
			return domain.AccessDecision{
				IP:            ip,
				Allowed:       false,
				Decision:      "deny",
				Reason:        "matched denylist rule",
				MatchedRuleID: rule.Rule.ID,
				MatchedValue:  rule.Rule.Value,
			}
		}
	}

	for _, rule := range snapshot.AllowRules {
		if rule.Matcher.Match(addr) {
			return domain.AccessDecision{
				IP:            ip,
				Allowed:       true,
				Decision:      "allow",
				Reason:        "matched allowlist rule",
				MatchedRuleID: rule.Rule.ID,
				MatchedValue:  rule.Rule.Value,
			}
		}
	}

	for _, rule := range snapshot.GrayRules {
		if rule.Matcher.Match(addr) {
			return defaultDecision(ip, snapshot.DefaultPolicy, "matched graylist rule", rule.Rule.ID, rule.Rule.Value)
		}
	}

	return defaultDecision(ip, snapshot.DefaultPolicy, "default policy", "", "")
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

func decisionLogFields(decision domain.AccessDecision) map[string]any {
	return map[string]any{
		"ip":        decision.IP,
		"decision":  decision.Decision,
		"reason":    decision.Reason,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	}
}
