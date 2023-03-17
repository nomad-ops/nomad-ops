package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "http"

	counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_count",
		Help:      "http request count.",
	}, []string{"app", "name", "method", "status"})

	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "request_duration_seconds",
		Help:      "Time taken to execute request.",
	}, []string{"app", "name", "method", "status"})
)

type metricResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricResponseWriter(w http.ResponseWriter) *metricResponseWriter {
	return &metricResponseWriter{w, http.StatusOK}
}

func (lrw *metricResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// NewPrometheusMiddleware
func NewPrometheusMiddleware(handler http.Handler, app string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newMetricResponseWriter(w)
		handler.ServeHTTP(lrw, r)
		statusCode := lrw.statusCode
		duration := time.Since(start)
		histogram.WithLabelValues(app, r.URL.String(), r.Method, fmt.Sprintf("%d", statusCode)).Observe(duration.Seconds())
		counter.WithLabelValues(app, r.URL.String(), r.Method, fmt.Sprintf("%d", statusCode)).Inc()
	})
	_ = prometheus.Register(histogram)
	_ = prometheus.Register(counter)
	return h
}
