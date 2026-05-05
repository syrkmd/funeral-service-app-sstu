package domain

import "errors"

var (
	ErrInvalidRuleType   = errors.New("invalid IP list type")
	ErrInvalidPolicy     = errors.New("invalid default policy")
	ErrRuleNotFound      = errors.New("rule not found")
	ErrDuplicateRuleID   = errors.New("duplicate rule id")
	ErrEmptyRuleValue    = errors.New("rule value is required")
	ErrInvalidIP         = errors.New("invalid IP address")
	ErrInvalidUpstream   = errors.New("invalid upstream URL")
	ErrNilAccessSnapshot = errors.New("access snapshot is not initialized")
)
