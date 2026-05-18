package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

const contextReservedDownloadBytesKey = "rate_limit_reserved_download_bytes"

func RateLimit(logger *zerolog.Logger, limiter usecase.RateLimiter, monitoring usecase.MonitoringRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := strings.TrimSpace(c.ClientIP())
		estimatedUploadBytes := normalizeContentLength(c.Request.ContentLength)
		decision, err := limiter.CheckRateLimit(c.Request.Context(), clientIP, estimatedUploadBytes)
		if err != nil {
			logger.Error().
				Err(err).
				Str("ip", clientIP).
				Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
				Str("reason", "rate_limit").
				Msg("failed to evaluate rate limit")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		rateDecision := "allow"
		if !decision.Allowed {
			rateDecision = "deny"
		}
		c.Set(contextRateLimitDecisionKey, rateDecision)
		monitoring.RecordRateLimitDecision(c.Request.Context(), usecase.RecordRateLimitInput{
			IP:           decision.IP,
			RuleID:       decision.RuleID,
			LimitType:    decision.LimitType,
			Allowed:      decision.Allowed,
			CurrentValue: decision.CurrentValue,
		})

		if !decision.Allowed {
			logger.Info().
				Str("ip", decision.IP).
				Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
				Str("reason", "rate_limit").
				Str("type", decision.LimitType).
				Str("rule_id", decision.RuleID).
				Int("current_value", decision.CurrentValue).
				Msg("rate limit exceeded")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"type":  decision.LimitType,
			})
			return
		}

		defer func() {
			downloadDelta := int64(c.Writer.Size()) - contextInt64(c, contextReservedDownloadBytesKey)
			if downloadDelta < 0 {
				downloadDelta = 0
			}
			if err := limiter.AccountTraffic(c.Request.Context(), clientIP, 0, downloadDelta); err != nil {
				logger.Error().
					Err(err).
					Str("ip", clientIP).
					Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
					Str("reason", "rate_limit_bandwidth_accounting").
					Msg("failed to account rate limit traffic")
			}

			if err := limiter.ReleaseConnections(c.Request.Context(), decision.ConnectionKeys); err != nil {
				logger.Error().
					Err(err).
					Str("ip", clientIP).
					Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
					Str("reason", "rate_limit_release").
					Msg("failed to release connection rate limit state")
			}
		}()

		c.Next()
	}
}

func contextInt64(c *gin.Context, key string) int64 {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}

	number, ok := value.(int64)
	if !ok {
		return 0
	}

	return number
}

func normalizeContentLength(contentLength int64) int64 {
	if contentLength < 0 {
		return 0
	}
	return contentLength
}
