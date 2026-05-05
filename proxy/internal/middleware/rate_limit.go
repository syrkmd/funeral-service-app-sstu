package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"proxy/internal/usecase"
)

func RateLimit(logger *zerolog.Logger, limiter usecase.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := strings.TrimSpace(c.ClientIP())
		decision, err := limiter.CheckRateLimit(c.Request.Context(), clientIP)
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

		c.Next()
	}
}
