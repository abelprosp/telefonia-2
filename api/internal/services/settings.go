package services

import (
	"context"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) GetCurrentUserProfile(ctx context.Context) (*models.UserProfileResponse, error) {
	u := auth.UserFromContext(ctx)
	if u == nil || u.ID == "" {
		return nil, httputil.ForbiddenError(notifications.N("UNAUTHORIZED", "Usuário não autenticado."))
	}

	fullName := strings.TrimSpace(u.Name)
	var firstName, lastName string
	parts := strings.Split(fullName, " ")
	if len(parts) > 0 {
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
		}
	}

	calcFullName := strings.TrimSpace(firstName + " " + lastName)
	if calcFullName == "" {
		calcFullName = u.Username
	}

	mfaEnrolled := false
	if s.Store != nil {
		if local, e := s.Store.LocalUserByID(ctx, u.ID); e == nil && local != nil {
			mfaEnrolled = local.PasskeyCount > 0
		}
	}
	accountURL := "/settings"

	return &models.UserProfileResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		FirstName:        firstName,
		LastName:         lastName,
		FullName:         calcFullName,
		Roles:            u.Roles,
		Profile:          profileFromRoles(u.Roles),
		MFAEnrolled:      mfaEnrolled,
		MFAVerified:      auth.MFAVerifiedFromClaims(u.Acr, u.Amr),
		MFAAccountURL:    accountURL,
		Acr:              u.Acr,
		Amr:              u.Amr,
		PrivilegedAccess: auth.CanApproveTwoLevel(ctx) || auth.IsMaster(ctx),
	}, nil
}

func (s *Service) UpdateCurrentUserProfile(ctx context.Context, input models.UpdateUserProfileInput) (*models.UserProfileResponse, error) {
	if input.NewPassword != nil || input.CurrentPassword != nil {
		return nil, httputil.ValidationError(notifications.N("PASSWORD_ACCOUNT_FLOW_REQUIRED", "Use a seção Segurança da conta para alterar a senha."))
	}
	u := auth.UserFromContext(ctx)
	if u == nil || u.ID == "" {
		return nil, httputil.ForbiddenError(notifications.N("UNAUTHORIZED", "Usuário não autenticado."))
	}
	local, err := s.Store.LocalUserByID(ctx, u.ID)
	if err != nil || local == nil {
		return nil, httputil.NotFoundError(notifications.N("USER_NOT_FOUND", "Usuário não encontrado."))
	}
	if input.FirstName != nil {
		local.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		local.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Email != nil {
		local.Email = strings.TrimSpace(*input.Email)
	}
	if err := validateLocalIdentity(local.Username, local.Email); err != nil {
		return nil, err
	}
	if err := s.Store.UpdateLocalUser(ctx, local); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError("Falha ao salvar perfil."))
	}
	u.Name = strings.TrimSpace(local.FirstName + " " + local.LastName)
	u.Email = local.Email
	return s.GetCurrentUserProfile(ctx)
}

func (s *Service) GetOrganizationSettings(ctx context.Context) (*models.OrganizationSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		// Público (login/whitelabel): branding Luxus.
		// Autenticado sem org no token: não mascarar como Luxus.
		if auth.UserFromContext(ctx) != nil {
			return nil, err
		}
		orgID = auth.DefaultLuxusOrganizationID
	}
	settings, err := s.Store.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return settings, nil
}

// PublicOrganizationBranding exposes only the information needed before login.
func (s *Service) PublicOrganizationBranding(ctx context.Context) (*models.PublicOrganizationBrandingResponse, error) {
	settings, err := s.GetOrganizationSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &models.PublicOrganizationBrandingResponse{Whitelabel: settings.Whitelabel}, nil
}

func (s *Service) UpdateCompanySettings(ctx context.Context, input models.UpdateCompanySettingsInput) (*models.OrganizationSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	current, err := s.Store.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if input.CompanyName != nil {
		current.Company.CompanyName = strings.TrimSpace(*input.CompanyName)
	}
	if input.TradingName != nil {
		current.Company.TradingName = strings.TrimSpace(*input.TradingName)
	}
	if input.Cnpj != nil {
		current.Company.Cnpj = strings.TrimSpace(*input.Cnpj)
	}
	if input.StateRegistration != nil {
		current.Company.StateRegistration = strings.TrimSpace(*input.StateRegistration)
	}
	if input.Email != nil {
		current.Company.Email = strings.TrimSpace(*input.Email)
	}
	if input.Phone != nil {
		current.Company.Phone = strings.TrimSpace(*input.Phone)
	}
	if input.Website != nil {
		current.Company.Website = strings.TrimSpace(*input.Website)
	}
	if input.ZipCode != nil {
		current.Company.ZipCode = strings.TrimSpace(*input.ZipCode)
	}
	if input.Street != nil {
		current.Company.Street = strings.TrimSpace(*input.Street)
	}
	if input.Number != nil {
		current.Company.Number = strings.TrimSpace(*input.Number)
	}
	if input.Complement != nil {
		current.Company.Complement = strings.TrimSpace(*input.Complement)
	}
	if input.Neighborhood != nil {
		current.Company.Neighborhood = strings.TrimSpace(*input.Neighborhood)
	}
	if input.City != nil {
		current.Company.City = strings.TrimSpace(*input.City)
	}
	if input.State != nil {
		current.Company.State = strings.TrimSpace(*input.State)
	}

	var updatedBy *string
	if u := auth.UserFromContext(ctx); u != nil {
		updatedBy = &u.ID
	}

	if err := s.Store.UpsertOrganizationSettings(ctx, orgID, updatedBy, current); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	return s.Store.GetOrganizationSettings(ctx, orgID)
}

func (s *Service) UpdateWhitelabelSettings(ctx context.Context, input models.UpdateWhitelabelSettingsInput) (*models.OrganizationSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	current, err := s.Store.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if input.AppName != nil {
		current.Whitelabel.AppName = strings.TrimSpace(*input.AppName)
	}
	if input.AppSlogan != nil {
		current.Whitelabel.AppSlogan = strings.TrimSpace(*input.AppSlogan)
	}
	if input.LogoUrl != nil {
		current.Whitelabel.LogoUrl = strings.TrimSpace(*input.LogoUrl)
	}
	if input.DarkLogoUrl != nil {
		current.Whitelabel.DarkLogoUrl = strings.TrimSpace(*input.DarkLogoUrl)
	}
	if input.FaviconUrl != nil {
		current.Whitelabel.FaviconUrl = strings.TrimSpace(*input.FaviconUrl)
	}
	if input.PrimaryColor != nil && strings.TrimSpace(*input.PrimaryColor) != "" {
		current.Whitelabel.PrimaryColor = strings.TrimSpace(*input.PrimaryColor)
	}
	if input.SupportEmail != nil {
		current.Whitelabel.SupportEmail = strings.TrimSpace(*input.SupportEmail)
	}
	if input.SupportPhone != nil {
		current.Whitelabel.SupportPhone = strings.TrimSpace(*input.SupportPhone)
	}
	if input.FooterText != nil {
		current.Whitelabel.FooterText = strings.TrimSpace(*input.FooterText)
	}

	var updatedBy *string
	if u := auth.UserFromContext(ctx); u != nil {
		updatedBy = &u.ID
	}

	if err := s.Store.UpsertOrganizationSettings(ctx, orgID, updatedBy, current); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	return s.Store.GetOrganizationSettings(ctx, orgID)
}

func (s *Service) UpdateSystemSettings(ctx context.Context, input models.UpdateSystemSettingsInput) (*models.OrganizationSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	current, err := s.Store.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if input.DefaultDueDay != nil {
		current.System.DefaultDueDay = *input.DefaultDueDay
	}
	if input.LateFeePercentage != nil {
		current.System.LateFeePercentage = *input.LateFeePercentage
	}
	if input.InterestRateMonthly != nil {
		current.System.InterestRateMonthly = *input.InterestRateMonthly
	}
	if input.DaysBeforeDueReminder != nil {
		current.System.DaysBeforeDueReminder = *input.DaysBeforeDueReminder
	}
	if input.DaysAfterDueReminder != nil {
		current.System.DaysAfterDueReminder = *input.DaysAfterDueReminder
	}
	if input.AutoSendInvoiceEmail != nil {
		current.System.AutoSendInvoiceEmail = *input.AutoSendInvoiceEmail
	}
	if input.AutoSendCollectionReminder != nil {
		current.System.AutoSendCollectionReminder = *input.AutoSendCollectionReminder
	}
	if input.ProrataDivisor != nil {
		d := *input.ProrataDivisor
		if d < 1 || d > 31 {
			return nil, httputil.ValidationError(notifications.N("PRORATA_DIVISOR_INVALID", "O divisor de pró-rata deve estar entre 1 e 31."))
		}
		current.System.ProrataDivisor = d
	}

	var updatedBy *string
	if u := auth.UserFromContext(ctx); u != nil {
		updatedBy = &u.ID
	}

	if err := s.Store.UpsertOrganizationSettings(ctx, orgID, updatedBy, current); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	return s.Store.GetOrganizationSettings(ctx, orgID)
}

func (s *Service) UpdateSicrediSettings(ctx context.Context, input models.UpdateSicrediSettingsInput) (*models.OrganizationSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	sec, err := s.Store.GetOrganizationSicrediSecrets(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sec == nil {
		sec = &store.OrgSicrediSecrets{Sandbox: true}
	}

	if input.Enabled != nil {
		sec.Enabled = *input.Enabled
	}
	if input.Sandbox != nil {
		sec.Sandbox = *input.Sandbox
	}
	if input.APIKey != nil {
		v := strings.TrimSpace(*input.APIKey)
		if v != "" && v != "********" {
			if sicredi.LooksLikeForeignAPIKey(v) {
				return nil, httputil.ValidationError(notifications.N(
					"SICREDI_API_KEY_INVALID",
					"A API Key informada parece ser da OpenAI (sk-proj/sk-), não da Sicredi. Use a chave do portal API Parceiro Sicredi.",
				))
			}
			sec.APIKey = v
		}
	}
	if input.Username != nil {
		sec.Username = strings.TrimSpace(*input.Username)
	}
	if input.Password != nil {
		v := strings.TrimSpace(*input.Password)
		if v != "" && v != "********" {
			sec.Password = v
		}
	}
	if input.Cooperativa != nil {
		sec.Cooperativa = strings.TrimSpace(*input.Cooperativa)
	}
	if input.Posto != nil {
		sec.Posto = strings.TrimSpace(*input.Posto)
	}
	if input.CodigoBeneficiario != nil {
		sec.CodigoBeneficiario = strings.TrimSpace(*input.CodigoBeneficiario)
	}
	if input.AccountNumber != nil {
		sec.AccountNumber = strings.TrimSpace(*input.AccountNumber)
	}
	if input.WebhookToken != nil {
		v := strings.TrimSpace(*input.WebhookToken)
		if v != "" && v != "********" {
			sec.WebhookToken = v
		}
	}
	if input.PublicAPIURL != nil {
		sec.PublicAPIURL = strings.TrimRight(strings.TrimSpace(*input.PublicAPIURL), "/")
	}

	var updatedBy *string
	if u := auth.UserFromContext(ctx); u != nil {
		updatedBy = &u.ID
	}

	if err := s.Store.UpsertOrganizationSicrediSettings(ctx, orgID, updatedBy, *sec); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	return s.Store.GetOrganizationSettings(ctx, orgID)
}
