package gintransport

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"proxy/internal/domain"
	"proxy/internal/usecase"
)

type ControllerIPAccessUseCase interface {
	ListRules(ctx context.Context) ([]domain.IPRule, error)
	AddRule(ctx context.Context, input usecase.AddRuleInput) (domain.IPRule, error)
	DeleteRule(ctx context.Context, id string) error
	CheckIP(ctx context.Context, rawIP string) (domain.AccessDecision, error)
}

type CreateRuleRequest struct {
	ID          string `json:"id"`
	Type        string `json:"type" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type IPAccessController struct {
	useCase ControllerIPAccessUseCase
}

func NewIPAccessController(useCase ControllerIPAccessUseCase) *IPAccessController {
	return &IPAccessController{useCase: useCase}
}

// ListRules godoc
// @Summary List IP rules
// @Tags ip_access
// @Produce json
// @Success 200 {array} domain.IPRule
// @Failure 500 {object} errorResponse
// @Router /api/ip_access/lists [get]
func (h *IPAccessController) ListRules(c *gin.Context) {
	rules, err := h.useCase.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// CreateRule godoc
// @Summary Create IP rule
// @Tags ip_access
// @Accept json
// @Produce json
// @Param request body CreateRuleRequest true "Rule payload"
// @Success 201 {object} domain.IPRule
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Router /api/ip_access/lists [post]
func (h *IPAccessController) CreateRule(c *gin.Context) {
	var request CreateRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	rule, err := h.useCase.AddRule(c.Request.Context(), usecase.AddRuleInput{
		ID:          request.ID,
		Type:        request.Type,
		Value:       request.Value,
		Description: request.Description,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, domain.ErrDuplicateRuleID) {
			status = http.StatusConflict
		}
		c.JSON(status, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// DeleteRule godoc
// @Summary Delete IP rule
// @Tags ip_access
// @Produce json
// @Param id path string true "Rule ID"
// @Success 204
// @Failure 404 {object} errorResponse
// @Router /api/ip_access/lists/{id} [delete]
func (h *IPAccessController) DeleteRule(c *gin.Context) {
	if err := h.useCase.DeleteRule(c.Request.Context(), c.Param("id")); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrRuleNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, errorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CheckIP godoc
// @Summary Check IP access decision
// @Tags ip_access
// @Produce json
// @Param ip query string true "IP address"
// @Success 200 {object} domain.AccessDecision
// @Failure 400 {object} errorResponse
// @Router /api/ip_access/check [get]
func (h *IPAccessController) CheckIP(c *gin.Context) {
	decision, err := h.useCase.CheckIP(c.Request.Context(), c.Query("ip"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, decision)
}
