package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

// NewFileServerMiddleware ...
func NewFileServerMiddleware(next http.Handler, baseURL string, staticFolder string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, baseURL) {
			http.StripPrefix(baseURL, http.FileServer(http.Dir(staticFolder))).ServeHTTP(w, r)
			return
		}

		// Move to closer to API
		next.ServeHTTP(w, r)
	})
}

// NewSwaggerJSONMiddleware ...
// Use like this:
//
//	func setupGlobalMiddleware(handler http.Handler) http.Handler {
//		fh := NewFileServerMiddleware(handler, "/coordinator/api/ui/", "wwwroot")
//		sh := NewSwaggerJSONMiddleware(fh, "/coordinator/api/ui/swagger.json", SwaggerJSON)
//		lh := NewLoggingMiddleware(sh, log.CreateLogger(context.Background(), "REST API", false))
//		return lh
//	}
func NewSwaggerJSONMiddleware(next http.Handler, swaggerURL string, swaggerJSON []byte) http.Handler {
	envHost := os.Getenv("HTTP_SWAGGER_JSON_HOST")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == swaggerURL {
			h := w.Header()
			h.Set("Content-Type", "application/json")

			scheme := "https"
			ref := r.Header.Get("Referer")
			if strings.HasPrefix(ref, "https://") {
				scheme = "https"
			}
			if strings.HasPrefix(ref, "http://") {
				scheme = "http"
			}
			newV := scheme + "://" + r.Host

			if envHost != "" {
				newV = envHost
			}

			_, _ = w.Write([]byte(strings.ReplaceAll(string(swaggerJSON), "http://XXXHOSTXXX", newV)))
			return
		}

		// Move to closer to API
		next.ServeHTTP(w, r)
	})
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         200,
	}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// NewLoggingMiddleware ...
func NewLoggingMiddleware(next http.Handler, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Move to closer to API
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		logger.LogInfo(context.Background(), "%s %s %s - %v %v",
			r.Method, r.URL.EscapedPath(), r.Proto, wrapped.status, time.Since(start))
	})
}

type LoggingOptions struct {
	PrefixesToIgnore []string
}

// NewLoggingMiddlewareWithOptions ...
func NewLoggingMiddlewareWithOptions(next http.Handler, logger log.Logger, opts LoggingOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Move to closer to API
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		escPath := r.URL.EscapedPath()
		for _, ig := range opts.PrefixesToIgnore {
			if strings.HasPrefix(escPath, ig) {
				return
			}
		}

		logger.LogInfo(context.Background(), "%s %s %s - %v %v",
			r.Method, r.URL.EscapedPath(), r.Proto, wrapped.status, time.Since(start))
	})
}
