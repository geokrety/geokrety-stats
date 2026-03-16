package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	"github.com/geokrety/geokrety-stats-api/internal/handlers"
	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

func NewRouter(
	cfg config.Config,
	logger *zap.Logger,
	metricsCollector *metrics.Collector,
	gatherer prometheus.Gatherer,
	statsHandler *handlers.StatsHandler,
	systemHandler *handlers.SystemHandler,
	hub *ws.Hub,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(requestLogging(logger))
	r.Use(httpMetrics(metricsCollector))
	r.Use(cors)

	r.Get("/health", systemHandler.Health)
	r.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	r.Get("/ws", hub.ServeWS)

	r.Route("/api/v3", func(api chi.Router) {
		api.Route("/stats", func(sr chi.Router) {
			sr.Get("/kpis", statsHandler.GetKPIs)
			sr.Get("/countries", statsHandler.GetCountries)
			sr.Get("/leaderboard", statsHandler.GetLeaderboard)
			sr.Get("/hourly-heatmap", statsHandler.GetHourlyHeatmap)
			sr.Get("/country-flows", statsHandler.GetCountryFlows)
			sr.Get("/top-caches", statsHandler.GetTopCaches)
			sr.Get("/first-finder-leaderboard", statsHandler.GetFirstFinderLeaderboard)
			sr.Get("/distance-records", statsHandler.GetDistanceRecords)
		})
		api.Route("/geokrety", func(gr chi.Router) {
			gr.Get("/recent-moves", statsHandler.GetRecentMoves)
			gr.Get("/recent-born", statsHandler.GetRecentBorn)
			gr.Get("/recent-loved", statsHandler.GetRecentLoved)
			gr.Get("/recent-watched", statsHandler.GetRecentWatched)
			gr.Get("/{id}/timeline", statsHandler.GetGeokretTimeline)
			gr.Get("/{id}/circulation", statsHandler.GetGeokretCirculation)
		})
		api.Route("/countries", func(cr chi.Router) {
			cr.Get("/recent-active", statsHandler.GetRecentActiveCountries)
		})
		api.Route("/waypoints", func(wr chi.Router) {
			wr.Get("/recent-active", statsHandler.GetRecentActiveWaypoints)
		})
		api.Route("/users", func(ur chi.Router) {
			ur.Get("/recent-registered", statsHandler.GetRecentRegisteredUsers)
			ur.Get("/recent-active", statsHandler.GetRecentActiveUsers)
			ur.Get("/{id}/network", statsHandler.GetUserNetwork)
		})
	})

	if cfg.EnableSwagger {
		r.Get("/openapi.yaml", serveOpenAPISpec)
		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs/index.html", http.StatusTemporaryRedirect)
		})
		r.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/openapi.yaml"),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DefaultModelsExpandDepth(-1),
		))
	}

	return r
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		http.Error(w, "failed to read working directory", http.StatusInternalServerError)
		return
	}
	path := filepath.Join(wd, "docs", "openapi.yaml")
	if _, err := os.Stat(path); err != nil {
		http.Error(w, "openapi spec not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeFile(w, r, path)
}

func requestLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			started := time.Now()
			next.ServeHTTP(ww, r)
			logger.Info("http_request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", time.Since(started)),
			)
		})
	}
}

func httpMetrics(collector *metrics.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			started := time.Now()
			next.ServeHTTP(ww, r)
			status := ww.Status()
			collector.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(status)).Inc()
			collector.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(started).Seconds())
		})
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
