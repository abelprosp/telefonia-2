package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/config"
)

func TestIntakeObjectIsUploadedPrivatelyWithOriginalBytes(t *testing.T) {
	data := []byte("%PDF-1.4\n%%EOF")
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != "PUT" || r.URL.Path != "/files/tenant/intake/unique/00.pdf" {
			t.Errorf("unexpected destination: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/pdf" || r.Header.Get("Content-Disposition") != "attachment" || r.Header.Get("X-Amz-Acl") != "" {
			t.Error("incorrect or public object metadata")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != string(data) {
			t.Errorf("object bytes changed: %q %v", body, err)
		}
		w.Header().Set("ETag", `"fake"`)
		w.WriteHeader(200)
	}))
	defer server.Close()
	c, err := NewClient(config.Config{ObjectStorageServiceURL: server.URL, ObjectStorageAccessKeyID: "test", ObjectStorageSecretKey: "test-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if err = c.PutIntakeObject(context.Background(), "files", "tenant/intake/unique/00.pdf", "application/pdf", data); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("object not uploaded")
	}
}
