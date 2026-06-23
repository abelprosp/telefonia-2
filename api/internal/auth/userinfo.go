package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type userInfoResponse struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
}

func (m *Middleware) enrichUserFromUserInfo(ctx context.Context, tokenStr string, user *User) *User {
	if user == nil || strings.TrimSpace(user.ID) != "" {
		return user
	}
	if m.cfg.KeycloakAuthServerURL == "" || m.cfg.KeycloakRealm == "" {
		return user
	}

	info, err := m.fetchUserInfo(ctx, tokenStr)
	if err != nil {
		m.logger.Warn("failed to resolve user id from userinfo", "error", err)
		return user
	}

	if strings.TrimSpace(user.ID) == "" && strings.TrimSpace(info.Sub) != "" {
		user.ID = info.Sub
	}
	if strings.TrimSpace(user.Email) == "" && strings.TrimSpace(info.Email) != "" {
		user.Email = info.Email
	}
	if strings.TrimSpace(user.Name) == "" && strings.TrimSpace(info.Name) != "" {
		user.Name = info.Name
	}
	if strings.TrimSpace(user.Username) == "" && strings.TrimSpace(info.PreferredUsername) != "" {
		user.Username = info.PreferredUsername
	}
	return user
}

func (m *Middleware) fetchUserInfo(ctx context.Context, tokenStr string) (*userInfoResponse, error) {
	baseURL := m.cfg.KeycloakPublicAuthServerURL
	if baseURL == "" {
		baseURL = m.cfg.KeycloakAuthServerURL
	}
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", baseURL, m.cfg.KeycloakRealm)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var info userInfoResponse
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
