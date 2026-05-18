package repository

import (
	"context"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type MonitoringRepository struct {
	startedAt atomic.Value

	totalRequests            atomic.Uint64
	activeRequests           atomic.Int64
	activeConnections        atomic.Int64
	totalBlockedRequests     atomic.Uint64
	totalRateLimitedRequests atomic.Uint64
	totalUpstreamErrors      atomic.Uint64
	cacheHits                atomic.Uint64
	cacheMisses              atomic.Uint64
	cacheStores              atomic.Uint64
	bytesIn                  atomic.Uint64
	bytesOut                 atomic.Uint64
	totalLatencyNanos        atomic.Uint64
	completedRequests        atomic.Uint64

	mu                 sync.RWMutex
	requestsByStatus   map[string]uint64
	requestsByMethod   map[string]uint64
	requestsByPath     map[string]uint64
	clients            map[string]*domain.ClientSnapshot
	accessStats        map[string]uint64
	matchedRuleStats   map[string]uint64
	rateViolations     map[string]uint64
	rateLimitedClients map[string]uint64
	cacheInvalidations map[string]uint64
	upstream           domain.UpstreamSnapshot
}

func NewMonitoringRepository() *MonitoringRepository {
	repo := &MonitoringRepository{
		requestsByStatus:   make(map[string]uint64),
		requestsByMethod:   make(map[string]uint64),
		requestsByPath:     make(map[string]uint64),
		clients:            make(map[string]*domain.ClientSnapshot),
		accessStats:        make(map[string]uint64),
		matchedRuleStats:   make(map[string]uint64),
		rateViolations:     make(map[string]uint64),
		rateLimitedClients: make(map[string]uint64),
		cacheInvalidations: make(map[string]uint64),
	}
	repo.startedAt.Store(time.Now().UTC())
	return repo
}

func (r *MonitoringRepository) BeginRequest(_ context.Context) {
	r.activeRequests.Add(1)
	r.activeConnections.Add(1)
}

func (r *MonitoringRepository) FinishRequest(_ context.Context, record domain.RequestRecord) {
	r.totalRequests.Add(1)
	r.completedRequests.Add(1)
	r.activeRequests.Add(-1)
	r.activeConnections.Add(-1)

	if record.BytesIn > 0 {
		r.bytesIn.Add(uint64(record.BytesIn))
	}
	if record.BytesOut > 0 {
		r.bytesOut.Add(uint64(record.BytesOut))
	}
	r.totalLatencyNanos.Add(uint64(record.Latency))

	statusKey := strconv.Itoa(record.Status)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestsByStatus[statusKey]++
	r.requestsByMethod[record.Method]++
	r.requestsByPath[record.Path]++

	client := r.getOrCreateClientLocked(record.ClientIP)
	client.Requests++
	if record.BytesIn > 0 {
		client.BytesIn += uint64(record.BytesIn)
	}
	if record.BytesOut > 0 {
		client.BytesOut += uint64(record.BytesOut)
	}
}

func (r *MonitoringRepository) RecordAccessDecision(_ context.Context, event domain.AccessDecisionEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accessStats[event.Decision]++
	if event.RuleID != "" {
		r.matchedRuleStats[event.RuleID]++
	}

	if event.Decision == "deny" {
		r.totalBlockedRequests.Add(1)
		client := r.getOrCreateClientLocked(event.IP)
		client.BlockedRequests++
	}
}

func (r *MonitoringRepository) RecordRateLimitDecision(_ context.Context, event domain.RateLimitDecisionEvent) {
	if event.Allowed {
		return
	}

	r.totalBlockedRequests.Add(1)
	r.totalRateLimitedRequests.Add(1)

	r.mu.Lock()
	defer r.mu.Unlock()

	key := event.LimitType
	if event.RuleID != "" {
		key = event.RuleID + ":" + event.LimitType
	}
	r.rateViolations[key]++
	r.rateLimitedClients[event.IP]++

	client := r.getOrCreateClientLocked(event.IP)
	client.BlockedRequests++
	client.RateLimitedRequests++
}

func (r *MonitoringRepository) RecordUpstream(_ context.Context, event domain.UpstreamEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.upstream.TotalRequests++
	r.upstream.LastStatus = event.StatusCode
	r.upstream.LastUpdatedAt = time.Now().UTC()
	r.upstream.LatencyMS = float64(event.Latency.Microseconds()) / 1000

	if event.Error != "" {
		r.totalUpstreamErrors.Add(1)
		r.upstream.Errors++
		r.upstream.LastError = event.Error
		r.upstream.Healthy = false
		return
	}

	r.upstream.LastError = ""
	r.upstream.Healthy = true
}

func (r *MonitoringRepository) RecordCacheEvent(_ context.Context, event domain.CacheEvent) {
	switch event.Action {
	case "hit":
		r.cacheHits.Add(1)
	case "miss":
		r.cacheMisses.Add(1)
	case "store":
		r.cacheStores.Add(1)
	default:
		r.mu.Lock()
		r.cacheInvalidations[event.Action]++
		r.mu.Unlock()
	}
}

func (r *MonitoringRepository) Snapshot(_ context.Context) (domain.MonitoringSnapshot, error) {
	snapshot := domain.MonitoringSnapshot{
		StartedAt:                r.startedAt.Load().(time.Time),
		TotalRequests:            r.totalRequests.Load(),
		ActiveRequests:           r.activeRequests.Load(),
		ActiveConnections:        r.activeConnections.Load(),
		TotalBlockedRequests:     r.totalBlockedRequests.Load(),
		TotalRateLimitedRequests: r.totalRateLimitedRequests.Load(),
		TotalUpstreamErrors:      r.totalUpstreamErrors.Load(),
		CacheHits:                r.cacheHits.Load(),
		CacheMisses:              r.cacheMisses.Load(),
		CacheStores:              r.cacheStores.Load(),
		BytesIn:                  r.bytesIn.Load(),
		BytesOut:                 r.bytesOut.Load(),
	}

	completed := r.completedRequests.Load()
	if completed > 0 {
		snapshot.AverageLatencyMS = float64(r.totalLatencyNanos.Load()) / float64(completed) / float64(time.Millisecond)
	}
	uptimeSeconds := time.Since(snapshot.StartedAt).Seconds()
	if uptimeSeconds > 0 {
		snapshot.RequestsPerSecond = float64(snapshot.TotalRequests) / uptimeSeconds
	}

	r.mu.RLock()
	snapshot.RequestsByStatus = cloneUint64Map(r.requestsByStatus)
	snapshot.RequestsByMethod = cloneUint64Map(r.requestsByMethod)
	snapshot.RequestsByPath = cloneUint64Map(r.requestsByPath)
	snapshot.AccessDecisionStats = cloneUint64Map(r.accessStats)
	snapshot.MatchedRuleStats = cloneUint64Map(r.matchedRuleStats)
	snapshot.RateLimitViolations = cloneUint64Map(r.rateViolations)
	snapshot.RateLimitedClients = cloneUint64Map(r.rateLimitedClients)
	snapshot.CacheInvalidations = cloneUint64Map(r.cacheInvalidations)
	snapshot.Upstream = r.upstream
	snapshot.Clients = cloneClients(r.clients)
	r.mu.RUnlock()

	return snapshot, nil
}

func (r *MonitoringRepository) getOrCreateClientLocked(ip string) *domain.ClientSnapshot {
	client, exists := r.clients[ip]
	if !exists {
		client = &domain.ClientSnapshot{IP: ip}
		r.clients[ip] = client
	}
	return client
}

func cloneUint64Map(source map[string]uint64) map[string]uint64 {
	cloned := make(map[string]uint64, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneClients(source map[string]*domain.ClientSnapshot) []domain.ClientSnapshot {
	clients := make([]domain.ClientSnapshot, 0, len(source))
	for _, client := range source {
		clients = append(clients, *client)
	}
	return clients
}
