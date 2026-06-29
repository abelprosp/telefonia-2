package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
)

// keycloakSetupClient is a lightweight, self-contained Keycloak admin client
// used exclusively by the /setup/keycloak endpoint. It authenticates against
// the master realm and operates on the target realm directly via the Admin REST API.
type keycloakSetupClient struct {
	baseURL    string
	adminUser  string
	adminPass  string
	httpClient *http.Client
}

func newKeycloakSetupClient(baseURL, adminUser, adminPass string) *keycloakSetupClient {
	return &keycloakSetupClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		adminUser:  adminUser,
		adminPass:  adminPass,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *keycloakSetupClient) getAdminToken(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", "admin-cli")
	form.Set("username", c.adminUser)
	form.Set("password", c.adminPass)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("keycloak admin token (status %d): %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}
	return parsed.AccessToken, nil
}

func (c *keycloakSetupClient) do(ctx context.Context, token, method, path string, payload any) (*http.Response, error) {
	var reader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

// ensureRealm creates the realm if it does not already exist (409 → idempotent).
func (c *keycloakSetupClient) ensureRealm(ctx context.Context, token, realmName string) error {
	payload := map[string]any{
		"realm":   realmName,
		"enabled": true,
		"sslRequired": "none",
		"registrationAllowed": false,
		"loginWithEmailAllowed": true,
		"duplicateEmailsAllowed": false,
		"resetPasswordAllowed": true,
		"editUsernameAllowed": false,
		"bruteForceProtected": true,
	}
	resp, err := c.do(ctx, token, http.MethodPost, "/admin/realms", payload)
	if err != nil {
		return fmt.Errorf("create realm: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return nil // already exists
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create realm (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// ensureRealmRole creates a realm role if it does not already exist.
// Returns the role ID.
func (c *keycloakSetupClient) ensureRealmRole(ctx context.Context, token, realm, roleName string) (string, error) {
	// Try to get existing role first.
	getResp, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/roles/%s", realm, url.PathEscape(roleName)), nil)
	if err != nil {
		return "", fmt.Errorf("get role %s: %w", roleName, err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode == http.StatusOK {
		var role struct {
			ID string `json:"id"`
		}
		body, _ := io.ReadAll(getResp.Body)
		if err := json.Unmarshal(body, &role); err != nil {
			return "", fmt.Errorf("parse role %s: %w", roleName, err)
		}
		return role.ID, nil
	}

	// Role does not exist — create it.
	createResp, err := c.do(ctx, token, http.MethodPost,
		fmt.Sprintf("/admin/realms/%s/roles", realm),
		map[string]any{"name": roleName})
	if err != nil {
		return "", fmt.Errorf("create role %s: %w", roleName, err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode >= 300 && createResp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(createResp.Body)
		return "", fmt.Errorf("create role %s (status %d): %s", roleName, createResp.StatusCode, string(body))
	}

	// Fetch the newly created role to get its ID.
	getResp2, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/roles/%s", realm, url.PathEscape(roleName)), nil)
	if err != nil {
		return "", fmt.Errorf("get role %s after create: %w", roleName, err)
	}
	defer getResp2.Body.Close()
	var role struct {
		ID string `json:"id"`
	}
	body, _ := io.ReadAll(getResp2.Body)
	if err := json.Unmarshal(body, &role); err != nil {
		return "", fmt.Errorf("parse role %s after create: %w", roleName, err)
	}
	return role.ID, nil
}

// ensureClientScope creates a client scope if it does not already exist.
// Returns the scope ID.
func (c *keycloakSetupClient) ensureClientScope(ctx context.Context, token, realm, scopeName, protocol string) (string, error) {
	listResp, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/client-scopes", realm), nil)
	if err != nil {
		return "", fmt.Errorf("list client scopes: %w", err)
	}
	defer listResp.Body.Close()

	var scopes []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	body, _ := io.ReadAll(listResp.Body)
	if err := json.Unmarshal(body, &scopes); err != nil {
		return "", fmt.Errorf("parse client scopes: %w", err)
	}
	for _, s := range scopes {
		if s.Name == scopeName {
			return s.ID, nil
		}
	}

	// Create the scope.
	payload := map[string]any{
		"name":     scopeName,
		"protocol": protocol,
	}
	createResp, err := c.do(ctx, token, http.MethodPost,
		fmt.Sprintf("/admin/realms/%s/client-scopes", realm), payload)
	if err != nil {
		return "", fmt.Errorf("create client scope %s: %w", scopeName, err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode >= 300 {
		b, _ := io.ReadAll(createResp.Body)
		return "", fmt.Errorf("create client scope %s (status %d): %s", scopeName, createResp.StatusCode, string(b))
	}

	// Re-fetch to get the ID.
	listResp2, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/client-scopes", realm), nil)
	if err != nil {
		return "", fmt.Errorf("list client scopes after create: %w", err)
	}
	defer listResp2.Body.Close()
	body2, _ := io.ReadAll(listResp2.Body)
	var scopes2 []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body2, &scopes2); err != nil {
		return "", fmt.Errorf("parse client scopes after create: %w", err)
	}
	for _, s := range scopes2 {
		if s.Name == scopeName {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("client scope %s not found after creation", scopeName)
}

// ensureClient creates the OIDC client if it does not already exist.
func (c *keycloakSetupClient) ensureClient(ctx context.Context, token, realm string, clientPayload map[string]any) error {
	clientID, _ := clientPayload["clientId"].(string)

	listResp, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/clients?clientId=%s", realm, url.QueryEscape(clientID)), nil)
	if err != nil {
		return fmt.Errorf("list clients: %w", err)
	}
	defer listResp.Body.Close()

	var clients []struct {
		ID string `json:"id"`
	}
	body, _ := io.ReadAll(listResp.Body)
	if err := json.Unmarshal(body, &clients); err != nil {
		return fmt.Errorf("parse clients: %w", err)
	}
	if len(clients) > 0 {
		return nil // already exists
	}

	createResp, err := c.do(ctx, token, http.MethodPost,
		fmt.Sprintf("/admin/realms/%s/clients", realm), clientPayload)
	if err != nil {
		return fmt.Errorf("create client %s: %w", clientID, err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode >= 300 && createResp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create client %s (status %d): %s", clientID, createResp.StatusCode, string(b))
	}
	return nil
}

// ensureUser creates a user if it does not already exist, then assigns realm roles.
func (c *keycloakSetupClient) ensureUser(ctx context.Context, token, realm, username, password string, roleIDs map[string]string, roleNames []string) error {
	// Check if user already exists.
	searchResp, err := c.do(ctx, token, http.MethodGet,
		fmt.Sprintf("/admin/realms/%s/users?username=%s&exact=true", realm, url.QueryEscape(username)), nil)
	if err != nil {
		return fmt.Errorf("search user %s: %w", username, err)
	}
	defer searchResp.Body.Close()

	var users []struct {
		ID string `json:"id"`
	}
	body, _ := io.ReadAll(searchResp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		return fmt.Errorf("parse users: %w", err)
	}

	var userID string
	if len(users) > 0 {
		userID = users[0].ID
	} else {
		// Create user.
		userPayload := map[string]any{
			"username":      username,
			"enabled":       true,
			"emailVerified": true,
			"credentials": []map[string]any{
				{"type": "password", "value": password, "temporary": false},
			},
		}
		createResp, err := c.do(ctx, token, http.MethodPost,
			fmt.Sprintf("/admin/realms/%s/users", realm), userPayload)
		if err != nil {
			return fmt.Errorf("create user %s: %w", username, err)
		}
		defer createResp.Body.Close()
		if createResp.StatusCode >= 300 {
			b, _ := io.ReadAll(createResp.Body)
			return fmt.Errorf("create user %s (status %d): %s", username, createResp.StatusCode, string(b))
		}

		location := createResp.Header.Get("Location")
		if location == "" {
			return fmt.Errorf("create user %s: missing Location header", username)
		}
		parts := strings.Split(strings.TrimRight(location, "/"), "/")
		userID = parts[len(parts)-1]
	}

	// Assign roles.
	if len(roleNames) == 0 {
		return nil
	}
	var roleMappings []map[string]any
	for _, name := range roleNames {
		id, ok := roleIDs[name]
		if !ok {
			return fmt.Errorf("role %s not found in role map", name)
		}
		roleMappings = append(roleMappings, map[string]any{"id": id, "name": name})
	}

	assignResp, err := c.do(ctx, token, http.MethodPost,
		fmt.Sprintf("/admin/realms/%s/users/%s/role-mappings/realm", realm, userID), roleMappings)
	if err != nil {
		return fmt.Errorf("assign roles to %s: %w", username, err)
	}
	defer assignResp.Body.Close()
	if assignResp.StatusCode >= 300 {
		b, _ := io.ReadAll(assignResp.Body)
		return fmt.Errorf("assign roles to %s (status %d): %s", username, assignResp.StatusCode, string(b))
	}
	return nil
}

// setupKeycloak is the handler for POST /setup/keycloak.
func (h *Handler) setupKeycloak(w http.ResponseWriter, r *http.Request) {
	const (
		keycloakURL = "https://keycloak-production-734c.up.railway.app"
		adminUser   = "admin"
		adminPass   = "admin"
		realm       = "luxus"
		frontendURL = "https://connect-web-production-e247.up.railway.app"
	)

	ctx := r.Context()
	client := newKeycloakSetupClient(keycloakURL, adminUser, adminPass)

	// 1. Authenticate.
	token, err := client.getAdminToken(ctx)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("keycloak authentication failed: %v", err),
		})
		return
	}

	// 2. Ensure realm exists.
	if err := client.ensureRealm(ctx, token, realm); err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("ensure realm: %v", err),
		})
		return
	}

	// 3. Create realm roles.
	roleNames := []string{"admin", "user", "partner", "master", "employee", "financial"}
	roleIDs := make(map[string]string, len(roleNames))
	for _, name := range roleNames {
		id, err := client.ensureRealmRole(ctx, token, realm, name)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("ensure role %s: %v", name, err),
			})
			return
		}
		roleIDs[name] = id
	}

	// 4. Create client scopes.
	scopes := []struct{ name, protocol string }{
		{"organization", "openid-connect"},
		{"luxus-roles", "openid-connect"},
	}
	for _, s := range scopes {
		if _, err := client.ensureClientScope(ctx, token, realm, s.name, s.protocol); err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("ensure client scope %s: %v", s.name, err),
			})
			return
		}
	}

	// 5. Create connect-cli client.
	connectClient := map[string]any{
		"clientId":                  "connect-cli",
		"name":                      "Connect CLI",
		"enabled":                   true,
		"publicClient":              true,
		"standardFlowEnabled":       true,
		"directAccessGrantsEnabled": true,
		"protocol":                  "openid-connect",
		"redirectUris": []string{
			frontendURL + "/*",
			"http://localhost:5173/*",
			"http://localhost:3000/*",
		},
		"webOrigins": []string{
			frontendURL,
			"http://localhost:5173",
			"http://localhost:3000",
		},
		"defaultClientScopes": []string{
			"openid", "profile", "email", "roles", "web-origins",
		},
		"optionalClientScopes": []string{
			"organization", "luxus-roles",
		},
	}
	if err := client.ensureClient(ctx, token, realm, connectClient); err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("ensure client connect-cli: %v", err),
		})
		return
	}

	// 6. Create test users.
	type testUser struct {
		username string
		password string
		roles    []string
	}
	testUsers := []testUser{
		{username: "dev", password: "dev", roles: []string{"admin", "master", "employee", "financial", "user"}},
		{username: "parceiro", password: "parceiro", roles: []string{"partner", "user"}},
		{username: "funcionario", password: "funcionario", roles: []string{"employee", "user"}},
		{username: "financeiro", password: "financeiro", roles: []string{"financial", "user"}},
	}
	for _, u := range testUsers {
		if err := client.ensureUser(ctx, token, realm, u.username, u.password, roleIDs, u.roles); err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("ensure user %s: %v", u.username, err),
			})
			return
		}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Keycloak realm '%s' configured successfully", realm),
		"details": map[string]any{
			"realm":        realm,
			"roles":        roleNames,
			"clientScopes": []string{"organization", "luxus-roles"},
			"client":       "connect-cli",
			"users":        []string{"dev", "parceiro", "funcionario", "financeiro"},
		},
	})
}
