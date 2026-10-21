package localauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

type Manager struct {
	Store      *store.Store
	Secret     string
	Production bool
	Origins    []string
	RPID       string
	RPName     string
	wa         *webauthn.WebAuthn
	dummy      string
}
type Session struct {
	User     *store.LocalUser
	Hash     string
	Verified bool
}

// WebAuthn requires a stable RP ID and every accepted frontend origin.
func New(st *store.Store, secret string, production bool, origins []string, rpID, rpName string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("AUTH_SECRET must contain at least 32 characters")
	}
	dummy, e := HashPassword("dummy-never-valid-password")
	if e != nil {
		return nil, e
	}
	wa, e := webauthn.New(&webauthn.Config{RPID: rpID, RPDisplayName: rpName, RPOrigins: origins})
	if e != nil {
		return nil, fmt.Errorf("webauthn: %w", e)
	}
	return &Manager{Store: st, Secret: secret, Production: production, Origins: origins, RPID: rpID, RPName: rpName, wa: wa, dummy: dummy}, nil
}

// webauthnUser adapts a local account and its stored credentials to the library.
type webauthnUser struct {
	user        *store.LocalUser
	credentials []webauthn.Credential
}

func (u webauthnUser) WebAuthnID() []byte { return []byte(u.user.ID) }
func (u webauthnUser) WebAuthnName() string {
	return u.user.Username
}
func (u webauthnUser) WebAuthnDisplayName() string {
	if name := strings.TrimSpace(u.user.FirstName + " " + u.user.LastName); name != "" {
		return name
	}
	return u.user.Username
}
func (u webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

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
				m.Store.Pool().Exec(cleanup, `DELETE FROM auth_webauthn_challenges WHERE expires_at<now()`)
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
func (m *Manager) NeedsMFA(s *Session) bool {
	return !s.Verified && (m.Production || s.User.PasskeyCount > 0)
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
	reply(w, 200, map[string]any{
		"user":          map[string]any{"profile": map[string]string{"sub": u.ID, "name": strings.TrimSpace(u.FirstName + " " + u.LastName), "email": u.Email, "preferred_username": u.Username}, "roles": u.Roles, "organization_id": u.OrganizationID},
		"mfa_required":  m.NeedsMFA(s),
		"mfa_enrolled":  u.PasskeyCount > 0,
		"passkey_count": u.PasskeyCount,
		"mfa_method":    "passkey",
	})
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
	return []string{
		digest("login-user:" + strings.ToLower(u.Username)),
		digest("mfa:" + u.ID),
		digest("passkey-register:" + u.ID),
		digest("passkey-login:" + u.ID),
		digest("recovery:" + u.ID),
		digest("password:" + u.ID),
	}
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
		r.Post("/webauthn/register/begin", m.registerBegin)
		r.Post("/webauthn/register/finish", m.registerFinish)
		r.Post("/webauthn/login/begin", m.loginBegin)
		r.Post("/webauthn/login/finish", m.loginFinish)
		r.Get("/webauthn/credentials", m.listCredentials)
		r.Delete("/webauthn/credentials/{id}", m.deleteCredential)
		r.Post("/recovery/verify", m.recoveryVerify)
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
	// Production always requires a passkey (or a recovery code) to elevate.
	verified := !m.Production && u.PasskeyCount == 0
	// Do not resurrect a password reset, disabled account, or role change racing login.
	tag, e := m.Store.Pool().Exec(r.Context(), `INSERT INTO auth_sessions(token_hash,user_id,mfa_verified,expires_at,user_version) SELECT $1,id,$3,now()+interval '12 hours',version FROM auth_users WHERE id=$2 AND enabled AND password_hash=$4 AND roles=$5 AND version=$6`, digest(token), u.ID, verified, u.PasswordHash, u.Roles, u.Version)
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

// elevateTx rotates the token, marks the session verified, and revokes the
// user's other sessions in the same transaction as the caller's writes.
func (m *Manager) elevateTx(ctx context.Context, tx pgx.Tx, s *Session, event string) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	tag, err := tx.Exec(ctx, `UPDATE auth_sessions SET token_hash=$2,mfa_verified=true,user_version=$3,last_seen=now() WHERE token_hash=$1 AND expires_at>now()`, s.Hash, digest(token), s.User.Version)
	if err != nil {
		return "", err
	}
	if tag.RowsAffected() != 1 {
		return "", pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `DELETE FROM auth_sessions WHERE user_id=$1 AND token_hash<>$2`, s.User.ID, digest(token)); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO auth_events(user_id,event) VALUES($1,$2)`, s.User.ID, event); err != nil {
		return "", err
	}
	return token, nil
}

func (m *Manager) markVerified(w http.ResponseWriter, r *http.Request, s *Session, event string) bool {
	tx, err := m.Store.Pool().Begin(r.Context())
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return false
	}
	defer tx.Rollback(r.Context())
	token, err := m.elevateTx(r.Context(), tx, s, event)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return false
	}
	if err = tx.Commit(r.Context()); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return false
	}
	m.cookie(w, token, 12*3600)
	s.Hash = digest(token)
	s.Verified = true
	return true
}

func (m *Manager) loadCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	// Credentials are scoped to the RP ID so a hostname change cannot reuse them.
	rows, err := m.Store.Pool().Query(ctx, `SELECT credential FROM auth_webauthn_credentials WHERE user_id=$1 AND rp_id=$2 ORDER BY created_at`, userID, m.RPID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []webauthn.Credential{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		var credential webauthn.Credential
		if err = json.Unmarshal(raw, &credential); err != nil {
			return nil, err
		}
		out = append(out, credential)
	}
	return out, rows.Err()
}

// Challenges are stored server-side, keyed by the session token hash and kind,
// and are single-use.
func (m *Manager) saveChallenge(ctx context.Context, sessionHash, kind, userID string, data *webauthn.SessionData) error {
	expires := data.Expires
	if expires.IsZero() {
		expires = time.Now().Add(5 * time.Minute)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = m.Store.Pool().Exec(ctx, `INSERT INTO auth_webauthn_challenges(session_hash,kind,user_id,session_data,expires_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(session_hash,kind) DO UPDATE SET user_id=EXCLUDED.user_id,session_data=EXCLUDED.session_data,expires_at=EXCLUDED.expires_at`, sessionHash, kind, userID, raw, expires)
	return err
}
func (m *Manager) consumeChallenge(ctx context.Context, sessionHash, kind string) (*webauthn.SessionData, error) {
	var raw []byte
	if err := m.Store.Pool().QueryRow(ctx, `DELETE FROM auth_webauthn_challenges WHERE session_hash=$1 AND kind=$2 AND expires_at>now() RETURNING session_data`, sessionHash, kind).Scan(&raw); err != nil {
		return nil, err
	}
	var data webauthn.SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (m *Manager) registerBegin(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.limited(r.Context(), "passkey-register:"+s.User.ID, 10) {
		fail(w, 429, "PASSKEY_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	// A password-only session may only enroll its first passkey.
	if !s.Verified && s.User.PasskeyCount > 0 {
		fail(w, 409, "PASSKEY_ALREADY_ENROLLED", "Confirme sua passkey atual antes de cadastrar outra.")
		return
	}
	credentials, err := m.loadCredentials(r.Context(), s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	user := webauthnUser{user: s.User, credentials: credentials}
	creation, session, err := m.wa.BeginRegistration(user, webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
		ResidentKey:      protocol.ResidentKeyRequirementDiscouraged,
		UserVerification: protocol.VerificationRequired,
	}))
	if err != nil {
		fail(w, 400, "PASSKEY_REGISTRATION_FAILED", "Não foi possível cadastrar a passkey. Tente novamente.")
		return
	}
	if err = m.saveChallenge(r.Context(), s.Hash, "register", s.User.ID, session); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	reply(w, 200, map[string]any{"options": creation})
}

func (m *Manager) registerFinish(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.limited(r.Context(), "passkey-register:"+s.User.ID, 10) {
		fail(w, 429, "PASSKEY_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	// Re-check at finish time: a password-only session must never add a second
	// credential if one was enrolled concurrently after BeginRegistration.
	if !s.Verified && s.User.PasskeyCount > 0 {
		fail(w, 409, "PASSKEY_ALREADY_ENROLLED", "Confirme sua passkey atual antes de cadastrar outra.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	session, err := m.consumeChallenge(r.Context(), s.Hash, "register")
	if err != nil {
		fail(w, 400, "PASSKEY_REGISTRATION_FAILED", "Não foi possível cadastrar a passkey. Tente novamente.")
		return
	}
	credentials, err := m.loadCredentials(r.Context(), s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	user := webauthnUser{user: s.User, credentials: credentials}
	credential, err := m.wa.FinishRegistration(user, *session, r)
	if err != nil {
		fail(w, 400, "PASSKEY_REGISTRATION_FAILED", "Não foi possível cadastrar a passkey. Tente novamente.")
		return
	}
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	raw, err := json.Marshal(credential)
	if err != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	tx, err := m.Store.Pool().Begin(r.Context())
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	defer tx.Rollback(r.Context())
	// Serialize first-time enrollment per user so two concurrent password-only
	// sessions cannot each register a different passkey.
	if _, err = tx.Exec(r.Context(), `SELECT pg_advisory_xact_lock(hashtext($1))`, "webauthn-register:"+s.User.ID); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if !s.Verified {
		var existing int
		if err = tx.QueryRow(r.Context(), `SELECT count(*)::int FROM auth_webauthn_credentials WHERE user_id=$1 AND rp_id=$2`, s.User.ID, m.RPID).Scan(&existing); err != nil {
			fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
			return
		}
		if existing > 0 {
			fail(w, 409, "PASSKEY_ALREADY_ENROLLED", "Confirme sua passkey atual antes de cadastrar outra.")
			return
		}
	}
	tag, err := tx.Exec(r.Context(), `INSERT INTO auth_webauthn_credentials(credential_id,user_id,rp_id,name,credential) VALUES($1,$2,$3,$4,$5) ON CONFLICT (credential_id) DO NOTHING`, credentialID, s.User.ID, m.RPID, "Passkey", raw)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if tag.RowsAffected() == 0 {
		fail(w, 409, "PASSKEY_ALREADY_ENROLLED", "Esta passkey já está cadastrada.")
		return
	}
	var count int
	if err = tx.QueryRow(r.Context(), `SELECT count(*)::int FROM auth_webauthn_credentials WHERE user_id=$1 AND rp_id=$2`, s.User.ID, m.RPID).Scan(&count); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	codes := []string{}
	if count == 1 {
		for i := 0; i < 8; i++ {
			code, err := newRecoveryCode()
			if err != nil {
				fail(w, 500, "AUTH_ERROR", "Tente novamente.")
				return
			}
			codes = append(codes, code)
			if _, err = tx.Exec(r.Context(), `INSERT INTO auth_recovery_codes(user_id,code_hash) VALUES($1,$2)`, s.User.ID, recoveryCodeHash(code)); err != nil {
				fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
				return
			}
		}
	}
	token, err := m.elevateTx(r.Context(), tx, s, "passkey_registered")
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	m.cookie(w, token, 12*3600)
	s.Hash = digest(token)
	s.Verified = true
	reply(w, 200, map[string]any{"ok": true, "recovery_codes": codes})
}

func (m *Manager) loginBegin(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.limited(r.Context(), "passkey-login:"+s.User.ID, 20) {
		fail(w, 429, "PASSKEY_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	if s.Verified {
		fail(w, 409, "ALREADY_VERIFIED", "A sessão já está verificada.")
		return
	}
	if s.User.PasskeyCount == 0 {
		fail(w, 409, "PASSKEY_NOT_ENROLLED", "Cadastre uma passkey para continuar.")
		return
	}
	credentials, err := m.loadCredentials(r.Context(), s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	user := webauthnUser{user: s.User, credentials: credentials}
	assertion, session, err := m.wa.BeginLogin(user, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		fail(w, 400, "PASSKEY_LOGIN_FAILED", "Não foi possível validar a passkey.")
		return
	}
	if err = m.saveChallenge(r.Context(), s.Hash, "login", s.User.ID, session); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	reply(w, 200, map[string]any{"options": assertion})
}

func (m *Manager) loginFinish(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if s.Verified {
		fail(w, 409, "ALREADY_VERIFIED", "A sessão já está verificada.")
		return
	}
	if m.limited(r.Context(), "passkey-login:"+s.User.ID, 20) {
		fail(w, 429, "PASSKEY_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	session, err := m.consumeChallenge(r.Context(), s.Hash, "login")
	if err != nil {
		fail(w, 401, "PASSKEY_LOGIN_FAILED", "Não foi possível validar a passkey.")
		return
	}
	credentials, err := m.loadCredentials(r.Context(), s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	user := webauthnUser{user: s.User, credentials: credentials}
	credential, err := m.wa.FinishLogin(user, *session, r)
	if err != nil {
		fail(w, 401, "PASSKEY_LOGIN_FAILED", "Não foi possível validar a passkey.")
		return
	}
	raw, err := json.Marshal(credential)
	if err != nil {
		fail(w, 500, "AUTH_ERROR", "Tente novamente.")
		return
	}
	tag, err := m.Store.Pool().Exec(r.Context(), `UPDATE auth_webauthn_credentials SET credential=$3,last_used_at=now() WHERE credential_id=$1 AND user_id=$2`, base64.RawURLEncoding.EncodeToString(credential.ID), s.User.ID, raw)
	if err != nil || tag.RowsAffected() != 1 {
		fail(w, 401, "PASSKEY_LOGIN_FAILED", "Não foi possível validar a passkey.")
		return
	}
	if !m.markVerified(w, r, s, "passkey_verified") {
		return
	}
	m.sessionReply(w, s)
}

func (m *Manager) recoveryVerify(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if s.Verified {
		fail(w, 409, "ALREADY_VERIFIED", "A sessão já está verificada.")
		return
	}
	if m.limited(r.Context(), "recovery:"+s.User.ID, 10) {
		fail(w, 429, "RECOVERY_LIMIT", "Muitas tentativas. Aguarde 15 minutos.")
		return
	}
	var in struct {
		Code string `json:"code"`
	}
	if !body(w, r, &in) {
		return
	}
	tag, err := m.Store.Pool().Exec(r.Context(), `DELETE FROM auth_recovery_codes WHERE user_id=$1 AND code_hash=$2`, s.User.ID, recoveryCodeHash(in.Code))
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if tag.RowsAffected() != 1 {
		fail(w, 401, "INVALID_RECOVERY_CODE", "Código de recuperação inválido ou já utilizado.")
		return
	}
	if !m.markVerified(w, r, s, "recovery_verified") {
		return
	}
	m.sessionReply(w, s)
}

type credentialInfo struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func (m *Manager) listCredentials(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.NeedsMFA(s) {
		fail(w, 403, "MFA_REQUIRED", "Confirme sua identidade com uma passkey na página de segurança da conta e entre novamente.")
		return
	}
	rows, err := m.Store.Pool().Query(r.Context(), `SELECT credential_id,name,created_at,last_used_at FROM auth_webauthn_credentials WHERE user_id=$1 ORDER BY created_at`, s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	defer rows.Close()
	out := []credentialInfo{}
	for rows.Next() {
		var item credentialInfo
		if err = rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.LastUsedAt); err != nil {
			fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
			return
		}
		out = append(out, item)
	}
	if err = rows.Err(); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	reply(w, 200, map[string]any{"credentials": out})
}

func (m *Manager) deleteCredential(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.NeedsMFA(s) {
		fail(w, 403, "MFA_REQUIRED", "Confirme sua identidade com uma passkey na página de segurança da conta e entre novamente.")
		return
	}
	id := chi.URLParam(r, "id")
	tx, err := m.Store.Pool().Begin(r.Context())
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	defer tx.Rollback(r.Context())
	// Lock the user's credentials so two concurrent deletes cannot remove the last one.
	rows, err := tx.Query(r.Context(), `SELECT credential_id FROM auth_webauthn_credentials WHERE user_id=$1 FOR UPDATE`, s.User.ID)
	if err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	count := 0
	exists := false
	for rows.Next() {
		var credentialID string
		if err = rows.Scan(&credentialID); err != nil {
			rows.Close()
			fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
			return
		}
		count++
		if credentialID == id {
			exists = true
		}
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if !exists {
		fail(w, 404, "PASSKEY_NOT_FOUND", "Passkey não encontrada.")
		return
	}
	if m.Production && count <= 1 {
		fail(w, 409, "LAST_PASSKEY", "Cadastre outra passkey antes de remover a última.")
		return
	}
	if _, err = tx.Exec(r.Context(), `DELETE FROM auth_webauthn_credentials WHERE credential_id=$1 AND user_id=$2`, id, s.User.ID); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO auth_events(user_id,event) VALUES($1,'passkey_removed')`, s.User.ID); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		fail(w, 503, "AUTH_UNAVAILABLE", "Tente novamente.")
		return
	}
	reply(w, 200, map[string]bool{"ok": true})
}

func (m *Manager) password(w http.ResponseWriter, r *http.Request) {
	s, ok := m.require(w, r)
	if !ok {
		return
	}
	if m.NeedsMFA(s) {
		fail(w, 403, "MFA_REQUIRED", "Confirme sua identidade com uma passkey antes de alterar a senha.")
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
