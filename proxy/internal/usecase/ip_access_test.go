package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/pkg/ipmatch"
)

type mockIPAccessRepository struct {
	mu sync.Mutex

	listRulesFn   func(context.Context) ([]domain.IPRule, error)
	addRuleFn     func(context.Context, domain.IPRule) (domain.IPRule, error)
	deleteRuleFn  func(context.Context, string) error
	getSnapshotFn func(context.Context) (domain.AccessSnapshot, error)

	listRulesCalls   int
	addRuleCalls     []domain.IPRule
	deleteRuleCalls  []string
	getSnapshotCalls int
}

func (m *mockIPAccessRepository) ListRules(ctx context.Context) ([]domain.IPRule, error) {
	m.mu.Lock()
	m.listRulesCalls++
	fn := m.listRulesFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return nil, nil
}

func (m *mockIPAccessRepository) AddRule(ctx context.Context, rule domain.IPRule) (domain.IPRule, error) {
	m.mu.Lock()
	m.addRuleCalls = append(m.addRuleCalls, rule)
	fn := m.addRuleFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, rule)
	}
	return rule, nil
}

func (m *mockIPAccessRepository) DeleteRule(ctx context.Context, id string) error {
	m.mu.Lock()
	m.deleteRuleCalls = append(m.deleteRuleCalls, id)
	fn := m.deleteRuleFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return nil
}

func (m *mockIPAccessRepository) GetSnapshot(ctx context.Context) (domain.AccessSnapshot, error) {
	m.mu.Lock()
	m.getSnapshotCalls++
	fn := m.getSnapshotFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return domain.AccessSnapshot{}, domain.ErrNilAccessSnapshot
}

type mockVerifiedIPRepository struct {
	mu sync.Mutex

	verified  map[string]time.Time
	ttl       time.Duration
	now       func() time.Time
	isCalls   []string
	markCalls []string
	isErr     error
	markErr   error
}

func newMockVerifiedIPRepository(ttl time.Duration) *mockVerifiedIPRepository {
	return &mockVerifiedIPRepository{
		verified: make(map[string]time.Time),
		ttl:      ttl,
		now:      time.Now,
	}
}

func (m *mockVerifiedIPRepository) IsVerified(_ context.Context, ip string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isCalls = append(m.isCalls, ip)
	if m.isErr != nil {
		return false, m.isErr
	}
	expiresAt, exists := m.verified[ip]
	if !exists {
		return false, nil
	}
	if m.now().UTC().After(expiresAt) {
		delete(m.verified, ip)
		return false, nil
	}
	return true, nil
}

func (m *mockVerifiedIPRepository) MarkVerified(_ context.Context, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markCalls = append(m.markCalls, ip)
	if m.markErr != nil {
		return m.markErr
	}
	m.verified[ip] = m.now().UTC().Add(m.ttl)
	return nil
}

type countingLookup struct {
	matches atomic.Int64
	next    domain.IPRuleLookup
}

func (l *countingLookup) Match(addr netip.Addr) (domain.CompiledIPRule, bool) {
	l.matches.Add(1)
	if l.next == nil {
		return domain.CompiledIPRule{}, false
	}
	return l.next.Match(addr)
}

func buildAccessSnapshot(t *testing.T, version uint64, policy domain.DefaultPolicy, rules []domain.IPRule) domain.AccessSnapshot {
	t.Helper()

	denyLookup := ipmatch.NewRuleSet[domain.CompiledIPRule]()
	allowLookup := ipmatch.NewRuleSet[domain.CompiledIPRule]()
	grayLookup := ipmatch.NewRuleSet[domain.CompiledIPRule]()

	snapshot := domain.AccessSnapshot{
		Version:       version,
		DefaultPolicy: policy,
		DenyLookup:    denyLookup,
		AllowLookup:   allowLookup,
		GrayLookup:    grayLookup,
	}

	for idx, rule := range rules {
		matcher, err := ipmatch.Parse(rule.Value)
		if err != nil {
			t.Fatalf("parse matcher for rule %s: %v", rule.ID, err)
		}
		prefixes, err := ipmatch.Prefixes(rule.Value)
		if err != nil {
			t.Fatalf("parse prefixes for rule %s: %v", rule.ID, err)
		}
		compiled := domain.CompiledIPRule{
			Rule:    rule,
			Matcher: matcher,
		}

		switch rule.Type {
		case domain.ListTypeDeny:
			snapshot.DenyRules = append(snapshot.DenyRules, compiled)
			denyLookup.InsertPrefixes(prefixes, idx, compiled)
		case domain.ListTypeAllow:
			snapshot.AllowRules = append(snapshot.AllowRules, compiled)
			allowLookup.InsertPrefixes(prefixes, idx, compiled)
		case domain.ListTypeGray:
			snapshot.GrayRules = append(snapshot.GrayRules, compiled)
			grayLookup.InsertPrefixes(prefixes, idx, compiled)
		default:
			t.Fatalf("unexpected rule type: %s", rule.Type)
		}
	}

	return snapshot
}

func wrapLookupCounters(snapshot domain.AccessSnapshot) (domain.AccessSnapshot, *countingLookup, *countingLookup, *countingLookup) {
	deny := &countingLookup{next: snapshot.DenyLookup}
	allow := &countingLookup{next: snapshot.AllowLookup}
	gray := &countingLookup{next: snapshot.GrayLookup}
	snapshot.DenyLookup = deny
	snapshot.AllowLookup = allow
	snapshot.GrayLookup = gray
	return snapshot, deny, allow, gray
}

func TestIPAccessUseCaseDecisionTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ip            string
		defaultPolicy domain.DefaultPolicy
		rules         []domain.IPRule
		verified      bool
		wantAllowed   bool
		wantDecision  string
		wantReason    string
		wantRuleID    string
		wantVerify    bool
	}{
		{
			name:          "blacklist deny exact ipv4",
			ip:            "10.0.0.1",
			defaultPolicy: domain.DefaultPolicyAllow,
			rules: []domain.IPRule{
				{ID: "deny-exact", Type: domain.ListTypeDeny, Value: "10.0.0.1"},
			},
			wantAllowed:  false,
			wantDecision: "deny",
			wantReason:   "matched denylist rule",
			wantRuleID:   "deny-exact",
		},
		{
			name:          "whitelist allow cidr",
			ip:            "192.168.1.42",
			defaultPolicy: domain.DefaultPolicyDeny,
			rules: []domain.IPRule{
				{ID: "allow-cidr", Type: domain.ListTypeAllow, Value: "192.168.1.0/24"},
			},
			wantAllowed:  true,
			wantDecision: "allow",
			wantReason:   "matched allowlist rule",
			wantRuleID:   "allow-cidr",
		},
		{
			name:          "graylist challenge range",
			ip:            "172.16.0.15",
			defaultPolicy: domain.DefaultPolicyAllow,
			rules: []domain.IPRule{
				{ID: "gray-range", Type: domain.ListTypeGray, Value: "172.16.0.10-172.16.0.20"},
			},
			wantAllowed:  false,
			wantDecision: "captcha_required",
			wantReason:   "captcha verification required",
			wantRuleID:   "gray-range",
			wantVerify:   true,
		},
		{
			name:          "default allow",
			ip:            "203.0.113.1",
			defaultPolicy: domain.DefaultPolicyAllow,
			wantAllowed:   true,
			wantDecision:  "allow",
			wantReason:    "default policy",
		},
		{
			name:          "default deny",
			ip:            "203.0.113.2",
			defaultPolicy: domain.DefaultPolicyDeny,
			wantAllowed:   false,
			wantDecision:  "deny",
			wantReason:    "default policy",
		},
		{
			name:          "blacklist precedence over whitelist",
			ip:            "10.0.0.42",
			defaultPolicy: domain.DefaultPolicyAllow,
			rules: []domain.IPRule{
				{ID: "allow-net", Type: domain.ListTypeAllow, Value: "10.0.0.0/8"},
				{ID: "deny-net", Type: domain.ListTypeDeny, Value: "10.0.0.0/8"},
			},
			wantAllowed:  false,
			wantDecision: "deny",
			wantReason:   "matched denylist rule",
			wantRuleID:   "deny-net",
		},
		{
			name:          "overlapping allow exact inside deny cidr still deny",
			ip:            "10.1.1.1",
			defaultPolicy: domain.DefaultPolicyAllow,
			rules: []domain.IPRule{
				{ID: "deny-broad", Type: domain.ListTypeDeny, Value: "10.0.0.0/8"},
				{ID: "allow-exact", Type: domain.ListTypeAllow, Value: "10.1.1.1"},
			},
			wantAllowed:  false,
			wantDecision: "deny",
			wantReason:   "matched denylist rule",
			wantRuleID:   "deny-broad",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snapshot := buildAccessSnapshot(t, 1, tt.defaultPolicy, tt.rules)
			repo := &mockIPAccessRepository{
				getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
					return snapshot, nil
				},
			}
			verified := newMockVerifiedIPRepository(time.Minute)
			if tt.verified {
				verified.verified[tt.ip] = verified.now().UTC().Add(time.Minute)
			}

			useCase := NewIPAccessUseCase(repo, verified, 32)
			decision, err := useCase.CheckIP(context.Background(), tt.ip)
			if err != nil {
				t.Fatalf("CheckIP error: %v", err)
			}
			if decision.Allowed != tt.wantAllowed || decision.Decision != tt.wantDecision || decision.Reason != tt.wantReason || decision.MatchedRuleID != tt.wantRuleID || decision.VerificationRequired != tt.wantVerify {
				t.Fatalf("unexpected decision: got %+v", decision)
			}
		})
	}
}

func TestIPAccessUseCaseInvalidIPHandling(t *testing.T) {
	t.Parallel()

	repo := &mockIPAccessRepository{}
	verified := newMockVerifiedIPRepository(time.Minute)
	useCase := NewIPAccessUseCase(repo, verified, 8)

	if _, err := useCase.CheckIP(context.Background(), "not-an-ip"); err == nil || !strings.Contains(err.Error(), domain.ErrInvalidIP.Error()) {
		t.Fatalf("expected invalid ip error, got %v", err)
	}
}

func TestIPAccessUseCaseRuleManagementTable(t *testing.T) {
	t.Parallel()

	existingRules := []domain.IPRule{{ID: "allow-1", Type: domain.ListTypeAllow, Value: "127.0.0.1"}}

	tests := []struct {
		name string
		run  func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository)
	}{
		{
			name: "list rules",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				repo.listRulesFn = func(context.Context) ([]domain.IPRule, error) {
					return existingRules, nil
				}
				rules, err := useCase.ListRules(context.Background())
				if err != nil {
					t.Fatalf("ListRules error: %v", err)
				}
				if len(rules) != 1 || rules[0].ID != "allow-1" {
					t.Fatalf("unexpected rules: %+v", rules)
				}
			},
		},
		{
			name: "add rule normalizes type and generates id",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				repo.addRuleFn = func(_ context.Context, rule domain.IPRule) (domain.IPRule, error) {
					return rule, nil
				}
				rule, err := useCase.AddRule(context.Background(), AddRuleInput{
					Type:        " AllowList ",
					Value:       " 192.168.1.1 ",
					Description: " desc ",
				})
				if err != nil {
					t.Fatalf("AddRule error: %v", err)
				}
				if !strings.HasPrefix(rule.ID, "runtime-") {
					t.Fatalf("expected generated runtime id, got %q", rule.ID)
				}
				if rule.Type != domain.ListTypeAllow || rule.Value != "192.168.1.1" || rule.Description != "desc" {
					t.Fatalf("unexpected normalized rule: %+v", rule)
				}
				if len(repo.addRuleCalls) != 1 {
					t.Fatalf("expected one repo AddRule call, got %d", len(repo.addRuleCalls))
				}
			},
		},
		{
			name: "unsupported rule type rejected",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				_, err := useCase.AddRule(context.Background(), AddRuleInput{Type: "unsupported", Value: "1.1.1.1"})
				if !errors.Is(err, domain.ErrInvalidRuleType) {
					t.Fatalf("expected invalid rule type, got %v", err)
				}
				if len(repo.addRuleCalls) != 0 {
					t.Fatalf("expected repo not called on invalid type, got %d calls", len(repo.addRuleCalls))
				}
			},
		},
		{
			name: "empty rule value rejected",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				_, err := useCase.AddRule(context.Background(), AddRuleInput{Type: "allowlist", Value: "  "})
				if !errors.Is(err, domain.ErrEmptyRuleValue) {
					t.Fatalf("expected empty rule value error, got %v", err)
				}
				if len(repo.addRuleCalls) != 0 {
					t.Fatalf("expected repo not called on empty value, got %d calls", len(repo.addRuleCalls))
				}
			},
		},
		{
			name: "duplicate rule handling propagated",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				repo.addRuleFn = func(_ context.Context, rule domain.IPRule) (domain.IPRule, error) {
					return domain.IPRule{}, fmt.Errorf("%w: %s", domain.ErrDuplicateRuleID, rule.ID)
				}
				_, err := useCase.AddRule(context.Background(), AddRuleInput{ID: "dup", Type: "allowlist", Value: "1.1.1.1"})
				if !errors.Is(err, domain.ErrDuplicateRuleID) {
					t.Fatalf("expected duplicate rule error, got %v", err)
				}
			},
		},
		{
			name: "delete rule",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				repo.deleteRuleFn = func(_ context.Context, id string) error {
					if id != "rule-1" {
						t.Fatalf("unexpected delete id %q", id)
					}
					return nil
				}
				if err := useCase.DeleteRule(context.Background(), " rule-1 "); err != nil {
					t.Fatalf("DeleteRule error: %v", err)
				}
				if len(repo.deleteRuleCalls) != 1 {
					t.Fatalf("expected one delete call, got %d", len(repo.deleteRuleCalls))
				}
			},
		},
		{
			name: "delete empty id rejected",
			run: func(t *testing.T, useCase *IPAccessUseCase, repo *mockIPAccessRepository) {
				err := useCase.DeleteRule(context.Background(), " ")
				if !errors.Is(err, domain.ErrRuleNotFound) {
					t.Fatalf("expected rule not found, got %v", err)
				}
				if len(repo.deleteRuleCalls) != 0 {
					t.Fatalf("expected repo not called on empty id, got %d calls", len(repo.deleteRuleCalls))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockIPAccessRepository{}
			verified := newMockVerifiedIPRepository(time.Minute)
			useCase := NewIPAccessUseCase(repo, verified, 8)
			tt.run(t, useCase, repo)
		})
	}
}

func TestIPAccessUseCaseDecisionCacheHitMissIsolationAndVersionInvalidation(t *testing.T) {
	t.Parallel()

	base := buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, nil)
	base, denyCount, allowCount, grayCount := wrapLookupCounters(base)

	current := atomic.Value{}
	current.Store(base)
	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return current.Load().(domain.AccessSnapshot), nil
		},
	}

	useCase := NewIPAccessUseCase(repo, newMockVerifiedIPRepository(time.Minute), 32)
	useCase.SetDecisionCacheTTL(time.Minute)

	first, err := useCase.CheckIP(context.Background(), "127.0.0.1")
	if err != nil || !first.Allowed {
		t.Fatalf("expected first decision allowed, got decision=%+v err=%v", first, err)
	}
	if denyCount.matches.Load() != 1 || allowCount.matches.Load() != 1 || grayCount.matches.Load() != 1 {
		t.Fatalf("expected first evaluation to touch all lookups once, got deny=%d allow=%d gray=%d", denyCount.matches.Load(), allowCount.matches.Load(), grayCount.matches.Load())
	}

	second, err := useCase.CheckIP(context.Background(), "127.0.0.1")
	if err != nil || !second.Allowed {
		t.Fatalf("expected cached decision allowed, got decision=%+v err=%v", second, err)
	}
	if denyCount.matches.Load() != 1 || allowCount.matches.Load() != 1 || grayCount.matches.Load() != 1 {
		t.Fatalf("expected cache hit to skip lookup reevaluation, got deny=%d allow=%d gray=%d", denyCount.matches.Load(), allowCount.matches.Load(), grayCount.matches.Load())
	}

	third, err := useCase.CheckIP(context.Background(), "127.0.0.2")
	if err != nil || !third.Allowed {
		t.Fatalf("expected different IP allowed, got decision=%+v err=%v", third, err)
	}
	if denyCount.matches.Load() != 2 || allowCount.matches.Load() != 2 || grayCount.matches.Load() != 2 {
		t.Fatalf("expected different IP to miss cache, got deny=%d allow=%d gray=%d", denyCount.matches.Load(), allowCount.matches.Load(), grayCount.matches.Load())
	}

	updated := buildAccessSnapshot(t, 2, domain.DefaultPolicyAllow, []domain.IPRule{
		{ID: "deny-local", Type: domain.ListTypeDeny, Value: "127.0.0.1"},
	})
	updated, denyCountV2, _, _ := wrapLookupCounters(updated)
	current.Store(updated)

	fourth, err := useCase.CheckIP(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("version invalidation CheckIP error: %v", err)
	}
	if fourth.Allowed || fourth.Decision != "deny" || fourth.MatchedRuleID != "deny-local" {
		t.Fatalf("expected cache invalidation after version change, got %+v", fourth)
	}
	if denyCountV2.matches.Load() != 1 {
		t.Fatalf("expected version change to force reevaluation, got %d deny matches", denyCountV2.matches.Load())
	}
}

func TestIPAccessUseCaseDecisionCacheTTLExpires(t *testing.T) {
	t.Parallel()

	snapshot := buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, nil)
	snapshot, denyLookup, _, _ := wrapLookupCounters(snapshot)
	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return snapshot, nil
		},
	}

	useCase := NewIPAccessUseCase(repo, newMockVerifiedIPRepository(time.Minute), 16)
	useCase.SetDecisionCacheTTL(20 * time.Millisecond)

	if _, err := useCase.CheckIP(context.Background(), "127.0.0.1"); err != nil {
		t.Fatalf("first CheckIP failed: %v", err)
	}
	if got := denyLookup.matches.Load(); got != 1 {
		t.Fatalf("expected first decision evaluation, got %d deny lookups", got)
	}

	if _, err := useCase.CheckIP(context.Background(), "127.0.0.1"); err != nil {
		t.Fatalf("second CheckIP failed: %v", err)
	}
	if got := denyLookup.matches.Load(); got != 1 {
		t.Fatalf("expected cached decision reuse before TTL expiry, got %d deny lookups", got)
	}

	time.Sleep(30 * time.Millisecond)

	if _, err := useCase.CheckIP(context.Background(), "127.0.0.1"); err != nil {
		t.Fatalf("third CheckIP failed: %v", err)
	}
	if got := denyLookup.matches.Load(); got != 2 {
		t.Fatalf("expected expired cache entry to force re-evaluation, got %d deny lookups", got)
	}
}

func TestIPAccessUseCaseGraylistVerificationFlow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	verifiedRepo := newMockVerifiedIPRepository(20 * time.Millisecond)
	verifiedRepo.now = func() time.Time { return now }

	snapshot := buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, []domain.IPRule{
		{ID: "gray-office", Type: domain.ListTypeGray, Value: "192.168.1.10-192.168.1.20"},
	})
	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return snapshot, nil
		},
	}

	useCase := NewIPAccessUseCase(repo, verifiedRepo, 8)

	initial, err := useCase.CheckIP(context.Background(), "192.168.1.15")
	if err != nil {
		t.Fatalf("initial graylist CheckIP error: %v", err)
	}
	if initial.Allowed || initial.Decision != "captcha_required" || !initial.VerificationRequired {
		t.Fatalf("expected captcha challenge, got %+v", initial)
	}

	if err := useCase.VerifyCaptcha(context.Background(), "192.168.1.15", "bad"); !errors.Is(err, domain.ErrInvalidCaptcha) {
		t.Fatalf("expected invalid captcha error, got %v", err)
	}
	if err := useCase.VerifyCaptcha(context.Background(), "not-an-ip", "1234"); err == nil || !strings.Contains(err.Error(), domain.ErrInvalidIP.Error()) {
		t.Fatalf("expected invalid IP error, got %v", err)
	}
	if err := useCase.VerifyCaptcha(context.Background(), "192.168.1.15", "1234"); err != nil {
		t.Fatalf("VerifyCaptcha error: %v", err)
	}
	if len(verifiedRepo.markCalls) != 1 || verifiedRepo.markCalls[0] != "192.168.1.15" {
		t.Fatalf("unexpected mark verified calls: %+v", verifiedRepo.markCalls)
	}

	allowed, err := useCase.CheckIP(context.Background(), "192.168.1.15")
	if err != nil {
		t.Fatalf("verified CheckIP error: %v", err)
	}
	if !allowed.Allowed || allowed.Decision != "allow" || allowed.MatchedRuleID != "gray-office" {
		t.Fatalf("expected verified graylist IP allowed, got %+v", allowed)
	}

	now = now.Add(25 * time.Millisecond)
	expired, err := useCase.CheckIP(context.Background(), "192.168.1.15")
	if err != nil {
		t.Fatalf("expired verification CheckIP error: %v", err)
	}
	if expired.Allowed || expired.Decision != "captcha_required" {
		t.Fatalf("expected verification ttl expiration to require captcha again, got %+v", expired)
	}
	if len(verifiedRepo.isCalls) < 3 {
		t.Fatalf("expected IsVerified to be checked repeatedly for graylist flow, got %d calls", len(verifiedRepo.isCalls))
	}
}

func TestIPAccessUseCaseVerifiedRepoErrorsPropagate(t *testing.T) {
	t.Parallel()

	verified := newMockVerifiedIPRepository(time.Minute)
	verified.isErr = errors.New("verification lookup failed")

	snapshot := buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, []domain.IPRule{
		{ID: "gray", Type: domain.ListTypeGray, Value: "192.168.1.0/24"},
	})
	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return snapshot, nil
		},
	}

	useCase := NewIPAccessUseCase(repo, verified, 8)
	if _, err := useCase.CheckIP(context.Background(), "192.168.1.10"); err == nil || !strings.Contains(err.Error(), "verification lookup failed") {
		t.Fatalf("expected verified repo lookup error, got %v", err)
	}

	verified = newMockVerifiedIPRepository(time.Minute)
	verified.markErr = errors.New("mark failed")
	useCase = NewIPAccessUseCase(repo, verified, 8)
	if err := useCase.VerifyCaptcha(context.Background(), "192.168.1.10", "1234"); err == nil || !strings.Contains(err.Error(), "mark failed") {
		t.Fatalf("expected verified repo mark error, got %v", err)
	}
}

func TestIPAccessUseCaseSnapshotErrorsAndEmptyRules(t *testing.T) {
	t.Parallel()

	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return domain.AccessSnapshot{}, errors.New("snapshot failed")
		},
	}
	useCase := NewIPAccessUseCase(repo, newMockVerifiedIPRepository(time.Minute), 8)
	if _, err := useCase.CheckIP(context.Background(), "127.0.0.1"); err == nil || !strings.Contains(err.Error(), "snapshot failed") {
		t.Fatalf("expected snapshot error, got %v", err)
	}

	repo = &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, nil), nil
		},
	}
	useCase = NewIPAccessUseCase(repo, newMockVerifiedIPRepository(time.Minute), 8)
	decision, err := useCase.CheckIP(context.Background(), "198.51.100.1")
	if err != nil {
		t.Fatalf("empty rules CheckIP error: %v", err)
	}
	if !decision.Allowed || decision.Reason != "default policy" {
		t.Fatalf("expected empty rules to fall back to default allow, got %+v", decision)
	}
}

func TestIPAccessUseCaseConcurrentDecisionsAndSnapshotUpdates(t *testing.T) {
	t.Parallel()

	allowSnapshot := buildAccessSnapshot(t, 1, domain.DefaultPolicyAllow, nil)
	denySnapshot := buildAccessSnapshot(t, 2, domain.DefaultPolicyAllow, []domain.IPRule{
		{ID: "deny-local", Type: domain.ListTypeDeny, Value: "127.0.0.1"},
	})

	current := atomic.Value{}
	current.Store(allowSnapshot)

	repo := &mockIPAccessRepository{
		getSnapshotFn: func(context.Context) (domain.AccessSnapshot, error) {
			return current.Load().(domain.AccessSnapshot), nil
		},
	}
	useCase := NewIPAccessUseCase(repo, newMockVerifiedIPRepository(time.Minute), 64)
	useCase.SetDecisionCacheTTL(50 * time.Millisecond)

	var wg sync.WaitGroup
	errs := make(chan error, 80)
	results := make(chan string, 80)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if j%2 == 0 {
					current.Store(allowSnapshot)
				} else {
					current.Store(denySnapshot)
				}
				decision, err := useCase.CheckIP(context.Background(), "127.0.0.1")
				if err != nil {
					errs <- err
					return
				}
				results <- decision.Decision
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	close(results)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}

	seen := map[string]bool{}
	for decision := range results {
		if decision != "allow" && decision != "deny" {
			t.Fatalf("unexpected concurrent decision value %q", decision)
		}
		seen[decision] = true
	}

	if len(seen) == 0 {
		t.Fatal("expected at least one decision result")
	}
}
