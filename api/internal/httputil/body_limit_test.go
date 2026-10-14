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

func TestIntakeBodyLimitDoesNotExpandOtherRoutes(t *testing.T) {
	for _, test := range []struct {
		method, path string
		length       int64
		want         int
	}{
		{"POST", "/v1/intake/customers", 3 << 20, 204},
		{"POST", "/v1/customers", 3 << 20, 204},
		{"POST", "/v1/customers/", 3 << 20, 204},
		{"PATCH", "/v1/customers", 3 << 20, 413},
		{"POST", "/v1/providers", 3 << 20, 413},
		{"POST", "/v1/intake/customers", 25 << 20, 413},
		{"POST", "/v1/integrations/luxus-parceiros/sales", 31 << 20, 204},
		{"POST", "/v1/integrations/luxus-parceiros/sales", 33 << 20, 413},
	} {
		r := httptest.NewRequest(test.method, test.path, nil)
		r.ContentLength = test.length
		w := httptest.NewRecorder()
		CustomerIntakeBodyLimit(2<<20)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })).ServeHTTP(w, r)
		if w.Code != test.want {
			t.Fatalf("%s %s: %d want %d", test.method, test.path, w.Code, test.want)
		}
	}
}
