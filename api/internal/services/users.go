package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func profileFromRoles(roles []string) string {
	set := map[string]struct{}{}
	for _, r := range roles {
		set[r] = struct{}{}
	}
	switch {
	case hasRole(set, auth.RoleMaster) || hasRole(set, auth.RoleAdmin):
		return auth.RoleMaster
	case hasRole(set, auth.RoleFinancial):
		return auth.RoleFinancial
	case hasRole(set, auth.RoleEmployee):
		return auth.RoleEmployee
	case hasRole(set, auth.RolePartner):
		return auth.RolePartner
	default:
		return auth.RoleUser
	}
}

func hasRole(set map[string]struct{}, role string) bool {
	_, ok := set[role]
	return ok
}

func rolesForProfile(profile string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case auth.RoleMaster, auth.RoleAdmin:
		return []string{auth.RoleMaster, auth.RoleUser}, nil
	case auth.RoleEmployee:
		return []string{auth.RoleEmployee, auth.RoleUser}, nil
	case auth.RoleFinancial:
		return []string{auth.RoleFinancial, auth.RoleUser}, nil
	case auth.RolePartner:
		return []string{auth.RolePartner, auth.RoleUser}, nil
	default:
		return nil, fmt.Errorf("invalid profile")
	}
}

func toListUser(u keycloak.UserRecord) models.ListOrganizationUserResponse {
	fullName := strings.TrimSpace(strings.TrimSpace(u.FirstName + " " + u.LastName))
	if fullName == "" {
		fullName = u.Username
	}
	return models.ListOrganizationUserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		FullName:  fullName,
		Profile:   profileFromRoles(u.Roles),
		Enabled:   u.Enabled,
	}
}

func (s *Service) ListOrganizationUsers(ctx context.Context, search string) ([]models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.InternalError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
	}
	users, err := s.Keycloak.ListUsers(ctx, search, 200)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items := make([]models.ListOrganizationUserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, toListUser(u))
	}
	return items, nil
}

func (s *Service) CreateOrganizationUser(ctx context.Context, input models.CreateOrganizationUserInput) (*models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.InternalError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
	}
	org := auth.OrganizationFromContext(ctx)
	if org == nil || org.ID == "" {
		return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
	}

	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)
	if username == "" || email == "" || password == "" {
		return nil, httputil.ValidationError(notifications.N("USER_FIELDS_REQUIRED", "Username, email and password are required."))
	}
	if len(password) < 6 {
		return nil, httputil.ValidationError(notifications.N("USER_PASSWORD_TOO_SHORT", "Password must be at least 6 characters."))
	}

	roleNames, err := rolesForProfile(input.Profile)
	if err != nil {
		return nil, httputil.ValidationError(notifications.N("USER_PROFILE_INVALID", "Invalid user profile."))
	}

	orgName := org.Name
	if orgName == "" {
		orgName = "Luxus Connect"
	}

	userID, err := s.Keycloak.CreateUser(ctx, keycloak.CreateUserPayload{
		Username:      username,
		Email:         email,
		FirstName:     strings.TrimSpace(input.FirstName),
		LastName:      strings.TrimSpace(input.LastName),
		Enabled:       true,
		EmailVerified: true,
		Attributes:    keycloak.DefaultOrganizationAttribute(org.ID, orgName),
		Credentials: []keycloak.CredentialPayload{
			{Type: "password", Value: password, Temporary: false},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, httputil.BusinessError(notifications.N("USER_USERNAME_DUPLICATED", "Username already exists."))
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if err := s.Keycloak.ReplaceUserRealmRoles(ctx, userID, roleNames); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	users, err := s.Keycloak.ListUsers(ctx, username, 5)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, u := range users {
		if u.ID == userID {
			item := toListUser(u)
			return &item, nil
		}
	}
	return nil, httputil.InternalError(notifications.SharedUnexpectedError("created user not found"))
}

func (s *Service) UpdateOrganizationUser(ctx context.Context, userID string, input models.UpdateOrganizationUserInput) (*models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.InternalError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, httputil.ValidationError(notifications.N("USER_NOT_FOUND", "User was not found."))
	}

	if input.Enabled != nil {
		if err := s.Keycloak.SetUserEnabled(ctx, userID, *input.Enabled); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	if input.Profile != nil {
		roleNames, err := rolesForProfile(*input.Profile)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("USER_PROFILE_INVALID", "Invalid user profile."))
		}
		if err := s.Keycloak.ReplaceUserRealmRoles(ctx, userID, roleNames); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	if input.Password != nil && strings.TrimSpace(*input.Password) != "" {
		if len(strings.TrimSpace(*input.Password)) < 6 {
			return nil, httputil.ValidationError(notifications.N("USER_PASSWORD_TOO_SHORT", "Password must be at least 6 characters."))
		}
		if err := s.Keycloak.ResetPassword(ctx, userID, strings.TrimSpace(*input.Password), false); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	users, err := s.Keycloak.ListUsers(ctx, "", 200)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, u := range users {
		if u.ID == userID {
			item := toListUser(u)
			return &item, nil
		}
	}
	return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "User was not found."))
}
