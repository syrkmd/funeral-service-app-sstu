package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

func IPAccess(logger *zerolog.Logger, checker usecase.IPAccessChecker, monitoring usecase.MonitoringRecorder) gin.HandlerFunc {
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

		c.Set(contextAccessDecisionKey, decision.Decision)
		monitoring.RecordAccessDecision(c.Request.Context(), usecase.RecordAccessDecisionInput{
			IP:       decision.IP,
			Decision: decision.Decision,
			RuleID:   decision.MatchedRuleID,
		})

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
			if decision.VerificationRequired {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "captcha verification required",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":    "access denied",
				"decision": decision,
			})
			return
		}

		c.Next()
	}
}
