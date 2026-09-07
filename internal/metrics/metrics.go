package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Evaluations *prometheus.CounterVec
	Latency     *prometheus.HistogramVec
	HTTP        *prometheus.CounterVec
}

func New(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		Evaluations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "feature_flag_evaluations_total",
			Help: "Total number of feature flag evaluations.",
		}, []string{"result"}),
		Latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "feature_flag_evaluation_duration_seconds",
			Help: "Duration of feature flag evaluations.",
		}, []string{"result"}),
		HTTP: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "feature_flag_http_requests_total",
			Help: "Total HTTP requests handled by the service.",
		}, []string{"method", "status"}),
	}
	registerer.MustRegister(m.Evaluations, m.Latency, m.HTTP)
	return m
}

func (m *Metrics) ObserveEvaluation(found, enabled bool, started time.Time) {
	result := "missing"
	if found {
		result = "disabled"
		if enabled {
			result = "enabled"
		}
	}
	m.Evaluations.WithLabelValues(result).Inc()
	m.Latency.WithLabelValues(result).Observe(time.Since(started).Seconds())
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()
		next.ServeHTTP(writer, r)
		m.HTTP.WithLabelValues(r.Method, strconv.Itoa(writer.status)).Inc()
		_ = started
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
