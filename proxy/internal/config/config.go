package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"proxy/internal/domain"
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
	DefaultPolicy domain.DefaultPolicy `yaml:"default_policy"`
	Lists         []domain.IPRule      `yaml:"lists"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type RateLimitSubnetConfig struct {
	ID          string `yaml:"id"`
	CIDR        string `yaml:"cidr"`
	RPS         int    `yaml:"rps"`
	RPM         int    `yaml:"rpm"`
	RPH         int    `yaml:"rph"`
	RPD         int    `yaml:"rpd"`
	Description string `yaml:"description"`
}

type RateLimitConfig struct {
	Enabled bool                    `yaml:"enabled"`
	RPS     int                     `yaml:"rps"`
	RPM     int                     `yaml:"rpm"`
	RPH     int                     `yaml:"rph"`
	RPD     int                     `yaml:"rpd"`
	Subnets []RateLimitSubnetConfig `yaml:"subnets"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Proxy     ProxyConfig     `yaml:"proxy"`
	Access    AccessConfig    `yaml:"access"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
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

	if c.Proxy.UpstreamURL == "" {
		return fmt.Errorf("proxy.upstream_url is required")
	}

	if c.RateLimit.RPS < 0 || c.RateLimit.RPM < 0 || c.RateLimit.RPH < 0 || c.RateLimit.RPD < 0 {
		return fmt.Errorf("rate_limit values must be >= 0")
	}

	for _, subnet := range c.RateLimit.Subnets {
		if subnet.CIDR == "" {
			return fmt.Errorf("rate_limit subnet cidr is required")
		}
		if subnet.RPS < 0 || subnet.RPM < 0 || subnet.RPH < 0 || subnet.RPD < 0 {
			return fmt.Errorf("rate_limit subnet values must be >= 0")
		}
	}

	return nil
}
