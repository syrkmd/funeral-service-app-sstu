package usecase

import (
	"context"
	"net/url"
	"time"

	"proxy/internal/config"
	"proxy/internal/domain"
)

type IPAccessRepository interface {
	ListRules(ctx context.Context) ([]domain.IPRule, error)
	AddRule(ctx context.Context, rule domain.IPRule) (domain.IPRule, error)
	DeleteRule(ctx context.Context, id string) error
	GetSnapshot(ctx context.Context) (domain.AccessSnapshot, error)
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
}

type RateLimitStateRepository interface {
	GetBucket(ctx context.Context, key string) (domain.BucketState, bool, error)
	SaveBucket(ctx context.Context, key string, state domain.BucketState) error
	DeleteStaleBuckets(ctx context.Context, before time.Time) error
}

type RateLimiter interface {
	CheckRateLimit(ctx context.Context, rawIP string) (domain.RateLimitDecision, error)
}

type ProxyResolver interface {
	ResolveUpstream(ctx context.Context) (*url.URL, error)
	FlushInterval() int64
}
