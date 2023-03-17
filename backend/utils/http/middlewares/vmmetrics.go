package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

// NewVictoriaMetricsMiddleware
func NewVictoriaMetricsMiddleware(handler http.Handler, app string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newMetricResponseWriter(w)
		handler.ServeHTTP(lrw, r)
		statusCode := lrw.statusCode
		duration := time.Since(start)
		metrics.GetOrCreateHistogram(fmt.Sprintf(`http_request_duration_seconds{app="%s",name="%s",method="%s",status="%d"}`,
			app, r.URL.String(), r.Method, statusCode)).Update(duration.Seconds())
		metrics.GetOrCreateCounter(fmt.Sprintf(`http_request_count{app="%s",name="%s",method="%s",status="%d"}`,
			app, r.URL.String(), r.Method, statusCode)).Inc()
	})
	return h
}
