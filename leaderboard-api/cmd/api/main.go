package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/geokrety/leaderboard-api/internal/config"
	"github.com/geokrety/leaderboard-api/internal/db"
	"github.com/geokrety/leaderboard-api/internal/handlers"
	"github.com/geokrety/leaderboard-api/internal/middleware"
	wsHub "github.com/geokrety/leaderboard-api/internal/websocket"
)

func main() {
	// ── Logging ──────────────────────────────────────────────────────────────
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// ── Config ───────────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── Database ─────────────────────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.Database.DSN, cfg.Database.PoolMax, cfg.Database.PoolMin)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection failed")
	}
	defer pool.Close()

	// ── Handlers & WebSocket hub ──────────────────────────────────────────────
	h := handlers.New(pool)
	hub := wsHub.New()
	go hub.Run()
	go h.StartBroadcaster(ctx, hub, time.Duration(cfg.Server.RefreshInterval)*time.Second)

	// ── Materialized view refresh scheduler (every 15 min) ───────────────────
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				refCtx, refCancel := context.WithTimeout(context.Background(), 5*time.Minute)
				if _, err := pool.Exec(refCtx, `SELECT geokrety_stats.refresh_leaderboard_views()`); err != nil {
					log.Error().Err(err).Msg("materialized view refresh failed")
				} else {
					log.Info().Msg("materialized views refreshed")
				}
				refCancel()
			case <-ctx.Done():
				return
			}
		}
	}()

	// ── Router ───────────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Server.AllowOrigins))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	// WebSocket
	r.GET("/ws", h.ServeWS(hub))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// ── Leaderboard ──────────────────────────────────────────────────
		v1.GET("/leaderboard", h.Leaderboard)

		// ── Users ────────────────────────────────────────────────────────
		v1.GET("/users", h.ListUsers)
		v1.GET("/users/:id", h.GetUser)
		v1.GET("/users/:id/moves", h.UserMoves)
		v1.GET("/users/:id/geokrety", h.UserGeokrety)
		v1.GET("/users/:id/countries", h.UserCountries)
		v1.GET("/users/:id/points/timeline", h.UserPointsTimeline)
		v1.GET("/users/:id/points/breakdown", h.UserPointsBreakdown)
		v1.GET("/users/:id/points/awards", h.UserPointsAwards)
		v1.GET("/users/:id/rank/history", h.UserRankHistory)

		// ── GeoKrety ─────────────────────────────────────────────────────
		v1.GET("/geokrety", h.ListGeokrety)
		v1.GET("/geokrety/:id", h.GetGeoKret)
		v1.GET("/geokrety/:id/moves", h.GeoKretMoves)
		v1.GET("/geokrety/:id/countries", h.GeoKretCountries)
		v1.GET("/geokrety/:id/holders", h.GeoKretHolderHistory)
		v1.GET("/geokrety/:id/points/timeline", h.GeoKretPointsTimeline)

		// ── Global Stats ─────────────────────────────────────────────────
		v1.GET("/stats", h.GlobalStats)
		v1.GET("/stats/activity/daily", h.DailyActivity)
		v1.GET("/stats/countries", h.TopCountries)
		v1.GET("/stats/points/breakdown", h.PointsBreakdownGlobal)
		v1.GET("/stats/periods", h.AvailablePeriods)
	}

	// ── Server ───────────────────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", addr).Msg("leaderboard-api listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down...")
	cancel()
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("stopped")
}
