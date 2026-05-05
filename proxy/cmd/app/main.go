package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	_ "proxy/docs"

	"proxy/internal/config"
	"proxy/internal/logger"
	"proxy/internal/repository"
	_jsii "proxy/internal/transport/http/gin"
	"proxy/internal/usecase"
)

// @title Proxy Service API
// @version 1.0
// @description Production-ready HTTP proxy service with IP access control.
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

	cfgManager.Subscribe(func(cfg config.Config) {
		if err := ipRepo.ApplyConfig(cfg); err != nil {
			appLogger.Error().Err(err).Msg("failed to apply reloaded config")
			return
		}

		if updatedLogger, err := logger.New(cfg.Logging.Level); err == nil {
			log.Logger = updatedLogger
		} else {
			appLogger.Error().Err(err).Msg("failed to reconfigure logger level")
		}
	})

	ipAccessUseCase := usecase.NewIPAccessUseCase(ipRepo, 4096)
	rateRepo := repository.NewRateLimitRepository()
	rateUseCase := usecase.NewRateLimitUseCase(rateRepo, cfgManager)
	proxyUseCase := usecase.NewProxyUseCase(cfgManager)

	router := _jsii.NewRouter(_jsii.Dependencies{
		Logger:           &log.Logger,
		IPAccessUseCase:  ipAccessUseCase,
		RateLimitUseCase: rateUseCase,
		ProxyUseCase:     proxyUseCase,
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
