package httputil

import (
	"net/http"
	"strings"
)

const IntakeBodyLimit int64 = 24 << 20

// Intake-aware body limit keeps the normal 2 MiB ceiling on all other routes.
func CustomerIntakeBodyLimit(defaultBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := defaultBytes
			p := strings.TrimSuffix(r.URL.Path, "/")
			if r.Method == http.MethodPost && (p == "/v1/intake/customers" || p == "/v1/customers") {
				limit = IntakeBodyLimit
			}
			RequestBodyLimit(limit)(next).ServeHTTP(w, r)
		})
	}
}

func RequestBodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w, "Corpo da requisição excede o limite permitido.", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
