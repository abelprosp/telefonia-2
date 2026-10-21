package localauth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"github.com/skip2/go-qrcode"
)

type Manager struct {
	Store      *store.Store
	Secret     string
	Production bool
	Origins    []string
	dummy      string
}
type Session struct {
	User     *store.LocalUser
	Hash     string
	Verified bool
}

func New(st *store.Store, secret string, production bool, origins []string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("AUTH_SECRET must contain at least 32 characters")
	}
	dummy, e := HashPassword("dummy-never-valid-password")
	if e != nil {
		return nil, e
	}
	return &Manager{Store: st, Secret: secret, Production: production, Origins: origins, dummy: dummy}, nil
}
func (m *Manager) cookieName() string {
	if m.Production {
		return "__Host-telefonia_session"
	}
	return "telefonia_session"
}

// Bound the persistent rate-limit/session tables without holding user requests.
func (m *Manager) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanup, cancel := context.WithTimeout(ctx, 30*time.Second)
				m.Store.Pool().Exec(cleanup, `DELETE FROM auth_sessions WHERE expires_at<now() OR last_seen<now()-interval '30 minutes'`)
				m.Store.Pool().Exec(cleanup, `DELETE FROM auth_attempts WHERE expires_at<now()`)
				m.Store.Pool().Exec(cleanup, `DELETE FROM auth_events WHERE created_at<now()-interval '90 days'`)
				cancel()
			}
		}
	}()
}
func (m *Manager) cookie(w http.ResponseWriter, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: m.cookieName(), Value: value, Path: "/", HttpOnly: true, Secure: m.Production, SameSite: http.SameSiteStrictMode, MaxAge: age})
}
func (m *Manager) AllowedOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	for _, allowed := range m.Origins {
		if origin == allowed {
			return true
		}
	}
	return false
}
func (m *Manager) Resolve(r *http.Request) (*Session, error) {
	c, e := r.Cookie(m.cookieName())
	if e != nil || len(c.Value) != 43 {
		return nil, fmt.Errorf("sessão ausente")
	}
	hash := digest(c.Value)
	var id string
	var verified bool
	var version int64
	// Pending password-only sessions are valid for at most ten minutes.
	e = m.Store.Pool().QueryRow(r.Context(), `UPDATE auth_sessions SET last_seen=now() WHERE token_hash=$1 AND expires_at>now() AND last_seen>now()-interval '30 minutes' AND (mfa_verified OR created_at>now()-interval '10 minutes') RETURNING user_id,mfa_verified,user_version`, hash).Scan(&id, &verified, &version)
	if e != nil {
		return nil, fmt.Errorf("sessão expirada")
	}
	u, e := m.Store.LocalUserByID(r.Context(), id)
	if e != nil || u == nil || !u.Enabled || u.Version != version {
		return nil, fmt.Errorf("sessão inválida")
	}
	return &Session{User: u, Hash: hash, Verified: verified}, nil
}
func privileged(u *store.LocalUser) bool {
	for _, r := range u.Roles {
		if r == "master" || r == "admin" || r == "financial" {
			return true
		}
	}
	return false
}
func (m *Manager) NeedsMFA(s *Session) bool {
	return !s.Verified && (s.User.MFASecret != "" || m.Production && privileged(s.User))
}
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	reply(w, status, map[string]any{"messages": []any{map[string]string{"code": code, "message": message}}})
}
func body(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		fail(w, 400, "INVALID_REQUEST", "Dados inválidos.")
		return false
	}
	if e := d.Decode(new(any)); e != io.EOF {
		fail(w, 400, "INVALID_REQUEST", "Dados inválidos.")
		return false
	}
	return true
}
func (m *Manager) sessionReply(w http.ResponseWriter, s *Session) {
	u := s.User
	reply(w, 200, map[string]any{"user": map[string]any{"profile": map[string]string{"sub": u.ID, "name": strings.TrimSpace(u.FirstName + " " + u.LastName), "email": u.Email, "preferred_username": u.Username}, "roles": u.Roles, "organization_id": u.OrganizationID}, "mfa_required": m.NeedsMFA(s), "mfa_enrolled": u.MFASecret != ""})
}
func (m *Manager) limited(ctx context.Context, bucket string, limit int) bool {
	var count int
	err := m.Store.Pool().QueryRow(ctx, `INSERT INTO auth_attempts(bucket,attempts,expires_at) VALUES($1,1,now()+interval '15 minutes') ON CONFLICT(bucket) DO UPDATE SET attempts=CASE WHEN auth_attempts.expires_at<now() THEN 1 ELSE auth_attempts.attempts+1 END, expires_at=CASE WHEN auth_attempts.expires_at<now() THEN now()+interval '15 minutes' ELSE auth_attempts.expires_at END RETURNING attempts`, digest(bucket)).Scan(&count)
	return err != nil || count > limit
}
func (m *Manager) event(ctx context.Context, id, event string) {
	m.Store.Pool().Exec(ctx, `INSERT INTO auth_events(user_id,event) VALUES($1,$2)`, id, event)
}

// Administrative recovery also clears the account's temporary attempt lockouts.
func AccountAttemptBuckets(u *store.LocalUser) []string {
	return []string{digest("login-user:" + strings.ToLower(u.Username)), digest("mfa:" + u.ID), digest("password:" + u.ID)}
}
func (m *Manager) Register(r chi.Router) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Use(middleware.Throttle(4))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-store")
				if r.Method != "GET" && !m.AllowedOrigin(r) {
					fail(w, 403, "CSRF_REJECTED", "Origem da solicitação não permitida.")
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Post("/login", m.login)
		r.Get("/session", func(w http.ResponseWriter, r *http.Request) {
			s, e := m.Resolve(r)
			if e != nil {
				fail(w, 401, "UNAUTHORIZED", "Entre novamente.")
				return
			}
			m.sessionReply(w, s)
		})
		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			if c, e := r.Cookie(m.cookieName()); e == nil {
				if _, e = m.Store.Pool().Exec(r.Context(), `DELETE FROM auth_sessions WHERE token_hash=$1`, digest(c.Value)); e != nil {
					fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
					return
				}
			}
			m.cookie(w, "", -1)
			reply(w, 200, map[string]bool{"ok": true})
		})
		r.Post("/mfa/setup", m.setup)
		r.Post("/mfa/confirm", m.confirm)
		r.Post("/mfa/verify", m.verify)
		r.Post("/password", m.password)
	})
}
func (m *Manager) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !body(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Username)
	peer, _, _ := net.SplitHostPort(r.RemoteAddr)
	if len(name) > 128 || m.limited(r.Context(), "login-user:"+strings.ToLower(name), 10) || m.limited(r.Context(), "login-peer:"+peer, 1000) {
		fail(w, 429, "LOGIN_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	u, e := m.Store.LocalUserByUsername(r.Context(), name)
	if e != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	hash := m.dummy
	if u != nil && u.PasswordHash != "" {
		hash = u.PasswordHash
	}
	valid := CheckPassword(hash, in.Password)
	if !valid || u == nil || !u.Enabled || u.PasswordHash == "" {
		m.event(r.Context(), "", "login_failed")
		fail(w, 401, "INVALID_CREDENTIALS", "Usuário ou senha inválidos.")
		return
	}
	token, e := randomToken()
	if e != nil {
		fail(w, 500, "AUTH_ERROR", "Falha ao iniciar sessão.")
		return
	}
	verified := u.MFASecret == "" && !(m.Production && privileged(u))
	// Do not resurrect a password reset, disabled account, or role change racing login.
	tag, e := m.Store.Pool().Exec(r.Context(), `INSERT INTO auth_sessions(token_hash,user_id,mfa_verified,expires_at,user_version) SELECT $1,id,$3,now()+interval '12 hours',version FROM auth_users WHERE id=$2 AND enabled AND password_hash=$4 AND roles=$5 AND mfa_secret=$6 AND version=$7`, digest(token), u.ID, verified, u.PasswordHash, u.Roles, u.MFASecret, u.Version)
	if e != nil || tag.RowsAffected() != 1 {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	m.cookie(w, token, 12*3600)
	m.event(r.Context(), u.ID, "login_password")
	m.sessionReply(w, &Session{User: u, Hash: digest(token), Verified: verified})
}
func (m *Manager) require(w http.ResponseWriter, r *http.Request) (*Session, bool) {
	s, e := m.Resolve(r)
	if e != nil {
		fail(w, 401, "UNAUTHORIZED", "Entre novamente.")
		return nil, false
	}
	return s, true
}
func (m *Manager) setup(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if s.User.MFASecret != "" {
		fail(w, 409, "MFA_ENROLLED", "Autenticador já configurado.")
		return
	}
	b := make([]byte, 20)
	if _, e := rand.Read(b); e != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	encrypted, e := seal(m.Secret, secret)
	if e != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	tag, e := m.Store.Pool().Exec(r.Context(), `UPDATE auth_sessions SET pending_secret=$2,pending_expires=now()+interval '10 minutes' WHERE token_hash=$1 AND created_at>now()-interval '10 minutes'`, s.Hash, encrypted)
	if e != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if tag.RowsAffected() != 1 {
		fail(w, 403, "REAUTH_REQUIRED", "Saia e entre novamente antes de configurar o autenticador.")
		return
	}
	uri := "otpauth://totp/" + url.PathEscape("Telefonia:"+s.User.Username) + "?" + url.Values{"secret": {secret}, "issuer": {"Telefonia"}, "algorithm": {"SHA1"}, "digits": {"6"}, "period": {"30"}}.Encode()
	png, e := qrcode.Encode(uri, qrcode.Medium, 256)
	if e != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	reply(w, 200, map[string]string{"secret": secret, "uri": uri, "qr": base64.StdEncoding.EncodeToString(png)})
}
func (m *Manager) confirm(w http.ResponseWriter, r *http.Request) { m.checkMFA(w, r, true) }
func (m *Manager) verify(w http.ResponseWriter, r *http.Request)  { m.checkMFA(w, r, false) }
func (m *Manager) checkMFA(w http.ResponseWriter, r *http.Request, enroll bool) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Code string `json:"code"`
	}
	if !body(w, r, &in) {
		return
	}
	if m.limited(r.Context(), "mfa:"+s.User.ID, 10) {
		fail(w, 429, "MFA_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	tx, e := m.Store.Pool().Begin(r.Context())
	if e != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	defer tx.Rollback(r.Context())
	var encrypted string
	var last int64
	e = tx.QueryRow(r.Context(), `SELECT mfa_secret,mfa_last_step FROM auth_users WHERE id=$1 AND enabled AND version=$2 FOR UPDATE`, s.User.ID, s.User.Version).Scan(&encrypted, &last)
	if e != nil {
		fail(w, 401, "UNAUTHORIZED", "Entre novamente.")
		return
	}
	if enroll {
		if encrypted != "" {
			fail(w, 409, "MFA_ENROLLED", "Autenticador já configurado.")
			return
		}
		e = tx.QueryRow(r.Context(), `SELECT pending_secret FROM auth_sessions WHERE token_hash=$1 AND pending_expires>now()`, s.Hash).Scan(&encrypted)
		if e != nil || encrypted == "" {
			fail(w, 400, "MFA_SETUP_EXPIRED", "Inicie novamente a configuração.")
			return
		}
	}
	secret, e := unseal(m.Secret, encrypted)
	if e != nil {
		fail(w, 503, "MFA_UNAVAILABLE", "Não foi possível ler o autenticador. Contate o administrador.")
		return
	}
	step, valid := verifyTOTP(secret, strings.TrimSpace(in.Code), time.Now(), last)
	if !valid && !enroll {
		result, err := tx.Exec(r.Context(), `DELETE FROM auth_recovery_codes WHERE user_id=$1 AND code_hash=$2`, s.User.ID, recoveryCodeHash(in.Code))
		if err == nil && result.RowsAffected() == 1 {
			valid = true
			step = last
		}
	}
	if !valid {
		fail(w, 401, "INVALID_OTP", "Código inválido, expirado ou já utilizado.")
		return
	}
	codes := []string{}
	if enroll {
		for i := 0; i < 8; i++ {
			code, err := newRecoveryCode()
			if err != nil {
				fail(w, 500, "AUTH_ERROR", "Tente novamente.")
				return
			}
			codes = append(codes, code)
			if _, e = tx.Exec(r.Context(), `INSERT INTO auth_recovery_codes(user_id,code_hash) VALUES($1,$2)`, s.User.ID, recoveryCodeHash(code)); e != nil {
				fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
				return
			}
		}
	}
	var version int64
	e = tx.QueryRow(r.Context(), `UPDATE auth_users SET mfa_secret=$2,mfa_last_step=$3,version=version+CASE WHEN $4 THEN 1 ELSE 0 END WHERE id=$1 RETURNING version`, s.User.ID, encrypted, step, enroll).Scan(&version)
	if e == nil && enroll {
		_, e = tx.Exec(r.Context(), `DELETE FROM auth_sessions WHERE user_id=$1 AND token_hash<>$2`, s.User.ID, s.Hash)
	}
	// Rotate the session token on elevation; the old cookie immediately loses access.
	token, tokenErr := randomToken()
	if tokenErr != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	var result interface{ RowsAffected() int64 }
	if e == nil {
		result, e = tx.Exec(r.Context(), `UPDATE auth_sessions SET token_hash=$2,mfa_verified=true,pending_secret='',pending_expires=NULL,user_version=$3 WHERE token_hash=$1 AND expires_at>now()`, s.Hash, digest(token), version)
		if e == nil && result.RowsAffected() != 1 {
			e = pgx.ErrNoRows
		}
	}
	if e != nil || tx.Commit(r.Context()) != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	m.cookie(w, token, 12*3600)
	m.event(r.Context(), s.User.ID, "mfa_verified")
	reply(w, 200, map[string]any{"ok": true, "recovery_codes": codes})
}
func (m *Manager) password(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.NeedsMFA(s) {
		fail(w, 403, "MFA_REQUIRED", "Confirme o autenticador.")
		return
	}
	var in struct {
		Current string `json:"current_password"`
		New     string `json:"new_password"`
	}
	if !body(w, r, &in) {
		return
	}
	if m.limited(r.Context(), "password:"+s.User.ID, 10) {
		fail(w, 429, "LOGIN_LIMIT", "Aguarde 15 minutos.")
		return
	}
	if !CheckPassword(s.User.PasswordHash, in.Current) {
		fail(w, 401, "INVALID_CREDENTIALS", "Senha atual incorreta.")
		return
	}
	hash, e := HashPassword(in.New)
	if e != nil {
		fail(w, 400, "PASSWORD_POLICY", e.Error())
		return
	}
	s.User.PasswordHash = hash
	if e = m.Store.UpdateLocalUser(r.Context(), s.User); e != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	m.event(r.Context(), s.User.ID, "password_changed")
	m.cookie(w, "", -1)
	reply(w, 200, map[string]bool{"ok": true})
}
