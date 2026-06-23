package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type userAttributesRecord struct {
	Attributes map[string][]string `json:"attributes"`
}

// GetUserOrganizationAttribute returns the raw JSON organization attribute for a Keycloak user.
func (c *AdminClient) GetUserOrganizationAttribute(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}

	path := fmt.Sprintf("/admin/realms/%s/users/%s", c.realm, userID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get user: %s", string(body))
	}

	var user userAttributesRecord
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	values := user.Attributes["organization"]
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return "", fmt.Errorf("organization attribute not found")
	}
	return values[0], nil
}
