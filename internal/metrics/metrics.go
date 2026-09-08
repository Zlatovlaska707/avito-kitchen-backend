package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	WorkerRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_runs_total",
		Help: "Total worker run attempts",
	}, []string{"name", "result"})

	WorkerProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_processed_total",
		Help: "Total items processed by workers",
	}, []string{"name"})

	OrdersCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total orders created",
	})

	WebhookDeliveriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_deliveries_total",
		Help: "Total webhook delivery attempts",
	}, []string{"result"})
)

func Handler() http.Handler {
	return promhttp.Handler()
}

func ObserveHTTP(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)
	HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func ObserveWorkerRun(name, result string, processed int) {
	WorkerRunsTotal.WithLabelValues(name, result).Inc()
	if processed > 0 {
		WorkerProcessedTotal.WithLabelValues(name).Add(float64(processed))
	}
}
