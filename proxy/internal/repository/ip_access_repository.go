package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"proxy/internal/config"
	"proxy/internal/domain"
	"proxy/pkg/ipmatch"
)

type IPAccessRepository struct {
	mu sync.RWMutex

	systemRules        map[string]domain.IPRule
	runtimeRules       map[string]domain.IPRule
	disabledSystemRule map[string]struct{}

	defaultPolicy atomic.Value
	version       atomic.Uint64
	snapshot      atomic.Value
}

func NewIPAccessRepository(cfg config.Config) (*IPAccessRepository, error) {
	repo := &IPAccessRepository{
		systemRules:        make(map[string]domain.IPRule),
		runtimeRules:       make(map[string]domain.IPRule),
		disabledSystemRule: make(map[string]struct{}),
	}

	if err := repo.ApplyConfig(cfg); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *IPAccessRepository) ApplyConfig(cfg config.Config) error {
	rules := make(map[string]domain.IPRule, len(cfg.Access.Lists))
	for idx, rule := range cfg.Access.Lists {
		if strings.TrimSpace(rule.ID) == "" {
			rule.ID = fmt.Sprintf("config-%d", idx+1)
		}

		if _, exists := rules[rule.ID]; exists {
			return fmt.Errorf("%w: %s", domain.ErrDuplicateRuleID, rule.ID)
		}

		rules[rule.ID] = normalizeRule(rule)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.systemRules = rules
	r.defaultPolicy.Store(cfg.Access.DefaultPolicy)

	if err := r.rebuildSnapshotLocked(); err != nil {
		return err
	}

	return nil
}

func (r *IPAccessRepository) ListRules(_ context.Context) ([]domain.IPRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	merged := r.collectRulesLocked()
	return merged, nil
}

func (r *IPAccessRepository) AddRule(_ context.Context, rule domain.IPRule) (domain.IPRule, error) {
	normalized := normalizeRule(rule)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.runtimeRules[normalized.ID]; exists {
		return domain.IPRule{}, fmt.Errorf("%w: %s", domain.ErrDuplicateRuleID, normalized.ID)
	}
	if _, exists := r.systemRules[normalized.ID]; exists {
		delete(r.disabledSystemRule, normalized.ID)
		return domain.IPRule{}, fmt.Errorf("%w: %s", domain.ErrDuplicateRuleID, normalized.ID)
	}

	r.runtimeRules[normalized.ID] = normalized
	if err := r.rebuildSnapshotLocked(); err != nil {
		delete(r.runtimeRules, normalized.ID)
		return domain.IPRule{}, err
	}

	return normalized, nil
}

func (r *IPAccessRepository) DeleteRule(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.runtimeRules[id]; exists {
		delete(r.runtimeRules, id)
		return r.rebuildSnapshotLocked()
	}

	if _, exists := r.systemRules[id]; exists {
		r.disabledSystemRule[id] = struct{}{}
		return r.rebuildSnapshotLocked()
	}

	return domain.ErrRuleNotFound
}

func (r *IPAccessRepository) GetSnapshot(_ context.Context) (domain.AccessSnapshot, error) {
	value := r.snapshot.Load()
	if value == nil {
		return domain.AccessSnapshot{}, domain.ErrNilAccessSnapshot
	}

	return value.(domain.AccessSnapshot), nil
}

func (r *IPAccessRepository) rebuildSnapshotLocked() error {
	compiled := make([]domain.CompiledIPRule, 0, len(r.systemRules)+len(r.runtimeRules))
	for _, rule := range r.collectRulesLocked() {
		matcher, err := ipmatch.Parse(rule.Value)
		if err != nil {
			return fmt.Errorf("compile rule %s: %w", rule.ID, err)
		}

		compiled = append(compiled, domain.CompiledIPRule{
			Rule:    rule,
			Matcher: matcher,
		})
	}

	snapshot := domain.AccessSnapshot{
		Version:       r.version.Add(1),
		DefaultPolicy: r.defaultPolicy.Load().(domain.DefaultPolicy),
	}

	for _, rule := range compiled {
		switch rule.Rule.Type {
		case domain.ListTypeDeny:
			snapshot.DenyRules = append(snapshot.DenyRules, rule)
		case domain.ListTypeAllow:
			snapshot.AllowRules = append(snapshot.AllowRules, rule)
		case domain.ListTypeGray:
			snapshot.GrayRules = append(snapshot.GrayRules, rule)
		default:
			return domain.ErrInvalidRuleType
		}
	}

	r.snapshot.Store(snapshot)
	return nil
}

func (r *IPAccessRepository) collectRulesLocked() []domain.IPRule {
	rules := make([]domain.IPRule, 0, len(r.systemRules)+len(r.runtimeRules))

	ids := make([]string, 0, len(r.systemRules))
	for id := range r.systemRules {
		if _, disabled := r.disabledSystemRule[id]; disabled {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		rules = append(rules, r.systemRules[id])
	}

	ids = ids[:0]
	for id := range r.runtimeRules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		rules = append(rules, r.runtimeRules[id])
	}

	return rules
}

func normalizeRule(rule domain.IPRule) domain.IPRule {
	rule.ID = strings.TrimSpace(rule.ID)
	rule.Value = strings.TrimSpace(rule.Value)
	rule.Description = strings.TrimSpace(rule.Description)
	rule.Type = domain.ListType(strings.ToLower(strings.TrimSpace(string(rule.Type))))
	return rule
}
