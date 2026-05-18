package domain

import (
	"regexp"
	"time"
)

type CacheEntry struct {
	Metadata CacheMetadata `json:"metadata"`
	Body     []byte        `json:"body"`
}

type CacheMetadata struct {
	Key         string              `json:"key"`
	Headers     map[string][]string `json:"headers"`
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	TTL         time.Duration       `json:"ttl"`
	CreatedAt   time.Time           `json:"created_at"`
	ExpiresAt   time.Time           `json:"expires_at"`
	Tags        []string            `json:"tags,omitempty"`
	Domain      string              `json:"domain,omitempty"`
	Path        string              `json:"path,omitempty"`
	Size        int64               `json:"size"`
}

type CacheRule struct {
	Name                 string        `json:"name"`
	Enabled              bool          `json:"enabled"`
	Domains              []string      `json:"domains,omitempty"`
	Paths                []string      `json:"paths,omitempty"`
	DefaultTTL           time.Duration `json:"default_ttl"`
	TTL2XX               time.Duration `json:"ttl_2xx"`
	TTL3XX               time.Duration `json:"ttl_3xx"`
	TTL4XX               time.Duration `json:"ttl_4xx"`
	TTL5XX               time.Duration `json:"ttl_5xx"`
	MinSize              int64         `json:"min_size,omitempty"`
	MaxSize              int64         `json:"max_size,omitempty"`
	Cache2XX             bool          `json:"cache_2xx"`
	Cache3XX             bool          `json:"cache_3xx"`
	Cache4XX             bool          `json:"cache_4xx"`
	Cache5XX             bool          `json:"cache_5xx"`
	ExcludedHeaders      []string      `json:"excluded_headers,omitempty"`
	ExcludedContentTypes []string      `json:"excluded_content_types,omitempty"`
	Tags                 []string      `json:"tags,omitempty"`
}

type CacheDecision struct {
	Allowed     bool          `json:"allowed"`
	Key         string        `json:"key,omitempty"`
	TTL         time.Duration `json:"ttl,omitempty"`
	Reason      string        `json:"reason"`
	MatchedRule string        `json:"matched_rule,omitempty"`
}

type CacheRequest struct {
	Method   string              `json:"method"`
	Scheme   string              `json:"scheme"`
	Domain   string              `json:"domain"`
	Path     string              `json:"path"`
	RawQuery string              `json:"raw_query,omitempty"`
	Headers  map[string][]string `json:"headers,omitempty"`
}

type CacheCandidate struct {
	Request     CacheRequest        `json:"request"`
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	Headers     map[string][]string `json:"headers,omitempty"`
	Size        int64               `json:"size"`
	Tags        []string            `json:"tags,omitempty"`
}

type CacheInvalidationEvent struct {
	Keys     []string       `json:"keys,omitempty"`
	Prefixes []string       `json:"prefixes,omitempty"`
	Tags     []string       `json:"tags,omitempty"`
	Regex    *regexp.Regexp `json:"-"`
}
