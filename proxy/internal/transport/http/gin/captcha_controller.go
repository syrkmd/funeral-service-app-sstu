package gintransport

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type ControllerCaptchaUseCase interface {
	VerifyCaptcha(ctx context.Context, rawIP string, answer string) error
}

type verifyCaptchaRequest struct {
	Answer string `json:"answer" binding:"required"`
}

type CaptchaController struct {
	logger  *zerolog.Logger
	useCase ControllerCaptchaUseCase
}

func NewCaptchaController(logger *zerolog.Logger, useCase ControllerCaptchaUseCase) *CaptchaController {
	return &CaptchaController{
		logger:  logger,
		useCase: useCase,
	}
}

// Verify godoc
// @Summary Verify graylist captcha
// @Tags captcha
// @Accept json
// @Produce json
// @Param request body verifyCaptchaRequest true "Captcha payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Router /api/captcha/verify [post]
func (h *CaptchaController) Verify(c *gin.Context) {
	var request verifyCaptchaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	if err := h.useCase.VerifyCaptcha(c.Request.Context(), clientIP, request.Answer); err != nil {
		if errors.Is(err, domain.ErrInvalidCaptcha) {
			h.logger.Info().
				Str("ip", clientIP).
				Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
				Str("event", "captcha_verification").
				Str("result", "failed").
				Msg("graylist captcha verification failed")

			c.JSON(http.StatusForbidden, errorResponse{Error: "captcha verification required"})
			return
		}

		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().
		Str("ip", clientIP).
		Str("timestamp", time.Now().UTC().Format(time.RFC3339Nano)).
		Str("event", "captcha_verification").
		Str("result", "verified").
		Msg("graylist captcha verification succeeded")

	c.JSON(http.StatusOK, gin.H{
		"status": "verified",
	})
}
