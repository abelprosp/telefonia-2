package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyLimitKnownAndChunked(t *testing.T) {
	for _, length := range []int64{20, -1} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("x", 20)))
		r.ContentLength = length
		w := httptest.NewRecorder()
		handler := RequestBodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.ReadAll(r.Body); err != nil {
				w.WriteHeader(413)
				return
			}
			w.WriteHeader(204)
		}))
		handler.ServeHTTP(w, r)
		if w.Code != 413 {
			t.Errorf("length=%d status=%d", length, w.Code)
		}
	}
}
