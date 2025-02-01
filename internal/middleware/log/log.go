package log

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := statusRecorder{ResponseWriter: w}

		start := time.Now()
		defer func() {
			log.Printf("%s %s status %d, processed in %v", r.Method, r.URL.Path, wrapper.status, time.Since(start))
		}()

		next.ServeHTTP(&wrapper, r)
	})
}
