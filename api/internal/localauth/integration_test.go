package localauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestPostgresPasskeyRecoveryRevocationAndExpiry(t *testing.T) {
	database := os.Getenv("AUTH_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	ctx := context.Background()
	st, e := store.New(ctx, database)
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	password := "senha de teste segura 2026"
	hash, _ := HashPassword(password)
	user := &store.LocalUser{ID: uuid.NewString(), Username: "auth-test-" + uuid.NewString(), Email: "test@example.com", OrganizationID: uuid.NewString(), Roles: []string{"master", "user"}, Enabled: true, PasswordHash: hash}
	if e = st.CreateLocalUser(ctx, user); e != nil {
		t.Fatal(e)
	}
	defer st.Pool().Exec(ctx, `DELETE FROM auth_users WHERE id=$1`, user.ID)
	m, e := New(st, strings.Repeat("s", 64), true, []string{"https://telefonia.example"}, "telefonia.example", "Telefonia")
	if e != nil {
		t.Fatal(e)
	}
	router := chi.NewRouter()
	m.Register(router)
	call := func(path string, data any, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(data)
		req := httptest.NewRequest("POST", "https://telefonia.example/v1/auth/"+path, strings.NewReader(string(raw)))
		req.Header.Set("Origin", origin)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	get := func(path string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("GET", "https://telefonia.example/v1/auth/"+path, nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	getSession := func(cookie *http.Cookie) (*Session, error) {
		req := httptest.NewRequest("GET", "https://telefonia.example/", nil)
		req.AddCookie(cookie)
		return m.Resolve(req)
	}
	addRecovery := func(code string) {
		t.Helper()
		if _, err := st.Pool().Exec(ctx, `INSERT INTO auth_recovery_codes(user_id,code_hash) VALUES($1,$2)`, user.ID, recoveryCodeHash(code)); err != nil {
			t.Fatal(err)
		}
	}
	origin := "https://telefonia.example"
	login := func() *http.Cookie {
		t.Helper()
		w := call("login", map[string]string{"username": user.Username, "password": password}, nil, origin)
		if w.Code != 200 {
			t.Fatalf("login: %d %s", w.Code, w.Body)
		}
		return w.Result().Cookies()[0]
	}
	if w := call("login", map[string]string{"username": user.Username, "password": password}, nil, "https://evil.example"); w.Code != 403 {
		t.Fatal("login CSRF accepted")
	}
	if w := call("login", map[string]string{"username": user.Username, "password": "wrong"}, nil, origin); w.Code != 401 {
		t.Fatal("wrong password accepted")
	}
	cookie := login()
	session, e := getSession(cookie)
	if e != nil || !m.NeedsMFA(session) || session.User.OrganizationID != user.OrganizationID {
		t.Fatal("missing tenant or passkey gate", e)
	}
	if w := get("webauthn/credentials", cookie); w.Code != 403 {
		t.Fatal("pending session listed passkeys")
	}
	// A password-only session can begin enrolling its first passkey.
	w := call("webauthn/register/begin", map[string]any{}, cookie, origin)
	if w.Code != 200 {
		t.Fatalf("register begin: %d %s", w.Code, w.Body)
	}
	var begin struct {
		Options struct {
			PublicKey struct {
				Challenge string `json:"challenge"`
			} `json:"publicKey"`
		} `json:"options"`
	}
	if e = json.Unmarshal(w.Body.Bytes(), &begin); e != nil || begin.Options.PublicKey.Challenge == "" {
		t.Fatalf("missing challenge: %s", w.Body)
	}
	if w = call("webauthn/login/begin", map[string]any{}, cookie, origin); w.Code != 409 || !strings.Contains(w.Body.String(), "PASSKEY_NOT_ENROLLED") {
		t.Fatalf("login begin without passkey: %d %s", w.Code, w.Body)
	}
	// Recovery codes elevate the session and are single-use.
	code, e := newRecoveryCode()
	if e != nil {
		t.Fatal(e)
	}
	addRecovery(code)
	w = call("recovery/verify", map[string]string{"code": strings.ToLower(strings.ReplaceAll(code, "-", " "))}, cookie, origin)
	if w.Code != 200 {
		t.Fatalf("recovery: %d %s", w.Code, w.Body)
	}
	if _, e = getSession(cookie); e == nil {
		t.Fatal("pre-verification cookie still valid after rotation")
	}
	cookie = w.Result().Cookies()[0]
	session, e = getSession(cookie)
	if e != nil || !session.Verified || m.NeedsMFA(session) {
		t.Fatal("verified session invalid", e)
	}
	if w = call("recovery/verify", map[string]string{"code": code}, cookie, origin); w.Code != 401 {
		t.Fatal("recovery code reused")
	}
	if w = get("webauthn/credentials", cookie); w.Code != 200 || !strings.Contains(w.Body.String(), `"credentials":[]`) {
		t.Fatalf("credentials: %d %s", w.Code, w.Body)
	}
	// Existing installations retain their original case-sensitive recovery codes.
	legacy, e := randomToken()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = st.Pool().Exec(ctx, `INSERT INTO auth_recovery_codes(user_id,code_hash) VALUES($1,$2)`, user.ID, digest(legacy)); e != nil {
		t.Fatal(e)
	}
	pending := login()
	if w = call("recovery/verify", map[string]string{"code": legacy}, pending, origin); w.Code != 200 {
		t.Fatalf("legacy recovery: %d %s", w.Code, w.Body)
	}
	verified := w.Result().Cookies()[0]
	if _, e = getSession(pending); e == nil {
		t.Fatal("old cookie survived recovery elevation")
	}
	if w = call("logout", map[string]string{}, verified, "https://evil.example"); w.Code != 403 {
		t.Fatal("logout CSRF accepted")
	}
	if _, e = getSession(verified); e != nil {
		t.Fatal("CSRF removed session")
	}
	if w = call("logout", map[string]string{}, verified, origin); w.Code != 200 {
		t.Fatal("logout failed")
	}
	if _, e = getSession(verified); e == nil {
		t.Fatal("logout did not revoke server session")
	}
	expired := login()
	st.Pool().Exec(ctx, `UPDATE auth_sessions SET expires_at=now()-interval '1 second' WHERE token_hash=$1`, digest(expired.Value))
	if _, e = getSession(expired); e == nil {
		t.Fatal("expired session accepted")
	}
	idle := login()
	st.Pool().Exec(ctx, `UPDATE auth_sessions SET last_seen=now()-interval '31 minutes' WHERE token_hash=$1`, digest(idle.Value))
	if _, e = getSession(idle); e == nil {
		t.Fatal("idle session accepted")
	}
	change := login()
	if w = call("password", map[string]string{"current_password": password, "new_password": "new safe password 2026"}, change, origin); w.Code != 403 {
		t.Fatal("password changed before passkey verification")
	}
	code2, e := newRecoveryCode()
	if e != nil {
		t.Fatal(e)
	}
	addRecovery(code2)
	if w = call("recovery/verify", map[string]string{"code": code2}, change, origin); w.Code != 200 {
		t.Fatalf("second recovery: %d %s", w.Code, w.Body)
	}
	change = w.Result().Cookies()[0]
	if w = call("password", map[string]string{"current_password": "wrong", "new_password": "new safe password 2026"}, change, origin); w.Code != 401 {
		t.Fatal("password changed without current password")
	}
	if w = call("password", map[string]string{"current_password": password, "new_password": "new safe password 2026"}, change, origin); w.Code != 200 {
		t.Fatalf("password change: %d %s", w.Code, w.Body)
	}
	if _, e = getSession(change); e == nil {
		t.Fatal("password change did not revoke current session")
	}
	password = "new safe password 2026"
	active := login()
	loaded, e := st.LocalUserByID(ctx, user.ID)
	if e != nil {
		t.Fatal(e)
	}
	stale := *loaded
	loaded.Enabled = false
	if e = st.UpdateLocalUser(ctx, loaded); e != nil {
		t.Fatal(e)
	}
	if _, e = getSession(active); e == nil {
		t.Fatal("disabled account session accepted")
	}
	if e = st.UpdateLocalUser(ctx, &stale); e == nil {
		t.Fatal("stale account update undid revocation")
	}
	var rawToken bool
	st.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth_sessions WHERE token_hash=$1)`, active.Value).Scan(&rawToken)
	if rawToken {
		t.Fatal("raw session cookie stored")
	}
	unknown := "unknown-" + uuid.NewString()
	for attempt := 1; attempt <= 11; attempt++ {
		w = call("login", map[string]string{"username": unknown, "password": "wrong-password"}, nil, origin)
		expected := 401
		if attempt == 11 {
			expected = 429
		}
		if w.Code != expected {
			t.Fatalf("attempt %d: %d want %d", attempt, w.Code, expected)
		}
	}
}
