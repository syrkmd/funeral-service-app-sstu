package usecase

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/repository"
)

type staticConfigProvider struct {
	current config.Config
}

func (p *staticConfigProvider) Current() config.Config {
	return p.current
}

func (p *staticConfigProvider) Subscribe(config.Subscriber) {}

func TestCacheUseCaseGlobalFallback(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())

	decision := useCase.DecideCachePolicy(testCandidate("example.com", "/resource", 200))

	if !decision.Allowed {
		t.Fatalf("expected global fallback to allow caching, got reason=%s", decision.Reason)
	}
	if decision.MatchedRule != "global-default" {
		t.Fatalf("expected global-default rule, got %s", decision.MatchedRule)
	}
	if decision.TTL != 5*time.Minute {
		t.Fatalf("expected 5m ttl, got %s", decision.TTL)
	}
}

func TestCacheUseCaseSpecificOverride(t *testing.T) {
	cfg := testCacheConfig()
	cache4xx := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:     "errors",
			Domains:  []string{"example.com"},
			Paths:    []string{"/errors/*"},
			Cache4XX: &cache4xx,
			TTL4XX:   config.Duration{Duration: 45 * time.Second},
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("example.com", "/errors/404", 404))

	if !decision.Allowed {
		t.Fatalf("expected specific rule to enable 4xx caching, got reason=%s", decision.Reason)
	}
	if decision.MatchedRule != "errors" {
		t.Fatalf("expected errors rule, got %s", decision.MatchedRule)
	}
	if decision.TTL != 45*time.Second {
		t.Fatalf("expected 45s ttl, got %s", decision.TTL)
	}
}

func TestCacheUseCaseExactDomainOverridesWildcard(t *testing.T) {
	cfg := testCacheConfig()
	disable := false
	enable := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "wildcard",
			Domains: []string{"*.example.com"},
			Enabled: &disable,
		},
		{
			Name:    "exact",
			Domains: []string{"api.example.com"},
			Enabled: &enable,
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("api.example.com", "/v1", 200))

	if !decision.Allowed || decision.MatchedRule != "exact" {
		t.Fatalf("expected exact domain rule to win, got allowed=%v rule=%s reason=%s", decision.Allowed, decision.MatchedRule, decision.Reason)
	}
}

func TestCacheUseCaseExactPathOverridesPrefix(t *testing.T) {
	cfg := testCacheConfig()
	disable := false
	enable := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "prefix",
			Paths:   []string{"/api/*"},
			Enabled: &disable,
		},
		{
			Name:    "exact",
			Paths:   []string{"/api/resource"},
			Enabled: &enable,
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("example.com", "/api/resource", 200))

	if !decision.Allowed || decision.MatchedRule != "exact" {
		t.Fatalf("expected exact path rule to win, got allowed=%v rule=%s reason=%s", decision.Allowed, decision.MatchedRule, decision.Reason)
	}
}

func TestCacheUseCaseMoreSpecificPathWins(t *testing.T) {
	cfg := testCacheConfig()
	disable := false
	enable := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "short-prefix",
			Domains: []string{"example.com"},
			Paths:   []string{"/api/*"},
			Enabled: &disable,
		},
		{
			Name:    "long-prefix",
			Domains: []string{"example.com"},
			Paths:   []string{"/api/private/*"},
			Enabled: &enable,
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("example.com", "/api/private/item", 200))

	if !decision.Allowed || decision.MatchedRule != "long-prefix" {
		t.Fatalf("expected longer prefix rule to win, got allowed=%v rule=%s reason=%s", decision.Allowed, decision.MatchedRule, decision.Reason)
	}
}

func TestCacheUseCaseDeterministicPriorityByOrder(t *testing.T) {
	cfg := testCacheConfig()
	disable := false
	enable := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "first",
			Domains: []string{"*.example.com"},
			Enabled: &disable,
		},
		{
			Name:    "second",
			Domains: []string{"*.com"},
			Enabled: &enable,
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("api.example.com", "/v1", 200))

	if decision.MatchedRule != "first" || decision.Allowed {
		t.Fatalf("expected first matching rule to win deterministically, got allowed=%v rule=%s", decision.Allowed, decision.MatchedRule)
	}
}

func TestCacheUseCaseInheritance(t *testing.T) {
	cfg := testCacheConfig()
	cfg.Cache.ExcludedHeaders = []string{"Authorization"}
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "inherits",
			Domains: []string{"example.com"},
		},
	}

	useCase := newTestCacheUseCase(cfg)
	candidate := testCandidate("example.com", "/secure", 200)
	candidate.Headers = map[string][]string{"Authorization": {"Bearer token"}}
	decision := useCase.DecideCachePolicy(candidate)

	if decision.Allowed || decision.Reason != "sensitive_headers" || decision.MatchedRule != "inherits" {
		t.Fatalf("expected inherited excluded header rule, got allowed=%v reason=%s rule=%s", decision.Allowed, decision.Reason, decision.MatchedRule)
	}
}

func TestCacheUseCaseEnableDisablePrecedence(t *testing.T) {
	cfg := testCacheConfig()
	cfg.Cache.Enabled = false
	enable := true
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "re-enable",
			Domains: []string{"example.com"},
			Enabled: &enable,
		},
	}

	useCase := newTestCacheUseCase(cfg)
	decision := useCase.DecideCachePolicy(testCandidate("example.com", "/resource", 200))
	if !decision.Allowed || decision.MatchedRule != "re-enable" {
		t.Fatalf("expected specific rule to re-enable cache, got allowed=%v rule=%s reason=%s", decision.Allowed, decision.MatchedRule, decision.Reason)
	}

	cfg = testCacheConfig()
	disable := false
	cfg.Cache.Rules = []config.CacheRuleConfig{
		{
			Name:    "disable",
			Domains: []string{"example.com"},
			Enabled: &disable,
		},
	}

	useCase = newTestCacheUseCase(cfg)
	decision = useCase.DecideCachePolicy(testCandidate("example.com", "/resource", 200))
	if decision.Allowed || decision.MatchedRule != "disable" || decision.Reason != "cache_disabled" {
		t.Fatalf("expected specific rule to disable cache, got allowed=%v rule=%s reason=%s", decision.Allowed, decision.MatchedRule, decision.Reason)
	}
}

func newTestCacheUseCase(cfg config.Config) *CacheUseCase {
	tRepo := repository.NewCacheRepository(time.Minute)
	return NewCacheUseCase(tRepo, &staticConfigProvider{current: cfg})
}

func newTestCacheUseCaseWithRepo(cfg config.Config) (*CacheUseCase, *repository.CacheRepository) {
	tRepo := repository.NewCacheRepository(time.Minute)
	return NewCacheUseCase(tRepo, &staticConfigProvider{current: cfg}), tRepo
}

func testCacheConfig() config.Config {
	return config.Config{
		Proxy: config.ProxyConfig{UpstreamURL: "https://example.com"},
		Cache: config.CacheConfig{
			Enabled:         true,
			DefaultTTL:      config.Duration{Duration: 5 * time.Minute},
			CleanupInterval: config.Duration{Duration: time.Minute},
			MinEntrySize:    0,
			MaxEntrySize:    1 << 20,
			Cache2XX:        true,
			Cache3XX:        false,
			Cache4XX:        false,
			Cache5XX:        false,
			TTL2XX:          config.Duration{Duration: 5 * time.Minute},
			TTL3XX:          config.Duration{Duration: 10 * time.Minute},
			TTL4XX:          config.Duration{Duration: 30 * time.Second},
			TTL5XX:          config.Duration{Duration: 10 * time.Second},
		},
	}
}

func testCandidate(domainName, path string, status int) domain.CacheCandidate {
	return domain.CacheCandidate{
		Request: domain.CacheRequest{
			Method: "GET",
			Scheme: "https",
			Domain: domainName,
			Path:   path,
		},
		StatusCode:  status,
		ContentType: "application/json",
		Headers:     map[string][]string{},
		Size:        256,
	}
}

func TestCacheUseCaseSetGetRoundTrip(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidate := testCandidate("example.com", "/data", 200)
	body := []byte(`{"ok":true}`)

	entry, stored, err := useCase.Set(context.Background(), candidate, body)
	if err != nil {
		t.Fatalf("unexpected set error: %v", err)
	}
	if !stored {
		t.Fatal("expected entry to be stored")
	}

	got, found, err := useCase.Lookup(context.Background(), candidate.Request)
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit after store")
	}
	if got.Metadata.Key != entry.Metadata.Key {
		t.Fatalf("unexpected key: got %s want %s", got.Metadata.Key, entry.Metadata.Key)
	}
	if string(got.Body) != string(body) {
		t.Fatalf("unexpected body: got %q want %q", string(got.Body), string(body))
	}
}

func TestCacheUseCaseLookupBypassesNoCacheRequests(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidate := testCandidate("example.com", "/data", 200)

	if _, stored, err := useCase.Set(context.Background(), candidate, []byte(`{"ok":true}`)); err != nil || !stored {
		t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
	}

	request := candidate.Request
	request.Headers = map[string][]string{
		"Cache-Control": {"no-cache"},
	}

	_, found, err := useCase.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}
	if found {
		t.Fatal("expected cache lookup to bypass no-cache request")
	}
}

func TestCacheUseCaseLookupBypassesSensitiveRequestHeaders(t *testing.T) {
	cfg := testCacheConfig()
	cfg.Cache.ExcludedHeaders = []string{"Authorization", "Cookie", "Set-Cookie"}
	useCase := newTestCacheUseCase(cfg)
	candidate := testCandidate("example.com", "/data", 200)

	if _, stored, err := useCase.Set(context.Background(), candidate, []byte(`{"ok":true}`)); err != nil || !stored {
		t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
	}

	request := candidate.Request
	request.Headers = map[string][]string{
		"Authorization": {"Bearer token"},
	}

	_, found, err := useCase.Lookup(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}
	if found {
		t.Fatal("expected cache lookup to bypass sensitive request headers")
	}
}

func TestCacheUseCaseInvalidateByKey(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidate := testCandidate("example.com", "/data", 200)
	_, stored, err := useCase.Set(context.Background(), candidate, []byte("payload"))
	if err != nil || !stored {
		t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
	}

	key := useCase.GenerateCacheKey(candidate.Request)
	deleted, err := useCase.InvalidateByKey(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected invalidate error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
	if _, found, _ := useCase.Lookup(context.Background(), candidate.Request); found {
		t.Fatal("expected cache miss after key invalidation")
	}
}

func TestCacheUseCaseInvalidateByPrefix(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	entries := []domain.CacheCandidate{
		testCandidate("example.com", "/one", 200),
		testCandidate("example.com", "/two", 200),
		testCandidate("other.com", "/three", 200),
	}
	keys := make([]string, 0, len(entries))
	for _, candidate := range entries {
		_, stored, err := useCase.Set(context.Background(), candidate, []byte(candidate.Request.Path))
		if err != nil || !stored {
			t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
		}
		keys = append(keys, useCase.GenerateCacheKey(candidate.Request))
	}

	prefix := keys[0][:8]
	if keys[1][:8] != prefix {
		prefix = keys[0][:4]
	}
	deleted, err := useCase.InvalidateByPrefix(context.Background(), prefix)
	if err != nil {
		t.Fatalf("unexpected prefix invalidate error: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("expected at least one deleted entry, got %d", deleted)
	}
}

func TestCacheUseCaseInvalidateByRegex(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidate := testCandidate("example.com", "/regex", 200)
	_, stored, err := useCase.Set(context.Background(), candidate, []byte("payload"))
	if err != nil || !stored {
		t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
	}

	key := useCase.GenerateCacheKey(candidate.Request)
	pattern := "^" + regexp.QuoteMeta(key[:12])
	deleted, err := useCase.InvalidateByRegex(context.Background(), pattern)
	if err != nil {
		t.Fatalf("unexpected regex invalidate error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
}

func TestCacheUseCaseInvalidateByTags(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidateA := testCandidate("example.com", "/users", 200)
	candidateA.Tags = []string{"users", "users:list"}
	candidateB := testCandidate("example.com", "/posts", 200)
	candidateB.Tags = []string{"posts"}

	for _, candidate := range []domain.CacheCandidate{candidateA, candidateB} {
		if _, stored, err := useCase.Set(context.Background(), candidate, []byte(candidate.Request.Path)); err != nil || !stored {
			t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
		}
	}

	deleted, err := useCase.InvalidateByTags(context.Background(), []string{"users"})
	if err != nil {
		t.Fatalf("unexpected tag invalidate error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
	if _, found, _ := useCase.Lookup(context.Background(), candidateA.Request); found {
		t.Fatal("expected tagged entry to be removed")
	}
	if _, found, _ := useCase.Lookup(context.Background(), candidateB.Request); !found {
		t.Fatal("expected unrelated entry to remain")
	}
}

func TestCacheUseCaseClear(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	for _, candidate := range []domain.CacheCandidate{
		testCandidate("example.com", "/one", 200),
		testCandidate("example.com", "/two", 200),
	} {
		if _, stored, err := useCase.Set(context.Background(), candidate, []byte(candidate.Request.Path)); err != nil || !stored {
			t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
		}
	}

	deleted, err := useCase.Clear(context.Background())
	if err != nil {
		t.Fatalf("unexpected clear error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deleted entries, got %d", deleted)
	}
}

func TestCacheUseCaseInvalidateExpired(t *testing.T) {
	useCase, repo := newTestCacheUseCaseWithRepo(testCacheConfig())

	expired := domain.CacheEntry{
		Metadata: domain.CacheMetadata{
			Key:        "expired",
			Headers:    map[string][]string{},
			StatusCode: 200,
			TTL:        time.Second,
			CreatedAt:  time.Now().UTC().Add(-2 * time.Second),
			ExpiresAt:  time.Now().UTC().Add(-time.Second),
			Domain:     "example.com",
			Path:       "/expired",
			Tags:       []string{"expired"},
			Size:       10,
		},
		Body: []byte("old"),
	}
	active := domain.CacheEntry{
		Metadata: domain.CacheMetadata{
			Key:        "active",
			Headers:    map[string][]string{},
			StatusCode: 200,
			TTL:        time.Minute,
			CreatedAt:  time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(time.Minute),
			Domain:     "example.com",
			Path:       "/active",
			Tags:       []string{"active"},
			Size:       10,
		},
		Body: []byte("new"),
	}

	if err := repo.Set(context.Background(), expired); err != nil {
		t.Fatalf("failed to seed expired entry: %v", err)
	}
	if err := repo.Set(context.Background(), active); err != nil {
		t.Fatalf("failed to seed active entry: %v", err)
	}

	deleted, err := useCase.InvalidateExpired(context.Background())
	if err != nil {
		t.Fatalf("unexpected expired invalidate error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 expired deletion, got %d", deleted)
	}
}

func TestCacheUseCaseInvalidRegexRejection(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	if _, err := useCase.InvalidateByRegex(context.Background(), "[invalid"); err == nil {
		t.Fatal("expected invalid regex error")
	}
}

func TestCacheUseCaseHandleInvalidationEvent(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	candidate := testCandidate("example.com", "/users", 200)
	candidate.Tags = []string{"users", "users:list"}
	if _, stored, err := useCase.Set(context.Background(), candidate, []byte("payload")); err != nil || !stored {
		t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
	}

	deleted, err := useCase.HandleInvalidationEvent(context.Background(), domain.CacheInvalidationEvent{
		Tags: []string{"users"},
	})
	if err != nil {
		t.Fatalf("unexpected invalidation event error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
}

func TestCacheUseCaseCascadeInvalidationBehavior(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	userEntry := testCandidate("example.com", "/users/42", 200)
	userEntry.Tags = []string{"users", "users:item"}
	listEntry := testCandidate("example.com", "/users", 200)
	listEntry.Tags = []string{"users", "users:list"}

	for _, candidate := range []domain.CacheCandidate{userEntry, listEntry} {
		if _, stored, err := useCase.Set(context.Background(), candidate, []byte(candidate.Request.Path)); err != nil || !stored {
			t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
		}
	}

	deleted, err := useCase.HandleInvalidationEvent(context.Background(), domain.CacheInvalidationEvent{
		Tags: []string{"users", "users:list"},
	})
	if err != nil {
		t.Fatalf("unexpected cascade invalidation error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deleted entries, got %d", deleted)
	}
}

func TestCacheUseCasePartialInvalidation(t *testing.T) {
	useCase := newTestCacheUseCase(testCacheConfig())
	entryA := testCandidate("example.com", "/users/1", 200)
	entryA.Tags = []string{"users"}
	entryB := testCandidate("example.com", "/posts/1", 200)
	entryB.Tags = []string{"posts"}

	for _, candidate := range []domain.CacheCandidate{entryA, entryB} {
		if _, stored, err := useCase.Set(context.Background(), candidate, []byte(candidate.Request.Path)); err != nil || !stored {
			t.Fatalf("setup store failed: stored=%v err=%v", stored, err)
		}
	}

	deleted, err := useCase.HandleInvalidationEvent(context.Background(), domain.CacheInvalidationEvent{
		Prefixes: []string{useCase.GenerateCacheKey(entryA.Request)[:10]},
	})
	if err != nil {
		t.Fatalf("unexpected partial invalidation error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted entry, got %d", deleted)
	}
	if _, found, _ := useCase.Lookup(context.Background(), entryB.Request); !found {
		t.Fatal("expected unrelated entry to remain after partial invalidation")
	}
}
