package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/config"
)

type AdminClient struct {
	baseURL           string
	realm             string
	adminUser         string
	adminPass         string
	httpClient        *http.Client
	tokenMu           sync.Mutex
	accessToken       string
	tokenExpires      time.Time
	clientMu          sync.Mutex
	connectClientName string
	connectClientID   string
	sessionCache      *sessionUserCache
}

func NewAdminClient(cfg config.Config) *AdminClient {
	return &AdminClient{
		baseURL:      strings.TrimRight(cfg.KeycloakAuthServerURL, "/"),
		realm:        cfg.KeycloakRealm,
		adminUser:    cfg.KeycloakAdminUsername,
		adminPass:    cfg.KeycloakAdminPassword,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
		sessionCache: &sessionUserCache{entries: map[string]sessionUserCacheEntry{}},
	}
}

func (c *AdminClient) Enabled() bool {
	return c.baseURL != "" && c.realm != "" && c.adminPass != ""
}

func (c *AdminClient) AccountSecurityURL() string {
	if c.baseURL == "" || c.realm == "" {
		return ""
	}
	return strings.TrimRight(c.baseURL, "/") + "/realms/" + c.realm + "/account/#/security/signingin"
}

type UserRecord struct {
	ID         string              `json:"id"`
	Username   string              `json:"username"`
	Email      string              `json:"email"`
	FirstName  string              `json:"firstName"`
	LastName   string              `json:"lastName"`
	Enabled    bool                `json:"enabled"`
	Attributes map[string][]string `json:"attributes,omitempty"`
	Roles      []string            `json:"-"`
}

type CreateUserPayload struct {
	Username      string              `json:"username"`
	Email         string              `json:"email"`
	FirstName     string              `json:"firstName"`
	LastName      string              `json:"lastName"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	Attributes    map[string][]string `json:"attributes,omitempty"`
	Credentials   []CredentialPayload `json:"credentials"`
}

type CredentialPayload struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

type roleRepresentation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *AdminClient) token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpires.Add(-30*time.Second)) {
		return c.accessToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", "admin-cli")
	form.Set("username", c.adminUser)
	form.Set("password", c.adminPass)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("keycloak admin token: %s", string(body))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	c.accessToken = parsed.AccessToken
	c.tokenExpires = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *AdminClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

func (c *AdminClient) ListUsers(ctx context.Context, search string, max int) ([]UserRecord, error) {
	if max <= 0 {
		max = 100
	}
	// briefRepresentation=false is required so Keycloak returns user attributes
	// (organization / organization_id) used for tenant isolation in the UI.
	path := fmt.Sprintf("/admin/realms/%s/users?max=%d&briefRepresentation=false", c.realm, max)
	if strings.TrimSpace(search) != "" {
		path += "&search=" + url.QueryEscape(strings.TrimSpace(search))
	}

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list users: %s", string(body))
	}

	var users []UserRecord
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	for i := range users {
		roles, err := c.GetUserRealmRoles(ctx, users[i].ID)
		if err != nil {
			return nil, err
		}
		users[i].Roles = roles
	}
	return users, nil
}

func (c *AdminClient) GetUserRealmRoles(ctx context.Context, userID string) ([]string, error) {
	path := fmt.Sprintf("/admin/realms/%s/users/%s/role-mappings/realm", c.realm, userID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user roles: %s", string(body))
	}

	var roles []roleRepresentation
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, r.Name)
	}
	return out, nil
}

func (c *AdminClient) CreateUser(ctx context.Context, payload CreateUserPayload) (string, error) {
	path := fmt.Sprintf("/admin/realms/%s/users", c.realm)
	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", fmt.Errorf("username already exists")
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create user: %s", string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("create user: missing location header")
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	return parts[len(parts)-1], nil
}

func (c *AdminClient) SetUserEnabled(ctx context.Context, userID string, enabled bool) error {
	current, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	return c.putUserRepresentation(ctx, userID, map[string]any{
		"username":      current.Username,
		"email":         current.Email,
		"firstName":     current.FirstName,
		"lastName":      current.LastName,
		"enabled":       enabled,
		"emailVerified": true,
		"attributes":    current.Attributes,
	})
}

func (c *AdminClient) UpdateUserProfile(ctx context.Context, userID, firstName, lastName, email string) error {
	current, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"username":      current.Username,
		"firstName":     firstName,
		"lastName":      lastName,
		"enabled":       current.Enabled,
		"emailVerified": true,
		"attributes":    current.Attributes,
	}
	if strings.TrimSpace(email) != "" {
		payload["email"] = strings.TrimSpace(email)
	} else {
		payload["email"] = current.Email
	}
	return c.putUserRepresentation(ctx, userID, payload)
}

func (c *AdminClient) putUserRepresentation(ctx context.Context, userID string, payload map[string]any) error {
	// Keycloak user profile may require non-empty names on update.
	if v, _ := payload["firstName"].(string); strings.TrimSpace(v) == "" {
		payload["firstName"] = "-"
	}
	if v, _ := payload["lastName"].(string); strings.TrimSpace(v) == "" {
		payload["lastName"] = "-"
	}
	if payload["attributes"] == nil {
		payload["attributes"] = map[string][]string{}
	}
	payload["id"] = userID

	path := fmt.Sprintf("/admin/realms/%s/users/%s", c.realm, userID)
	resp, err := c.do(ctx, http.MethodPut, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update user: %s", string(body))
	}
	return nil
}

func (c *AdminClient) GetUserByID(ctx context.Context, userID string) (*UserRecord, error) {
	path := fmt.Sprintf("/admin/realms/%s/users/%s", c.realm, userID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user %s: %s", userID, string(body))
	}
	var u UserRecord
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	roles, err := c.GetUserRealmRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return &u, nil
}

type credentialRepresentation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *AdminClient) UserHasOTP(ctx context.Context, userID string) (bool, error) {
	path := fmt.Sprintf("/admin/realms/%s/users/%s/credentials", c.realm, userID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("list credentials: %s", string(body))
	}
	var creds []credentialRepresentation
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return false, err
	}
	for _, cred := range creds {
		switch strings.ToLower(cred.Type) {
		case "otp", "totp":
			return true, nil
		}
	}
	return false, nil
}

func (c *AdminClient) ReplaceUserRealmRoles(ctx context.Context, userID string, roleNames []string) error {
	current, err := c.GetUserRealmRoles(ctx, userID)
	if err != nil {
		return err
	}

	manageable := map[string]struct{}{
		"master": {}, "admin": {}, "employee": {}, "financial": {}, "partner": {},
		"operator": {}, "sales": {}, "viewer": {}, "user": {},
	}
	var toRemove []roleRepresentation
	for _, name := range current {
		if _, ok := manageable[name]; ok {
			toRemove = append(toRemove, roleRepresentation{Name: name})
		}
	}
	if len(toRemove) > 0 {
		path := fmt.Sprintf("/admin/realms/%s/users/%s/role-mappings/realm", c.realm, userID)
		resp, err := c.do(ctx, http.MethodDelete, path, toRemove)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("remove roles: status %d", resp.StatusCode)
		}
	}

	var toAdd []roleRepresentation
	for _, name := range roleNames {
		role, err := c.getRealmRole(ctx, name)
		if err != nil {
			return err
		}
		toAdd = append(toAdd, *role)
	}
	if len(toAdd) == 0 {
		return nil
	}

	path := fmt.Sprintf("/admin/realms/%s/users/%s/role-mappings/realm", c.realm, userID)
	resp, err := c.do(ctx, http.MethodPost, path, toAdd)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("assign roles: %s", string(body))
	}
	return nil
}

func (c *AdminClient) getRealmRole(ctx context.Context, name string) (*roleRepresentation, error) {
	path := fmt.Sprintf("/admin/realms/%s/roles/%s", c.realm, name)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get role %s: %s", name, string(body))
	}
	var role roleRepresentation
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (c *AdminClient) ResetPassword(ctx context.Context, userID, password string, temporary bool) error {
	path := fmt.Sprintf("/admin/realms/%s/users/%s/reset-password", c.realm, userID)
	resp, err := c.do(ctx, http.MethodPut, path, CredentialPayload{
		Type: "password", Value: password, Temporary: temporary,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("reset password: %s", string(body))
	}
	return nil
}

func DefaultOrganizationAttribute(orgID, orgName string) map[string][]string {
	orgID = strings.TrimSpace(orgID)
	orgName = strings.TrimSpace(orgName)
	alias := organizationAlias(orgID, orgName)
	payload := map[string]any{
		alias: map[string]any{
			"id":   orgID,
			"name": []string{orgName},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		// Extremely unlikely; keep a minimal fallback without string interpolation risks.
		raw = []byte(`{"org":{"id":"` + orgID + `","name":[""]}}`)
	}
	return map[string][]string{
		"organization":    {string(raw)},
		"organization_id": {orgID},
	}
<<<<<<< HEAD
}

// organizationAlias builds a stable claim key. Never use a raw UUID as the only
// identity signal without nested id — always embed orgID in the JSON value.
func organizationAlias(orgID, orgName string) string {
	const luxusID = "00000000-0000-0000-0000-000000000001"
	if orgID == luxusID {
		return "luxus"
	}
	alias := slugifyOrgName(orgName)
	if alias == "" || alias == "luxus" || strings.HasPrefix(alias, "luxus") {
		compact := strings.ReplaceAll(orgID, "-", "")
		if len(compact) >= 8 {
			return "org-" + compact[:8]
		}
		return "org"
	}
	return alias
}

func slugifyOrgName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 48 {
		out = out[:48]
		out = strings.Trim(out, "-")
	}
	return out
=======
	payload := map[string]map[string]any{
		alias: {"id": strings.TrimSpace(orgID), "name": []string{strings.TrimSpace(orgName)}},
	}
	rawBytes, _ := json.Marshal(payload)
	raw := string(rawBytes)
	return map[string][]string{"organization": {raw}}
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)
}
