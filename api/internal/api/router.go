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
	r.Get("/healtz", systemHandler.Healtz)
	r.Get("/health", systemHandler.Health)
	r.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	r.Get("/ws", hub.ServeWS)

	r.Route("/api/v3", func(api chi.Router) {
		api.Route("/geokrety", func(gr chi.Router) {
			gr.Get("/", statsHandler.GetGeokretyList)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}", statsHandler.GetGeokretyDetailsByGkId)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/stats", statsHandler.GetGeokretStats)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/moves", statsHandler.GetGeokretyMoves)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/loved-by", statsHandler.GetGeokretyLovedBy)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/watched-by", statsHandler.GetGeokretyWatchedBy)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/finders", statsHandler.GetGeokretyFinders)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/pictures", statsHandler.GetGeokretyPictures)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/countries", statsHandler.GetGeokretyCountries)
			gr.Get("/{gkid:(?:[Gg][Kk])?[0-9A-Fa-f]+}/waypoints", statsHandler.GetGeokretyWaypoints)
		})
		api.Route("/moves", func(mr chi.Router) {
			mr.Get("/", statsHandler.GetMoveList)
			mr.Get("/{id:[0-9]+}", statsHandler.GetMoveDetails)
		})
		api.Route("/countries", func(cr chi.Router) {
			cr.Get("/", statsHandler.GetCountryList)
			cr.Get("/{code:[A-Za-z]{2}}", statsHandler.GetCountryDetails)
			cr.Get("/{code:[A-Za-z]{2}}/geokrety", statsHandler.GetCountryGeokrety)
		})
		api.Route("/waypoints", func(wr chi.Router) {
			wr.Get("/{code:[A-Za-z0-9_-]{2,32}}", statsHandler.GetWaypoint)
			wr.Get("/{code:[A-Za-z0-9_-]{2,32}}/geokrety-current", statsHandler.GetWaypointCurrentGeokrety)
			wr.Get("/{code:[A-Za-z0-9_-]{2,32}}/geokrety-past", statsHandler.GetWaypointPastGeokrety)
		})
		api.Route("/users", func(ur chi.Router) {
			ur.Get("/", statsHandler.GetUserList)
			ur.Get("/{id:[0-9]+}", statsHandler.GetUserDetails)
			ur.Get("/{id:[0-9]+}/stats", statsHandler.GetUserStats)
			ur.Get("/{id:[0-9]+}/geokrety-owned", statsHandler.GetUserOwnedGeokrety)
			ur.Get("/{id:[0-9]+}/geokrety-found", statsHandler.GetUserFoundGeokrety)
			ur.Get("/{id:[0-9]+}/geokrety-loved", statsHandler.GetUserLovedGeokrety)
			ur.Get("/{id:[0-9]+}/geokrety-watched", statsHandler.GetUserWatchedGeokrety)
			ur.Get("/{id:[0-9]+}/pictures", statsHandler.GetUserPictures)
			ur.Get("/{id:[0-9]+}/countries", statsHandler.GetUserCountries)
			ur.Get("/{id:[0-9]+}/waypoints", statsHandler.GetUserWaypoints)
		})
		api.Route("/pictures", func(pr chi.Router) {
			pr.Get("/", statsHandler.GetPictureList)
			pr.Get("/{id:[0-9]+}", statsHandler.GetPictureDetails)
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
