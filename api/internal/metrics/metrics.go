package metrics

import "github.com/prometheus/client_golang/prometheus"

type Collector struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	WSConnections       prometheus.Gauge
	WSBroadcastsTotal   prometheus.Counter
}

func New(registerer prometheus.Registerer) *Collector {
	c := &Collector{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "geokrety_stats_api",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "geokrety_stats_api",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		WSConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "geokrety_stats_api",
				Name:      "ws_connections",
				Help:      "Current number of active websocket connections.",
			},
		),
		WSBroadcastsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "geokrety_stats_api",
				Name:      "ws_broadcasts_total",
				Help:      "Total websocket broadcast messages sent.",
			},
		),
	}

	registerer.MustRegister(
		c.HTTPRequestsTotal,
		c.HTTPRequestDuration,
		c.WSConnections,
		c.WSBroadcastsTotal,
	)

	return c
}
