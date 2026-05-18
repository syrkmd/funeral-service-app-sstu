package repository

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type verifiedIPEntry struct {
	expiresAt time.Time
}

type VerifiedIPRepository struct {
	mu       sync.RWMutex
	ttlNanos atomic.Int64
	entries  map[string]verifiedIPEntry
}

func NewVerifiedIPRepository(ttl time.Duration) *VerifiedIPRepository {
	repo := &VerifiedIPRepository{
		entries: make(map[string]verifiedIPEntry),
	}
	repo.SetTTL(ttl)
	return repo
}

func (r *VerifiedIPRepository) SetTTL(ttl time.Duration) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	r.ttlNanos.Store(int64(ttl))
}

func (r *VerifiedIPRepository) IsVerified(_ context.Context, ip string) (bool, error) {
	now := time.Now().UTC()

	r.mu.RLock()
	entry, exists := r.entries[ip]
	r.mu.RUnlock()
	if !exists {
		return false, nil
	}

	if !entry.expiresAt.After(now) {
		r.mu.Lock()
		current, exists := r.entries[ip]
		if exists && !current.expiresAt.After(now) {
			delete(r.entries, ip)
		}
		r.mu.Unlock()
		return false, nil
	}

	return true, nil
}

func (r *VerifiedIPRepository) MarkVerified(_ context.Context, ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[ip] = verifiedIPEntry{
		expiresAt: time.Now().UTC().Add(time.Duration(r.ttlNanos.Load())),
	}
	return nil
}

func (r *VerifiedIPRepository) StartCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.cleanupExpired(time.Now().UTC())
		}
	}
}

func (r *VerifiedIPRepository) cleanupExpired(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for ip, entry := range r.entries {
		if !entry.expiresAt.After(now) {
			delete(r.entries, ip)
		}
	}
}
