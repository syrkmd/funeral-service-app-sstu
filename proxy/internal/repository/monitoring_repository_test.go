package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

func TestNewMonitoringRepositoryInitialSnapshot(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}

	if snapshot.StartedAt.IsZero() {
		t.Fatal("expected StartedAt to be initialized")
	}
	if snapshot.TotalRequests != 0 || snapshot.ActiveRequests != 0 || snapshot.ActiveConnections != 0 {
		t.Fatalf("expected empty counters, got %+v", snapshot)
	}
	if len(snapshot.RequestsByStatus) != 0 || len(snapshot.RequestsByMethod) != 0 || len(snapshot.RequestsByPath) != 0 {
		t.Fatalf("expected empty request maps, got %+v", snapshot)
	}
	if len(snapshot.Clients) != 0 || len(snapshot.AccessDecisionStats) != 0 || len(snapshot.RateLimitViolations) != 0 {
		t.Fatalf("expected empty aggregates, got %+v", snapshot)
	}
	if len(snapshot.CacheInvalidations) != 0 || snapshot.CacheHits != 0 || snapshot.CacheMisses != 0 || snapshot.CacheStores != 0 {
		t.Fatalf("expected empty cache stats, got %+v", snapshot)
	}
}

func TestMonitoringRepositoryBeginAndFinishRequest(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	repo.BeginRequest(context.Background())

	mid, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot mid error: %v", err)
	}
	if mid.ActiveRequests != 1 || mid.ActiveConnections != 1 {
		t.Fatalf("expected active counters incremented, got %+v", mid)
	}

	repo.FinishRequest(context.Background(), domain.RequestRecord{
		Timestamp:         time.Now().UTC(),
		Method:            "GET",
		Path:              "/resource",
		ClientIP:          "127.0.0.1",
		Status:            201,
		Latency:           50 * time.Millisecond,
		BytesIn:           123,
		BytesOut:          456,
		AccessDecision:    "allow",
		RateLimitDecision: "allow",
		CacheStatus:       "hit",
		UpstreamLatency:   20 * time.Millisecond,
	})

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if snapshot.TotalRequests != 1 || snapshot.ActiveRequests != 0 || snapshot.ActiveConnections != 0 {
		t.Fatalf("unexpected request counters: %+v", snapshot)
	}
	if snapshot.BytesIn != 123 || snapshot.BytesOut != 456 {
		t.Fatalf("unexpected traffic totals: %+v", snapshot)
	}
	if snapshot.RequestsByStatus["201"] != 1 || snapshot.RequestsByMethod["GET"] != 1 || snapshot.RequestsByPath["/resource"] != 1 {
		t.Fatalf("unexpected request aggregates: %+v", snapshot)
	}
	if snapshot.AverageLatencyMS <= 0 {
		t.Fatalf("expected positive average latency, got %f", snapshot.AverageLatencyMS)
	}
	if len(snapshot.Clients) != 1 || snapshot.Clients[0].Requests != 1 || snapshot.Clients[0].BytesIn != 123 || snapshot.Clients[0].BytesOut != 456 {
		t.Fatalf("unexpected client stats: %+v", snapshot.Clients)
	}
}

func TestMonitoringRepositoryRecordAccessDecision(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	repo.RecordAccessDecision(context.Background(), domain.AccessDecisionEvent{
		IP:       "10.0.0.1",
		Decision: "deny",
		RuleID:   "deny-private",
	})
	repo.RecordAccessDecision(context.Background(), domain.AccessDecisionEvent{
		IP:       "127.0.0.1",
		Decision: "allow",
		RuleID:   "allow-local",
	})

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if snapshot.AccessDecisionStats["deny"] != 1 || snapshot.AccessDecisionStats["allow"] != 1 {
		t.Fatalf("unexpected access decision stats: %+v", snapshot.AccessDecisionStats)
	}
	if snapshot.MatchedRuleStats["deny-private"] != 1 || snapshot.MatchedRuleStats["allow-local"] != 1 {
		t.Fatalf("unexpected matched rule stats: %+v", snapshot.MatchedRuleStats)
	}
	if snapshot.TotalBlockedRequests != 1 {
		t.Fatalf("expected one blocked request, got %d", snapshot.TotalBlockedRequests)
	}
	if len(snapshot.Clients) != 1 || snapshot.Clients[0].IP != "10.0.0.1" || snapshot.Clients[0].BlockedRequests != 1 {
		t.Fatalf("expected only denied client snapshot to be tracked, got %+v", snapshot.Clients)
	}
}

func TestMonitoringRepositoryRecordRateLimitDecision(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	repo.RecordRateLimitDecision(context.Background(), domain.RateLimitDecisionEvent{
		IP:           "192.168.1.10",
		RuleID:       "office-subnet",
		LimitType:    "rps",
		Allowed:      false,
		CurrentValue: 25,
	})
	repo.RecordRateLimitDecision(context.Background(), domain.RateLimitDecisionEvent{
		IP:           "192.168.1.10",
		RuleID:       "office-subnet",
		LimitType:    "rps",
		Allowed:      true,
		CurrentValue: 1,
	})

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if snapshot.TotalBlockedRequests != 1 || snapshot.TotalRateLimitedRequests != 1 {
		t.Fatalf("unexpected blocked/rate-limited totals: %+v", snapshot)
	}
	if snapshot.RateLimitViolations["office-subnet:rps"] != 1 {
		t.Fatalf("unexpected rate violation aggregates: %+v", snapshot.RateLimitViolations)
	}
	if snapshot.RateLimitedClients["192.168.1.10"] != 1 {
		t.Fatalf("unexpected rate limited clients: %+v", snapshot.RateLimitedClients)
	}
}

func TestMonitoringRepositoryRecordUpstream(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	repo.RecordUpstream(context.Background(), domain.UpstreamEvent{
		Path:       "/ok",
		StatusCode: 200,
		Latency:    25 * time.Millisecond,
	})
	repo.RecordUpstream(context.Background(), domain.UpstreamEvent{
		Path:       "/bad",
		StatusCode: 502,
		Latency:    30 * time.Millisecond,
		Error:      "upstream failed",
	})

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if snapshot.Upstream.TotalRequests != 2 || snapshot.TotalUpstreamErrors != 1 {
		t.Fatalf("unexpected upstream totals: %+v", snapshot)
	}
	if snapshot.Upstream.Errors != 1 || snapshot.Upstream.LastError != "upstream failed" || snapshot.Upstream.Healthy {
		t.Fatalf("unexpected upstream state after error: %+v", snapshot.Upstream)
	}
	if snapshot.Upstream.LatencyMS <= 0 {
		t.Fatalf("expected upstream latency recorded, got %+v", snapshot.Upstream)
	}
}

func TestMonitoringRepositoryRecordCacheEventTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action string
		verify func(t *testing.T, snapshot domain.MonitoringSnapshot)
	}{
		{
			action: "hit",
			verify: func(t *testing.T, snapshot domain.MonitoringSnapshot) {
				if snapshot.CacheHits != 1 {
					t.Fatalf("expected cache hit count 1, got %+v", snapshot)
				}
			},
		},
		{
			action: "miss",
			verify: func(t *testing.T, snapshot domain.MonitoringSnapshot) {
				if snapshot.CacheMisses != 1 {
					t.Fatalf("expected cache miss count 1, got %+v", snapshot)
				}
			},
		},
		{
			action: "store",
			verify: func(t *testing.T, snapshot domain.MonitoringSnapshot) {
				if snapshot.CacheStores != 1 {
					t.Fatalf("expected cache store count 1, got %+v", snapshot)
				}
			},
		},
		{
			action: "invalidate_prefix",
			verify: func(t *testing.T, snapshot domain.MonitoringSnapshot) {
				if snapshot.CacheInvalidations["invalidate_prefix"] != 1 {
					t.Fatalf("expected cache invalidation aggregate, got %+v", snapshot.CacheInvalidations)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()
			repo := NewMonitoringRepository()
			repo.RecordCacheEvent(context.Background(), domain.CacheEvent{Action: tt.action})
			snapshot, err := repo.Snapshot(context.Background())
			if err != nil {
				t.Fatalf("Snapshot error: %v", err)
			}
			tt.verify(t, snapshot)
		})
	}
}

func TestMonitoringRepositorySnapshotDefensiveCopy(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	repo.FinishRequest(context.Background(), domain.RequestRecord{
		Timestamp: time.Now().UTC(),
		Method:    "GET",
		Path:      "/path",
		ClientIP:  "127.0.0.1",
		Status:    200,
		Latency:   10 * time.Millisecond,
	})
	repo.RecordAccessDecision(context.Background(), domain.AccessDecisionEvent{
		IP:       "127.0.0.1",
		Decision: "allow",
		RuleID:   "allow-local",
	})

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}

	snapshot.RequestsByStatus["200"] = 99
	snapshot.RequestsByMethod["GET"] = 99
	snapshot.RequestsByPath["/path"] = 99
	snapshot.AccessDecisionStats["allow"] = 99
	snapshot.MatchedRuleStats["allow-local"] = 99
	snapshot.Clients[0].Requests = 99

	again, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if again.RequestsByStatus["200"] != 1 || again.RequestsByMethod["GET"] != 1 || again.RequestsByPath["/path"] != 1 {
		t.Fatalf("request maps were mutated through snapshot: %+v", again)
	}
	if again.AccessDecisionStats["allow"] != 1 || again.MatchedRuleStats["allow-local"] != 1 {
		t.Fatalf("access maps were mutated through snapshot: %+v", again)
	}
	if again.Clients[0].Requests != 1 {
		t.Fatalf("client slice was mutated through snapshot: %+v", again.Clients)
	}
}

func TestMonitoringRepositoryConcurrency(t *testing.T) {
	t.Parallel()

	repo := NewMonitoringRepository()
	var wg sync.WaitGroup

	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			repo.BeginRequest(context.Background())
			repo.RecordAccessDecision(context.Background(), domain.AccessDecisionEvent{
				IP:       "127.0.0.1",
				Decision: "allow",
				RuleID:   "allow-local",
			})
			repo.RecordRateLimitDecision(context.Background(), domain.RateLimitDecisionEvent{
				IP:        "127.0.0.1",
				RuleID:    "default:rps",
				LimitType: "rps",
				Allowed:   i%2 == 0,
			})
			repo.RecordCacheEvent(context.Background(), domain.CacheEvent{Action: "hit"})
			repo.FinishRequest(context.Background(), domain.RequestRecord{
				Timestamp: time.Now().UTC(),
				Method:    "GET",
				Path:      "/resource",
				ClientIP:  "127.0.0.1",
				Status:    200,
				Latency:   time.Millisecond,
				BytesIn:   10,
				BytesOut:  20,
			})
		}(i)
	}

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.Snapshot(context.Background()); err != nil {
				t.Errorf("Snapshot error during concurrency test: %v", err)
			}
		}()
	}

	wg.Wait()

	snapshot, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("final Snapshot error: %v", err)
	}
	if snapshot.TotalRequests != 32 || snapshot.ActiveRequests != 0 || snapshot.ActiveConnections != 0 {
		t.Fatalf("unexpected concurrency totals: %+v", snapshot)
	}
	if snapshot.CacheHits != 32 {
		t.Fatalf("unexpected cache hit count after concurrency: %+v", snapshot)
	}
}
