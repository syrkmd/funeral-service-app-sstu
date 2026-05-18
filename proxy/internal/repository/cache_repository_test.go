package repository

import (
	"context"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

func testCacheEntry(key string, ttl time.Duration, expiresAt time.Time, tags ...string) domain.CacheEntry {
	return domain.CacheEntry{
		Metadata: domain.CacheMetadata{
			Key:         key,
			Headers:     map[string][]string{"Content-Type": {"application/json"}},
			StatusCode:  200,
			ContentType: "application/json",
			TTL:         ttl,
			CreatedAt:   expiresAt.Add(-ttl),
			ExpiresAt:   expiresAt,
			Tags:        append([]string(nil), tags...),
			Domain:      "example.com",
			Path:        "/" + key,
			Size:        5,
		},
		Body: []byte("body:" + key),
	}
}

func seedEntries(t *testing.T, repo *CacheRepository, entries ...domain.CacheEntry) {
	t.Helper()
	for _, entry := range entries {
		if err := repo.Set(context.Background(), entry); err != nil {
			t.Fatalf("seed Set(%s): %v", entry.Metadata.Key, err)
		}
	}
}

func TestCacheRepositorySetGetAndMiss(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	entry := testCacheEntry("alpha", time.Minute, now.Add(time.Minute), "users")

	if err := repo.Set(context.Background(), entry); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, found, err := repo.Get(context.Background(), "alpha")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if got.Metadata.Key != entry.Metadata.Key || string(got.Body) != string(entry.Body) {
		t.Fatalf("unexpected entry returned: got=%+v want=%+v", got.Metadata, entry.Metadata)
	}

	_, found, err = repo.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Get missing: %v", err)
	}
	if found {
		t.Fatal("expected miss for absent key")
	}
}

func TestCacheRepositoryOverwriteExistingEntry(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	first := testCacheEntry("same", time.Minute, now.Add(time.Minute), "one")
	second := testCacheEntry("same", 2*time.Minute, now.Add(2*time.Minute), "two")
	second.Body = []byte("second")

	seedEntries(t, repo, first, second)

	got, found, err := repo.Get(context.Background(), "same")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected hit")
	}
	if string(got.Body) != "second" {
		t.Fatalf("expected overwritten body, got %q", string(got.Body))
	}
	if _, exists := repo.tagIndex["one"]; exists {
		t.Fatal("expected old tag index to be removed after overwrite")
	}
	if _, exists := repo.tagIndex["two"]; !exists {
		t.Fatal("expected new tag index to exist after overwrite")
	}
}

func TestCacheRepositoryGetReturnsDefensiveCopies(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	entry := testCacheEntry("copy", time.Minute, now.Add(time.Minute), "tag")
	seedEntries(t, repo, entry)

	got, found, err := repo.Get(context.Background(), "copy")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected hit")
	}

	got.Body[0] = 'X'
	got.Metadata.Headers["Content-Type"][0] = "text/plain"
	got.Metadata.Tags[0] = "mutated"

	again, found, err := repo.Get(context.Background(), "copy")
	if err != nil {
		t.Fatalf("Get again: %v", err)
	}
	if !found {
		t.Fatal("expected hit on second read")
	}
	if string(again.Body) != string(entry.Body) {
		t.Fatalf("body mutation leaked into storage: got %q want %q", string(again.Body), string(entry.Body))
	}
	if again.Metadata.Headers["Content-Type"][0] != "application/json" {
		t.Fatalf("header mutation leaked into storage: got %q", again.Metadata.Headers["Content-Type"][0])
	}
	if again.Metadata.Tags[0] != "tag" {
		t.Fatalf("tag mutation leaked into storage: got %q", again.Metadata.Tags[0])
	}
}

func TestCacheRepositoryTTLBehavior(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()

	tests := []struct {
		name       string
		entry      domain.CacheEntry
		wantFound  bool
		wantExists bool
	}{
		{
			name:       "non expired entry available",
			entry:      testCacheEntry("active", time.Minute, now.Add(time.Minute)),
			wantFound:  true,
			wantExists: true,
		},
		{
			name:       "expired entry unavailable",
			entry:      testCacheEntry("expired", time.Minute, now.Add(-time.Second)),
			wantFound:  false,
			wantExists: false,
		},
		{
			name:       "zero ttl zero expiry treated expired",
			entry:      testCacheEntry("zero", 0, time.Time{}),
			wantFound:  false,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := repo.Set(context.Background(), tt.entry); err != nil {
				t.Fatalf("Set: %v", err)
			}

			_, found, err := repo.Get(context.Background(), tt.entry.Metadata.Key)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if found != tt.wantFound {
				t.Fatalf("unexpected found state: got %v want %v", found, tt.wantFound)
			}

			exists, err := repo.Exists(context.Background(), tt.entry.Metadata.Key)
			if err != nil {
				t.Fatalf("Exists: %v", err)
			}
			if exists != tt.wantExists {
				t.Fatalf("unexpected exists state: got %v want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestCacheRepositoryCleanupExpiredRemovesOnlyExpired(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	expiredA := testCacheEntry("expired-a", time.Minute, now.Add(-time.Second), "users")
	expiredB := testCacheEntry("expired-b", time.Minute, now.Add(-2*time.Second), "posts")
	active := testCacheEntry("active", time.Minute, now.Add(time.Minute), "users")
	seedEntries(t, repo, expiredA, expiredB, active)

	deleted, err := repo.CleanupExpired(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpired: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 expired deletions, got %d", deleted)
	}
	if _, found, _ := repo.Get(context.Background(), "active"); !found {
		t.Fatal("expected active entry to remain")
	}
	if _, exists := repo.tagIndex["posts"]; exists {
		t.Fatal("expected dangling post tag index to be cleaned")
	}
}

func TestCacheRepositoryDeleteTable(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("a", time.Minute, now.Add(time.Minute), "shared"),
		testCacheEntry("b", time.Minute, now.Add(time.Minute), "shared"),
	)

	tests := []struct {
		name        string
		key         string
		wantDeleted int
		wantA       bool
		wantB       bool
	}{
		{name: "delete exact existing", key: "a", wantDeleted: 1, wantA: false, wantB: true},
		{name: "delete absent key", key: "missing", wantDeleted: 0, wantA: false, wantB: true},
		{name: "delete second key", key: "b", wantDeleted: 1, wantA: false, wantB: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.Delete(context.Background(), tt.key)
			if err != nil {
				t.Fatalf("Delete: %v", err)
			}
			if deleted != tt.wantDeleted {
				t.Fatalf("unexpected deleted count: got %d want %d", deleted, tt.wantDeleted)
			}
			a, _ := repo.Exists(context.Background(), "a")
			b, _ := repo.Exists(context.Background(), "b")
			if a != tt.wantA || b != tt.wantB {
				t.Fatalf("unexpected key presence: got a=%v b=%v want a=%v b=%v", a, b, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestCacheRepositoryDeleteByPrefixTable(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("users:1", time.Minute, now.Add(time.Minute), "users"),
		testCacheEntry("users:2", time.Minute, now.Add(time.Minute), "users"),
		testCacheEntry("posts:1", time.Minute, now.Add(time.Minute), "posts"),
	)

	tests := []struct {
		name        string
		prefix      string
		wantDeleted int
		wantUsers1  bool
		wantUsers2  bool
		wantPosts1  bool
	}{
		{name: "empty prefix no-op", prefix: "", wantDeleted: 0, wantUsers1: true, wantUsers2: true, wantPosts1: true},
		{name: "matching prefix deletes subset", prefix: "users:", wantDeleted: 2, wantUsers1: false, wantUsers2: false, wantPosts1: true},
		{name: "no match prefix", prefix: "missing:", wantDeleted: 0, wantUsers1: false, wantUsers2: false, wantPosts1: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.DeleteByPrefix(context.Background(), tt.prefix)
			if err != nil {
				t.Fatalf("DeleteByPrefix: %v", err)
			}
			if deleted != tt.wantDeleted {
				t.Fatalf("unexpected deleted count: got %d want %d", deleted, tt.wantDeleted)
			}
			users1, _ := repo.Exists(context.Background(), "users:1")
			users2, _ := repo.Exists(context.Background(), "users:2")
			posts1, _ := repo.Exists(context.Background(), "posts:1")
			if users1 != tt.wantUsers1 || users2 != tt.wantUsers2 || posts1 != tt.wantPosts1 {
				t.Fatalf("unexpected key presence: got users1=%v users2=%v posts1=%v", users1, users2, posts1)
			}
		})
	}
}

func TestCacheRepositoryDeleteByRegexTable(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("users:1", time.Minute, now.Add(time.Minute), "users"),
		testCacheEntry("users:22", time.Minute, now.Add(time.Minute), "users"),
		testCacheEntry("posts:1", time.Minute, now.Add(time.Minute), "posts"),
	)

	tests := []struct {
		name        string
		pattern     *regexp.Regexp
		wantDeleted int
		wantUsers1  bool
		wantUsers22 bool
		wantPosts1  bool
	}{
		{name: "nil regex no-op", pattern: nil, wantDeleted: 0, wantUsers1: true, wantUsers22: true, wantPosts1: true},
		{name: "partial regex match", pattern: regexp.MustCompile(`users:\d+$`), wantDeleted: 2, wantUsers1: false, wantUsers22: false, wantPosts1: true},
		{name: "no-match regex", pattern: regexp.MustCompile(`^comments:`), wantDeleted: 0, wantUsers1: false, wantUsers22: false, wantPosts1: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.DeleteByRegex(context.Background(), tt.pattern)
			if err != nil {
				t.Fatalf("DeleteByRegex: %v", err)
			}
			if deleted != tt.wantDeleted {
				t.Fatalf("unexpected deleted count: got %d want %d", deleted, tt.wantDeleted)
			}
			users1, _ := repo.Exists(context.Background(), "users:1")
			users22, _ := repo.Exists(context.Background(), "users:22")
			posts1, _ := repo.Exists(context.Background(), "posts:1")
			if users1 != tt.wantUsers1 || users22 != tt.wantUsers22 || posts1 != tt.wantPosts1 {
				t.Fatalf("unexpected key presence: got users1=%v users22=%v posts1=%v", users1, users22, posts1)
			}
		})
	}
}

func TestCacheRepositoryDeleteByTagsTable(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("users:list", time.Minute, now.Add(time.Minute), "users", "list"),
		testCacheEntry("users:item", time.Minute, now.Add(time.Minute), "users", "item"),
		testCacheEntry("posts:list", time.Minute, now.Add(time.Minute), "posts", "list"),
		testCacheEntry("empty-tags", time.Minute, now.Add(time.Minute)),
		testCacheEntry("duplicate-tags", time.Minute, now.Add(time.Minute), "dup", "dup"),
	)

	tests := []struct {
		name        string
		tags        []string
		wantDeleted int
		wantKeys    map[string]bool
	}{
		{
			name:        "single tag invalidation",
			tags:        []string{"item"},
			wantDeleted: 1,
			wantKeys: map[string]bool{
				"users:list":     true,
				"users:item":     false,
				"posts:list":     true,
				"empty-tags":     true,
				"duplicate-tags": true,
			},
		},
		{
			name:        "shared tag invalidation",
			tags:        []string{"list"},
			wantDeleted: 2,
			wantKeys: map[string]bool{
				"users:list":     false,
				"users:item":     false,
				"posts:list":     false,
				"empty-tags":     true,
				"duplicate-tags": true,
			},
		},
		{
			name:        "duplicate tag key deleted once",
			tags:        []string{"dup"},
			wantDeleted: 1,
			wantKeys: map[string]bool{
				"users:list":     false,
				"users:item":     false,
				"posts:list":     false,
				"empty-tags":     true,
				"duplicate-tags": false,
			},
		},
		{
			name:        "unrelated tag no-op",
			tags:        []string{"missing"},
			wantDeleted: 0,
			wantKeys: map[string]bool{
				"users:list":     false,
				"users:item":     false,
				"posts:list":     false,
				"empty-tags":     true,
				"duplicate-tags": false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.DeleteByTags(context.Background(), tt.tags)
			if err != nil {
				t.Fatalf("DeleteByTags: %v", err)
			}
			if deleted != tt.wantDeleted {
				t.Fatalf("unexpected deleted count: got %d want %d", deleted, tt.wantDeleted)
			}
			for key, want := range tt.wantKeys {
				got, _ := repo.Exists(context.Background(), key)
				if got != want {
					t.Fatalf("unexpected key presence for %s: got %v want %v", key, got, want)
				}
			}
		})
	}
}

func TestCacheRepositoryClearRemovesEverything(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("a", time.Minute, now.Add(time.Minute), "x"),
		testCacheEntry("b", time.Minute, now.Add(time.Minute), "y"),
	)

	deleted, err := repo.Clear(context.Background())
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("unexpected deleted count: got %d want 2", deleted)
	}
	if len(repo.entries) != 0 || len(repo.tagIndex) != 0 {
		t.Fatalf("expected empty storage after clear, got entries=%d tags=%d", len(repo.entries), len(repo.tagIndex))
	}
}

func TestCacheRepositoryInternalIndexesConsistency(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	entry := testCacheEntry("users:1", time.Minute, now.Add(time.Minute), "users", "shared")
	seedEntries(t, repo, entry)

	if deleted, err := repo.Delete(context.Background(), "users:1"); err != nil || deleted != 1 {
		t.Fatalf("Delete: deleted=%d err=%v", deleted, err)
	}
	if len(repo.entries) != 0 {
		t.Fatalf("expected empty entries after delete, got %d", len(repo.entries))
	}
	if len(repo.tagIndex) != 0 {
		t.Fatalf("expected empty tag index after delete, got %d", len(repo.tagIndex))
	}

	if deleted, err := repo.Delete(context.Background(), "users:1"); err != nil || deleted != 0 {
		t.Fatalf("repeated delete: deleted=%d err=%v", deleted, err)
	}

	expired := testCacheEntry("expired", time.Minute, now.Add(-time.Second), "expired-tag")
	seedEntries(t, repo, expired)
	if deleted, err := repo.CleanupExpired(context.Background()); err != nil || deleted != 1 {
		t.Fatalf("CleanupExpired: deleted=%d err=%v", deleted, err)
	}
	if len(repo.tagIndex) != 0 {
		t.Fatalf("expected cleanup to remove dangling tag refs, got %d tags", len(repo.tagIndex))
	}
}

func TestCacheRepositoryEmptyCacheBehavior(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)

	if _, found, err := repo.Get(context.Background(), "missing"); err != nil || found {
		t.Fatalf("unexpected Get result on empty repo: found=%v err=%v", found, err)
	}
	if exists, err := repo.Exists(context.Background(), "missing"); err != nil || exists {
		t.Fatalf("unexpected Exists result on empty repo: exists=%v err=%v", exists, err)
	}
	if deleted, err := repo.DeleteByPrefix(context.Background(), "x"); err != nil || deleted != 0 {
		t.Fatalf("unexpected DeleteByPrefix on empty repo: deleted=%d err=%v", deleted, err)
	}
	if deleted, err := repo.DeleteByRegex(context.Background(), regexp.MustCompile(`.`)); err != nil || deleted != 0 {
		t.Fatalf("unexpected DeleteByRegex on empty repo: deleted=%d err=%v", deleted, err)
	}
	if deleted, err := repo.DeleteByTags(context.Background(), []string{"tag"}); err != nil || deleted != 0 {
		t.Fatalf("unexpected DeleteByTags on empty repo: deleted=%d err=%v", deleted, err)
	}
	if deleted, err := repo.Clear(context.Background()); err != nil || deleted != 0 {
		t.Fatalf("unexpected Clear on empty repo: deleted=%d err=%v", deleted, err)
	}
}

func TestCacheRepositoryStartCleanupRemovesExpiredEntries(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(5 * time.Millisecond)
	now := time.Now().UTC()
	seedEntries(t, repo,
		testCacheEntry("expired", time.Minute, now.Add(-time.Second), "expired"),
		testCacheEntry("active", time.Minute, now.Add(time.Minute), "active"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go repo.StartCleanup(ctx)
	time.Sleep(20 * time.Millisecond)
	cancel()
	time.Sleep(5 * time.Millisecond)

	if _, found, _ := repo.Get(context.Background(), "expired"); found {
		t.Fatal("expected cleanup worker to remove expired entry")
	}
	if _, found, _ := repo.Get(context.Background(), "active"); !found {
		t.Fatal("expected cleanup worker to keep active entry")
	}
}

func TestCacheRepositoryConcurrentAccess(t *testing.T) {
	t.Parallel()

	repo := NewCacheRepository(time.Minute)
	now := time.Now().UTC()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key-" + time.Now().Add(time.Duration(i)*time.Nanosecond).Format("150405.000000000")
			entry := testCacheEntry(key, time.Minute, now.Add(time.Minute), "shared")
			_ = repo.Set(ctx, entry)
			_, _, _ = repo.Get(ctx, key)
			_, _ = repo.Exists(ctx, key)
			_, _ = repo.DeleteByTags(ctx, []string{"missing"})
			_, _ = repo.DeleteByPrefix(ctx, "no-match")
		}(i)
	}

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = repo.Get(ctx, "missing")
			_, _ = repo.DeleteByRegex(ctx, regexp.MustCompile(`^never$`))
			_, _ = repo.CleanupExpired(ctx)
		}()
	}

	wg.Wait()
}
