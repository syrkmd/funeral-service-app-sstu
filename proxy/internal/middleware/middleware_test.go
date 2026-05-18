package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

type stubIPAccessChecker struct {
	mu        sync.Mutex
	checkFn   func(ctx context.Context, rawIP string) (domain.AccessDecision, error)
	lastIPs   []string
	checkHits int
}

func (s *stubIPAccessChecker) ListRules(context.Context) ([]domain.IPRule, error) { return nil, nil }
func (s *stubIPAccessChecker) AddRule(context.Context, usecase.AddRuleInput) (domain.IPRule, error) {
	return domain.IPRule{}, nil
}
func (s *stubIPAccessChecker) DeleteRule(context.Context, string) error            { return nil }
func (s *stubIPAccessChecker) VerifyCaptcha(context.Context, string, string) error { return nil }
func (s *stubIPAccessChecker) CheckIP(ctx context.Context, rawIP string) (domain.AccessDecision, error) {
	s.mu.Lock()
	s.lastIPs = append(s.lastIPs, rawIP)
	s.checkHits++
	s.mu.Unlock()
	if s.checkFn != nil {
		return s.checkFn(ctx, rawIP)
	}
	return domain.AccessDecision{}, nil
}

type stubRateLimiter struct {
	mu                  sync.Mutex
	checkFn             func(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error)
	accountCalls        []trafficCall
	releaseCalls        [][]string
	checkCalls          []rateCheckCall
	reserveResponseHits int
}

type rateCheckCall struct {
	ip          string
	uploadBytes int64
}

type trafficCall struct {
	ip            string
	uploadBytes   int64
	downloadBytes int64
}

func (s *stubRateLimiter) CheckRateLimit(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
	s.mu.Lock()
	s.checkCalls = append(s.checkCalls, rateCheckCall{ip: rawIP, uploadBytes: uploadBytes})
	s.mu.Unlock()
	if s.checkFn != nil {
		return s.checkFn(ctx, rawIP, uploadBytes)
	}
	return domain.RateLimitDecision{IP: rawIP, Allowed: true}, nil
}

func (s *stubRateLimiter) ReserveResponseBandwidth(context.Context, string, int64) (domain.RateLimitDecision, error) {
	s.mu.Lock()
	s.reserveResponseHits++
	s.mu.Unlock()
	return domain.RateLimitDecision{Allowed: true}, nil
}

func (s *stubRateLimiter) AccountTraffic(_ context.Context, rawIP string, uploadBytes int64, downloadBytes int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accountCalls = append(s.accountCalls, trafficCall{ip: rawIP, uploadBytes: uploadBytes, downloadBytes: downloadBytes})
	return nil
}

func (s *stubRateLimiter) ReleaseConnections(_ context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cloned := append([]string(nil), keys...)
	s.releaseCalls = append(s.releaseCalls, cloned)
	return nil
}

type stubMonitoring struct {
	mu             sync.Mutex
	beginCalls     int
	finishCalls    []usecase.FinishRequestInput
	accessCalls    []usecase.RecordAccessDecisionInput
	rateLimitCalls []usecase.RecordRateLimitInput
	upstreamCalls  []usecase.RecordUpstreamInput
	cacheCalls     []usecase.RecordCacheInput
}

func (s *stubMonitoring) BeginRequest(context.Context, usecase.BeginRequestInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.beginCalls++
}

func (s *stubMonitoring) FinishRequest(_ context.Context, input usecase.FinishRequestInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishCalls = append(s.finishCalls, input)
}

func (s *stubMonitoring) RecordAccessDecision(_ context.Context, input usecase.RecordAccessDecisionInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessCalls = append(s.accessCalls, input)
}

func (s *stubMonitoring) RecordRateLimitDecision(_ context.Context, input usecase.RecordRateLimitInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rateLimitCalls = append(s.rateLimitCalls, input)
}

func (s *stubMonitoring) RecordUpstream(_ context.Context, input usecase.RecordUpstreamInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upstreamCalls = append(s.upstreamCalls, input)
}

func (s *stubMonitoring) RecordCacheEvent(_ context.Context, input usecase.RecordCacheInput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cacheCalls = append(s.cacheCalls, input)
}

func testLogger() *zerolog.Logger {
	logger := zerolog.New(httptest.NewRecorder())
	return &logger
}

func newTestEngine(handlers ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/", handlers...)
	return engine
}

func performRequest(engine *gin.Engine, remoteAddr string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestIPAccessMiddlewareTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		remoteAddr         string
		checkFn            func(ctx context.Context, rawIP string) (domain.AccessDecision, error)
		wantStatus         int
		wantBodyContains   string
		wantAborted        bool
		wantNextCalled     bool
		wantMonitoringCall bool
		wantDecision       string
	}{
		{
			name:       "allow request continues",
			remoteAddr: "127.0.0.1:1234",
			checkFn: func(context.Context, string) (domain.AccessDecision, error) {
				return domain.AccessDecision{IP: "127.0.0.1", Allowed: true, Decision: "allow", Reason: "matched allowlist"}, nil
			},
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantMonitoringCall: true,
			wantDecision:       "allow",
		},
		{
			name:       "deny request aborts forbidden",
			remoteAddr: "10.0.0.1:1234",
			checkFn: func(context.Context, string) (domain.AccessDecision, error) {
				return domain.AccessDecision{IP: "10.0.0.1", Allowed: false, Decision: "deny", Reason: "matched denylist rule"}, nil
			},
			wantStatus:         http.StatusForbidden,
			wantBodyContains:   "access denied",
			wantAborted:        true,
			wantMonitoringCall: true,
			wantDecision:       "deny",
		},
		{
			name:       "graylist captcha required",
			remoteAddr: "192.168.1.15:1234",
			checkFn: func(context.Context, string) (domain.AccessDecision, error) {
				return domain.AccessDecision{IP: "192.168.1.15", Allowed: false, Decision: "captcha_required", Reason: "captcha verification required", VerificationRequired: true}, nil
			},
			wantStatus:         http.StatusForbidden,
			wantBodyContains:   "captcha verification required",
			wantAborted:        true,
			wantMonitoringCall: true,
			wantDecision:       "captcha_required",
		},
		{
			name:       "default policy allow",
			remoteAddr: "203.0.113.10:1234",
			checkFn: func(context.Context, string) (domain.AccessDecision, error) {
				return domain.AccessDecision{IP: "203.0.113.10", Allowed: true, Decision: "allow", Reason: "default policy"}, nil
			},
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantMonitoringCall: true,
			wantDecision:       "allow",
		},
		{
			name:       "invalid client ip handling",
			remoteAddr: "not-a-real-addr",
			checkFn: func(_ context.Context, rawIP string) (domain.AccessDecision, error) {
				if rawIP != "" {
					t.Fatalf("expected empty client ip for malformed remote addr, got %q", rawIP)
				}
				return domain.AccessDecision{}, errors.New("invalid ip")
			},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: "invalid ip",
			wantAborted:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checker := &stubIPAccessChecker{checkFn: tt.checkFn}
			monitoring := &stubMonitoring{}
			var nextCalled bool

			engine := newTestEngine(
				IPAccess(testLogger(), checker, monitoring),
				func(c *gin.Context) {
					nextCalled = true
					if value, exists := c.Get(contextAccessDecisionKey); !exists || value != tt.wantDecision {
						t.Fatalf("unexpected access decision context: exists=%v value=%v", exists, value)
					}
					c.String(http.StatusOK, "ok")
				},
			)

			rec := performRequest(engine, tt.remoteAddr, nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBodyContains != "" && !strings.Contains(rec.Body.String(), tt.wantBodyContains) {
				t.Fatalf("unexpected body: got %q missing %q", rec.Body.String(), tt.wantBodyContains)
			}
			if nextCalled != tt.wantNextCalled {
				t.Fatalf("unexpected continuation: got %v want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantAborted && !strings.Contains(rec.Body.String(), "error") {
				t.Fatalf("expected aborted JSON error response, got %q", rec.Body.String())
			}
			if tt.wantMonitoringCall && len(monitoring.accessCalls) != 1 {
				t.Fatalf("expected 1 access monitoring call, got %d", len(monitoring.accessCalls))
			}
			if !tt.wantMonitoringCall && len(monitoring.accessCalls) != 0 {
				t.Fatalf("expected 0 access monitoring calls, got %d", len(monitoring.accessCalls))
			}
		})
	}
}

func TestIPAccessMiddlewareConcurrentAllowedRequests(t *testing.T) {
	t.Parallel()

	checker := &stubIPAccessChecker{
		checkFn: func(_ context.Context, rawIP string) (domain.AccessDecision, error) {
			return domain.AccessDecision{IP: rawIP, Allowed: true, Decision: "allow", Reason: "ok"}, nil
		},
	}
	monitoring := &stubMonitoring{}
	engine := newTestEngine(
		IPAccess(testLogger(), checker, monitoring),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := performRequest(engine, "127.0.0.1:1234", nil)
			if rec.Code != http.StatusNoContent {
				t.Errorf("unexpected status: %d", rec.Code)
			}
		}()
	}
	wg.Wait()

	if len(monitoring.accessCalls) != 24 {
		t.Fatalf("expected 24 monitoring access calls, got %d", len(monitoring.accessCalls))
	}
}

func TestRateLimitMiddlewareTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		remoteAddr           string
		contentLength        int64
		checkFn              func(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error)
		wantStatus           int
		wantBodyContains     string
		wantNextCalled       bool
		wantRateCalls        int
		wantAccountCalls     int
		wantReleaseCalls     int
		wantRateDecisionCtx  string
		expectRetryAfterMiss bool
	}{
		{
			name:          "request allowed under limit",
			remoteAddr:    "127.0.0.1:4321",
			contentLength: 128,
			checkFn: func(_ context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
				if uploadBytes != 128 {
					t.Fatalf("expected upload bytes 128, got %d", uploadBytes)
				}
				return domain.RateLimitDecision{IP: rawIP, Allowed: true, ConnectionKeys: []string{"ip:127.0.0.1"}}, nil
			},
			wantStatus:          http.StatusCreated,
			wantNextCalled:      true,
			wantRateCalls:       1,
			wantAccountCalls:    1,
			wantReleaseCalls:    1,
			wantRateDecisionCtx: "allow",
		},
		{
			name:       "request blocked over limit",
			remoteAddr: "127.0.0.2:4321",
			checkFn: func(_ context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
				return domain.RateLimitDecision{IP: rawIP, Allowed: false, LimitType: "rps", RuleID: "global", CurrentValue: 11}, nil
			},
			wantStatus:           http.StatusTooManyRequests,
			wantBodyContains:     "rate limit exceeded",
			wantRateCalls:        1,
			wantRateDecisionCtx:  "deny",
			expectRetryAfterMiss: true,
		},
		{
			name:       "middleware returns bad request on limiter error",
			remoteAddr: "127.0.0.3:4321",
			checkFn: func(context.Context, string, int64) (domain.RateLimitDecision, error) {
				return domain.RateLimitDecision{}, errors.New("limiter failed")
			},
			wantStatus:           http.StatusBadRequest,
			wantBodyContains:     "limiter failed",
			expectRetryAfterMiss: true,
		},
		{
			name:       "subnet limit surfaced as deny",
			remoteAddr: "192.168.1.10:4321",
			checkFn: func(_ context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
				return domain.RateLimitDecision{IP: rawIP, Allowed: false, LimitType: "subnet_rps", RuleID: "office-subnet", CurrentValue: 26}, nil
			},
			wantStatus:           http.StatusTooManyRequests,
			wantBodyContains:     "subnet_rps",
			wantRateCalls:        1,
			wantRateDecisionCtx:  "deny",
			expectRetryAfterMiss: true,
		},
		{
			name:          "missing headers and negative content length normalize to zero",
			remoteAddr:    "127.0.0.4:4321",
			contentLength: -1,
			checkFn: func(_ context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error) {
				if uploadBytes != 0 {
					t.Fatalf("expected normalized upload bytes 0, got %d", uploadBytes)
				}
				return domain.RateLimitDecision{IP: rawIP, Allowed: true}, nil
			},
			wantStatus:          http.StatusCreated,
			wantNextCalled:      true,
			wantRateCalls:       1,
			wantAccountCalls:    1,
			wantReleaseCalls:    1,
			wantRateDecisionCtx: "allow",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limiter := &stubRateLimiter{checkFn: tt.checkFn}
			monitoring := &stubMonitoring{}
			var nextCalled bool

			engine := newTestEngine(
				RateLimit(testLogger(), limiter, monitoring),
				func(c *gin.Context) {
					nextCalled = true
					if value, exists := c.Get(contextRateLimitDecisionKey); !exists || value != tt.wantRateDecisionCtx {
						t.Fatalf("unexpected rate decision context: exists=%v value=%v", exists, value)
					}
					c.String(http.StatusCreated, "created")
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			req.ContentLength = tt.contentLength
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBodyContains != "" && !strings.Contains(rec.Body.String(), tt.wantBodyContains) {
				t.Fatalf("unexpected body: got %q missing %q", rec.Body.String(), tt.wantBodyContains)
			}
			if nextCalled != tt.wantNextCalled {
				t.Fatalf("unexpected continuation: got %v want %v", nextCalled, tt.wantNextCalled)
			}
			if len(monitoring.rateLimitCalls) != tt.wantRateCalls {
				t.Fatalf("unexpected monitoring rate call count: got %d want %d", len(monitoring.rateLimitCalls), tt.wantRateCalls)
			}
			if len(limiter.accountCalls) != tt.wantAccountCalls {
				t.Fatalf("unexpected account call count: got %d want %d", len(limiter.accountCalls), tt.wantAccountCalls)
			}
			if len(limiter.releaseCalls) != tt.wantReleaseCalls {
				t.Fatalf("unexpected release call count: got %d want %d", len(limiter.releaseCalls), tt.wantReleaseCalls)
			}
			if tt.expectRetryAfterMiss && rec.Header().Get("Retry-After") != "" {
				t.Fatalf("expected no Retry-After header, got %q", rec.Header().Get("Retry-After"))
			}
		})
	}
}

func TestRateLimitMiddlewareDifferentClientsIsolation(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{
		checkFn: func(_ context.Context, rawIP string, _ int64) (domain.RateLimitDecision, error) {
			if rawIP == "127.0.0.2" {
				return domain.RateLimitDecision{IP: rawIP, Allowed: false, LimitType: "rps"}, nil
			}
			return domain.RateLimitDecision{IP: rawIP, Allowed: true}, nil
		},
	}
	monitoring := &stubMonitoring{}
	engine := newTestEngine(
		RateLimit(testLogger(), limiter, monitoring),
		func(c *gin.Context) { c.Status(http.StatusAccepted) },
	)

	recA := performRequest(engine, "127.0.0.1:1001", nil)
	recB := performRequest(engine, "127.0.0.2:1002", nil)

	if recA.Code != http.StatusAccepted {
		t.Fatalf("client A unexpected status: %d", recA.Code)
	}
	if recB.Code != http.StatusTooManyRequests {
		t.Fatalf("client B unexpected status: %d", recB.Code)
	}
	if len(limiter.checkCalls) != 2 {
		t.Fatalf("expected two limiter calls, got %d", len(limiter.checkCalls))
	}
	if limiter.checkCalls[0].ip == limiter.checkCalls[1].ip {
		t.Fatalf("expected different client IPs, got %+v", limiter.checkCalls)
	}
}

func TestRateLimitMiddlewareConcurrentRequests(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{
		checkFn: func(_ context.Context, rawIP string, _ int64) (domain.RateLimitDecision, error) {
			return domain.RateLimitDecision{IP: rawIP, Allowed: true, ConnectionKeys: []string{"ip:" + rawIP}}, nil
		},
	}
	monitoring := &stubMonitoring{}
	engine := newTestEngine(
		RateLimit(testLogger(), limiter, monitoring),
		func(c *gin.Context) { c.String(http.StatusOK, "ok") },
	)

	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := performRequest(engine, "127.0.0.1:2000", nil)
			if rec.Code != http.StatusOK {
				t.Errorf("unexpected status: %d", rec.Code)
			}
		}()
	}
	wg.Wait()

	if len(limiter.checkCalls) != 24 {
		t.Fatalf("expected 24 limiter calls, got %d", len(limiter.checkCalls))
	}
	if len(limiter.releaseCalls) != 24 {
		t.Fatalf("expected 24 release calls, got %d", len(limiter.releaseCalls))
	}
}

func TestRequestLoggerMiddlewareRecordsAndPreservesResponse(t *testing.T) {
	t.Parallel()

	monitoring := &stubMonitoring{}
	engine := newTestEngine(
		RequestLogger(testLogger(), monitoring),
		func(c *gin.Context) {
			c.Set(contextAccessDecisionKey, "allow")
			c.Set(contextRateLimitDecisionKey, "allow")
			c.Set(contextCacheStatusKey, "hit")
			c.Set(contextUpstreamLatencyKey, 25*time.Millisecond)
			c.String(http.StatusCreated, "payload")
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	req.ContentLength = 123
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "payload" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
	if monitoring.beginCalls != 1 {
		t.Fatalf("expected begin request hook once, got %d", monitoring.beginCalls)
	}
	if len(monitoring.finishCalls) != 1 {
		t.Fatalf("expected finish request hook once, got %d", len(monitoring.finishCalls))
	}
	record := monitoring.finishCalls[0]
	if record.Status != http.StatusCreated || record.AccessDecision != "allow" || record.RateLimitDecision != "allow" || record.CacheStatus != "hit" {
		t.Fatalf("unexpected finish request record: %+v", record)
	}
	if record.BytesIn != 123 || record.BytesOut != int64(len("payload")) {
		t.Fatalf("unexpected byte accounting: %+v", record)
	}
}

func TestRequestLoggerMiddlewareDefaultsWithoutContextValues(t *testing.T) {
	t.Parallel()

	monitoring := &stubMonitoring{}
	engine := newTestEngine(
		RequestLogger(testLogger(), monitoring),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	rec := performRequest(engine, "127.0.0.1:9001", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if len(monitoring.finishCalls) != 1 {
		t.Fatalf("expected one finish call, got %d", len(monitoring.finishCalls))
	}
	record := monitoring.finishCalls[0]
	if record.AccessDecision != "not_applicable" || record.RateLimitDecision != "not_applicable" || record.CacheStatus != "not_applicable" {
		t.Fatalf("unexpected default context values: %+v", record)
	}
}

func TestRecoveryMiddlewareRecoversPanics(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Recovery(testLogger()))
	engine.GET("/", func(c *gin.Context) {
		panic("boom")
	})

	rec := performRequest(engine, "127.0.0.1:9999", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestRecoveryMiddlewareIgnoresAbortHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Recovery(testLogger()))
	engine.GET("/", func(c *gin.Context) {
		panic(http.ErrAbortHandler)
	})

	rec := performRequest(engine, "127.0.0.1:9998", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status for abort handler panic: got %d want %d", rec.Code, http.StatusOK)
	}
}
