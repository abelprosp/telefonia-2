package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestLocalSessionKeepsTenantAndRolesOnProtectedAPIs(t *testing.T) {
	database := os.Getenv("AUTH_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a disposable database")
	}
	ctx := context.Background()
	st, e := store.New(ctx, database)
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	hash, _ := localauth.HashPassword("safe-test-password-2026")
	user := &store.LocalUser{ID: uuid.NewString(), Username: "sales-" + uuid.NewString(), Email: "sales@example.com", OrganizationID: uuid.NewString(), Roles: []string{"sales", "user"}, Enabled: true, PasswordHash: hash}
	if e = st.CreateLocalUser(ctx, user); e != nil {
		t.Fatal(e)
	}
	defer st.Pool().Exec(ctx, `DELETE FROM auth_users WHERE id=$1`, user.ID)
	local, e := localauth.New(st, strings.Repeat("a", 64), true, []string{"https://telefonia.example"}, "telefonia.example", "Telefonia")
	if e != nil {
		t.Fatal(e)
	}
	mw, e := auth.NewMiddleware(config.Config{Environment: "Production"}, slog.Default(), local)
	if e != nil {
		t.Fatal(e)
	}
	router := chi.NewRouter()
	local.Register(router)
	router.With(mw.Authenticate, mw.RequireOperational).Get("/operation", func(w http.ResponseWriter, r *http.Request) {
		if auth.OrganizationFromContext(r.Context()).ID != user.OrganizationID {
			t.Error("tenant replaced by caller header")
		}
		w.WriteHeader(204)
	})
	router.With(mw.Authenticate, mw.RequireMaster).Get("/master", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	router.With(mw.Authenticate, mw.RequireOperational).Post("/operation", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	req := httptest.NewRequest("POST", "https://telefonia.example/v1/auth/login", strings.NewReader(`{"username":"`+user.Username+`","password":"safe-test-password-2026"}`))
	req.Header.Set("Origin", "https://telefonia.example")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	// Production passwords only create a pending session; protected routes stay blocked.
	req = httptest.NewRequest("GET", "https://telefonia.example/operation", nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("pending session reached protected route: %d", w.Code)
	}
	// Simulate a completed passkey ceremony by marking the session verified.
	tokenHash := sha256.Sum256([]byte(cookie.Value))
	if _, e = st.Pool().Exec(ctx, `UPDATE auth_sessions SET mfa_verified=true WHERE token_hash=$1`, hex.EncodeToString(tokenHash[:])); e != nil {
		t.Fatal(e)
	}
	for _, test := range []struct {
		method, path, origin string
		cookie               bool
		status               int
	}{
		{"GET", "/operation", "", true, 204}, {"GET", "/master", "", true, 403},
		{"POST", "/operation", "https://evil.example", true, 403}, {"POST", "/operation", "https://telefonia.example", true, 204},
		{"GET", "/operation", "", false, 401},
	} {
		req = httptest.NewRequest(test.method, "https://telefonia.example"+test.path, nil)
		req.Header.Set("Origin", test.origin)
		req.Header.Set("X-Organization-ID", "another-tenant")
		req.Header.Set("Authorization", "Bearer legacy-or-forged-jwt")
		if test.cookie {
			req.AddCookie(cookie)
		}
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != test.status {
			t.Errorf("%s %s: %d want %d", test.method, test.path, w.Code, test.status)
		}
	}
}
