package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	case hasRole(set, auth.RoleSales):
		return auth.RoleSales
	case hasRole(set, auth.RoleOperator):
		return auth.RoleOperator
	case hasRole(set, auth.RoleEmployee):
		return auth.RoleEmployee
	case hasRole(set, auth.RoleViewer):
		return auth.RoleViewer
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
	case auth.RoleOperator:
		return []string{auth.RoleOperator, auth.RoleEmployee, auth.RoleUser}, nil
	case auth.RoleFinancial:
		return []string{auth.RoleFinancial, auth.RoleUser}, nil
	case auth.RoleSales:
		return []string{auth.RoleSales, auth.RoleUser}, nil
	case auth.RoleViewer:
		return []string{auth.RoleViewer, auth.RoleUser}, nil
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

	orgID := ""
	orgName := ""
	if u.Attributes != nil {
		if orgAttr, ok := u.Attributes["organization"]; ok && len(orgAttr) > 0 {
			if parsed, err := auth.ParseOrganizationClaim(orgAttr[0]); err == nil && parsed != nil {
				parsed = auth.NormalizeOrganization(parsed)
				orgID = parsed.ID
				orgName = parsed.Name
			}
		}
		if orgID == "" {
			if ids, ok := u.Attributes["organization_id"]; ok && len(ids) > 0 {
				orgID = strings.TrimSpace(ids[0])
			}
		}
	}

	return models.ListOrganizationUserResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		FullName:         fullName,
		Profile:          profileFromRoles(u.Roles),
		Enabled:          u.Enabled,
		OrganizationID:   orgID,
		OrganizationName: orgName,
	}
}

func (s *Service) ListOrganizationUsers(ctx context.Context, search string) ([]models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.UnavailableError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
	}
	callerOrg, err := requireCallerOrganization(ctx)
	if err != nil {
		return nil, err
	}
	users, err := s.Keycloak.ListUsers(ctx, search, 200)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

<<<<<<< HEAD
	items := make([]models.ListOrganizationUserResponse, 0, len(users))
	for _, u := range users {
		item := toListUser(u)
		if item.OrganizationID != callerOrg.ID {
=======
	currentOrg := auth.OrganizationFromContext(ctx)
	if currentOrg == nil || strings.TrimSpace(currentOrg.ID) == "" {
		return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
	}

	items := make([]models.ListOrganizationUserResponse, 0, len(users))
	for _, u := range users {
		item := toListUser(u)
		if item.OrganizationID != currentOrg.ID {
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) CreateOrganizationUser(ctx context.Context, input models.CreateOrganizationUserInput) (*models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.UnavailableError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
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

	isNewUserMaster := strings.EqualFold(strings.TrimSpace(input.Profile), auth.RoleMaster)

	var targetOrgID string
	var targetOrgName string

	if isNewUserMaster {
		callerOrg, err := requireCallerOrganization(ctx)
		if err != nil {
			return nil, err
		}
		// Somente a org plataforma Luxus pode provisionar novas empresas/masters.
		if callerOrg.ID != auth.DefaultLuxusOrganizationID {
			return nil, httputil.ForbiddenError(notifications.N(
				"FORBIDDEN",
				"Apenas o administrador da plataforma pode criar usuários Master com nova empresa.",
			))
		}
		// Novo Master = nova organização isolada (UUID próprio no token/atributo).
		targetOrgID = uuid.NewString()
		if input.OrganizationName == nil || strings.TrimSpace(*input.OrganizationName) == "" {
			return nil, httputil.ValidationError(notifications.N(
				"ORGANIZATION_NAME_REQUIRED",
				"Informe o nome da nova empresa para o usuário Master.",
			))
		}
		targetOrgName = strings.TrimSpace(*input.OrganizationName)

		blankSettings := &models.OrganizationSettingsResponse{
			OrganizationID: targetOrgID,
			Company: models.CompanySettingsDto{
				CompanyName: targetOrgName,
				TradingName: targetOrgName,
			},
			Whitelabel: models.WhitelabelSettingsDto{
				AppName:      targetOrgName,
				PrimaryColor: "#0f766e",
			},
			System: models.SystemSettingsDto{
				DefaultDueDay:         10,
				LateFeePercentage:     2.0,
				InterestRateMonthly:   1.0,
				DaysBeforeDueReminder: 3,
				DaysAfterDueReminder:  2,
				ProrataDivisor:        30,
			},
			Sicredi: models.SicrediSettingsDto{
				Sandbox: true,
			},
		}
		if err := s.Store.UpsertOrganizationSettings(ctx, targetOrgID, nil, blankSettings); err != nil {
			return nil, httputil.BusinessError(notifications.N("ORG_SETTINGS_CREATE_FAILED", truncateErr("Falha ao criar organização no banco", err)))
		}
	} else {
		// Usuário comum herda a organização do usuário autenticado.
		org := auth.OrganizationFromContext(ctx)
		if org != nil {
			org = auth.NormalizeOrganization(org)
		}
		if org != nil && strings.TrimSpace(org.ID) != "" {
			targetOrgID = strings.TrimSpace(org.ID)
			targetOrgName = strings.TrimSpace(org.Name)
		}
		if targetOrgID == "" || targetOrgID == "luxus" || targetOrgID == "default" {
			return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
		}
		if targetOrgName == "" {
			targetOrgName = targetOrgID
		}
	}

	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)
	if firstName == "" {
		firstName = username
	}
	if lastName == "" {
		lastName = "-"
	}

	userID, err := s.Keycloak.CreateUser(ctx, keycloak.CreateUserPayload{
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: true,
		Attributes:    keycloak.DefaultOrganizationAttribute(targetOrgID, targetOrgName),
		Credentials: []keycloak.CredentialPayload{
			{Type: "password", Value: password, Temporary: false},
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, httputil.BusinessError(notifications.N("USER_USERNAME_DUPLICATED", "Username already exists."))
		}
		return nil, httputil.BusinessError(notifications.N("USER_CREATE_FAILED", truncateErr("Falha ao criar usuário no Keycloak", err)))
	}

	// Garante o atributo organization no Keycloak (PUT parcial de outros fluxos já não apaga mais).
	if err := s.Keycloak.SetUserOrganizationAttribute(ctx, userID, targetOrgID, targetOrgName); err != nil {
		return nil, httputil.BusinessError(notifications.N("USER_ORG_ATTRIBUTE_FAILED", truncateErr("Falha ao gravar organização do usuário", err)))
	}

	if err := s.Keycloak.ReplaceUserRealmRoles(ctx, userID, roleNames); err != nil {
		return nil, httputil.BusinessError(notifications.N("USER_ROLES_FAILED", truncateErr("Falha ao atribuir perfil do usuário", err)))
	}

	item := models.ListOrganizationUserResponse{
		ID:               userID,
		Username:         username,
		Email:            email,
		FirstName:        firstName,
		LastName:         lastName,
		FullName:         strings.TrimSpace(firstName + " " + lastName),
		Profile:          profileFromRoles(roleNames),
		Enabled:          true,
		OrganizationID:   targetOrgID,
		OrganizationName: targetOrgName,
	}

	// Confirma no Keycloak; se o GET omitir attributes, ainda devolvemos o que gravamos.
	if created, err := s.Keycloak.GetUserByID(ctx, userID); err == nil && created != nil {
		listed := toListUser(*created)
		if listed.OrganizationID != "" {
			item.OrganizationID = listed.OrganizationID
			item.OrganizationName = listed.OrganizationName
		}
	}
	if raw, err := s.Keycloak.GetUserOrganizationAttribute(ctx, userID); err != nil || !strings.Contains(raw, targetOrgID) {
		return nil, httputil.BusinessError(notifications.N(
			"USER_ORG_ATTRIBUTE_MISSING",
			"Usuário criado, mas o atributo organization não ficou gravado no Keycloak. Verifique o user profile (organization).",
		))
	}
	return &item, nil
}

func truncateErr(prefix string, err error) string {
	if err == nil {
		return prefix
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 240 {
		msg = msg[:240] + "…"
	}
	return prefix + ": " + msg
}

func (s *Service) UpdateOrganizationUser(ctx context.Context, userID string, input models.UpdateOrganizationUserInput) (*models.ListOrganizationUserResponse, error) {
	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.UnavailableError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured."))
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, httputil.ValidationError(notifications.N("USER_NOT_FOUND", "User was not found."))
	}
	currentOrg := auth.OrganizationFromContext(ctx)
	if currentOrg == nil || strings.TrimSpace(currentOrg.ID) == "" {
		return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	owner, err := s.Keycloak.GetUserByID(ctx, userID)
	if err != nil || owner == nil || toListUser(*owner).OrganizationID != currentOrg.ID {
		return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "User was not found."))
	}

	callerOrg, err := requireCallerOrganization(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.Keycloak.GetUserByID(ctx, userID)
	if err != nil {
		return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "User was not found."))
	}
	targetItem := toListUser(*target)
	isPlatformAdmin := callerOrg.ID == auth.DefaultLuxusOrganizationID
	if !isPlatformAdmin && targetItem.OrganizationID != callerOrg.ID {
		return nil, httputil.ForbiddenError(notifications.N("FORBIDDEN", "Usuário fora da organização autenticada."))
	}

	if input.FirstName != nil || input.LastName != nil || input.Email != nil {
		firstName := target.FirstName
		lastName := target.LastName
		email := target.Email
		if input.FirstName != nil {
			firstName = strings.TrimSpace(*input.FirstName)
		}
		if input.LastName != nil {
			lastName = strings.TrimSpace(*input.LastName)
		}
		if input.Email != nil {
			email = strings.TrimSpace(*input.Email)
		}
		if err := s.Keycloak.UpdateUserProfile(ctx, userID, firstName, lastName, email); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
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
		newPass := strings.TrimSpace(*input.Password)
		if len(newPass) < 6 {
			return nil, httputil.ValidationError(notifications.N("USER_PASSWORD_TOO_SHORT", "Password must be at least 6 characters."))
		}
		if err := s.Keycloak.ResetPassword(ctx, userID, newPass, false); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	if input.OrganizationName != nil {
		orgName := strings.TrimSpace(*input.OrganizationName)
		if orgName == "" {
			return nil, httputil.ValidationError(notifications.N(
				"ORGANIZATION_NAME_REQUIRED",
				"Informe o nome da empresa.",
			))
		}
		if !isPlatformAdmin {
			return nil, httputil.ForbiddenError(notifications.N("FORBIDDEN", "Only the platform admin can reassign organizations."))
		}

		orgID := strings.TrimSpace(targetItem.OrganizationID)
		if orgID == "" || orgID == "luxus" || orgID == "default" {
			// Usuário sem tenant próprio: provisiona organização nova.
			orgID = uuid.NewString()
			blankSettings := &models.OrganizationSettingsResponse{
				OrganizationID: orgID,
				Company: models.CompanySettingsDto{
					CompanyName: orgName,
					TradingName: orgName,
				},
				Whitelabel: models.WhitelabelSettingsDto{
					AppName:      orgName,
					PrimaryColor: "#0f766e",
				},
				System: models.SystemSettingsDto{
					DefaultDueDay:         10,
					LateFeePercentage:     2.0,
					InterestRateMonthly:   1.0,
					DaysBeforeDueReminder: 3,
					DaysAfterDueReminder:  2,
					ProrataDivisor:        30,
				},
				Sicredi: models.SicrediSettingsDto{Sandbox: true},
			}
			if err := s.Store.UpsertOrganizationSettings(ctx, orgID, nil, blankSettings); err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
		}
		if err := s.Keycloak.SetUserOrganizationAttribute(ctx, userID, orgID, orgName); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	created, err := s.Keycloak.GetUserByID(ctx, userID)
	if err != nil {
		return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "User was not found."))
	}
	item := toListUser(*created)
	return &item, nil
}

func requireCallerOrganization(ctx context.Context) (*auth.Organization, error) {
	org := auth.NormalizeOrganization(auth.OrganizationFromContext(ctx))
	if org == nil || strings.TrimSpace(org.ID) == "" {
		return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	return org, nil
}
