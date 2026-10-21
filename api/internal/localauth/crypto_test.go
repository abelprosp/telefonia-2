package localauth

import (
	"github.com/luxus-connect/telefonia/api/internal/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPasswordHashSaltAndPolicy(t *testing.T) {
	for _, p := range []string{"", "short", strings.Repeat("a", 257)} {
		if _, e := HashPassword(p); e == nil {
			t.Fatal("weak password accepted")
		}
	}
	password := "correta com espaços e acentuação"
	a, e := HashPassword(password)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := HashPassword(password)
	if a == b || !CheckPassword(a, password) || CheckPassword(a, "wrong") || CheckPassword("$argon2id$bad", password) {
		t.Fatal("invalid hashing or verification")
	}
	if CheckPassword(strings.Replace(a, "m=19456", "m=999999999", 1), password) {
		t.Fatal("unbounded hashing parameters")
	}
}
func TestRecoveryCodeFormattingAndLegacyCompatibility(t *testing.T) {
	a, err := newRecoveryCode()
	if err != nil {
		t.Fatal(err)
	}
	b, err := newRecoveryCode()
	if err != nil || a == b {
		t.Fatal("recovery codes must be independently random", err)
	}
	groups := strings.Split(a, "-")
	if len(groups) != 4 {
		t.Fatal("recovery code must contain four groups")
	}
	for _, group := range groups {
		if len(group) != 8 || strings.Trim(group, "0123456789ABCDEF") != "" {
			t.Fatal("unexpected recovery code alphabet or length")
		}
	}
	for _, formatted := range []string{a, strings.ToLower(a), strings.ReplaceAll(a, "-", ""), " \n" + strings.ReplaceAll(a, "-", " \t") + "\n"} {
		if recoveryCodeHash(formatted) != recoveryCodeHash(a) {
			t.Fatal("formatted recovery code did not match")
		}
	}
	legacy := "aBcDef-ghij_Klmnopqrstuvwxyz01234567890123456"
	if recoveryCodeHash(" "+legacy+"\n") != digest(legacy) {
		t.Fatal("old recovery code was changed")
	}
	if recoveryCodeHash(strings.ToUpper(legacy)) == digest(legacy) {
		t.Fatal("old recovery codes must remain case-sensitive")
	}
	if recoveryCodeHash("123456") == recoveryCodeHash(a) {
		t.Fatal("short code matched a recovery code")
	}
}
func TestCookieCSRFAndPrivilegedMFA(t *testing.T) {
	m := &Manager{Production: true, Origins: []string{"https://telefonia.example"}}
	w := httptest.NewRecorder()
	m.cookie(w, "token", 100)
	c := w.Result().Cookies()[0]
	if !c.Secure || !c.HttpOnly || c.Path != "/" || c.Domain != "" || c.SameSite != http.SameSiteStrictMode || !strings.HasPrefix(c.Name, "__Host-") {
		t.Fatal("unsafe session cookie")
	}
	req := httptest.NewRequest("POST", "https://telefonia.example/v1/auth/login", nil)
	for _, origin := range []string{"", "https://evil.example", "https://telefonia.example.evil.test"} {
		req.Header.Set("Origin", origin)
		if m.AllowedOrigin(req) {
			t.Fatal("CSRF origin accepted")
		}
	}
	req.Header.Set("Origin", "https://telefonia.example")
	if !m.AllowedOrigin(req) {
		t.Fatal("exact origin rejected")
	}
	s := &Session{User: &store.LocalUser{Roles: []string{"master"}}}
	if !m.NeedsMFA(s) {
		t.Fatal("privileged access without MFA")
	}
	s.Verified = true
	if m.NeedsMFA(s) {
		t.Fatal("verified session rejected")
	}
	m.Production = false
	s.Verified = false
	s.User.PasskeyCount = 1
	if !m.NeedsMFA(s) {
		t.Fatal("enrolled passkey must be checked even in development")
	}
	s.User.PasskeyCount = 0
	if m.NeedsMFA(s) {
		t.Fatal("development without passkeys should not force MFA")
	}
}
