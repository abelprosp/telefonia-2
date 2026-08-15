package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) GetOrganizationSettings(ctx context.Context, orgID string) (*models.OrganizationSettingsResponse, error) {
	q := s.q(ctx)
	query := `
		SELECT "OrganizationId",
			"CompanyName", "TradingName", "Cnpj", "StateRegistration", "Email", "Phone", "Website",
			"ZipCode", "Street", "Number", "Complement", "Neighborhood", "City", "State",
			"AppName", "AppSlogan", "LogoUrl", "DarkLogoUrl", "FaviconUrl", "PrimaryColor",
			"SupportEmail", "SupportPhone", "FooterText",
			"DefaultDueDay", "LateFeePercentage", "InterestRateMonthly", "DaysBeforeDueReminder",
			"DaysAfterDueReminder", "AutoSendInvoiceEmail", "AutoSendCollectionReminder",
			"UpdatedAt", "UpdatedBy"
		FROM "OrganizationSettings"
		WHERE "OrganizationId" = $1
	`
	var (
		res models.OrganizationSettingsResponse
		comp models.CompanySettingsDto
		white models.WhitelabelSettingsDto
		sys models.SystemSettingsDto
	)

	err := q.QueryRow(ctx, query, orgID).Scan(
		&res.OrganizationID,
		&comp.CompanyName, &comp.TradingName, &comp.Cnpj, &comp.StateRegistration, &comp.Email, &comp.Phone, &comp.Website,
		&comp.ZipCode, &comp.Street, &comp.Number, &comp.Complement, &comp.Neighborhood, &comp.City, &comp.State,
		&white.AppName, &white.AppSlogan, &white.LogoUrl, &white.DarkLogoUrl, &white.FaviconUrl, &white.PrimaryColor,
		&white.SupportEmail, &white.SupportPhone, &white.FooterText,
		&sys.DefaultDueDay, &sys.LateFeePercentage, &sys.InterestRateMonthly, &sys.DaysBeforeDueReminder,
		&sys.DaysAfterDueReminder, &sys.AutoSendInvoiceEmail, &sys.AutoSendCollectionReminder,
		&res.UpdatedAt, &res.UpdatedBy,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		// Retorna defaults se ainda não houver registro
		return &models.OrganizationSettingsResponse{
			OrganizationID: orgID,
			Company: models.CompanySettingsDto{
				CompanyName:       "Luxus Telefonia Ltda",
				TradingName:       "Luxus Connect",
				Cnpj:              "11.309.896/0001-01",
				StateRegistration: "",
				Email:             "contato@luxusconnect.com.br",
				Phone:             "(11) 99999-9999",
				Website:           "https://telefonia.redobrai.online",
				ZipCode:           "",
				Street:            "",
				Number:            "",
				Complement:        "",
				Neighborhood:      "",
				City:              "",
				State:             "",
			},
			Whitelabel: models.WhitelabelSettingsDto{
				AppName:      "Luxus Connect",
				AppSlogan:    "Gestão Inteligente de Telefonia",
				LogoUrl:      "",
				DarkLogoUrl:  "",
				FaviconUrl:   "",
				PrimaryColor: "#0f766e",
				SupportEmail: "suporte@luxusconnect.com.br",
				SupportPhone: "(11) 99999-9999",
				FooterText:   "© 2026 Luxus Connect. Todos os direitos reservados.",
			},
			System: models.SystemSettingsDto{
				DefaultDueDay:              10,
				LateFeePercentage:          2.00,
				InterestRateMonthly:        1.00,
				DaysBeforeDueReminder:      3,
				DaysAfterDueReminder:       2,
				AutoSendInvoiceEmail:       true,
				AutoSendCollectionReminder: false,
			},
			UpdatedAt: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	res.Company = comp
	res.Whitelabel = white
	res.System = sys
	return &res, nil
}

func (s *Store) UpsertOrganizationSettings(ctx context.Context, orgID string, updatedBy *string, current *models.OrganizationSettingsResponse) error {
	q := s.q(ctx)
	query := `
		INSERT INTO "OrganizationSettings" (
			"OrganizationId",
			"CompanyName", "TradingName", "Cnpj", "StateRegistration", "Email", "Phone", "Website",
			"ZipCode", "Street", "Number", "Complement", "Neighborhood", "City", "State",
			"AppName", "AppSlogan", "LogoUrl", "DarkLogoUrl", "FaviconUrl", "PrimaryColor",
			"SupportEmail", "SupportPhone", "FooterText",
			"DefaultDueDay", "LateFeePercentage", "InterestRateMonthly", "DaysBeforeDueReminder",
			"DaysAfterDueReminder", "AutoSendInvoiceEmail", "AutoSendCollectionReminder",
			"UpdatedAt", "UpdatedBy"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			$22, $23, $24,
			$25, $26, $27, $28,
			$29, $30, $31,
			now(), $32
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
			"UpdatedAt" = now(),
			"UpdatedBy" = EXCLUDED."UpdatedBy"
	`
	_, err := q.Exec(ctx, query,
		orgID,
		current.Company.CompanyName, current.Company.TradingName, current.Company.Cnpj, current.Company.StateRegistration,
		current.Company.Email, current.Company.Phone, current.Company.Website, current.Company.ZipCode, current.Company.Street,
		current.Company.Number, current.Company.Complement, current.Company.Neighborhood, current.Company.City, current.Company.State,
		current.Whitelabel.AppName, current.Whitelabel.AppSlogan, current.Whitelabel.LogoUrl, current.Whitelabel.DarkLogoUrl,
		current.Whitelabel.FaviconUrl, current.Whitelabel.PrimaryColor, current.Whitelabel.SupportEmail, current.Whitelabel.SupportPhone,
		current.Whitelabel.FooterText,
		current.System.DefaultDueDay, current.System.LateFeePercentage, current.System.InterestRateMonthly,
		current.System.DaysBeforeDueReminder, current.System.DaysAfterDueReminder, current.System.AutoSendInvoiceEmail,
		current.System.AutoSendCollectionReminder,
		updatedBy,
	)
	return err
}
