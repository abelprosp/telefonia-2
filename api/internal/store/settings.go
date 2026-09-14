package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
)

const defaultLuxusOrgID = "00000000-0000-0000-0000-000000000001"

func (s *Store) resolveOrgID(ctx context.Context, orgID string) string {
	if orgID != "" && orgID != "default" {
		return orgID
	}

	return defaultLuxusOrgID
}

// blankOrganizationSettings returns empty company/whitelabel data so each
// tenant configures itself. System keeps only safe numeric defaults.
func blankOrganizationSettings(orgID string) *models.OrganizationSettingsResponse {
	return &models.OrganizationSettingsResponse{
		OrganizationID: orgID,
		Company:        models.CompanySettingsDto{},
		Whitelabel: models.WhitelabelSettingsDto{
			PrimaryColor: "#0f766e",
		},
		System: models.SystemSettingsDto{
			DefaultDueDay:         10,
			LateFeePercentage:     2.00,
			InterestRateMonthly:   1.00,
			DaysBeforeDueReminder: 3,
			DaysAfterDueReminder:  2,
			ProrataDivisor:        30,
		},
		Sicredi: models.SicrediSettingsDto{
			Sandbox: true,
		},
		UpdatedAt: time.Now(),
	}
}

type OrgSicrediSecrets struct {
	Enabled            bool
	Sandbox            bool
	APIKey             string
	Username           string
	Password           string
	Cooperativa        string
	Posto              string
	CodigoBeneficiario string
	AccountNumber      string
	WebhookToken       string
	PublicAPIURL       string
}

func (sec OrgSicrediSecrets) toPublicDTO() models.SicrediSettingsDto {
	username := strings.TrimSpace(sec.Username)
	if username == "" && sec.CodigoBeneficiario != "" && sec.Cooperativa != "" {
		username = strings.TrimSpace(sec.CodigoBeneficiario) + strings.TrimSpace(sec.Cooperativa)
	}
	cfg := sicredi.Config{
		Enabled:            sec.Enabled,
		Sandbox:            sec.Sandbox,
		APIKey:             strings.TrimSpace(sec.APIKey),
		Username:           username,
		Password:           sec.Password,
		Cooperativa:        strings.TrimSpace(sec.Cooperativa),
		Posto:              strings.TrimSpace(sec.Posto),
		CodigoBeneficiario: strings.TrimSpace(sec.CodigoBeneficiario),
		WebhookToken:       strings.TrimSpace(sec.WebhookToken),
		PublicAPIURL:       strings.TrimSpace(sec.PublicAPIURL),
	}
	return models.SicrediSettingsDto{
		Enabled:            sec.Enabled,
		Sandbox:            sec.Sandbox,
		APIKeySet:          strings.TrimSpace(sec.APIKey) != "",
		Username:           strings.TrimSpace(sec.Username),
		PasswordSet:        strings.TrimSpace(sec.Password) != "",
		Cooperativa:        strings.TrimSpace(sec.Cooperativa),
		Posto:              strings.TrimSpace(sec.Posto),
		CodigoBeneficiario: strings.TrimSpace(sec.CodigoBeneficiario),
		AccountNumber:      strings.TrimSpace(sec.AccountNumber),
		WebhookTokenSet:    strings.TrimSpace(sec.WebhookToken) != "",
		PublicAPIURL:       strings.TrimSpace(sec.PublicAPIURL),
		Configured:         cfg.EnabledAndConfigured(),
	}
}

func (sec OrgSicrediSecrets) toSicrediConfig(production bool) sicredi.Config {
	username := strings.TrimSpace(sec.Username)
	if username == "" && sec.CodigoBeneficiario != "" && sec.Cooperativa != "" {
		username = strings.TrimSpace(sec.CodigoBeneficiario) + strings.TrimSpace(sec.Cooperativa)
	}
	return sicredi.Config{
		Enabled:            sec.Enabled,
		Sandbox:            sec.Sandbox,
		Production:         production,
		APIKey:             strings.TrimSpace(sec.APIKey),
		Username:           username,
		Password:           sec.Password,
		Cooperativa:        strings.TrimSpace(sec.Cooperativa),
		Posto:              strings.TrimSpace(sec.Posto),
		CodigoBeneficiario: strings.TrimSpace(sec.CodigoBeneficiario),
		WebhookToken:       strings.TrimSpace(sec.WebhookToken),
		PublicAPIURL:       strings.TrimSpace(sec.PublicAPIURL),
	}
}

func (s *Store) scanOrganizationSettings(ctx context.Context, orgID string, withProrata, withSicredi bool) (*models.OrganizationSettingsResponse, *OrgSicrediSecrets, error) {
	q := s.q(ctx)
	prorataSelect := `30`
	if withProrata {
		prorataSelect = `COALESCE("ProrataDivisor", 30)`
	}
	sicrediSelect := `
			FALSE, TRUE, '', '', '', '', '', '', '', '', ''`
	if withSicredi {
		sicrediSelect = `
			COALESCE("SicrediEnabled", FALSE),
			COALESCE("SicrediSandbox", TRUE),
			COALESCE("SicrediAPIKey", ''),
			COALESCE("SicrediUsername", ''),
			COALESCE("SicrediPassword", ''),
			COALESCE("SicrediCooperativa", ''),
			COALESCE("SicrediPosto", ''),
			COALESCE("SicrediCodigoBeneficiario", ''),
			COALESCE("SicrediAccountNumber", ''),
			COALESCE("SicrediWebhookToken", ''),
			COALESCE("SicrediPublicAPIURL", '')`
	}
	query := `
		SELECT "OrganizationId",
			"CompanyName", "TradingName", "Cnpj", "StateRegistration", "Email", "Phone", "Website",
			"ZipCode", "Street", "Number", "Complement", "Neighborhood", "City", "State",
			"AppName", "AppSlogan", "LogoUrl", "DarkLogoUrl", "FaviconUrl", "PrimaryColor",
			"SupportEmail", "SupportPhone", "FooterText",
			"DefaultDueDay", "LateFeePercentage", "InterestRateMonthly", "DaysBeforeDueReminder",
			"DaysAfterDueReminder", "AutoSendInvoiceEmail", "AutoSendCollectionReminder",
			` + prorataSelect + `,
			` + sicrediSelect + `,
			"UpdatedAt", "UpdatedBy"
		FROM "OrganizationSettings"
		WHERE "OrganizationId" = $1
	`
	var (
		res   models.OrganizationSettingsResponse
		comp  models.CompanySettingsDto
		white models.WhitelabelSettingsDto
		sys   models.SystemSettingsDto
		sec   OrgSicrediSecrets
	)
	err := q.QueryRow(ctx, query, orgID).Scan(
		&res.OrganizationID,
		&comp.CompanyName, &comp.TradingName, &comp.Cnpj, &comp.StateRegistration, &comp.Email, &comp.Phone, &comp.Website,
		&comp.ZipCode, &comp.Street, &comp.Number, &comp.Complement, &comp.Neighborhood, &comp.City, &comp.State,
		&white.AppName, &white.AppSlogan, &white.LogoUrl, &white.DarkLogoUrl, &white.FaviconUrl, &white.PrimaryColor,
		&white.SupportEmail, &white.SupportPhone, &white.FooterText,
		&sys.DefaultDueDay, &sys.LateFeePercentage, &sys.InterestRateMonthly, &sys.DaysBeforeDueReminder,
		&sys.DaysAfterDueReminder, &sys.AutoSendInvoiceEmail, &sys.AutoSendCollectionReminder,
		&sys.ProrataDivisor,
		&sec.Enabled, &sec.Sandbox, &sec.APIKey, &sec.Username, &sec.Password,
		&sec.Cooperativa, &sec.Posto, &sec.CodigoBeneficiario, &sec.AccountNumber,
		&sec.WebhookToken, &sec.PublicAPIURL,
		&res.UpdatedAt, &res.UpdatedBy,
	)
	if err != nil {
		return nil, nil, err
	}
	res.Company = comp
	res.Whitelabel = white
	res.System = sys
	if res.System.ProrataDivisor < 1 {
		res.System.ProrataDivisor = 30
	}
	res.Sicredi = sec.toPublicDTO()
	return &res, &sec, nil
}

func (s *Store) GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettingsResponse, error) {
	effectiveOrgID := s.resolveOrgID(ctx, orgID)
	res, _, err := s.scanOrganizationSettings(ctx, effectiveOrgID, true, true)
	if err != nil && isUndefinedColumn(err) {
		_ = s.ensureOrganizationSettingsSchema(ctx)
		res, _, err = s.scanOrganizationSettings(ctx, effectiveOrgID, true, true)
		if err != nil && isUndefinedColumn(err) {
			res, _, err = s.scanOrganizationSettings(ctx, effectiveOrgID, true, false)
			if err != nil && isUndefinedColumn(err) {
				res, _, err = s.scanOrganizationSettings(ctx, effectiveOrgID, false, false)
			}
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return blankOrganizationSettings(effectiveOrgID), nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Store) GetOrganizationSicrediSecrets(ctx context.Context, orgID string) (*OrgSicrediSecrets, error) {
	effectiveOrgID := s.resolveOrgID(ctx, orgID)
	_, sec, err := s.scanOrganizationSettings(ctx, effectiveOrgID, true, true)
	if err != nil && isUndefinedColumn(err) {
		_ = s.ensureOrganizationSettingsSchema(ctx)
		_, sec, err = s.scanOrganizationSettings(ctx, effectiveOrgID, true, true)
		if err != nil && isUndefinedColumn(err) {
			return &OrgSicrediSecrets{Sandbox: true}, nil
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return &OrgSicrediSecrets{Sandbox: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (s *Store) FindOrganizationIDsBySicrediWebhookToken(ctx context.Context, token string) ([]string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "OrganizationId"
		FROM "OrganizationSettings"
		WHERE "SicrediWebhookToken" = $1`, token)
	if err != nil {
		if isUndefinedColumn(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) SicrediConfigForOrg(ctx context.Context, orgID string, production bool) (sicredi.Config, bool, error) {
	sec, err := s.GetOrganizationSicrediSecrets(ctx, orgID)
	if err != nil {
		return sicredi.Config{}, false, err
	}
	cfg := sec.toSicrediConfig(production)
	return cfg, cfg.EnabledAndConfigured(), nil
}

func (s *Store) GetProrataDivisor(ctx context.Context, orgID string) (int, error) {
	effectiveOrgID := s.resolveOrgID(ctx, orgID)
	var n int
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE("ProrataDivisor", 30) FROM "OrganizationSettings" WHERE "OrganizationId" = $1`, effectiveOrgID).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) || isUndefinedColumn(err) {
		return 30, nil
	}
	if err != nil {
		return 30, err
	}
	if n < 1 {
		return 30, nil
	}
	return n, nil
}

func (s *Store) UpsertOrganizationSettings(ctx context.Context, orgID string, updatedBy *string, current *models.OrganizationSettingsResponse) error {
	effectiveOrgID := s.resolveOrgID(ctx, orgID)
	q := s.q(ctx)
	_ = s.ensureOrganizationSettingsSchema(ctx)

	sec, _ := s.GetOrganizationSicrediSecrets(ctx, effectiveOrgID)
	if sec == nil {
		sec = &OrgSicrediSecrets{Sandbox: true}
	}
	// Preserve secrets unless the caller populated them via UpdateSicrediSettings
	// which writes through UpsertOrganizationSicrediSettings.

	query := `
		INSERT INTO "OrganizationSettings" (
			"OrganizationId",
			"CompanyName", "TradingName", "Cnpj", "StateRegistration", "Email", "Phone", "Website",
			"ZipCode", "Street", "Number", "Complement", "Neighborhood", "City", "State",
			"AppName", "AppSlogan", "LogoUrl", "DarkLogoUrl", "FaviconUrl", "PrimaryColor",
			"SupportEmail", "SupportPhone", "FooterText",
			"DefaultDueDay", "LateFeePercentage", "InterestRateMonthly", "DaysBeforeDueReminder",
			"DaysAfterDueReminder", "AutoSendInvoiceEmail", "AutoSendCollectionReminder",
			"ProrataDivisor",
			"SicrediEnabled", "SicrediSandbox", "SicrediAPIKey", "SicrediUsername", "SicrediPassword",
			"SicrediCooperativa", "SicrediPosto", "SicrediCodigoBeneficiario", "SicrediAccountNumber",
			"SicrediWebhookToken", "SicrediPublicAPIURL",
			"UpdatedAt", "UpdatedBy"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			$22, $23, $24,
			$25, $26, $27, $28,
			$29, $30, $31,
			$32,
			$33, $34, $35, $36, $37,
			$38, $39, $40, $41,
			$42, $43,
			now(), $44
		)
		ON CONFLICT ("OrganizationId") DO UPDATE SET
			"CompanyName" = EXCLUDED."CompanyName",
			"TradingName" = EXCLUDED."TradingName",
			"Cnpj" = EXCLUDED."Cnpj",
			"StateRegistration" = EXCLUDED."StateRegistration",
			"Email" = EXCLUDED."Email",
			"Phone" = EXCLUDED."Phone",
			"Website" = EXCLUDED."Website",
			"ZipCode" = EXCLUDED."ZipCode",
			"Street" = EXCLUDED."Street",
			"Number" = EXCLUDED."Number",
			"Complement" = EXCLUDED."Complement",
			"Neighborhood" = EXCLUDED."Neighborhood",
			"City" = EXCLUDED."City",
			"State" = EXCLUDED."State",
			"AppName" = EXCLUDED."AppName",
			"AppSlogan" = EXCLUDED."AppSlogan",
			"LogoUrl" = EXCLUDED."LogoUrl",
			"DarkLogoUrl" = EXCLUDED."DarkLogoUrl",
			"FaviconUrl" = EXCLUDED."FaviconUrl",
			"PrimaryColor" = EXCLUDED."PrimaryColor",
			"SupportEmail" = EXCLUDED."SupportEmail",
			"SupportPhone" = EXCLUDED."SupportPhone",
			"FooterText" = EXCLUDED."FooterText",
			"DefaultDueDay" = EXCLUDED."DefaultDueDay",
			"LateFeePercentage" = EXCLUDED."LateFeePercentage",
			"InterestRateMonthly" = EXCLUDED."InterestRateMonthly",
			"DaysBeforeDueReminder" = EXCLUDED."DaysBeforeDueReminder",
			"DaysAfterDueReminder" = EXCLUDED."DaysAfterDueReminder",
			"AutoSendInvoiceEmail" = EXCLUDED."AutoSendInvoiceEmail",
			"AutoSendCollectionReminder" = EXCLUDED."AutoSendCollectionReminder",
			"ProrataDivisor" = EXCLUDED."ProrataDivisor",
			"UpdatedAt" = now(),
			"UpdatedBy" = EXCLUDED."UpdatedBy"
	`
	_, err := q.Exec(ctx, query,
		effectiveOrgID,
		current.Company.CompanyName, current.Company.TradingName, current.Company.Cnpj, current.Company.StateRegistration,
		current.Company.Email, current.Company.Phone, current.Company.Website, current.Company.ZipCode, current.Company.Street,
		current.Company.Number, current.Company.Complement, current.Company.Neighborhood, current.Company.City, current.Company.State,
		current.Whitelabel.AppName, current.Whitelabel.AppSlogan, current.Whitelabel.LogoUrl, current.Whitelabel.DarkLogoUrl,
		current.Whitelabel.FaviconUrl, current.Whitelabel.PrimaryColor, current.Whitelabel.SupportEmail, current.Whitelabel.SupportPhone,
		current.Whitelabel.FooterText,
		current.System.DefaultDueDay, current.System.LateFeePercentage, current.System.InterestRateMonthly,
		current.System.DaysBeforeDueReminder, current.System.DaysAfterDueReminder, current.System.AutoSendInvoiceEmail,
		current.System.AutoSendCollectionReminder, current.System.ProrataDivisor,
		sec.Enabled, sec.Sandbox, sec.APIKey, sec.Username, sec.Password,
		sec.Cooperativa, sec.Posto, sec.CodigoBeneficiario, sec.AccountNumber,
		sec.WebhookToken, sec.PublicAPIURL,
		updatedBy,
	)
	return err
}

func (s *Store) UpsertOrganizationSicrediSettings(ctx context.Context, orgID string, updatedBy *string, sec OrgSicrediSecrets) error {
	effectiveOrgID := s.resolveOrgID(ctx, orgID)
	_ = s.ensureOrganizationSettingsSchema(ctx)
	current, err := s.GetOrganizationSettings(ctx, effectiveOrgID)
	if err != nil {
		return err
	}
	if err := s.UpsertOrganizationSettings(ctx, effectiveOrgID, updatedBy, current); err != nil {
		return err
	}
	_, err = s.q(ctx).Exec(ctx, `
		UPDATE "OrganizationSettings" SET
			"SicrediEnabled" = $2,
			"SicrediSandbox" = $3,
			"SicrediAPIKey" = $4,
			"SicrediUsername" = $5,
			"SicrediPassword" = $6,
			"SicrediCooperativa" = $7,
			"SicrediPosto" = $8,
			"SicrediCodigoBeneficiario" = $9,
			"SicrediAccountNumber" = $10,
			"SicrediWebhookToken" = $11,
			"SicrediPublicAPIURL" = $12,
			"UpdatedAt" = now(),
			"UpdatedBy" = $13
		WHERE "OrganizationId" = $1`,
		effectiveOrgID,
		sec.Enabled, sec.Sandbox, sec.APIKey, sec.Username, sec.Password,
		sec.Cooperativa, sec.Posto, sec.CodigoBeneficiario, sec.AccountNumber,
		sec.WebhookToken, sec.PublicAPIURL, updatedBy,
	)
	return err
}
