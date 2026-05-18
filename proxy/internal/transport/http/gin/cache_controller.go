package gintransport

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

type ControllerCacheManagementUseCase interface {
	InvalidateByKey(ctx context.Context, key string) (int, error)
	InvalidateByPrefix(ctx context.Context, prefix string) (int, error)
	InvalidateByRegex(ctx context.Context, pattern string) (int, error)
	InvalidateByTags(ctx context.Context, tags []string) (int, error)
	InvalidateExpired(ctx context.Context) (int, error)
	Clear(ctx context.Context) (int, error)
	HandleInvalidationEvent(ctx context.Context, event domain.CacheInvalidationEvent) (int, error)
}

type deleteCountResponse struct {
	DeletedCount int `json:"deleted_count"`
}

type CacheController struct {
	useCase    ControllerCacheManagementUseCase
	monitoring usecase.MonitoringRecorder
}

func NewCacheController(useCase ControllerCacheManagementUseCase, monitoring usecase.MonitoringRecorder) *CacheController {
	return &CacheController{useCase: useCase, monitoring: monitoring}
}

// DeleteByKey godoc
// @Summary Invalidate cache by exact key
// @Tags cache
// @Produce json
// @Param key path string true "Cache key"
// @Success 200 {object} deleteCountResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/key/{key} [delete]
func (h *CacheController) DeleteByKey(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "key is required"})
		return
	}

	deleted, err := h.useCase.InvalidateByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "invalidate_key", key, deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

// DeleteByPrefix godoc
// @Summary Invalidate cache by key prefix
// @Tags cache
// @Produce json
// @Param value query string true "Cache key prefix"
// @Success 200 {object} deleteCountResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/prefix [delete]
func (h *CacheController) DeleteByPrefix(c *gin.Context) {
	prefix := strings.TrimSpace(c.Query("value"))
	if prefix == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "prefix is required"})
		return
	}

	deleted, err := h.useCase.InvalidateByPrefix(c.Request.Context(), prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "invalidate_prefix", prefix, deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

// DeleteByRegex godoc
// @Summary Invalidate cache by regex
// @Tags cache
// @Produce json
// @Param pattern query string true "Go regexp pattern"
// @Success 200 {object} deleteCountResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/regex [delete]
func (h *CacheController) DeleteByRegex(c *gin.Context) {
	pattern := strings.TrimSpace(c.Query("pattern"))
	if pattern == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "pattern is required"})
		return
	}

	deleted, err := h.useCase.InvalidateByRegex(c.Request.Context(), pattern)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "invalidate_regex", pattern, deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

// DeleteByTag godoc
// @Summary Invalidate cache by tag
// @Tags cache
// @Produce json
// @Param tag path string true "Cache tag"
// @Success 200 {object} deleteCountResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/tags/{tag} [delete]
func (h *CacheController) DeleteByTag(c *gin.Context) {
	tag := strings.TrimSpace(c.Param("tag"))
	if tag == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "tag is required"})
		return
	}

	deleted, err := h.useCase.InvalidateByTags(c.Request.Context(), []string{tag})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "invalidate_tags", tag, deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

// DeleteExpired godoc
// @Summary Invalidate expired cache entries
// @Tags cache
// @Produce json
// @Success 200 {object} deleteCountResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/expired [delete]
func (h *CacheController) DeleteExpired(c *gin.Context) {
	deleted, err := h.useCase.InvalidateExpired(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "invalidate_expired", "", deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

// Clear godoc
// @Summary Clear all cache entries
// @Tags cache
// @Produce json
// @Success 200 {object} deleteCountResponse
// @Failure 500 {object} errorResponse
// @Router /api/cache/all [delete]
func (h *CacheController) Clear(c *gin.Context) {
	deleted, err := h.useCase.Clear(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	h.recordInvalidation(c, "clear_all", "", deleted)
	c.JSON(http.StatusOK, deleteCountResponse{DeletedCount: deleted})
}

func (h *CacheController) recordInvalidation(c *gin.Context, action string, key string, _ int) {
	h.monitoring.RecordCacheEvent(c.Request.Context(), usecase.RecordCacheInput{
		Key:    key,
		Domain: c.Request.Host,
		Path:   c.FullPath(),
		Action: action,
		Rule:   "",
	})
}
