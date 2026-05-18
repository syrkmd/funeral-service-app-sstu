package gintransport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/domain"
)

type ControllerRateLimitUseCase interface {
	CheckRateLimit(ctx context.Context, rawIP string, uploadBytes int64) (domain.RateLimitDecision, error)
	ReserveResponseBandwidth(ctx context.Context, rawIP string, downloadBytes int64) (domain.RateLimitDecision, error)
	AccountTraffic(ctx context.Context, rawIP string, uploadBytes int64, downloadBytes int64) error
	ReleaseConnections(ctx context.Context, keys []string) error
	ListRules(ctx context.Context) ([]domain.RateLimitRule, error)
	ListBuckets(ctx context.Context) ([]domain.RateLimitBucketSnapshot, error)
}

type Dependencies struct {
	Logger            *zerolog.Logger
	MetricsHandler    http.Handler
	MonitoringUseCase ControllerMonitoringUseCase
	IPAccessUseCase   ControllerIPAccessUseCase
	RateLimitUseCase  ControllerRateLimitUseCase
	CacheUseCase      ControllerCacheUseCase
	ProxyUseCase      ControllerProxyUseCase
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(Recovery(deps.Logger))
	router.Use(RequestLogger(deps.Logger, deps.MonitoringUseCase))

	router.GET("/metrics", gin.WrapH(deps.MetricsHandler))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	ipController := NewIPAccessController(deps.IPAccessUseCase)
	captchaController := NewCaptchaController(deps.Logger, deps.IPAccessUseCase)
	cacheController := NewCacheController(deps.CacheUseCase, deps.MonitoringUseCase)
	monitoringController := NewMonitoringController(deps.MonitoringUseCase)
	api := router.Group("/api")
	{
		ipGroup := api.Group("/ip_access")
		ipGroup.GET("/lists", ipController.ListRules)
		ipGroup.POST("/lists", ipController.CreateRule)
		ipGroup.DELETE("/lists/:id", ipController.DeleteRule)
		ipGroup.GET("/check", ipController.CheckIP)
		api.POST("/captcha/verify", captchaController.Verify)
		cacheGroup := api.Group("/cache")
		cacheGroup.DELETE("/key/:key", cacheController.DeleteByKey)
		cacheGroup.DELETE("/prefix", cacheController.DeleteByPrefix)
		cacheGroup.DELETE("/regex", cacheController.DeleteByRegex)
		cacheGroup.DELETE("/tags/:tag", cacheController.DeleteByTag)
		cacheGroup.DELETE("/expired", cacheController.DeleteExpired)
		cacheGroup.DELETE("/all", cacheController.Clear)

		api.GET("/metrics", monitoringController.Metrics)

		dashboard := api.Group("/dashboard")
		dashboard.GET("/overview", monitoringController.Overview)
		dashboard.GET("/clients", monitoringController.Clients)
		dashboard.GET("/upstream", monitoringController.Upstream)
		dashboard.GET("/rate_limits", monitoringController.RateLimits)
		dashboard.GET("/ip_access", monitoringController.IPAccess)
	}

	proxyController := NewProxyController(deps.ProxyUseCase, deps.MonitoringUseCase, deps.RateLimitUseCase, deps.CacheUseCase)

	router.NoRoute(
		IPAccess(deps.Logger, deps.IPAccessUseCase, deps.MonitoringUseCase),
		RateLimit(deps.Logger, deps.RateLimitUseCase, deps.MonitoringUseCase),
		proxyController.Proxy(),
	)

	return router
}

func newReverseProxyErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(writer http.ResponseWriter, request *http.Request, err error) {
		writer.WriteHeader(502)
		_, _ = writer.Write([]byte(`{"error":"upstream request failed"}`))
	}
}

func buildFlushInterval(raw int64) time.Duration {
	return time.Duration(raw)
}
