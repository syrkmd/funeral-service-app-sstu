package usecase

import (
	"context"
	"net/url"
	"regexp"
	"time"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type IPAccessRepository interface {
	ListRules(ctx context.Context) ([]domain.IPRule, error)
	AddRule(ctx context.Context, rule domain.IPRule) (domain.IPRule, error)
	DeleteRule(ctx context.Context, id string) error
	GetSnapshot(ctx context.Context) (domain.AccessSnapshot, error)
}

type VerifiedIPRepository interface {
	IsVerified(ctx context.Context, ip string) (bool, error)
	MarkVerified(ctx context.Context, ip string) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (domain.CacheEntry, bool, error)
	Set(ctx context.Context, entry domain.CacheEntry) error
	Delete(ctx context.Context, key string) (int, error)
	DeleteByPrefix(ctx context.Context, prefix string) (int, error)
	DeleteByRegex(ctx context.Context, pattern *regexp.Regexp) (int, error)
	DeleteByTags(ctx context.Context, tags []string) (int, error)
	Exists(ctx context.Context, key string) (bool, error)
	CleanupExpired(ctx context.Context) (int, error)
	Clear(ctx context.Context) (int, error)
}

type ConfigProvider interface {
	Current() config.Config
	Subscribe(config.Subscriber)
}

type IPAccessChecker interface {
	ListRules(ctx context.Context) ([]domain.IPRule, error)
	AddRule(ctx context.Context, input AddRuleInput) (domain.IPRule, error)
	DeleteRule(ctx context.Context, id string) error
	CheckIP(ctx context.Context, rawIP string) (domain.AccessDecision, error)
	VerifyCaptcha(ctx context.Context, rawIP string, answer string) error
}

type RateLimitStateRepository interface {
	GetBucket(ctx context.Context, key string) (domain.BucketState, bool, error)
	SaveBucket(ctx context.Context, key string, state domain.BucketState) error
	DeleteStaleBuckets(ctx context.Context, before time.Time) error
	SnapshotBuckets(ctx context.Context) ([]domain.RateLimitBucketSnapshot, error)
	TryAcquireConnection(ctx context.Context, key string, limit int) (bool, int, error)
	ReleaseConnection(ctx context.Context, key string) error
}

type RateLimiter interface {
	CheckRateLimit(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error)
	ReserveResponseBandwidth(ctx context.Context, rawIP string, downloadBytes int64) (domain.RateLimitDecision, error)
	AccountTraffic(ctx context.Context, rawIP string, uploadBytes int64, downloadBytes int64) error
	ReleaseConnections(ctx context.Context, keys []string) error
}

type MonitoringRecorder interface {
	BeginRequest(ctx context.Context, input BeginRequestInput)
	FinishRequest(ctx context.Context, input FinishRequestInput)
	RecordAccessDecision(ctx context.Context, input RecordAccessDecisionInput)
	RecordRateLimitDecision(ctx context.Context, input RecordRateLimitInput)
	RecordUpstream(ctx context.Context, input RecordUpstreamInput)
	RecordCacheEvent(ctx context.Context, input RecordCacheInput)
}

type ProxyResolver interface {
	ResolveUpstream(ctx context.Context) (*url.URL, error)
	FlushInterval() int64
}

type UseCaseLogger interface {
	Error(msg string, err error)
}
