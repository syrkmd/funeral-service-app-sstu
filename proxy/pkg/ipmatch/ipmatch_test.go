package ipmatch

import (
	"net/netip"
	"reflect"
	"sync"
	"testing"
)

func mustAddr(t *testing.T, raw string) netip.Addr {
	t.Helper()
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		t.Fatalf("parse addr %q: %v", raw, err)
	}
	return addr
}

func mustPrefix(t *testing.T, raw string) netip.Prefix {
	t.Helper()
	prefix, err := netip.ParsePrefix(raw)
	if err != nil {
		t.Fatalf("parse prefix %q: %v", raw, err)
	}
	return prefix.Masked()
}

func stringifyPrefixes(prefixes []netip.Prefix) []string {
	values := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		values = append(values, prefix.String())
	}
	return values
}

func TestParseMatcherTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantDesc    string
		matchIP     string
		noMatchIP   string
		expectError bool
	}{
		{
			name:      "ipv4 exact match",
			input:     "192.168.1.100",
			wantDesc:  "192.168.1.100",
			matchIP:   "192.168.1.100",
			noMatchIP: "192.168.1.101",
		},
		{
			name:      "cidr match",
			input:     "10.0.0.0/8",
			wantDesc:  "10.0.0.0/8",
			matchIP:   "10.20.30.40",
			noMatchIP: "11.0.0.1",
		},
		{
			name:      "ip range match",
			input:     "192.168.1.10-192.168.1.20",
			wantDesc:  "192.168.1.10-192.168.1.20",
			matchIP:   "192.168.1.15",
			noMatchIP: "192.168.1.21",
		},
		{
			name:        "empty value",
			input:       "   ",
			expectError: true,
		},
		{
			name:        "invalid ip",
			input:       "300.1.1.1",
			expectError: true,
		},
		{
			name:        "invalid cidr",
			input:       "10.0.0.0/99",
			expectError: true,
		},
		{
			name:        "invalid range start",
			input:       "nope-192.168.1.10",
			expectError: true,
		},
		{
			name:        "invalid range end",
			input:       "192.168.1.10-nope",
			expectError: true,
		},
		{
			name:        "mixed family range",
			input:       "192.168.1.10-2001:db8::1",
			expectError: true,
		},
		{
			name:        "reversed range",
			input:       "192.168.1.20-192.168.1.10",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher, err := Parse(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matcher.Description() != tt.wantDesc {
				t.Fatalf("unexpected description: got %q want %q", matcher.Description(), tt.wantDesc)
			}
			if tt.matchIP != "" && !matcher.Match(mustAddr(t, tt.matchIP)) {
				t.Fatalf("expected %s to match %q", tt.matchIP, tt.input)
			}
			if tt.noMatchIP != "" && matcher.Match(mustAddr(t, tt.noMatchIP)) {
				t.Fatalf("expected %s not to match %q", tt.noMatchIP, tt.input)
			}
		})
	}
}

func TestPrefixesTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		want        []string
		expectError bool
	}{
		{
			name:  "single ipv4",
			input: "192.168.1.100",
			want:  []string{"192.168.1.100/32"},
		},
		{
			name:  "cidr prefix",
			input: "10.0.0.0/8",
			want:  []string{"10.0.0.0/8"},
		},
		{
			name:  "aligned range to one prefix",
			input: "192.168.1.0-192.168.1.255",
			want:  []string{"192.168.1.0/24"},
		},
		{
			name:  "small range splits into minimal set",
			input: "192.168.1.10-192.168.1.15",
			want:  []string{"192.168.1.10/31", "192.168.1.12/30"},
		},
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid cidr",
			input:       "10.0.0.0/99",
			expectError: true,
		},
		{
			name:        "mixed family range",
			input:       "192.168.1.1-2001:db8::1",
			expectError: true,
		},
		{
			name:        "reversed range",
			input:       "192.168.1.15-192.168.1.10",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prefixes, err := Prefixes(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := stringifyPrefixes(prefixes); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected prefixes: got %v want %v", got, tt.want)
			}
		})
	}
}

func TestRuleSetMatchTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(t *testing.T) *RuleSet[string]
		addr  string
		want  string
		found bool
	}{
		{
			name: "exact ipv4 match",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "192.168.1.100/32")}, 10, "exact")
				return rs
			},
			addr:  "192.168.1.100",
			want:  "exact",
			found: true,
		},
		{
			name: "cidr match",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.0.0.0/8")}, 10, "private")
				return rs
			},
			addr:  "10.20.30.40",
			want:  "private",
			found: true,
		},
		{
			name: "nested cidr lower order wins",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.0.0.0/8")}, 20, "broad")
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.1.0.0/16")}, 10, "narrow")
				return rs
			},
			addr:  "10.1.2.3",
			want:  "narrow",
			found: true,
		},
		{
			name: "overlapping prefixes earlier order wins",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.0.0.0/8")}, 5, "first")
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.1.0.0/16")}, 10, "second")
				return rs
			},
			addr:  "10.1.2.3",
			want:  "first",
			found: true,
		},
		{
			name: "range prefixes match boundary start",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				prefixes, err := Prefixes("192.168.1.10-192.168.1.20")
				if err != nil {
					t.Fatalf("unexpected prefixes error: %v", err)
				}
				rs.InsertPrefixes(prefixes, 10, "range")
				return rs
			},
			addr:  "192.168.1.10",
			want:  "range",
			found: true,
		},
		{
			name: "range prefixes match boundary end",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				prefixes, err := Prefixes("192.168.1.10-192.168.1.20")
				if err != nil {
					t.Fatalf("unexpected prefixes error: %v", err)
				}
				rs.InsertPrefixes(prefixes, 10, "range")
				return rs
			},
			addr:  "192.168.1.20",
			want:  "range",
			found: true,
		},
		{
			name: "not found ipv4",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.0.0.0/8")}, 10, "private")
				return rs
			},
			addr:  "11.0.0.1",
			found: false,
		},
		{
			name: "ipv6 exact match",
			build: func(t *testing.T) *RuleSet[string] {
				rs := NewRuleSet[string]()
				rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "2001:db8::1/128")}, 10, "v6")
				return rs
			},
			addr:  "2001:db8::1",
			want:  "v6",
			found: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rs := tt.build(t)
			got, ok := rs.Match(mustAddr(t, tt.addr))
			if ok != tt.found {
				t.Fatalf("unexpected found state: got %v want %v", ok, tt.found)
			}
			if tt.found && got != tt.want {
				t.Fatalf("unexpected value: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestRuleSetMatchNotFoundForZeroAddr(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet[string]()
	if got, ok := rs.Match(netip.Addr{}); ok || got != "" {
		t.Fatalf("expected zero-value addr lookup to miss, got ok=%v value=%q", ok, got)
	}
}

func TestTrieHelpersTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		start         string
		end           string
		wantLargest   string
		wantLast      string
		wantNext      string
		bitChecks     map[int]int
		wantAddrRound string
	}{
		{
			name:          "ipv4 helpers",
			start:         "192.168.1.0",
			end:           "192.168.1.255",
			wantLargest:   "192.168.1.0/24",
			wantLast:      "192.168.1.255",
			wantNext:      "192.168.2.0",
			bitChecks:     map[int]int{0: 1, 1: 1, 2: 0},
			wantAddrRound: "192.168.1.0",
		},
		{
			name:          "ipv6 helpers",
			start:         "2001:db8::",
			end:           "2001:db8::ffff",
			wantLargest:   "2001:db8::/112",
			wantLast:      "2001:db8::ffff",
			wantNext:      "2001:db8::1:0",
			bitChecks:     map[int]int{0: 0, 1: 0, 2: 1},
			wantAddrRound: "2001:db8::",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start := mustAddr(t, tt.start)
			end := mustAddr(t, tt.end)

			largest := largestPrefixWithinRange(start, end)
			if largest.String() != tt.wantLargest {
				t.Fatalf("unexpected largest prefix: got %s want %s", largest, tt.wantLargest)
			}

			if got := lastAddrOfPrefix(largest).String(); got != tt.wantLast {
				t.Fatalf("unexpected last addr: got %s want %s", got, tt.wantLast)
			}

			if got := nextAddr(lastAddrOfPrefix(largest)).String(); got != tt.wantNext {
				t.Fatalf("unexpected next addr: got %s want %s", got, tt.wantNext)
			}

			bytes := addrBytes(start)
			roundTrip := bytesToAddr(start.BitLen(), bytes)
			if roundTrip.String() != tt.wantAddrRound {
				t.Fatalf("unexpected round trip addr: got %s want %s", roundTrip, tt.wantAddrRound)
			}

			for bit, want := range tt.bitChecks {
				if got := bitAt(start, bit); got != want {
					t.Fatalf("unexpected bit %d: got %d want %d", bit, got, want)
				}
			}
		})
	}
}

func TestRuleSetConcurrentReads(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet[string]()
	rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.0.0.0/8")}, 20, "broad")
	rs.InsertPrefixes([]netip.Prefix{mustPrefix(t, "10.1.0.0/16")}, 10, "narrow")

	addrs := []string{
		"10.1.1.1",
		"10.2.2.2",
		"11.0.0.1",
		"10.1.255.255",
	}

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		for _, raw := range addrs {
			wg.Add(1)
			go func(raw string) {
				defer wg.Done()
				addr := mustAddr(t, raw)
				_, _ = rs.Match(addr)
			}(raw)
		}
	}
	wg.Wait()
}
