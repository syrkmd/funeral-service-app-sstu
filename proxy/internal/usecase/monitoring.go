package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type MonitoringRepository interface {
	BeginRequest(ctx context.Context)
	FinishRequest(ctx context.Context, record domain.RequestRecord)
	RecordAccessDecision(ctx context.Context, event domain.AccessDecisionEvent)
	RecordRateLimitDecision(ctx context.Context, event domain.RateLimitDecisionEvent)
	RecordUpstream(ctx context.Context, event domain.UpstreamEvent)
	RecordCacheEvent(ctx context.Context, event domain.CacheEvent)
	Snapshot(ctx context.Context) (domain.MonitoringSnapshot, error)
}

type MonitoringMetricsRepository interface {
	BeginRequest(ctx context.Context)
	FinishRequest(ctx context.Context, record domain.RequestRecord)
	RecordAccessDecision(ctx context.Context, event domain.AccessDecisionEvent)
	RecordRateLimitDecision(ctx context.Context, event domain.RateLimitDecisionEvent)
	RecordUpstream(ctx context.Context, event domain.UpstreamEvent)
	RecordCacheEvent(ctx context.Context, event domain.CacheEvent)
}

type IPAccessRulesReader interface {
	ListRules(ctx context.Context) ([]domain.IPRule, error)
}

type RateLimitInspector interface {
	CheckRateLimit(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error)
	ListRules(ctx context.Context) ([]domain.RateLimitRule, error)
	ListBuckets(ctx context.Context) ([]domain.RateLimitBucketSnapshot, error)
}

type BeginRequestInput struct{}

type FinishRequestInput struct {
	Timestamp         time.Time
	Method            string
	Path              string
	ClientIP          string
	Status            int
	Latency           time.Duration
	BytesIn           int64
	BytesOut          int64
	AccessDecision    string
	RateLimitDecision string
	CacheStatus       string
	UpstreamLatency   time.Duration
}

type RecordAccessDecisionInput struct {
	IP       string
	Decision string
	RuleID   string
}

type RecordRateLimitInput struct {
	IP           string
	RuleID       string
	LimitType    string
	Allowed      bool
	CurrentValue int
}

type RecordUpstreamInput struct {
	Path       string
	StatusCode int
	Latency    time.Duration
	Error      string
}

type RecordCacheInput struct {
	Key    string
	Domain string
	Path   string
	Action string
	Rule   string
}

type MonitoringUseCase struct {
	repo      MonitoringRepository
	metrics   MonitoringMetricsRepository
	ipAccess  IPAccessRulesReader
	rateLimit RateLimitInspector
}

func NewMonitoringUseCase(repo MonitoringRepository, metrics MonitoringMetricsRepository, ipAccess IPAccessRulesReader, rateLimit RateLimitInspector) *MonitoringUseCase {
	return &MonitoringUseCase{
		repo:      repo,
		metrics:   metrics,
		ipAccess:  ipAccess,
		rateLimit: rateLimit,
	}
}

func (u *MonitoringUseCase) BeginRequest(ctx context.Context, _ BeginRequestInput) {
	u.repo.BeginRequest(ctx)
	u.metrics.BeginRequest(ctx)
}

func (u *MonitoringUseCase) FinishRequest(ctx context.Context, input FinishRequestInput) {
	record := domain.RequestRecord{
		Timestamp:         input.Timestamp,
		Method:            input.Method,
		Path:              input.Path,
		ClientIP:          input.ClientIP,
		Status:            input.Status,
		Latency:           input.Latency,
		BytesIn:           input.BytesIn,
		BytesOut:          input.BytesOut,
		AccessDecision:    input.AccessDecision,
		RateLimitDecision: input.RateLimitDecision,
		CacheStatus:       input.CacheStatus,
		UpstreamLatency:   input.UpstreamLatency,
	}
	u.repo.FinishRequest(ctx, record)
	u.metrics.FinishRequest(ctx, record)
}

func (u *MonitoringUseCase) RecordAccessDecision(ctx context.Context, input RecordAccessDecisionInput) {
	event := domain.AccessDecisionEvent{
		IP:       input.IP,
		Decision: input.Decision,
		RuleID:   input.RuleID,
	}
	u.repo.RecordAccessDecision(ctx, event)
	u.metrics.RecordAccessDecision(ctx, event)
}

func (u *MonitoringUseCase) RecordRateLimitDecision(ctx context.Context, input RecordRateLimitInput) {
	event := domain.RateLimitDecisionEvent{
		IP:           input.IP,
		RuleID:       input.RuleID,
		LimitType:    input.LimitType,
		Allowed:      input.Allowed,
		CurrentValue: input.CurrentValue,
	}
	u.repo.RecordRateLimitDecision(ctx, event)
	u.metrics.RecordRateLimitDecision(ctx, event)
}

func (u *MonitoringUseCase) RecordUpstream(ctx context.Context, input RecordUpstreamInput) {
	event := domain.UpstreamEvent{
		Path:       input.Path,
		StatusCode: input.StatusCode,
		Latency:    input.Latency,
		Error:      input.Error,
	}
	u.repo.RecordUpstream(ctx, event)
	u.metrics.RecordUpstream(ctx, event)
}

func (u *MonitoringUseCase) RecordCacheEvent(ctx context.Context, input RecordCacheInput) {
	event := domain.CacheEvent{
		Key:    input.Key,
		Domain: input.Domain,
		Path:   input.Path,
		Action: input.Action,
		Rule:   input.Rule,
	}
	u.repo.RecordCacheEvent(ctx, event)
	u.metrics.RecordCacheEvent(ctx, event)
}

func (u *MonitoringUseCase) GetMetrics(ctx context.Context) (domain.MetricsResponse, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.MetricsResponse{}, err
	}

	return domain.MetricsResponse{
		Uptime:                   time.Since(snapshot.StartedAt).String(),
		TotalRequests:            snapshot.TotalRequests,
		ActiveRequests:           snapshot.ActiveRequests,
		TotalBlockedRequests:     snapshot.TotalBlockedRequests,
		TotalRateLimitedRequests: snapshot.TotalRateLimitedRequests,
		TotalUpstreamErrors:      snapshot.TotalUpstreamErrors,
		RequestsPerSecond:        snapshot.RequestsPerSecond,
		AverageLatencyMS:         snapshot.AverageLatencyMS,
		BytesIn:                  snapshot.BytesIn,
		BytesOut:                 snapshot.BytesOut,
		ActiveConnections:        snapshot.ActiveConnections,
		RequestsByStatus:         snapshot.RequestsByStatus,
		RequestsByMethod:         snapshot.RequestsByMethod,
		RequestsByPath:           snapshot.RequestsByPath,
		CacheHits:                snapshot.CacheHits,
		CacheMisses:              snapshot.CacheMisses,
		CacheStores:              snapshot.CacheStores,
		CacheInvalidations:       snapshot.CacheInvalidations,
	}, nil
}

func (u *MonitoringUseCase) GetDashboardOverview(ctx context.Context) (domain.DashboardOverview, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.DashboardOverview{}, err
	}

	upstreamStatus := "unknown"
	if snapshot.Upstream.Healthy {
		upstreamStatus = "healthy"
	} else if snapshot.Upstream.TotalRequests > 0 {
		upstreamStatus = "degraded"
	}

	return domain.DashboardOverview{
		Uptime:            time.Since(snapshot.StartedAt).String(),
		TotalRequests:     snapshot.TotalRequests,
		CurrentRPS:        snapshot.RequestsPerSecond,
		ActiveConnections: snapshot.ActiveConnections,
		BlockedRequests:   snapshot.TotalBlockedRequests,
		UpstreamStatus:    upstreamStatus,
		AverageLatencyMS:  snapshot.AverageLatencyMS,
	}, nil
}

func (u *MonitoringUseCase) GetDashboardClients(ctx context.Context) (domain.DashboardClients, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.DashboardClients{}, err
	}

	return domain.DashboardClients{
		TopClientsByRequests: sortClients(snapshot.Clients, func(item domain.ClientSnapshot) uint64 {
			return item.Requests
		}),
		TopClientsByTraffic: sortClients(snapshot.Clients, func(item domain.ClientSnapshot) uint64 {
			return item.BytesIn + item.BytesOut
		}),
		BlockedClients: sortFilteredClients(snapshot.Clients, func(item domain.ClientSnapshot) bool {
			return item.BlockedRequests > 0
		}, func(item domain.ClientSnapshot) uint64 {
			return item.BlockedRequests
		}),
		RateLimitedClients: sortFilteredClients(snapshot.Clients, func(item domain.ClientSnapshot) bool {
			return item.RateLimitedRequests > 0
		}, func(item domain.ClientSnapshot) uint64 {
			return item.RateLimitedRequests
		}),
	}, nil
}

func (u *MonitoringUseCase) GetDashboardUpstream(ctx context.Context) (domain.DashboardUpstream, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.DashboardUpstream{}, err
	}

	return domain.DashboardUpstream{
		Healthy:       snapshot.Upstream.Healthy,
		LatencyMS:     snapshot.Upstream.LatencyMS,
		Errors:        snapshot.Upstream.Errors,
		LastError:     snapshot.Upstream.LastError,
		TotalRequests: snapshot.Upstream.TotalRequests,
		LastStatus:    snapshot.Upstream.LastStatus,
		LastUpdatedAt: snapshot.Upstream.LastUpdatedAt,
	}, nil
}

func (u *MonitoringUseCase) GetDashboardRateLimits(ctx context.Context) (domain.DashboardRateLimits, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.DashboardRateLimits{}, err
	}

	rules, err := u.rateLimit.ListRules(ctx)
	if err != nil {
		return domain.DashboardRateLimits{}, err
	}

	buckets, err := u.rateLimit.ListBuckets(ctx)
	if err != nil {
		return domain.DashboardRateLimits{}, err
	}

	blockedIPs := make([]domain.BlockedClient, 0, len(snapshot.RateLimitedClients))
	for ip, count := range snapshot.RateLimitedClients {
		blockedIPs = append(blockedIPs, domain.BlockedClient{
			IP:    ip,
			Count: count,
		})
	}
	sort.Slice(blockedIPs, func(i, j int) bool { return blockedIPs[i].Count > blockedIPs[j].Count })

	return domain.DashboardRateLimits{
		ActiveRules:        rules,
		CurrentBucketUsage: buckets,
		Violations:         snapshot.RateLimitViolations,
		BlockedIPs:         blockedIPs,
	}, nil
}

func (u *MonitoringUseCase) GetDashboardIPAccess(ctx context.Context) (domain.DashboardIPAccess, error) {
	snapshot, err := u.repo.Snapshot(ctx)
	if err != nil {
		return domain.DashboardIPAccess{}, err
	}

	rules, err := u.ipAccess.ListRules(ctx)
	if err != nil {
		return domain.DashboardIPAccess{}, err
	}

	response := domain.DashboardIPAccess{
		DenyStatistics:        cloneMap(snapshot.AccessDecisionStats),
		MatchedRuleStatistics: cloneMap(snapshot.MatchedRuleStats),
	}

	for _, rule := range rules {
		switch rule.Type {
		case domain.ListTypeAllow:
			response.Allowlist = append(response.Allowlist, rule)
		case domain.ListTypeDeny:
			response.Denylist = append(response.Denylist, rule)
		case domain.ListTypeGray:
			response.Graylist = append(response.Graylist, rule)
		}
	}

	return response, nil
}

func sortClients(clients []domain.ClientSnapshot, selector func(domain.ClientSnapshot) uint64) []domain.ClientSnapshot {
	cloned := append([]domain.ClientSnapshot(nil), clients...)
	sort.Slice(cloned, func(i, j int) bool {
		if selector(cloned[i]) == selector(cloned[j]) {
			return strings.Compare(cloned[i].IP, cloned[j].IP) < 0
		}
		return selector(cloned[i]) > selector(cloned[j])
	})
	if len(cloned) > 10 {
		cloned = cloned[:10]
	}
	return cloned
}

func sortFilteredClients(clients []domain.ClientSnapshot, filter func(domain.ClientSnapshot) bool, selector func(domain.ClientSnapshot) uint64) []domain.ClientSnapshot {
	filtered := make([]domain.ClientSnapshot, 0, len(clients))
	for _, client := range clients {
		if filter(client) {
			filtered = append(filtered, client)
		}
	}
	return sortClients(filtered, selector)
}

func cloneMap(source map[string]uint64) map[string]uint64 {
	cloned := make(map[string]uint64, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
