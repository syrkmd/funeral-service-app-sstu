package domain

import "net/netip"

type ListType string

const (
	ListTypeAllow ListType = "allowlist"
	ListTypeDeny  ListType = "denylist"
	ListTypeGray  ListType = "graylist"
)

type DefaultPolicy string

const (
	DefaultPolicyAllow DefaultPolicy = "allow"
	DefaultPolicyDeny  DefaultPolicy = "deny"
)

type IPRule struct {
	ID          string   `json:"id" yaml:"id"`
	Type        ListType `json:"type" yaml:"type"`
	Value       string   `json:"value" yaml:"value"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

type IPMatcher interface {
	Match(addr netip.Addr) bool
	Description() string
}

type CompiledIPRule struct {
	Rule    IPRule
	Matcher IPMatcher
}

type AccessSnapshot struct {
	Version       uint64
	DefaultPolicy DefaultPolicy
	DenyRules     []CompiledIPRule
	AllowRules    []CompiledIPRule
	GrayRules     []CompiledIPRule
}

type AccessDecision struct {
	IP            string `json:"ip"`
	Allowed       bool   `json:"allowed"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	MatchedRuleID string `json:"matched_rule_id,omitempty"`
	MatchedValue  string `json:"matched_value,omitempty"`
}
