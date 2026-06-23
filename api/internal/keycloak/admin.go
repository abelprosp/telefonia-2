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
	baseURL      string
	realm        string
	adminUser    string
	adminPass    string
	httpClient   *http.Client
	tokenMu      sync.Mutex
	accessToken  string
	tokenExpires time.Time
	clientMu         sync.Mutex
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

type UserRecord struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Enabled   bool     `json:"enabled"`
	Roles     []string `json:"-"`
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
	path := fmt.Sprintf("/admin/realms/%s/users?max=%d", c.realm, max)
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
	path := fmt.Sprintf("/admin/realms/%s/users/%s", c.realm, userID)
	resp, err := c.do(ctx, http.MethodPut, path, map[string]any{"enabled": enabled})
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

func (c *AdminClient) ReplaceUserRealmRoles(ctx context.Context, userID string, roleNames []string) error {
	current, err := c.GetUserRealmRoles(ctx, userID)
	if err != nil {
		return err
	}

	manageable := map[string]struct{}{
		"master": {}, "admin": {}, "employee": {}, "financial": {}, "partner": {},
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
	raw := fmt.Sprintf(`{"luxus":{"id":"%s","name":["%s"]}}`, orgID, orgName)
	return map[string][]string{"organization": {raw}}
}
