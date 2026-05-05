package repository

import (
	"context"
	"sync"
	"time"

	"proxy/internal/domain"
)

type RateLimitRepository struct {
	mu      sync.RWMutex
	buckets map[string]domain.BucketState
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
