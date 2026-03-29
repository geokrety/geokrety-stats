package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/api"
	"github.com/geokrety/geokrety-stats-api/internal/config"
	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/handlers"
	"github.com/geokrety/geokrety-stats-api/internal/logging"
	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	logger, err := logging.New(cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	defer func() { _ = logger.Sync() }()

	store, err := db.Open(cfg)
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer func() { _ = store.Close() }()

	registry := prometheus.NewRegistry()
	metricsCollector := metrics.New(registry)

	hub := ws.NewHub(logger, metricsCollector, time.Duration(cfg.WSBroadcastInterval)*time.Millisecond)
	statsHandler := handlers.NewStatsHandler(handlers.StatsHandlerStores{
		Geokrety:  store,
		Moves:     store,
		Countries: store,
		Waypoints: store,
		Users:     store,
		Pictures:  store,
	}, logger)
	systemHandler := handlers.NewSystemHandler(store, hub, logger)

	router := api.NewRouter(cfg, logger, metricsCollector, registry, statsHandler, systemHandler, hub)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting server",
			zap.String("addr", srv.Addr),
			zap.String("log_level", cfg.LogLevel),
			zap.Bool("swagger", cfg.EnableSwagger),
		)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	waitForShutdown(srv, logger)
}

func waitForShutdown(server *http.Server, logger *zap.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown gracefully", zap.Error(err))
		return
	}
	logger.Info("server shutdown complete")
}
