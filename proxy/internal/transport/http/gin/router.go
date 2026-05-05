package gintransport

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"proxy/internal/domain"
)

type Dependencies struct {
	Logger           *zerolog.Logger
	IPAccessUseCase  ControllerIPAccessUseCase
	RateLimitUseCase ControllerRateLimitUseCase
	ProxyUseCase     ControllerProxyUseCase
}

type ControllerRateLimitUseCase interface {
	CheckRateLimit(ctx context.Context, rawIP string) (domain.RateLimitDecision, error)
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(Recovery(deps.Logger))
	router.Use(RequestLogger(deps.Logger))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	ipController := NewIPAccessController(deps.IPAccessUseCase)
	api := router.Group("/api/ip_access")
	api.Use(
		IPAccess(deps.Logger, deps.IPAccessUseCase),
		RateLimit(deps.Logger, deps.RateLimitUseCase),
	)
	{
		api.GET("/lists", ipController.ListRules)
		api.POST("/lists", ipController.CreateRule)
		api.DELETE("/lists/:id", ipController.DeleteRule)
		api.GET("/check", ipController.CheckIP)
	}

	proxyController := NewProxyController(deps.ProxyUseCase)

	router.NoRoute(
		IPAccess(deps.Logger, deps.IPAccessUseCase),
		RateLimit(deps.Logger, deps.RateLimitUseCase),
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

func withErrorHandler(proxy *httputil.ReverseProxy) *httputil.ReverseProxy {
	proxy.ErrorHandler = newReverseProxyErrorHandler()
	return proxy
}
