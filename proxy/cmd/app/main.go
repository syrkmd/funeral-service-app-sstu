package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	_ "github.com/syrkmd/funeral-service-app-sstu/proxy/docs"

	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/config"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/logger"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/repository"
	_jsii "github.com/syrkmd/funeral-service-app-sstu/proxy/internal/transport/http/gin"
	"github.com/syrkmd/funeral-service-app-sstu/proxy/internal/usecase"
)

// @title Proxy Service API
// @version 1.0
// @description HTTP proxy service.
// @BasePath /
func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfgManager, err := config.NewManager(cfgPath)
	if err != nil {
		panic(err)
	}

	appLogger, err := logger.New(cfgManager.Current().Logging.Level)
	if err != nil {
		panic(err)
	}
	log.Logger = appLogger

	ipRepo, err := repository.NewIPAccessRepository(cfgManager.Current())
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to initialize IP access repository")
	}
	verifiedIPRepo := repository.NewVerifiedIPRepository(cfgManager.Current().Access.VerificationTTL.Duration)
	cacheRepo := repository.NewCacheRepository(cfgManager.Current().Cache.CleanupInterval.Duration)
	monitoringRepo := repository.NewMonitoringRepository()
	ipAccessUseCase := usecase.NewIPAccessUseCase(ipRepo, verifiedIPRepo, 4096)
	ipAccessUseCase.SetDecisionCacheTTL(cfgManager.Current().Access.DecisionCacheTTL.Duration)

	cfgManager.Subscribe(func(cfg config.Config) {
		if err := ipRepo.ApplyConfig(cfg); err != nil {
			appLogger.Error().Err(err).Msg("failed to apply reloaded config")
			return
		}
		verifiedIPRepo.SetTTL(cfg.Access.VerificationTTL.Duration)
		ipAccessUseCase.SetDecisionCacheTTL(cfg.Access.DecisionCacheTTL.Duration)
		cacheRepo.SetCleanupInterval(cfg.Cache.CleanupInterval.Duration)

		if updatedLogger, err := logger.New(cfg.Logging.Level); err == nil {
			log.Logger = updatedLogger
		} else {
			appLogger.Error().Err(err).Msg("failed to reconfigure logger level")
		}
	})

	rateRepo := repository.NewRateLimitRepository()
	rateUseCase := usecase.NewRateLimitUseCase(rateRepo, cfgManager, logger.NewUseCaseAdapter(&log.Logger))
	cacheUseCase := usecase.NewCacheUseCase(cacheRepo, cfgManager)
	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	prometheusRepo := repository.NewPrometheusRepository(promRegistry)
	monitoringUseCase := usecase.NewMonitoringUseCase(monitoringRepo, prometheusRepo, ipAccessUseCase, rateUseCase)
	proxyUseCase := usecase.NewProxyUseCase(cfgManager)

	router := _jsii.NewRouter(_jsii.Dependencies{
		Logger:            &log.Logger,
		MetricsHandler:    promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}),
		MonitoringUseCase: monitoringUseCase,
		IPAccessUseCase:   ipAccessUseCase,
		RateLimitUseCase:  rateUseCase,
		CacheUseCase:      cacheUseCase,
		ProxyUseCase:      proxyUseCase,
	})

	server := &http.Server{
		Addr:         cfgManager.Current().Server.Address,
		Handler:      router,
		ReadTimeout:  cfgManager.Current().Server.ReadTimeout.Duration,
		WriteTimeout: cfgManager.Current().Server.WriteTimeout.Duration,
		IdleTimeout:  cfgManager.Current().Server.IdleTimeout.Duration,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := cfgManager.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error().Err(err).Msg("config watcher stopped with error")
		}
	}()

	go rateRepo.StartCleanup(ctx, 10*time.Minute, 48*time.Hour)
	go verifiedIPRepo.StartCleanup(ctx, time.Minute)
	go cacheRepo.StartCleanup(ctx)

	go func() {
		log.Info().Str("address", server.Addr).Msg("starting server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfgManager.Current().Server.ShutdownTimeout.Duration)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
