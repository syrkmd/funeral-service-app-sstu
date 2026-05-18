package gintransport

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

type ControllerMonitoringUseCase interface {
	usecase.MonitoringRecorder
	GetMetrics(ctx context.Context) (domain.MetricsResponse, error)
	GetDashboardOverview(ctx context.Context) (domain.DashboardOverview, error)
	GetDashboardClients(ctx context.Context) (domain.DashboardClients, error)
	GetDashboardUpstream(ctx context.Context) (domain.DashboardUpstream, error)
	GetDashboardRateLimits(ctx context.Context) (domain.DashboardRateLimits, error)
	GetDashboardIPAccess(ctx context.Context) (domain.DashboardIPAccess, error)
}

type MonitoringController struct {
	useCase ControllerMonitoringUseCase
}

func NewMonitoringController(useCase ControllerMonitoringUseCase) *MonitoringController {
	return &MonitoringController{useCase: useCase}
}

// Metrics godoc
// @Summary Get service metrics
// @Tags monitoring
// @Produce json
// @Success 200 {object} domain.MetricsResponse
// @Failure 500 {object} errorResponse
// @Router /api/metrics [get]
func (h *MonitoringController) Metrics(c *gin.Context) {
	response, err := h.useCase.GetMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Overview godoc
// @Summary Get dashboard overview
// @Tags dashboard
// @Produce json
// @Success 200 {object} domain.DashboardOverview
// @Failure 500 {object} errorResponse
// @Router /api/dashboard/overview [get]
func (h *MonitoringController) Overview(c *gin.Context) {
	response, err := h.useCase.GetDashboardOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Clients godoc
// @Summary Get dashboard client stats
// @Tags dashboard
// @Produce json
// @Success 200 {object} domain.DashboardClients
// @Failure 500 {object} errorResponse
// @Router /api/dashboard/clients [get]
func (h *MonitoringController) Clients(c *gin.Context) {
	response, err := h.useCase.GetDashboardClients(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// Upstream godoc
// @Summary Get upstream status
// @Tags dashboard
// @Produce json
// @Success 200 {object} domain.DashboardUpstream
// @Failure 500 {object} errorResponse
// @Router /api/dashboard/upstream [get]
func (h *MonitoringController) Upstream(c *gin.Context) {
	response, err := h.useCase.GetDashboardUpstream(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// RateLimits godoc
// @Summary Get rate limit dashboard data
// @Tags dashboard
// @Produce json
// @Success 200 {object} domain.DashboardRateLimits
// @Failure 500 {object} errorResponse
// @Router /api/dashboard/rate_limits [get]
func (h *MonitoringController) RateLimits(c *gin.Context) {
	response, err := h.useCase.GetDashboardRateLimits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// IPAccess godoc
// @Summary Get IP access dashboard data
// @Tags dashboard
// @Produce json
// @Success 200 {object} domain.DashboardIPAccess
// @Failure 500 {object} errorResponse
// @Router /api/dashboard/ip_access [get]
func (h *MonitoringController) IPAccess(c *gin.Context) {
	response, err := h.useCase.GetDashboardIPAccess(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
