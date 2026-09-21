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

// GetUserOrganizationIDAttribute returns the plain UUID organization_id attribute.
func (c *AdminClient) GetUserOrganizationIDAttribute(ctx context.Context, userID string) (string, error) {
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

	values := user.Attributes["organization_id"]
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return "", fmt.Errorf("organization_id attribute not found")
	}
	return strings.TrimSpace(values[0]), nil
}

// SetUserOrganizationAttribute replaces the organization user attribute while
// preserving every other Keycloak attribute on the user representation.
func (c *AdminClient) SetUserOrganizationAttribute(ctx context.Context, userID, orgID, orgName string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	current, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	attrs := map[string][]string{}
	for k, v := range current.Attributes {
		attrs[k] = append([]string{}, v...)
	}
	for k, v := range DefaultOrganizationAttribute(orgID, orgName) {
		attrs[k] = v
	}
	return c.putUserRepresentation(ctx, userID, map[string]any{
		"username":      current.Username,
		"email":         current.Email,
		"firstName":     current.FirstName,
		"lastName":      current.LastName,
		"enabled":       current.Enabled,
		"emailVerified": true,
		"attributes":    attrs,
	})
}
