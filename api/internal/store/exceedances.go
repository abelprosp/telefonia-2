package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type ExceedanceTermRow struct {
	ID              string
	OrganizationID  string
	Term            string
	ChargeType      string
	TabulatedAmount *float64
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DetectedExceedanceRow struct {
	ID            string
	InvoiceID     string
	PhoneLineID   *string
	TermID        *string
	Term          string
	Description   string
	InvoiceAmount float64
	ChargedAmount float64
	ChargeType    string
	Applied       bool
	CreatedAt     time.Time
}

type PhoneLineExceedanceSettings struct {
	ID                   string
	ChargeExceedances    bool
	ExceedanceChargeType string
}

func NormalizeExceedanceChargeType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "tabulated", "tabelado":
		return "tabulated"
	case "mirrored", "mirroed", "espelhado":
		return "mirroed"
	default:
		return ""
	}
}

func ExceedanceChargeTypeAPI(dbValue string) string {
	switch strings.ToLower(strings.TrimSpace(dbValue)) {
	case "tabulated":
		return "tabulated"
	default:
		return "mirrored"
	}
}

func (s *Store) ListExceedanceTerms(ctx context.Context, orgID string, activeOnly bool) ([]models.ExceedanceTermResponse, error) {
	q := `SELECT "Id", "Term", "ChargeType"::text, "TabulatedAmount", "Active", "CreatedAt", "UpdatedAt"
		FROM "ExceedanceTerms" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if activeOnly {
		q += ` AND "Active" = true`
	}
	q += ` ORDER BY "Term"`
	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ExceedanceTermResponse
	for rows.Next() {
		var item models.ExceedanceTermResponse
		var chargeType string
		if err := rows.Scan(&item.ID, &item.Term, &chargeType, &item.TabulatedAmount, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.ChargeType = ExceedanceChargeTypeAPI(chargeType)
		items = append(items, item)
	}
	if items == nil {
		items = []models.ExceedanceTermResponse{}
	}
	return items, rows.Err()
}

func (s *Store) GetExceedanceTerm(ctx context.Context, orgID, id string) (*models.ExceedanceTermResponse, error) {
	var item models.ExceedanceTermResponse
	var chargeType string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Term", "ChargeType"::text, "TabulatedAmount", "Active", "CreatedAt", "UpdatedAt"
		FROM "ExceedanceTerms" WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(&item.ID, &item.Term, &chargeType, &item.TabulatedAmount, &item.Active, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.ChargeType = ExceedanceChargeTypeAPI(chargeType)
	return &item, nil
}

func (s *Store) CreateExceedanceTerm(ctx context.Context, row ExceedanceTermRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ExceedanceTerms" (
			"Id", "OrganizationId", "Term", "ChargeType", "TabulatedAmount", "Active", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4::exceedance_charge_type, $5, $6, $7, $8)`,
		row.ID, row.OrganizationID, row.Term, row.ChargeType, row.TabulatedAmount, row.Active, row.CreatedAt, row.UpdatedAt)
	return err
}

func (s *Store) UpdateExceedanceTerm(ctx context.Context, orgID, id, term, chargeType string, tabulated *float64, active *bool, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "ExceedanceTerms"
		SET "Term" = COALESCE(NULLIF($3, ''), "Term"),
			"ChargeType" = COALESCE(NULLIF($4, '')::exceedance_charge_type, "ChargeType"),
			"TabulatedAmount" = COALESCE($5, "TabulatedAmount"),
			"Active" = COALESCE($6, "Active"),
			"UpdatedAt" = $7
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, term, chargeType, tabulated, active, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetPhoneLineExceedanceSettings(ctx context.Context, phoneLineID string) (*PhoneLineExceedanceSettings, error) {
	var row PhoneLineExceedanceSettings
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", COALESCE("ChargeExceedances", true), COALESCE("ExceedanceChargeType"::text, 'mirroed')
		FROM "PhoneLines" WHERE "Id" = $1`, phoneLineID).
		Scan(&row.ID, &row.ChargeExceedances, &row.ExceedanceChargeType)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &row, err
}

func (s *Store) UpdatePhoneLineExceedanceSettings(ctx context.Context, phoneLineID string, charge *bool, chargeType *string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLines"
		SET "ChargeExceedances" = COALESCE($2, "ChargeExceedances"),
			"ExceedanceChargeType" = COALESCE($3::exceedance_charge_type, "ExceedanceChargeType")
		WHERE "Id" = $1`, phoneLineID, charge, chargeType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) InsertDetectedExceedance(ctx context.Context, row DetectedExceedanceRow) (bool, error) {
	tag, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "InvoiceDetectedExceedances" (
			"Id", "InvoiceId", "PhoneLineId", "TermId", "Term", "Description",
			"InvoiceAmount", "ChargedAmount", "ChargeType", "Applied", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT DO NOTHING`,
		row.ID, row.InvoiceID, row.PhoneLineID, row.TermID, row.Term, row.Description,
		row.InvoiceAmount, row.ChargedAmount, row.ChargeType, row.Applied, row.CreatedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) GetPrimaryProcessingIDForLine(ctx context.Context, phoneLineID string) (string, error) {
	var id string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT pr."Id"
		FROM "LineBillingProcessings" pr
		JOIN "PhoneLineCustomerLinks" l ON l."Id" = pr."PhoneLineCustomerLinkId" AND l."EndDate" IS NULL
		WHERE l."PhoneLineId" = $1
			AND pr."Perspective" = 'luxus_customer'::billing_processing_perspective
			AND pr."Active" = true
		LIMIT 1`, phoneLineID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) GetMirroredSecondaryProcessingID(ctx context.Context, phoneLineID string) (string, error) {
	var id string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT pr."Id"
		FROM "LineBillingProcessings" pr
		JOIN "PhoneLineCustomerLinks" l ON l."Id" = pr."PhoneLineCustomerLinkId" AND l."EndDate" IS NULL
		WHERE l."PhoneLineId" = $1
			AND pr."Perspective" = 'customer_end_user'::billing_processing_perspective
			AND pr."Active" = true AND pr."MirrorFromPrimary" = true
		LIMIT 1`, phoneLineID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) OrgIDForPhoneLine(ctx context.Context, phoneLineID string) (string, error) {
	var orgID string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT p."OrganizationId"
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE pl."Id" = $1`, phoneLineID).Scan(&orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return orgID, err
}
