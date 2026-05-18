package repository

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type RateLimitRepository struct {
	mu          sync.RWMutex
	buckets     map[string]domain.BucketState
	connections sync.Map
}

type connectionCounter struct {
	count atomic.Int64
}

func NewRateLimitRepository() *RateLimitRepository {
	return &RateLimitRepository{
		buckets: make(map[string]domain.BucketState),
	}
}

func (r *RateLimitRepository) GetBucket(_ context.Context, key string) (domain.BucketState, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.buckets[key]
	return state, ok, nil
}

func (r *RateLimitRepository) SaveBucket(_ context.Context, key string, state domain.BucketState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buckets[key] = state
	return nil
}

func (r *RateLimitRepository) DeleteStaleBuckets(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, state := range r.buckets {
		if state.LastRefill.Before(before) {
			delete(r.buckets, key)
		}
	}

	return nil
}

func (r *RateLimitRepository) SnapshotBuckets(_ context.Context) ([]domain.RateLimitBucketSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshots := make([]domain.RateLimitBucketSnapshot, 0, len(r.buckets))
	for key, state := range r.buckets {
		scope, limitType, value := parseBucketKey(key)
		snapshots = append(snapshots, domain.RateLimitBucketSnapshot{
			Key:             key,
			Scope:           scope,
			LimitType:       limitType,
			Value:           value,
			TokensRemaining: state.Tokens,
			LastRefill:      state.LastRefill,
		})
	}

	return snapshots, nil
}

func (r *RateLimitRepository) StartCleanup(ctx context.Context, interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = r.DeleteStaleBuckets(ctx, time.Now().UTC().Add(-ttl))
		}
	}
}

func (r *RateLimitRepository) TryAcquireConnection(_ context.Context, key string, limit int) (bool, int, error) {
	if limit <= 0 {
		return true, 0, nil
	}

	value, _ := r.connections.LoadOrStore(key, &connectionCounter{})
	counter := value.(*connectionCounter)

	for {
		current := counter.count.Load()
		if current >= int64(limit) {
			return false, int(current), nil
		}
		if counter.count.CompareAndSwap(current, current+1) {
			return true, int(current + 1), nil
		}
	}
}

func (r *RateLimitRepository) ReleaseConnection(_ context.Context, key string) error {
	value, exists := r.connections.Load(key)
	if !exists {
		return nil
	}

	counter := value.(*connectionCounter)
	for {
		current := counter.count.Load()
		if current <= 0 {
			return nil
		}
		next := current - 1
		if counter.count.CompareAndSwap(current, next) {
			if next == 0 {
				r.connections.CompareAndDelete(key, counter)
			}
			return nil
		}
	}
}

func parseBucketKey(key string) (string, string, string) {
	parts := strings.SplitN(key, ":", 3)
	if len(parts) != 3 {
		return "", "", key
	}
	return parts[0], parts[1], parts[2]
}
