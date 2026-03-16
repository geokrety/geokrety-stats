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

	r.Get("/", Home)
	r.Get("/healtz", systemHandler.Healtz) // TODO simple liveness check without DB dependency
	r.Get("/health", systemHandler.Health) // TODO comprehensive health check with DB and dependencies
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
			gr.Get("/{id}", statsHandler.GetGeokrety)
			gr.Get("/{id}/moves", statsHandler.GetGeokretyMoves)
			gr.Get("/{id}/moves/{moveId}", statsHandler.GetGeokretyMoveDetails)
			gr.Get("/{id}/loved-by", statsHandler.GetGeokretyLovedBy)
			gr.Get("/{id}/watched-by", statsHandler.GetGeokretyWatchedBy)
			gr.Get("/{id}/pictures", statsHandler.GetGeokretyPictures)
			gr.Get("/search", statsHandler.SearchGeokrety)
			gr.Get("/{id}/loves", statsHandler.GetGeokretyLovedBy)
			gr.Get("/{id}/watches", statsHandler.GetGeokretyWatchedBy)
			gr.Get("/recent-moves", statsHandler.GetRecentMoves)
			gr.Get("/recent-born", statsHandler.GetRecentBorn)
			gr.Get("/recent-loved", statsHandler.GetRecentLoved)
			gr.Get("/recent-watched", statsHandler.GetRecentWatched)
			gr.Get("/{id}/timeline", statsHandler.GetGeokretTimeline)
			gr.Get("/{id}/countries/timeline", statsHandler.GetGeokretTimeline)
			gr.Get("/{id}/events/timeline", statsHandler.GetGeokretTimeline)
			gr.Get("/{id}/circulation", statsHandler.GetGeokretCirculation)
			gr.Get("/{id}/countries", statsHandler.GetGeokretyCountries)
			gr.Get("/{id}/waypoints", statsHandler.GetGeokretyWaypoints)
			gr.Get("/{id}/stats/map/countries", statsHandler.GetGeokretyStatsMapCountries)
			gr.Get("/{id}/stats/elevation", statsHandler.GetGeokretyStatsElevation)
			gr.Get("/{id}/stats/heatmap/days", statsHandler.GetGeokretyStatsHeatmapDays)
			gr.Get("/{id}/geojson/trip", statsHandler.GetGeokretyGeoJSONTrip)
			gr.Get("/{id}/world-choropleth", statsHandler.GetGeokretyStatsMapCountries)
			gr.Get("/{id}/WorldChoropleth", statsHandler.GetGeokretyWorldChoropleth)

		})
		api.Route("/countries", func(cr chi.Router) {
			cr.Get("/{code}", statsHandler.GetCountryDetails)
			cr.Get("/recent-active", statsHandler.GetRecentActiveCountries)
			cr.Get("/{code}/geokrety", statsHandler.GetCountryGeokrety)
			cr.Get("/{code}/spotted-geokrety", statsHandler.GetCountrySpottedGeokrety)
		})
		api.Route("/waypoints", func(wr chi.Router) {
			wr.Get("/recent-active", statsHandler.GetRecentActiveWaypoints)
			wr.Get("/{code}", statsHandler.GetWaypoint)
			wr.Get("/{code}/geokrety-current", statsHandler.GetWaypointCurrentGeokrety)
			wr.Get("/{code}/geokrety-past", statsHandler.GetWaypointPastGeokrety)
			wr.Get("/{code}/spotted-geokrety", statsHandler.GetWaypointSpottedGeokrety)
			wr.Get("/{code}/past-geokrety", statsHandler.GetWaypointPastGeokrety)
			wr.Get("/search", statsHandler.SearchWaypoints)
		})
		api.Route("/users", func(ur chi.Router) {
			ur.Get("/{id}", statsHandler.GetUserDetails)
			ur.Get("/{id}/geokrety-owned", statsHandler.GetUserOwnedGeokrety)
			ur.Get("/{id}/geokrety-found", statsHandler.GetUserFoundGeokrety)
			ur.Get("/{id}/geokrety-loved", statsHandler.GetUserLovedGeokrety)
			ur.Get("/{id}/geokrety-watched", statsHandler.GetUserWatchedGeokrety)
			ur.Get("/{id}/owned-geokrety", statsHandler.GetUserOwnedGeokrety)
			ur.Get("/{id}/found-geokrety", statsHandler.GetUserFoundGeokrety)
			ur.Get("/{id}/loved-geokrety", statsHandler.GetUserLovedGeokrety)
			ur.Get("/{id}/watched-geokrety", statsHandler.GetUserWatchedGeokrety)
			ur.Get("/{id}/pictures", statsHandler.GetUserPictures)
			ur.Get("/{id}/countries", statsHandler.GetUserCountries)
			ur.Get("/{id}/waypoints", statsHandler.GetUserWaypoints)
			ur.Get("/{id}/network", statsHandler.GetUserNetwork)
			ur.Get("/search", statsHandler.SearchUsers)
			ur.Get("/recent-registered", statsHandler.GetRecentRegisteredUsers)
			ur.Get("/recent-active", statsHandler.GetRecentActiveUsers)
			ur.Get("/{id}/stats/heatmap/days", statsHandler.GetUserStatsHeatmapDays)
			ur.Get("/{id}/stats/heatmap/hours", statsHandler.GetUserStatsHeatmapHours)
			ur.Get("/{id}/stats/map/countries", statsHandler.GetUserStatsMapCountries)
		})
		api.Route("/pictures", func(pr chi.Router) {
			pr.Get("/{id}", statsHandler.GetPicture)
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
