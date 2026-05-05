package domain

import "time"

type RateLimitScope string

const (
	RateLimitScopeIP     RateLimitScope = "ip"
	RateLimitScopeSubnet RateLimitScope = "subnet"
)

type RateLimitRule struct {
	ID          string         `json:"id" yaml:"id"`
	Scope       RateLimitScope `json:"scope" yaml:"scope"`
	Value       string         `json:"value" yaml:"value"`
	RPS         int            `json:"rps" yaml:"rps"`
	RPM         int            `json:"rpm" yaml:"rpm"`
	RPH         int            `json:"rph" yaml:"rph"`
	RPD         int            `json:"rpd" yaml:"rpd"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
}

type RateLimitDecision struct {
	IP           string `json:"ip"`
	Allowed      bool   `json:"allowed"`
	Reason       string `json:"reason"`
	RuleID       string `json:"rule_id,omitempty"`
	RuleValue    string `json:"rule_value,omitempty"`
	LimitType    string `json:"type,omitempty"`
	CurrentValue int    `json:"current_value,omitempty"`
}

type BucketState struct {
	Tokens     float64
	LastRefill time.Time
}
