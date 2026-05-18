package config

import (
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a scalar")
	}

	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("parse duration %q: %w", value.Value, err)
	}

	d.Duration = parsed
	return nil
}

type ServerConfig struct {
	Address         string   `yaml:"address"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	IdleTimeout     Duration `yaml:"idle_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
}

type ProxyConfig struct {
	UpstreamURL   string   `yaml:"upstream_url"`
	FlushInterval Duration `yaml:"flush_interval"`
}

type AccessConfig struct {
	DefaultPolicy    domain.DefaultPolicy `yaml:"default_policy"`
	VerificationTTL  Duration             `yaml:"verification_ttl"`
	DecisionCacheTTL Duration             `yaml:"decision_cache_ttl"`
	Lists            []domain.IPRule      `yaml:"lists"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type RateLimitSubnetConfig struct {
	ID             string   `yaml:"id"`
	CIDR           string   `yaml:"cidr"`
	RPS            int      `yaml:"rps"`
	RPM            int      `yaml:"rpm"`
	RPH            int      `yaml:"rph"`
	RPD            int      `yaml:"rpd"`
	CPS            int      `yaml:"cps"`
	MaxConnections int      `yaml:"max_connections"`
	UploadBPS      int64    `yaml:"upload_bps"`
	DownloadBPS    int64    `yaml:"download_bps"`
	TotalBytes     int64    `yaml:"total_bytes"`
	TotalWindow    Duration `yaml:"total_window"`
	Description    string   `yaml:"description"`
}

type RateLimitConfig struct {
	Enabled        bool                    `yaml:"enabled"`
	RPS            int                     `yaml:"rps"`
	RPM            int                     `yaml:"rpm"`
	RPH            int                     `yaml:"rph"`
	RPD            int                     `yaml:"rpd"`
	CPS            int                     `yaml:"cps"`
	MaxConnections int                     `yaml:"max_connections"`
	UploadBPS      int64                   `yaml:"upload_bps"`
	DownloadBPS    int64                   `yaml:"download_bps"`
	TotalBytes     int64                   `yaml:"total_bytes"`
	TotalWindow    Duration                `yaml:"total_window"`
	Subnets        []RateLimitSubnetConfig `yaml:"subnets"`
}

type CacheConfig struct {
	Enabled              bool              `yaml:"enabled"`
	DefaultTTL           Duration          `yaml:"default_ttl"`
	CleanupInterval      Duration          `yaml:"cleanup_interval"`
	MaxEntrySize         int64             `yaml:"max_entry_size"`
	MinEntrySize         int64             `yaml:"min_entry_size"`
	Cache2XX             bool              `yaml:"cache_2xx"`
	Cache3XX             bool              `yaml:"cache_3xx"`
	Cache4XX             bool              `yaml:"cache_4xx"`
	Cache5XX             bool              `yaml:"cache_5xx"`
	TTL2XX               Duration          `yaml:"ttl_2xx"`
	TTL3XX               Duration          `yaml:"ttl_3xx"`
	TTL4XX               Duration          `yaml:"ttl_4xx"`
	TTL5XX               Duration          `yaml:"ttl_5xx"`
	ExcludedHeaders      []string          `yaml:"excluded_headers"`
	ExcludedContentTypes []string          `yaml:"excluded_content_types"`
	Rules                []CacheRuleConfig `yaml:"rules"`
}

type CacheRuleConfig struct {
	Name                 string   `yaml:"name"`
	Enabled              *bool    `yaml:"enabled"`
	Domains              []string `yaml:"domains"`
	Paths                []string `yaml:"paths"`
	DefaultTTL           Duration `yaml:"default_ttl"`
	TTL2XX               Duration `yaml:"ttl_2xx"`
	TTL3XX               Duration `yaml:"ttl_3xx"`
	TTL4XX               Duration `yaml:"ttl_4xx"`
	TTL5XX               Duration `yaml:"ttl_5xx"`
	MinEntrySize         *int64   `yaml:"min_entry_size"`
	MaxEntrySize         *int64   `yaml:"max_entry_size"`
	Cache2XX             *bool    `yaml:"cache_2xx"`
	Cache3XX             *bool    `yaml:"cache_3xx"`
	Cache4XX             *bool    `yaml:"cache_4xx"`
	Cache5XX             *bool    `yaml:"cache_5xx"`
	ExcludedHeaders      []string `yaml:"excluded_headers"`
	ExcludedContentTypes []string `yaml:"excluded_content_types"`
	Tags                 []string `yaml:"tags"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Proxy     ProxyConfig     `yaml:"proxy"`
	Access    AccessConfig    `yaml:"access"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Cache     CacheConfig     `yaml:"cache"`
	Logging   LoggingConfig   `yaml:"logging"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config file: %w", err)
	}

	if cfg.Proxy.FlushInterval.Duration <= 0 {
		cfg.Proxy.FlushInterval.Duration = 100 * time.Millisecond
	}
	if cfg.Access.VerificationTTL.Duration <= 0 {
		cfg.Access.VerificationTTL.Duration = 15 * time.Minute
	}
	if cfg.Access.DecisionCacheTTL.Duration <= 0 {
		cfg.Access.DecisionCacheTTL.Duration = 5 * time.Minute
	}
	if cfg.Cache.DefaultTTL.Duration <= 0 {
		cfg.Cache.DefaultTTL.Duration = 5 * time.Minute
	}
	if cfg.Cache.CleanupInterval.Duration <= 0 {
		cfg.Cache.CleanupInterval.Duration = time.Minute
	}
	if cfg.Cache.TTL2XX.Duration <= 0 {
		cfg.Cache.TTL2XX.Duration = cfg.Cache.DefaultTTL.Duration
	}
	if cfg.Cache.TTL3XX.Duration <= 0 {
		cfg.Cache.TTL3XX.Duration = 10 * time.Minute
	}
	if cfg.Cache.TTL4XX.Duration <= 0 {
		cfg.Cache.TTL4XX.Duration = 30 * time.Second
	}
	if cfg.Cache.TTL5XX.Duration <= 0 {
		cfg.Cache.TTL5XX.Duration = 10 * time.Second
	}
	if cfg.Cache.MaxEntrySize <= 0 {
		cfg.Cache.MaxEntrySize = 1 << 20
	}
	if cfg.Cache.MinEntrySize < 0 {
		cfg.Cache.MinEntrySize = 0
	}
	if cfg.RateLimit.TotalBytes > 0 && cfg.RateLimit.TotalWindow.Duration <= 0 {
		cfg.RateLimit.TotalWindow.Duration = time.Hour
	}
	for idx := range cfg.RateLimit.Subnets {
		if cfg.RateLimit.Subnets[idx].TotalBytes > 0 && cfg.RateLimit.Subnets[idx].TotalWindow.Duration <= 0 {
			cfg.RateLimit.Subnets[idx].TotalWindow.Duration = time.Hour
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}

	switch c.Access.DefaultPolicy {
	case domain.DefaultPolicyAllow, domain.DefaultPolicyDeny:
	default:
		return domain.ErrInvalidPolicy
	}
	if c.Access.VerificationTTL.Duration < 0 || c.Access.DecisionCacheTTL.Duration < 0 {
		return fmt.Errorf("access durations must be >= 0")
	}

	if c.Proxy.UpstreamURL == "" {
		return fmt.Errorf("proxy.upstream_url is required")
	}
	if c.Cache.MaxEntrySize < 0 || c.Cache.MinEntrySize < 0 {
		return fmt.Errorf("cache entry sizes must be >= 0")
	}
	if c.Cache.MaxEntrySize > 0 && c.Cache.MinEntrySize > c.Cache.MaxEntrySize {
		return fmt.Errorf("cache min_entry_size must be <= max_entry_size")
	}
	if c.Cache.DefaultTTL.Duration < 0 || c.Cache.CleanupInterval.Duration < 0 || c.Cache.TTL2XX.Duration < 0 || c.Cache.TTL3XX.Duration < 0 || c.Cache.TTL4XX.Duration < 0 || c.Cache.TTL5XX.Duration < 0 {
		return fmt.Errorf("cache durations must be >= 0")
	}
	for _, rule := range c.Cache.Rules {
		if rule.MinEntrySize != nil && *rule.MinEntrySize < 0 {
			return fmt.Errorf("cache rule entry sizes must be >= 0")
		}
		if rule.MaxEntrySize != nil && *rule.MaxEntrySize < 0 {
			return fmt.Errorf("cache rule entry sizes must be >= 0")
		}
		if rule.MinEntrySize != nil && rule.MaxEntrySize != nil && *rule.MaxEntrySize > 0 && *rule.MinEntrySize > *rule.MaxEntrySize {
			return fmt.Errorf("cache rule min_entry_size must be <= max_entry_size")
		}
		if rule.DefaultTTL.Duration < 0 || rule.TTL2XX.Duration < 0 || rule.TTL3XX.Duration < 0 || rule.TTL4XX.Duration < 0 || rule.TTL5XX.Duration < 0 {
			return fmt.Errorf("cache rule durations must be >= 0")
		}
	}
	if err := validateCacheRules(c.Cache.Rules); err != nil {
		return err
	}

	if c.RateLimit.RPS < 0 || c.RateLimit.RPM < 0 || c.RateLimit.RPH < 0 || c.RateLimit.RPD < 0 || c.RateLimit.CPS < 0 || c.RateLimit.MaxConnections < 0 || c.RateLimit.UploadBPS < 0 || c.RateLimit.DownloadBPS < 0 || c.RateLimit.TotalBytes < 0 {
		return fmt.Errorf("rate_limit values must be >= 0")
	}
	if c.RateLimit.TotalBytes > 0 && c.RateLimit.TotalWindow.Duration <= 0 {
		return fmt.Errorf("rate_limit total_window must be > 0 when total_bytes is configured")
	}

	for _, subnet := range c.RateLimit.Subnets {
		if subnet.CIDR == "" {
			return fmt.Errorf("rate_limit subnet cidr is required")
		}
		if subnet.RPS < 0 || subnet.RPM < 0 || subnet.RPH < 0 || subnet.RPD < 0 || subnet.CPS < 0 || subnet.MaxConnections < 0 || subnet.UploadBPS < 0 || subnet.DownloadBPS < 0 || subnet.TotalBytes < 0 {
			return fmt.Errorf("rate_limit subnet values must be >= 0")
		}
		if subnet.TotalBytes > 0 && subnet.TotalWindow.Duration <= 0 {
			return fmt.Errorf("rate_limit subnet total_window must be > 0 when total_bytes is configured")
		}
	}

	return nil
}

func validateCacheRules(rules []CacheRuleConfig) error {
	seen := make(map[string]string, len(rules))

	for idx, rule := range rules {
		domains, err := normalizeRuleDomains(rule.Domains)
		if err != nil {
			return fmt.Errorf("cache rule %d domains: %w", idx+1, err)
		}
		paths, err := normalizeRulePaths(rule.Paths)
		if err != nil {
			return fmt.Errorf("cache rule %d paths: %w", idx+1, err)
		}
		if len(domains) == 0 && len(paths) == 0 {
			return fmt.Errorf("cache rule %d must define at least one domain or path matcher", idx+1)
		}

		signature := strings.Join(domains, ",") + "|" + strings.Join(paths, ",")
		if existing, exists := seen[signature]; exists {
			return fmt.Errorf("cache rule %d duplicates matcher scope of rule %s", idx+1, existing)
		}
		seen[signature] = defaultRuleDisplayName(rule.Name, idx+1)
	}

	return nil
}

func normalizeRuleDomains(patterns []string) ([]string, error) {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if strings.Contains(pattern, "*") {
			if !strings.HasPrefix(pattern, "*.") || strings.Count(pattern, "*") != 1 {
				return nil, fmt.Errorf("invalid wildcard domain pattern %q", pattern)
			}
			suffix := strings.TrimPrefix(pattern, "*.")
			if suffix == "" || !isValidHostLabelSequence(suffix) {
				return nil, fmt.Errorf("invalid wildcard domain pattern %q", pattern)
			}
			normalized = append(normalized, pattern)
			continue
		}
		if net.ParseIP(pattern) == nil && !isValidHostLabelSequence(pattern) {
			return nil, fmt.Errorf("invalid domain pattern %q", pattern)
		}
		normalized = append(normalized, pattern)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func normalizeRulePaths(patterns []string) ([]string, error) {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.Count(pattern, "*") > 1 || (strings.Contains(pattern, "*") && !strings.HasSuffix(pattern, "*")) {
			return nil, fmt.Errorf("invalid path pattern %q", pattern)
		}
		base := strings.TrimSuffix(pattern, "*")
		if strings.Contains(base, "*") {
			return nil, fmt.Errorf("invalid path pattern %q", pattern)
		}
		if !strings.HasPrefix(base, "/") {
			return nil, fmt.Errorf("invalid path pattern %q", pattern)
		}
		cleaned := path.Clean(base)
		if cleaned == "." {
			cleaned = "/"
		}
		if strings.HasSuffix(pattern, "*") {
			if !strings.HasSuffix(cleaned, "/") {
				cleaned += "/"
			}
			normalized = append(normalized, cleaned+"*")
			continue
		}
		normalized = append(normalized, cleaned)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func isValidHostLabelSequence(value string) bool {
	if value == "" || strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") {
		return false
	}
	labels := strings.Split(value, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		for i, r := range label {
			isAlphaNum := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
			if !isAlphaNum && r != '-' {
				return false
			}
			if (i == 0 || i == len(label)-1) && r == '-' {
				return false
			}
		}
	}
	return true
}

func defaultRuleDisplayName(name string, index int) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	return fmt.Sprintf("rule-%d", index)
}
