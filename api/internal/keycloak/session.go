package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type userSessionRecord struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

type sessionUserCacheEntry struct {
	userID  string
	expires time.Time
}

type sessionUserCache struct {
	mu      sync.RWMutex
	entries map[string]sessionUserCacheEntry
}

func (c *sessionUserCache) get(sessionID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[sessionID]
	if !ok || time.Now().After(entry.expires) {
		return "", false
	}
	return entry.userID, true
}

func (c *sessionUserCache) set(sessionID, userID string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = map[string]sessionUserCacheEntry{}
	}
	c.entries[sessionID] = sessionUserCacheEntry{
		userID:  userID,
		expires: time.Now().Add(ttl),
	}
}

// ResolveUserIDBySessionSID maps a Keycloak session id (sid claim) to the user UUID.
func (c *AdminClient) ResolveUserIDBySessionSID(ctx context.Context, sessionID, clientID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", fmt.Errorf("session id is required")
	}
	if clientID == "" {
		clientID = "connect-cli"
	}

	if c.sessionCache != nil {
		if userID, ok := c.sessionCache.get(sessionID); ok {
			return userID, nil
		}
	}

	internalClientID, err := c.resolveClientInternalID(ctx, clientID)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("/admin/realms/%s/clients/%s/user-sessions", c.realm, internalClientID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("list user sessions: %s", string(body))
	}

	var sessions []userSessionRecord
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return "", err
	}

	for _, session := range sessions {
		if session.ID == sessionID && strings.TrimSpace(session.UserID) != "" {
			if c.sessionCache != nil {
				c.sessionCache.set(sessionID, session.UserID, 5*time.Minute)
			}
			return session.UserID, nil
		}
	}
	return "", fmt.Errorf("session not found")
}

func (c *AdminClient) resolveClientInternalID(ctx context.Context, clientID string) (string, error) {
	c.clientMu.Lock()
	if c.connectClientID != "" && c.connectClientName == clientID {
		id := c.connectClientID
		c.clientMu.Unlock()
		return id, nil
	}
	c.clientMu.Unlock()

	path := fmt.Sprintf("/admin/realms/%s/clients?clientId=%s", c.realm, url.QueryEscape(clientID))
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("resolve client id: %s", string(body))
	}

	var clients []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return "", err
	}
	if len(clients) == 0 || strings.TrimSpace(clients[0].ID) == "" {
		return "", fmt.Errorf("client %s not found", clientID)
	}

	c.clientMu.Lock()
	c.connectClientName = clientID
	c.connectClientID = clients[0].ID
	id := c.connectClientID
	c.clientMu.Unlock()
	return id, nil
}
