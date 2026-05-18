package usecase

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type mockMonitoringRepository struct {
	beginCalls     int
	finishCalls    []domain.RequestRecord
	accessCalls    []domain.AccessDecisionEvent
	rateLimitCalls []domain.RateLimitDecisionEvent
	upstreamCalls  []domain.UpstreamEvent
	cacheCalls     []domain.CacheEvent
	snapshot       domain.MonitoringSnapshot
	snapshotErr    error
}

func (m *mockMonitoringRepository) BeginRequest(context.Context) { m.beginCalls++ }
func (m *mockMonitoringRepository) FinishRequest(_ context.Context, record domain.RequestRecord) {
	m.finishCalls = append(m.finishCalls, record)
}
func (m *mockMonitoringRepository) RecordAccessDecision(_ context.Context, event domain.AccessDecisionEvent) {
	m.accessCalls = append(m.accessCalls, event)
}
func (m *mockMonitoringRepository) RecordRateLimitDecision(_ context.Context, event domain.RateLimitDecisionEvent) {
	m.rateLimitCalls = append(m.rateLimitCalls, event)
}
func (m *mockMonitoringRepository) RecordUpstream(_ context.Context, event domain.UpstreamEvent) {
	m.upstreamCalls = append(m.upstreamCalls, event)
}
func (m *mockMonitoringRepository) RecordCacheEvent(_ context.Context, event domain.CacheEvent) {
	m.cacheCalls = append(m.cacheCalls, event)
}
func (m *mockMonitoringRepository) Snapshot(context.Context) (domain.MonitoringSnapshot, error) {
	if m.snapshotErr != nil {
		return domain.MonitoringSnapshot{}, m.snapshotErr
	}
	return m.snapshot, nil
}

type mockMonitoringMetricsRepository struct {
	beginCalls     int
	finishCalls    []domain.RequestRecord
	accessCalls    []domain.AccessDecisionEvent
	rateLimitCalls []domain.RateLimitDecisionEvent
	upstreamCalls  []domain.UpstreamEvent
	cacheCalls     []domain.CacheEvent
}

func (m *mockMonitoringMetricsRepository) BeginRequest(context.Context) { m.beginCalls++ }
func (m *mockMonitoringMetricsRepository) FinishRequest(_ context.Context, record domain.RequestRecord) {
	m.finishCalls = append(m.finishCalls, record)
}
func (m *mockMonitoringMetricsRepository) RecordAccessDecision(_ context.Context, event domain.AccessDecisionEvent) {
	m.accessCalls = append(m.accessCalls, event)
}
func (m *mockMonitoringMetricsRepository) RecordRateLimitDecision(_ context.Context, event domain.RateLimitDecisionEvent) {
	m.rateLimitCalls = append(m.rateLimitCalls, event)
}
func (m *mockMonitoringMetricsRepository) RecordUpstream(_ context.Context, event domain.UpstreamEvent) {
	m.upstreamCalls = append(m.upstreamCalls, event)
}
func (m *mockMonitoringMetricsRepository) RecordCacheEvent(_ context.Context, event domain.CacheEvent) {
	m.cacheCalls = append(m.cacheCalls, event)
}

type mockIPAccessRulesReader struct {
	rules []domain.IPRule
	err   error
}

func (m *mockIPAccessRulesReader) ListRules(context.Context) ([]domain.IPRule, error) {
	if m.err != nil {
		return nil, m.err
	}
	return append([]domain.IPRule(nil), m.rules...), nil
}

type mockRateLimitInspector struct {
	rules      []domain.RateLimitRule
	buckets    []domain.RateLimitBucketSnapshot
	rulesErr   error
	bucketsErr error
}

func (m *mockRateLimitInspector) CheckRateLimit(context.Context, string, int64) (domain.RateLimitDecision, error) {
	return domain.RateLimitDecision{}, nil
}

func (m *mockRateLimitInspector) ListRules(context.Context) ([]domain.RateLimitRule, error) {
	if m.rulesErr != nil {
		return nil, m.rulesErr
	}
	return append([]domain.RateLimitRule(nil), m.rules...), nil
}

func (m *mockRateLimitInspector) ListBuckets(context.Context) ([]domain.RateLimitBucketSnapshot, error) {
	if m.bucketsErr != nil {
		return nil, m.bucketsErr
	}
	return append([]domain.RateLimitBucketSnapshot(nil), m.buckets...), nil
}

func baseMonitoringSnapshot() domain.MonitoringSnapshot {
	return domain.MonitoringSnapshot{
		StartedAt:                time.Now().UTC().Add(-2 * time.Hour),
		TotalRequests:            20,
		ActiveRequests:           3,
		ActiveConnections:        4,
		TotalBlockedRequests:     5,
		TotalRateLimitedRequests: 2,
		TotalUpstreamErrors:      1,
		RequestsPerSecond:        1.5,
		AverageLatencyMS:         12.5,
		BytesIn:                  111,
		BytesOut:                 222,
		RequestsByStatus:         map[string]uint64{"200": 10},
		RequestsByMethod:         map[string]uint64{"GET": 20},
		RequestsByPath:           map[string]uint64{"/users": 7},
		Clients: []domain.ClientSnapshot{
			{IP: "10.0.0.2", Requests: 10, BytesIn: 50, BytesOut: 20, BlockedRequests: 0, RateLimitedRequests: 0},
			{IP: "10.0.0.1", Requests: 10, BytesIn: 10, BytesOut: 10, BlockedRequests: 5, RateLimitedRequests: 2},
			{IP: "10.0.0.3", Requests: 3, BytesIn: 100, BytesOut: 200, BlockedRequests: 1, RateLimitedRequests: 1},
		},
		Upstream: domain.UpstreamSnapshot{
			Healthy:       false,
			LatencyMS:     33.3,
			Errors:        1,
			LastError:     "boom",
			TotalRequests: 8,
			LastStatus:    502,
			LastUpdatedAt: time.Now().UTC(),
		},
		AccessDecisionStats: map[string]uint64{"allow": 7, "deny": 3},
		MatchedRuleStats:    map[string]uint64{"deny-private": 2},
		RateLimitViolations: map[string]uint64{"office:rps": 2},
		RateLimitedClients:  map[string]uint64{"10.0.0.1": 2, "10.0.0.3": 1},
		CacheHits:           4,
		CacheMisses:         5,
		CacheStores:         2,
		CacheInvalidations:  map[string]uint64{"invalidate_prefix": 1},
	}
}

func TestNewMonitoringUseCaseAndFanout(t *testing.T) {
	t.Parallel()

	repo := &mockMonitoringRepository{}
	metrics := &mockMonitoringMetricsRepository{}
	ipAccess := &mockIPAccessRulesReader{}
	rateLimit := &mockRateLimitInspector{}
	useCase := NewMonitoringUseCase(repo, metrics, ipAccess, rateLimit)

	useCase.BeginRequest(context.Background(), BeginRequestInput{})
	useCase.FinishRequest(context.Background(), FinishRequestInput{
		Timestamp:         time.Now().UTC(),
		Method:            "GET",
		Path:              "/path",
		ClientIP:          "127.0.0.1",
		Status:            200,
		Latency:           10 * time.Millisecond,
		BytesIn:           11,
		BytesOut:          12,
		AccessDecision:    "allow",
		RateLimitDecision: "allow",
		CacheStatus:       "hit",
		UpstreamLatency:   time.Millisecond,
	})
	useCase.RecordAccessDecision(context.Background(), RecordAccessDecisionInput{IP: "127.0.0.1", Decision: "deny", RuleID: "deny-1"})
	useCase.RecordRateLimitDecision(context.Background(), RecordRateLimitInput{IP: "127.0.0.1", RuleID: "global", LimitType: "rps", Allowed: false, CurrentValue: 10})
	useCase.RecordUpstream(context.Background(), RecordUpstreamInput{Path: "/path", StatusCode: 502, Latency: 12 * time.Millisecond, Error: "boom"})
	useCase.RecordCacheEvent(context.Background(), RecordCacheInput{Key: "k", Domain: "example.com", Path: "/path", Action: "hit", Rule: "global"})

	if repo.beginCalls != 1 || metrics.beginCalls != 1 {
		t.Fatalf("unexpected begin fanout: repo=%d metrics=%d", repo.beginCalls, metrics.beginCalls)
	}
	if len(repo.finishCalls) != 1 || len(metrics.finishCalls) != 1 {
		t.Fatalf("unexpected finish fanout: repo=%d metrics=%d", len(repo.finishCalls), len(metrics.finishCalls))
	}
	if len(repo.accessCalls) != 1 || len(metrics.accessCalls) != 1 {
		t.Fatalf("unexpected access fanout: repo=%d metrics=%d", len(repo.accessCalls), len(metrics.accessCalls))
	}
	if len(repo.rateLimitCalls) != 1 || len(metrics.rateLimitCalls) != 1 {
		t.Fatalf("unexpected rate limit fanout: repo=%d metrics=%d", len(repo.rateLimitCalls), len(metrics.rateLimitCalls))
	}
	if len(repo.upstreamCalls) != 1 || len(metrics.upstreamCalls) != 1 {
		t.Fatalf("unexpected upstream fanout: repo=%d metrics=%d", len(repo.upstreamCalls), len(metrics.upstreamCalls))
	}
	if len(repo.cacheCalls) != 1 || len(metrics.cacheCalls) != 1 {
		t.Fatalf("unexpected cache fanout: repo=%d metrics=%d", len(repo.cacheCalls), len(metrics.cacheCalls))
	}
	if repo.finishCalls[0].CacheStatus != "hit" || repo.cacheCalls[0].Action != "hit" || repo.accessCalls[0].Decision != "deny" {
		t.Fatalf("unexpected mapped events: finish=%+v cache=%+v access=%+v", repo.finishCalls[0], repo.cacheCalls[0], repo.accessCalls[0])
	}
}

func TestMonitoringUseCaseGetMetrics(t *testing.T) {
	t.Parallel()

	snapshot := baseMonitoringSnapshot()
	repo := &mockMonitoringRepository{snapshot: snapshot}
	useCase := NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})

	response, err := useCase.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error: %v", err)
	}
	if response.TotalRequests != snapshot.TotalRequests || response.ActiveConnections != snapshot.ActiveConnections || response.CacheHits != snapshot.CacheHits {
		t.Fatalf("unexpected metrics mapping: %+v", response)
	}
	if response.RequestsByStatus["200"] != 10 || response.CacheInvalidations["invalidate_prefix"] != 1 {
		t.Fatalf("unexpected map mapping: %+v", response)
	}
}

func TestMonitoringUseCaseDashboardOverview(t *testing.T) {
	t.Parallel()

	snapshot := baseMonitoringSnapshot()
	repo := &mockMonitoringRepository{snapshot: snapshot}
	useCase := NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})

	overview, err := useCase.GetDashboardOverview(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardOverview error: %v", err)
	}
	if overview.TotalRequests != snapshot.TotalRequests || overview.CurrentRPS != snapshot.RequestsPerSecond || overview.UpstreamStatus != "degraded" {
		t.Fatalf("unexpected overview: %+v", overview)
	}

	snapshot.Upstream.Healthy = true
	snapshot.Upstream.TotalRequests = 1
	repo.snapshot = snapshot
	overview, err = useCase.GetDashboardOverview(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardOverview healthy error: %v", err)
	}
	if overview.UpstreamStatus != "healthy" {
		t.Fatalf("expected healthy upstream status, got %+v", overview)
	}

	snapshot.Upstream.Healthy = false
	snapshot.Upstream.TotalRequests = 0
	repo.snapshot = snapshot
	overview, err = useCase.GetDashboardOverview(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardOverview unknown error: %v", err)
	}
	if overview.UpstreamStatus != "unknown" {
		t.Fatalf("expected unknown upstream status, got %+v", overview)
	}
}

func TestMonitoringUseCaseDashboardClientsAndSortHelpers(t *testing.T) {
	t.Parallel()

	snapshot := baseMonitoringSnapshot()
	repo := &mockMonitoringRepository{snapshot: snapshot}
	useCase := NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})

	clients, err := useCase.GetDashboardClients(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardClients error: %v", err)
	}
	if len(clients.TopClientsByRequests) != 3 || clients.TopClientsByRequests[0].IP != "10.0.0.1" {
		t.Fatalf("unexpected top clients by requests ordering: %+v", clients.TopClientsByRequests)
	}
	if len(clients.TopClientsByTraffic) != 3 || clients.TopClientsByTraffic[0].IP != "10.0.0.3" {
		t.Fatalf("unexpected top clients by traffic ordering: %+v", clients.TopClientsByTraffic)
	}
	if len(clients.BlockedClients) != 2 || clients.BlockedClients[0].IP != "10.0.0.1" {
		t.Fatalf("unexpected blocked clients ordering: %+v", clients.BlockedClients)
	}
	if len(clients.RateLimitedClients) != 2 || clients.RateLimitedClients[0].IP != "10.0.0.1" {
		t.Fatalf("unexpected rate limited clients ordering: %+v", clients.RateLimitedClients)
	}

	ordered := sortClients([]domain.ClientSnapshot{
		{IP: "b", Requests: 1},
		{IP: "a", Requests: 1},
	}, func(item domain.ClientSnapshot) uint64 { return item.Requests })
	if len(ordered) != 2 || ordered[0].IP != "a" {
		t.Fatalf("expected deterministic tie-break ordering, got %+v", ordered)
	}

	filtered := sortFilteredClients([]domain.ClientSnapshot{
		{IP: "a", BlockedRequests: 0},
		{IP: "b", BlockedRequests: 2},
	}, func(item domain.ClientSnapshot) bool { return item.BlockedRequests > 0 }, func(item domain.ClientSnapshot) uint64 { return item.BlockedRequests })
	if len(filtered) != 1 || filtered[0].IP != "b" {
		t.Fatalf("unexpected filtered ordering: %+v", filtered)
	}

	if got := sortClients(nil, func(domain.ClientSnapshot) uint64 { return 0 }); len(got) != 0 {
		t.Fatalf("expected empty sorted clients, got %+v", got)
	}
}

func TestMonitoringUseCaseDashboardUpstreamRateLimitsAndIPAccess(t *testing.T) {
	t.Parallel()

	snapshot := baseMonitoringSnapshot()
	repo := &mockMonitoringRepository{snapshot: snapshot}
	ipReader := &mockIPAccessRulesReader{
		rules: []domain.IPRule{
			{ID: "a", Type: domain.ListTypeAllow, Value: "127.0.0.1"},
			{ID: "d", Type: domain.ListTypeDeny, Value: "10.0.0.0/8"},
			{ID: "g", Type: domain.ListTypeGray, Value: "192.168.1.0/24"},
		},
	}
	rateInspector := &mockRateLimitInspector{
		rules: []domain.RateLimitRule{
			{ID: "default-ip", Scope: domain.RateLimitScopeIP, Value: "127.0.0.1", RPS: 10},
		},
		buckets: []domain.RateLimitBucketSnapshot{
			{Key: "ip:rps:127.0.0.1", Scope: "ip", LimitType: "rps", Value: "127.0.0.1", TokensRemaining: 2},
		},
	}
	useCase := NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, ipReader, rateInspector)

	upstream, err := useCase.GetDashboardUpstream(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardUpstream error: %v", err)
	}
	if upstream.LastError != "boom" || upstream.LastStatus != 502 {
		t.Fatalf("unexpected upstream dashboard response: %+v", upstream)
	}

	rateLimits, err := useCase.GetDashboardRateLimits(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardRateLimits error: %v", err)
	}
	if len(rateLimits.ActiveRules) != 1 || len(rateLimits.CurrentBucketUsage) != 1 {
		t.Fatalf("unexpected rate limits dashboard response: %+v", rateLimits)
	}
	if len(rateLimits.BlockedIPs) != 2 || rateLimits.BlockedIPs[0].IP != "10.0.0.1" {
		t.Fatalf("unexpected blocked IP ordering: %+v", rateLimits.BlockedIPs)
	}

	ipAccess, err := useCase.GetDashboardIPAccess(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardIPAccess error: %v", err)
	}
	if len(ipAccess.Allowlist) != 1 || len(ipAccess.Denylist) != 1 || len(ipAccess.Graylist) != 1 {
		t.Fatalf("unexpected ip access dashboard response: %+v", ipAccess)
	}
	if ipAccess.DenyStatistics["deny"] != 3 || ipAccess.MatchedRuleStatistics["deny-private"] != 2 {
		t.Fatalf("unexpected ip access stats mapping: %+v", ipAccess)
	}
}

func TestMonitoringUseCaseEmptyStateAndErrors(t *testing.T) {
	t.Parallel()

	emptyRepo := &mockMonitoringRepository{snapshot: domain.MonitoringSnapshot{StartedAt: time.Now().UTC()}}
	useCase := NewMonitoringUseCase(emptyRepo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})

	clients, err := useCase.GetDashboardClients(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardClients empty error: %v", err)
	}
	if len(clients.TopClientsByRequests) != 0 || len(clients.BlockedClients) != 0 {
		t.Fatalf("expected empty client dashboard, got %+v", clients)
	}

	repoErr := errors.New("snapshot failed")
	repo := &mockMonitoringRepository{snapshotErr: repoErr}
	useCase = NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})
	if _, err := useCase.GetMetrics(context.Background()); !errors.Is(err, repoErr) {
		t.Fatalf("expected snapshot error from GetMetrics, got %v", err)
	}

	repo = &mockMonitoringRepository{snapshot: baseMonitoringSnapshot()}
	useCase = NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{err: errors.New("rules failed")}, &mockRateLimitInspector{})
	if _, err := useCase.GetDashboardIPAccess(context.Background()); err == nil {
		t.Fatal("expected ip access list error")
	}

	useCase = NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{rulesErr: errors.New("rate rules failed")})
	if _, err := useCase.GetDashboardRateLimits(context.Background()); err == nil {
		t.Fatal("expected rate rules error")
	}

	useCase = NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{rules: []domain.RateLimitRule{{ID: "r"}}, bucketsErr: errors.New("bucket failed")})
	if _, err := useCase.GetDashboardRateLimits(context.Background()); err == nil {
		t.Fatal("expected rate buckets error")
	}
}

func TestMonitoringUseCaseGetMetricsLatencyMapping(t *testing.T) {
	t.Parallel()

	snapshot := baseMonitoringSnapshot()
	snapshot.AverageLatencyMS = 42.25
	repo := &mockMonitoringRepository{snapshot: snapshot}
	useCase := NewMonitoringUseCase(repo, &mockMonitoringMetricsRepository{}, &mockIPAccessRulesReader{}, &mockRateLimitInspector{})

	response, err := useCase.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error: %v", err)
	}
	if math.Abs(response.AverageLatencyMS-42.25) > 0.0001 {
		t.Fatalf("unexpected latency mapping: %+v", response)
	}
}
