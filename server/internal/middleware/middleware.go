package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// LoggingMiddleware logs incoming HTTP requests
func Logger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture the status code
		wrapped := &responseWriterWrapper{ResponseWriter: w, status: http.StatusOK}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		log.Printf(
			"method=%s path=%s remote_addr=%s status=%d duration=%s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrapped.status,
			duration,
		)
	}
}

func RecoveryMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				msg := "Caught panic: %v, Stack treace: %s"
				log.Printf(msg, err, string(debug.Stack()))

				er := http.StatusInternalServerError
				http.Error(w, "Internal Server Error", er)
			}
		}()

		next.ServeHTTP(w, r)
	}
}

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for _, middleware := range middlewares {
			next = middleware(next)
		}
		return next.ServeHTTP
	}
}
