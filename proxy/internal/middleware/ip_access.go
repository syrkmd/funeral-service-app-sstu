package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"proxy/internal/usecase"
)

func IPAccess(logger *zerolog.Logger, checker usecase.IPAccessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := strings.TrimSpace(c.ClientIP())
		decision, err := checker.CheckIP(c.Request.Context(), clientIP)
		if err != nil {
			logger.Error().
				Err(err).
				Str("ip", clientIP).
				Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
				Str("url", c.Request.URL.String()).
				Msg("failed to evaluate IP access")

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		event := logger.Info().
			Str("ip", decision.IP).
			Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
			Str("url", c.Request.URL.String()).
			Str("decision", decision.Decision).
			Str("reason", decision.Reason)

		if decision.MatchedRuleID != "" {
			event = event.Str("matched_rule_id", decision.MatchedRuleID)
		}

		event.Msg("IP access evaluated")

		if !decision.Allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":    "access denied",
				"decision": decision,
			})
			return
		}

		c.Next()
	}
}
