package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type mockConfigProvider struct {
	mu         sync.Mutex
	current    config.Config
	subscriber config.Subscriber
}

func (m *mockConfigProvider) Current() config.Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current
}

func (m *mockConfigProvider) Subscribe(fn config.Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriber = fn
}

func (m *mockConfigProvider) Update(cfg config.Config) {
	m.mu.Lock()
	m.current = cfg
	fn := m.subscriber
	m.mu.Unlock()
	if fn != nil {
		fn(cfg)
	}
}

type mockUseCaseLogger struct {
	mu     sync.Mutex
	errors []loggedError
}

type loggedError struct {
	msg string
	err error
}

func (m *mockUseCaseLogger) Error(msg string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, loggedError{msg: msg, err: err})
}

type mockRateLimitStateRepository struct {
	mu sync.Mutex

	buckets     map[string]domain.BucketState
	connections map[string]int

	getBucketCalls       int
	saveBucketCalls      int
	tryAcquireCalls      int
	releaseCalls         int
	snapshotBucketsCalls int

	getBucketErr  error
	saveBucketErr error
	acquireErr    error
	releaseErr    error

	snapshotBuckets []domain.RateLimitBucketSnapshot
}

func newMockRateLimitStateRepository() *mockRateLimitStateRepository {
	return &mockRateLimitStateRepository{
		buckets:     make(map[string]domain.BucketState),
		connections: make(map[string]int),
	}
}

func (m *mockRateLimitStateRepository) GetBucket(_ context.Context, key string) (domain.BucketState, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getBucketCalls++
	if m.getBucketErr != nil {
		return domain.BucketState{}, false, m.getBucketErr
	}
	state, ok := m.buckets[key]
	return state, ok, nil
}

func (m *mockRateLimitStateRepository) SaveBucket(_ context.Context, key string, state domain.BucketState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveBucketCalls++
	if m.saveBucketErr != nil {
		return m.saveBucketErr
	}
	m.buckets[key] = state
	return nil
}

func (m *mockRateLimitStateRepository) DeleteStaleBuckets(context.Context, time.Time) error {
	return nil
}

func (m *mockRateLimitStateRepository) SnapshotBuckets(context.Context) ([]domain.RateLimitBucketSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshotBucketsCalls++
	return append([]domain.RateLimitBucketSnapshot(nil), m.snapshotBuckets...), nil
}

func (m *mockRateLimitStateRepository) TryAcquireConnection(_ context.Context, key string, limit int) (bool, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tryAcquireCalls++
	if m.acquireErr != nil {
		return false, 0, m.acquireErr
	}
	current := m.connections[key]
	if current >= limit {
		return false, current, nil
	}
	current++
	m.connections[key] = current
	return true, current, nil
}

func (m *mockRateLimitStateRepository) ReleaseConnection(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseCalls++
	if m.releaseErr != nil {
		return m.releaseErr
	}
	current := m.connections[key]
	if current <= 1 {
		delete(m.connections, key)
		return nil
	}
	m.connections[key] = current - 1
	return nil
}

func baseRateLimitConfig() config.Config {
	return config.Config{
		Proxy: config.ProxyConfig{
			UpstreamURL: "https://example.com",
		},
		RateLimit: config.RateLimitConfig{
			Enabled: true,
		},
	}
}

func newTestRateLimitUseCase(cfg config.Config, repo *mockRateLimitStateRepository, logger *mockUseCaseLogger) (*RateLimitUseCase, *mockConfigProvider) {
	provider := &mockConfigProvider{current: cfg}
	useCase := NewRateLimitUseCase(repo, provider, logger)
	return useCase, provider
}

func TestRateLimitUseCaseDisabledAndInvalidIP(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.Enabled = false
	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	decision, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 123)
	if err != nil {
		t.Fatalf("unexpected error for disabled limiter: %v", err)
	}
	if !decision.Allowed || decision.Reason != "rate_limit_disabled" {
		t.Fatalf("unexpected disabled decision: %+v", decision)
	}

	cfg = baseRateLimitConfig()
	useCase, _ = newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
	if _, err := useCase.CheckRateLimit(context.Background(), "not-an-ip", 0); err == nil {
		t.Fatal("expected invalid ip error")
	}
}

func TestRateLimitUseCaseRequestRateLimitsTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		applyLimit    func(*config.Config)
		resetDuration time.Duration
	}{
		{
			name: "rps",
			applyLimit: func(cfg *config.Config) {
				cfg.RateLimit.RPS = 1
			},
			resetDuration: time.Second,
		},
		{
			name: "rpm",
			applyLimit: func(cfg *config.Config) {
				cfg.RateLimit.RPM = 1
			},
			resetDuration: time.Minute,
		},
		{
			name: "rph",
			applyLimit: func(cfg *config.Config) {
				cfg.RateLimit.RPH = 1
			},
			resetDuration: time.Hour,
		},
		{
			name: "rpd",
			applyLimit: func(cfg *config.Config) {
				cfg.RateLimit.RPD = 1
			},
			resetDuration: 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := baseRateLimitConfig()
			tt.applyLimit(&cfg)
			repo := newMockRateLimitStateRepository()
			useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

			now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			useCase.now = func() time.Time { return now }

			first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
			if err != nil {
				t.Fatalf("first request error: %v", err)
			}
			if !first.Allowed {
				t.Fatalf("expected first request allowed, got %+v", first)
			}

			second, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
			if err != nil {
				t.Fatalf("second request error: %v", err)
			}
			if second.Allowed {
				t.Fatalf("expected second request blocked, got %+v", second)
			}
			if second.LimitType != tt.name {
				t.Fatalf("unexpected limit type: got %q want %q", second.LimitType, tt.name)
			}
			if second.CurrentValue != 1 {
				t.Fatalf("unexpected current value: got %d want 1", second.CurrentValue)
			}

			now = now.Add(tt.resetDuration + time.Millisecond)
			third, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
			if err != nil {
				t.Fatalf("third request error: %v", err)
			}
			if !third.Allowed {
				t.Fatalf("expected reset request allowed, got %+v", third)
			}
		})
	}
}

func TestRateLimitUseCaseSubnetRuleAndIsolation(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.RPS = 100
	cfg.RateLimit.Subnets = []config.RateLimitSubnetConfig{
		{
			ID:   "office",
			CIDR: "192.168.1.0/24",
			RPS:  1,
		},
	}

	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	useCase.now = func() time.Time { return now }

	first, err := useCase.CheckRateLimit(context.Background(), "192.168.1.10", 0)
	if err != nil {
		t.Fatalf("first subnet request error: %v", err)
	}
	if !first.Allowed {
		t.Fatalf("expected first subnet request allowed, got %+v", first)
	}

	second, err := useCase.CheckRateLimit(context.Background(), "192.168.1.10", 0)
	if err != nil {
		t.Fatalf("second subnet request error: %v", err)
	}
	if second.Allowed {
		t.Fatalf("expected subnet override to block second request, got %+v", second)
	}
	if second.RuleID != "office" || second.LimitType != "rps" {
		t.Fatalf("unexpected subnet deny decision: %+v", second)
	}

	outside, err := useCase.CheckRateLimit(context.Background(), "203.0.113.10", 0)
	if err != nil {
		t.Fatalf("outside subnet request error: %v", err)
	}
	if !outside.Allowed {
		t.Fatalf("expected outside subnet client allowed, got %+v", outside)
	}
}

func TestRateLimitUseCaseConnectionAcquireRelease(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.MaxConnections = 1
	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("first request error: %v", err)
	}
	if !first.Allowed || len(first.ConnectionKeys) != 1 {
		t.Fatalf("expected first request to acquire one connection key, got %+v", first)
	}

	second, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("second request error: %v", err)
	}
	if second.Allowed || second.LimitType != "connections" {
		t.Fatalf("expected over-limit connections rejection, got %+v", second)
	}

	if err := useCase.ReleaseConnections(context.Background(), first.ConnectionKeys); err != nil {
		t.Fatalf("release connections error: %v", err)
	}
	if got := repo.connections[first.ConnectionKeys[0]]; got != 0 {
		t.Fatalf("expected released connection counter cleanup, got %d", got)
	}

	third, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("third request error: %v", err)
	}
	if !third.Allowed {
		t.Fatalf("expected request allowed after release, got %+v", third)
	}
}

func TestRateLimitUseCaseCPSLimitAndReset(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.CPS = 1
	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	useCase.now = func() time.Time { return now }

	first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !first.Allowed {
		t.Fatalf("expected first cps request allowed, got decision=%+v err=%v", first, err)
	}

	second, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("second cps request error: %v", err)
	}
	if second.Allowed || second.LimitType != "cps" {
		t.Fatalf("expected cps block, got %+v", second)
	}

	now = now.Add(time.Second + time.Millisecond)
	third, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !third.Allowed {
		t.Fatalf("expected cps window reset, got decision=%+v err=%v", third, err)
	}
}

func TestRateLimitUseCaseBandwidthTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*config.Config)
		exercise  func(t *testing.T, useCase *RateLimitUseCase)
	}{
		{
			name: "upload limit",
			configure: func(cfg *config.Config) {
				cfg.RateLimit.UploadBPS = 100
			},
			exercise: func(t *testing.T, useCase *RateLimitUseCase) {
				first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 60)
				if err != nil || !first.Allowed {
					t.Fatalf("expected first upload allowed, got decision=%+v err=%v", first, err)
				}
				second, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 50)
				if err != nil {
					t.Fatalf("second upload error: %v", err)
				}
				if second.Allowed || second.LimitType != "upload" {
					t.Fatalf("expected upload limit rejection, got %+v", second)
				}
			},
		},
		{
			name: "download limit",
			configure: func(cfg *config.Config) {
				cfg.RateLimit.DownloadBPS = 100
			},
			exercise: func(t *testing.T, useCase *RateLimitUseCase) {
				first, err := useCase.ReserveResponseBandwidth(context.Background(), "127.0.0.1", 60)
				if err != nil || !first.Allowed {
					t.Fatalf("expected first download allowed, got decision=%+v err=%v", first, err)
				}
				second, err := useCase.ReserveResponseBandwidth(context.Background(), "127.0.0.1", 50)
				if err != nil {
					t.Fatalf("second download error: %v", err)
				}
				if second.Allowed || second.LimitType != "download" {
					t.Fatalf("expected download limit rejection, got %+v", second)
				}
			},
		},
		{
			name: "total limit",
			configure: func(cfg *config.Config) {
				cfg.RateLimit.TotalBytes = 100
				cfg.RateLimit.TotalWindow = config.Duration{Duration: time.Hour}
			},
			exercise: func(t *testing.T, useCase *RateLimitUseCase) {
				first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 60)
				if err != nil || !first.Allowed {
					t.Fatalf("expected first total request allowed, got decision=%+v err=%v", first, err)
				}
				second, err := useCase.ReserveResponseBandwidth(context.Background(), "127.0.0.1", 50)
				if err != nil {
					t.Fatalf("second total reservation error: %v", err)
				}
				if second.Allowed || second.LimitType != "total" {
					t.Fatalf("expected total limit rejection, got %+v", second)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := baseRateLimitConfig()
			tt.configure(&cfg)
			repo := newMockRateLimitStateRepository()
			useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
			useCase.now = func() time.Time {
				return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			}
			tt.exercise(t, useCase)
		})
	}
}

func TestRateLimitUseCaseZeroAndNegativeTrafficHandling(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.UploadBPS = 100
	cfg.RateLimit.DownloadBPS = 100
	cfg.RateLimit.TotalBytes = 100
	cfg.RateLimit.TotalWindow = config.Duration{Duration: time.Hour}

	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	decision, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", -50)
	if err != nil || !decision.Allowed {
		t.Fatalf("expected negative upload to normalize and pass, got decision=%+v err=%v", decision, err)
	}

	decision, err = useCase.ReserveResponseBandwidth(context.Background(), "127.0.0.1", 0)
	if err != nil || !decision.Allowed {
		t.Fatalf("expected zero download reservation to pass, got decision=%+v err=%v", decision, err)
	}

	if err := useCase.AccountTraffic(context.Background(), "127.0.0.1", -10, 0); err != nil {
		t.Fatalf("expected negative account traffic to normalize and pass, got %v", err)
	}
	if repo.getBucketCalls != 0 && repo.saveBucketCalls != 0 {
		t.Fatalf("expected zero/negative traffic to avoid bucket writes, got get=%d save=%d", repo.getBucketCalls, repo.saveBucketCalls)
	}
}

func TestRateLimitUseCaseListRulesAndBuckets(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.RPS = 10
	cfg.RateLimit.Subnets = []config.RateLimitSubnetConfig{
		{
			ID:   "office",
			CIDR: "192.168.1.0/24",
			RPS:  25,
		},
	}

	repo := newMockRateLimitStateRepository()
	repo.snapshotBuckets = []domain.RateLimitBucketSnapshot{{Key: "ip:rps:127.0.0.1"}}
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	rules, err := useCase.ListRules(context.Background())
	if err != nil {
		t.Fatalf("ListRules error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].ID != "default-ip" || rules[1].ID != "office" {
		t.Fatalf("unexpected rules: %+v", rules)
	}

	buckets, err := useCase.ListBuckets(context.Background())
	if err != nil {
		t.Fatalf("ListBuckets error: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Key != "ip:rps:127.0.0.1" {
		t.Fatalf("unexpected buckets: %+v", buckets)
	}
}

func TestRateLimitUseCaseConfigReloadAndInvalidSubnetLogging(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.RPS = 1
	repo := newMockRateLimitStateRepository()
	logger := &mockUseCaseLogger{}
	useCase, provider := newTestRateLimitUseCase(cfg, repo, logger)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	useCase.now = func() time.Time { return now }

	first, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !first.Allowed {
		t.Fatalf("expected initial request allowed, got decision=%+v err=%v", first, err)
	}
	second, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || second.Allowed {
		t.Fatalf("expected second request blocked before reload, got decision=%+v err=%v", second, err)
	}

	updated := baseRateLimitConfig()
	updated.RateLimit.Enabled = true
	updated.RateLimit.RPS = 2
	updated.RateLimit.Subnets = []config.RateLimitSubnetConfig{
		{ID: "broken", CIDR: "not-a-cidr", RPS: 1},
	}
	provider.Update(updated)

	if len(logger.errors) != 1 {
		t.Fatalf("expected one invalid subnet log entry, got %d", len(logger.errors))
	}
	if logger.errors[0].msg != "invalid subnet in rate limit config" {
		t.Fatalf("unexpected log message: %+v", logger.errors[0])
	}

	now = now.Add(time.Second + time.Millisecond)
	third, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !third.Allowed {
		t.Fatalf("expected first request after reload allowed, got decision=%+v err=%v", third, err)
	}
	fourth, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !fourth.Allowed {
		t.Fatalf("expected second request after reload allowed with new limit, got decision=%+v err=%v", fourth, err)
	}
	fifth, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("fifth request error: %v", err)
	}
	if fifth.Allowed || fifth.LimitType != "rps" {
		t.Fatalf("expected third request after reload blocked, got %+v", fifth)
	}
}

func TestRateLimitUseCaseRepositoryErrors(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.RPS = 1
	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})

	repo.getBucketErr = errors.New("get failed")
	if _, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0); err == nil {
		t.Fatal("expected get bucket error")
	}
	repo.getBucketErr = nil

	repo.saveBucketErr = errors.New("save failed")
	if _, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0); err == nil {
		t.Fatal("expected save bucket error")
	}
	repo.saveBucketErr = nil

	cfg = baseRateLimitConfig()
	cfg.RateLimit.MaxConnections = 1
	repo = newMockRateLimitStateRepository()
	repo.acquireErr = errors.New("acquire failed")
	useCase, _ = newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
	if _, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0); err == nil {
		t.Fatal("expected acquire connection error")
	}

	repo = newMockRateLimitStateRepository()
	repo.releaseErr = errors.New("release failed")
	useCase, _ = newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
	decision, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
	if err != nil || !decision.Allowed {
		t.Fatalf("expected first connection acquisition allowed, got decision=%+v err=%v", decision, err)
	}
	if err := useCase.ReleaseConnections(context.Background(), decision.ConnectionKeys); err == nil {
		t.Fatal("expected release connections error")
	}
}

func TestRateLimitUseCaseConcurrentSameIP(t *testing.T) {
	t.Parallel()

	cfg := baseRateLimitConfig()
	cfg.RateLimit.RPS = 1
	repo := newMockRateLimitStateRepository()
	useCase, _ := newTestRateLimitUseCase(cfg, repo, &mockUseCaseLogger{})
	useCase.now = func() time.Time {
		return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var wg sync.WaitGroup
	results := make(chan domain.RateLimitDecision, 16)
	errorsCh := make(chan error, 16)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decision, err := useCase.CheckRateLimit(context.Background(), "127.0.0.1", 0)
			if err != nil {
				errorsCh <- err
				return
			}
			results <- decision
		}()
	}

	wg.Wait()
	close(results)
	close(errorsCh)

	for err := range errorsCh {
		if err != nil {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}

	allowed := 0
	blocked := 0
	for decision := range results {
		if decision.Allowed {
			allowed++
		} else {
			blocked++
		}
	}

	if allowed != 1 {
		t.Fatalf("expected exactly one allowed request under concurrent RPS=1, got %d", allowed)
	}
	if blocked != 7 {
		t.Fatalf("expected remaining requests blocked, got %d blocked", blocked)
	}
}
