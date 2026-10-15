package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
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
	case auth.RoleUser:
		return []string{auth.RoleUser}, nil
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

func localListUser(u store.LocalUser) models.ListOrganizationUserResponse {
	return models.ListOrganizationUserResponse{ID: u.ID, Username: u.Username, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName, FullName: strings.TrimSpace(u.FirstName + " " + u.LastName), Profile: profileFromRoles(u.Roles), Enabled: u.Enabled, OrganizationID: u.OrganizationID, OrganizationName: u.OrganizationName}
}
func validateLocalIdentity(username, email string) error {
	address, err := mail.ParseAddress(email)
	if len(username) == 0 || len(username) > 128 || strings.ContainsAny(username, " \t\r\n") || err != nil || address.Address != email || len(email) > 254 {
		return httputil.ValidationError(notifications.N("USER_FIELDS_INVALID", "Informe usuário sem espaços e e-mail válido."))
	}
	return nil
}
func localUserError(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return httputil.BusinessError(notifications.N("USER_USERNAME_DUPLICATED", "Usuário já existe."))
	}
	return httputil.InternalError(notifications.SharedUnexpectedError("Falha ao salvar usuário."))
}
func (s *Service) ListOrganizationUsers(ctx context.Context, search string) ([]models.ListOrganizationUserResponse, error) {
	org, err := requireCallerOrganization(ctx)
	if err != nil {
		return nil, err
	}
	users, err := s.Store.ListLocalUsers(ctx, org.ID, search)
	if err != nil {
		return nil, localUserError(err)
	}
	out := []models.ListOrganizationUserResponse{}
	for _, u := range users {
		out = append(out, localListUser(u))
	}
	return out, nil
}

func (s *Service) CreateOrganizationUser(ctx context.Context, input models.CreateOrganizationUserInput) (*models.ListOrganizationUserResponse, error) {

	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(input.Email)
	password := input.Password
	if username == "" || email == "" || password == "" {
		return nil, httputil.ValidationError(notifications.N("USER_FIELDS_REQUIRED", "Username, email and password are required."))
	}
	if err := validateLocalIdentity(username, email); err != nil {
		return nil, err
	}
	passwordHash, err := localauth.HashPassword(password)
	if err != nil {
		return nil, httputil.ValidationError(notifications.N("PASSWORD_POLICY", err.Error()))
	}

	roleNames, err := rolesForProfile(input.Profile)
	if err != nil {
		return nil, httputil.ValidationError(notifications.N("USER_PROFILE_INVALID", "Invalid user profile."))
	}

	isNewUserMaster := profileFromRoles(roleNames) == auth.RoleMaster

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

	user := store.LocalUser{ID: uuid.NewString(), Username: username, Email: email, FirstName: firstName, LastName: lastName, Enabled: true, Roles: roleNames, OrganizationID: targetOrgID, OrganizationName: targetOrgName, PasswordHash: passwordHash}
	if err := s.Store.CreateLocalUser(ctx, &user); err != nil {
		return nil, localUserError(err)
	}
	item := localListUser(user)
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
	org, err := requireCallerOrganization(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.Store.LocalUserByID(ctx, userID)
	if err != nil {
		return nil, localUserError(err)
	}
	if target == nil {
		return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "Usuário não encontrado."))
	}
	if target.OrganizationID != org.ID {
		return nil, httputil.ForbiddenError(notifications.N("FORBIDDEN", "Usuário fora da organização autenticada."))
	}
	caller := auth.UserFromContext(ctx)
	if caller != nil && caller.ID == userID && (input.Profile != nil && strings.ToLower(strings.TrimSpace(*input.Profile)) != profileFromRoles(target.Roles) || input.Password != nil || input.Enabled != nil && !*input.Enabled) {
		return nil, httputil.ValidationError(notifications.N("SELF_ACCOUNT_CHANGE", "Altere sua senha em Segurança da conta. Seu próprio perfil e acesso não podem ser removidos aqui."))
	}
	if input.FirstName != nil {
		target.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		target.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Email != nil {
		target.Email = strings.TrimSpace(*input.Email)
	}
	if err := validateLocalIdentity(target.Username, target.Email); err != nil {
		return nil, err
	}
	if input.Enabled != nil {
		target.Enabled = *input.Enabled
	}
	if input.Profile != nil {
		roles, err := rolesForProfile(*input.Profile)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("USER_PROFILE_INVALID", "Perfil inválido."))
		}
		target.Roles = roles
	}
	if input.Password != nil {
		hash, err := localauth.HashPassword(*input.Password)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("PASSWORD_POLICY", err.Error()))
		}
		target.PasswordHash = hash
	}
	if input.OrganizationName != nil {
		if org.ID != auth.DefaultLuxusOrganizationID {
			return nil, httputil.ForbiddenError(notifications.N("FORBIDDEN", "Apenas o administrador da plataforma pode alterar a empresa."))
		}
		if strings.TrimSpace(*input.OrganizationName) == "" {
			return nil, httputil.ValidationError(notifications.N("ORGANIZATION_NAME_REQUIRED", "Informe a empresa."))
		}
		target.OrganizationName = strings.TrimSpace(*input.OrganizationName)
	}
	if err := s.Store.UpdateLocalUser(ctx, target); err != nil {
		return nil, localUserError(err)
	}
	item := localListUser(*target)
	return &item, nil
}

func requireCallerOrganization(ctx context.Context) (*auth.Organization, error) {
	org := auth.NormalizeOrganization(auth.OrganizationFromContext(ctx))
	if org == nil || strings.TrimSpace(org.ID) == "" {
		return nil, httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	return org, nil
}
