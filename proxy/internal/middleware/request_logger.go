package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

const (
	contextAccessDecisionKey    = "access_decision"
	contextRateLimitDecisionKey = "rate_limit_decision"
	contextCacheStatusKey       = "cache_status"
	contextUpstreamLatencyKey   = "upstream_latency"
)

func RequestLogger(logger *zerolog.Logger, monitoring usecase.MonitoringRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		monitoring.BeginRequest(c.Request.Context(), usecase.BeginRequestInput{})
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		accessDecision := contextString(c, contextAccessDecisionKey, "not_applicable")
		rateLimitDecision := contextString(c, contextRateLimitDecisionKey, "not_applicable")
		cacheStatus := contextString(c, contextCacheStatusKey, "not_applicable")
		upstreamLatency := contextDuration(c, contextUpstreamLatencyKey)
		bytesIn := requestBytesIn(c.Request)
		bytesOut := int64(c.Writer.Size())
		latency := time.Since(start)

		monitoring.FinishRequest(c.Request.Context(), usecase.FinishRequestInput{
			Timestamp:         time.Now().UTC(),
			Method:            c.Request.Method,
			Path:              path,
			ClientIP:          c.ClientIP(),
			Status:            c.Writer.Status(),
			Latency:           latency,
			BytesIn:           bytesIn,
			BytesOut:          bytesOut,
			AccessDecision:    accessDecision,
			RateLimitDecision: rateLimitDecision,
			CacheStatus:       cacheStatus,
			UpstreamLatency:   upstreamLatency,
		})

		logger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Int64("bytes_in", bytesIn).
			Int64("bytes_out", bytesOut).
			Dur("upstream_latency", upstreamLatency).
			Str("access_decision", accessDecision).
			Str("rate_limit_decision", rateLimitDecision).
			Str("cache_status", cacheStatus).
			Msg("request completed")
	}
}

func contextString(c *gin.Context, key string, fallback string) string {
	value, exists := c.Get(key)
	if !exists {
		return fallback
	}

	text, ok := value.(string)
	if !ok || text == "" {
		return fallback
	}

	return text
}

func contextDuration(c *gin.Context, key string) time.Duration {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}

	duration, ok := value.(time.Duration)
	if !ok {
		return 0
	}

	return duration
}

func requestBytesIn(request *http.Request) int64 {
	if request.ContentLength < 0 {
		return 0
	}
	return request.ContentLength
}
