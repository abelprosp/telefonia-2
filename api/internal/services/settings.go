package services

import (
	"context"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
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

	// Tentar obter dados atualizados do Keycloak se disponível
	if s.Keycloak != nil && s.Keycloak.Enabled() {
		if kcUser, err := s.Keycloak.GetUserByID(ctx, u.ID); err == nil && kcUser != nil {
			if kcUser.FirstName != "" {
				firstName = kcUser.FirstName
			}
			if kcUser.LastName != "" {
				lastName = kcUser.LastName
			}
			if kcUser.Email != "" {
				u.Email = kcUser.Email
			}
			if len(kcUser.Roles) > 0 {
				u.Roles = kcUser.Roles
			}
		}
	}

	calcFullName := strings.TrimSpace(firstName + " " + lastName)
	if calcFullName == "" {
		calcFullName = u.Username
	}

	mfaEnrolled := auth.MFAVerifiedFromClaims(u.Acr, u.Amr)
	if !mfaEnrolled && s.Keycloak != nil && s.Keycloak.Enabled() {
		if hasOTP, err := s.Keycloak.UserHasOTP(ctx, u.ID); err == nil {
			mfaEnrolled = hasOTP
		}
	}
	accountURL := ""
	if s.Keycloak != nil {
		accountURL = s.Keycloak.AccountSecurityURL()
	}

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
	u := auth.UserFromContext(ctx)
	if u == nil || u.ID == "" {
		return nil, httputil.ForbiddenError(notifications.N("UNAUTHORIZED", "Usuário não autenticado."))
	}

	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return nil, httputil.InternalError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "Keycloak admin is not available."))
	}

	firstName := ""
	lastName := ""
	email := u.Email

	// Pegar dados atuais
	if kcUser, err := s.Keycloak.GetUserByID(ctx, u.ID); err == nil && kcUser != nil {
		firstName = kcUser.FirstName
		lastName = kcUser.LastName
		if kcUser.Email != "" {
			email = kcUser.Email
		}
	}

	if input.FirstName != nil {
		firstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		lastName = strings.TrimSpace(*input.LastName)
	}
	if input.Email != nil && strings.TrimSpace(*input.Email) != "" {
		email = strings.TrimSpace(*input.Email)
	}

	if err := s.Keycloak.UpdateUserProfile(ctx, u.ID, firstName, lastName, email); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if input.NewPassword != nil && strings.TrimSpace(*input.NewPassword) != "" {
		newPass := strings.TrimSpace(*input.NewPassword)
		if len(newPass) < 6 {
			return nil, httputil.ValidationError(notifications.N("PASSWORD_TOO_SHORT", "Nova senha deve conter no mínimo 6 caracteres."))
		}
		if err := s.Keycloak.ResetPassword(ctx, u.ID, newPass, false); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	return s.GetCurrentUserProfile(ctx)
}

func (s *Service) GetOrganizationSettings(ctx context.Context) (*models.OrganizationSettingsResponse, error) {
	org := auth.OrganizationFromContext(ctx)
	orgID := "default"
	if org != nil && org.ID != "" {
		orgID = org.ID
	}
	return s.Store.GetOrganizationSettings(ctx, orgID)
}

func (s *Service) UpdateCompanySettings(ctx context.Context, input models.UpdateCompanySettingsInput) (*models.OrganizationSettingsResponse, error) {
	org := auth.OrganizationFromContext(ctx)
	orgID := "default"
	if org != nil && org.ID != "" {
		orgID = org.ID
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
	org := auth.OrganizationFromContext(ctx)
	orgID := "default"
	if org != nil && org.ID != "" {
		orgID = org.ID
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
	org := auth.OrganizationFromContext(ctx)
	orgID := "default"
	if org != nil && org.ID != "" {
		orgID = org.ID
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
