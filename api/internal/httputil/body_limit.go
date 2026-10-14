package httputil

import (
	"net/http"
	"strings"
)

func RequestBodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := maxBytes
			if strings.Contains(r.URL.Path, "/integrations/luxus-parceiros/") {
				limit = 32 << 20
			}
			if r.ContentLength > limit {
				http.Error(w, "Corpo da requisição excede o limite permitido.", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
