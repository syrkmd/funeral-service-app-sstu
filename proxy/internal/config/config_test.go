package config

import (
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"testing"
)

func TestConfigValidateRejectsEmptyCacheRule(t *testing.T) {
	cfg := validConfig()
	cfg.Cache.Rules = []CacheRuleConfig{{Name: "empty"}}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected empty cache rule to be rejected")
	}
}

func TestConfigValidateRejectsInvalidWildcardDomain(t *testing.T) {
	cfg := validConfig()
	cfg.Cache.Rules = []CacheRuleConfig{{Name: "bad-domain", Domains: []string{"*example.com"}}}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid wildcard domain to be rejected")
	}
}

func TestConfigValidateRejectsInvalidPathPattern(t *testing.T) {
	cfg := validConfig()
	cfg.Cache.Rules = []CacheRuleConfig{{Name: "bad-path", Paths: []string{"api/*/v1"}}}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid path pattern to be rejected")
	}
}

func TestConfigValidateRejectsDuplicateRuleScope(t *testing.T) {
	cfg := validConfig()
	cfg.Cache.Rules = []CacheRuleConfig{
		{Name: "first", Domains: []string{"example.com"}, Paths: []string{"/api/*"}},
		{Name: "second", Domains: []string{"example.com"}, Paths: []string{"/api/*"}},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate rule scope to be rejected")
	}
}

func validConfig() Config {
	return Config{
		Server: ServerConfig{Address: ":8080"},
		Proxy:  ProxyConfig{UpstreamURL: "https://example.com"},
		Access: AccessConfig{DefaultPolicy: domain.DefaultPolicyAllow},
		Cache: CacheConfig{
			Enabled:      true,
			MinEntrySize: 0,
			MaxEntrySize: 1024,
			Cache2XX:     true,
			Cache3XX:     false,
			Cache4XX:     false,
			Cache5XX:     false,
		},
	}
}
