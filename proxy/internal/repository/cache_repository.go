package repository

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type CacheRepository struct {
	mu sync.RWMutex

	entries  map[string]domain.CacheEntry
	tagIndex map[string]map[string]struct{}

	cleanupIntervalNanos atomic.Int64
}

func NewCacheRepository(cleanupInterval time.Duration) *CacheRepository {
	repo := &CacheRepository{
		entries:  make(map[string]domain.CacheEntry),
		tagIndex: make(map[string]map[string]struct{}),
	}
	repo.SetCleanupInterval(cleanupInterval)
	return repo
}

func (r *CacheRepository) SetCleanupInterval(interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	r.cleanupIntervalNanos.Store(int64(interval))
}

func (r *CacheRepository) Get(_ context.Context, key string) (domain.CacheEntry, bool, error) {
	now := time.Now().UTC()

	r.mu.RLock()
	entry, exists := r.entries[key]
	r.mu.RUnlock()
	if !exists {
		return domain.CacheEntry{}, false, nil
	}

	if !entry.Metadata.ExpiresAt.After(now) {
		r.mu.Lock()
		current, exists := r.entries[key]
		if exists && !current.Metadata.ExpiresAt.After(now) {
			r.deleteLocked(key, current)
		}
		r.mu.Unlock()
		return domain.CacheEntry{}, false, nil
	}

	return cloneCacheEntry(entry), true, nil
}

func (r *CacheRepository) Set(_ context.Context, entry domain.CacheEntry) error {
	cloned := cloneCacheEntry(entry)

	r.mu.Lock()
	defer r.mu.Unlock()

	if current, exists := r.entries[cloned.Metadata.Key]; exists {
		r.deleteLocked(cloned.Metadata.Key, current)
	}

	r.entries[cloned.Metadata.Key] = cloned
	r.indexTagsLocked(cloned.Metadata.Key, cloned.Metadata.Tags)
	return nil
}

func (r *CacheRepository) Delete(_ context.Context, key string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry, exists := r.entries[key]; exists {
		r.deleteLocked(key, entry)
		return 1, nil
	}

	return 0, nil
}

func (r *CacheRepository) Exists(_ context.Context, key string) (bool, error) {
	now := time.Now().UTC()

	r.mu.RLock()
	entry, exists := r.entries[key]
	r.mu.RUnlock()
	if !exists {
		return false, nil
	}

	if !entry.Metadata.ExpiresAt.After(now) {
		r.mu.Lock()
		current, exists := r.entries[key]
		if exists && !current.Metadata.ExpiresAt.After(now) {
			r.deleteLocked(key, current)
		}
		r.mu.Unlock()
		return false, nil
	}

	return true, nil
}

func (r *CacheRepository) DeleteByPrefix(_ context.Context, prefix string) (int, error) {
	if prefix == "" {
		return 0, nil
	}

	r.mu.RLock()
	keys := make([]string, 0)
	for key := range r.entries {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		entry, exists := r.entries[key]
		if !exists {
			continue
		}
		r.deleteLocked(key, entry)
		deleted++
	}

	return deleted, nil
}

func (r *CacheRepository) DeleteByRegex(_ context.Context, pattern *regexp.Regexp) (int, error) {
	if pattern == nil {
		return 0, nil
	}

	r.mu.RLock()
	keys := make([]string, 0)
	for key := range r.entries {
		if pattern.MatchString(key) {
			keys = append(keys, key)
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		entry, exists := r.entries[key]
		if !exists {
			continue
		}
		r.deleteLocked(key, entry)
		deleted++
	}

	return deleted, nil
}

func (r *CacheRepository) DeleteByTags(_ context.Context, tags []string) (int, error) {
	uniqueKeys := make(map[string]struct{})
	r.mu.RLock()
	for _, tag := range tags {
		keys, exists := r.tagIndex[tag]
		if !exists {
			continue
		}
		for key := range keys {
			uniqueKeys[key] = struct{}{}
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for key := range uniqueKeys {
		entry, exists := r.entries[key]
		if !exists {
			continue
		}
		r.deleteLocked(key, entry)
		deleted++
	}

	return deleted, nil
}

func (r *CacheRepository) CleanupExpired(_ context.Context) (int, error) {
	return r.cleanupExpired(time.Now().UTC()), nil
}

func (r *CacheRepository) Clear(_ context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := len(r.entries)
	r.entries = make(map[string]domain.CacheEntry)
	r.tagIndex = make(map[string]map[string]struct{})
	return deleted, nil
}

func (r *CacheRepository) StartCleanup(ctx context.Context) {
	timer := time.NewTimer(time.Duration(r.cleanupIntervalNanos.Load()))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			r.cleanupExpired(time.Now().UTC())
			timer.Reset(time.Duration(r.cleanupIntervalNanos.Load()))
		}
	}
}

func (r *CacheRepository) cleanupExpired(now time.Time) int {
	r.mu.RLock()
	keys := make([]string, 0)
	for key, entry := range r.entries {
		if !entry.Metadata.ExpiresAt.After(now) {
			keys = append(keys, key)
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		entry, exists := r.entries[key]
		if !exists || entry.Metadata.ExpiresAt.After(now) {
			continue
		}
		r.deleteLocked(key, entry)
		deleted++
	}
	return deleted
}

func (r *CacheRepository) deleteLocked(key string, entry domain.CacheEntry) {
	delete(r.entries, key)
	for _, tag := range entry.Metadata.Tags {
		keys, exists := r.tagIndex[tag]
		if !exists {
			continue
		}
		delete(keys, key)
		if len(keys) == 0 {
			delete(r.tagIndex, tag)
		}
	}
}

func (r *CacheRepository) indexTagsLocked(key string, tags []string) {
	for _, tag := range tags {
		keys, exists := r.tagIndex[tag]
		if !exists {
			keys = make(map[string]struct{})
			r.tagIndex[tag] = keys
		}
		keys[key] = struct{}{}
	}
}

func cloneCacheEntry(entry domain.CacheEntry) domain.CacheEntry {
	return domain.CacheEntry{
		Metadata: cloneCacheMetadata(entry.Metadata),
		Body:     append([]byte(nil), entry.Body...),
	}
}

func cloneCacheMetadata(metadata domain.CacheMetadata) domain.CacheMetadata {
	headers := make(map[string][]string, len(metadata.Headers))
	for key, values := range metadata.Headers {
		headers[key] = append([]string(nil), values...)
	}

	return domain.CacheMetadata{
		Key:         metadata.Key,
		Headers:     headers,
		StatusCode:  metadata.StatusCode,
		ContentType: metadata.ContentType,
		TTL:         metadata.TTL,
		CreatedAt:   metadata.CreatedAt,
		ExpiresAt:   metadata.ExpiresAt,
		Tags:        append([]string(nil), metadata.Tags...),
		Domain:      metadata.Domain,
		Path:        metadata.Path,
		Size:        metadata.Size,
	}
}
