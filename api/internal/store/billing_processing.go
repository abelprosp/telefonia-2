package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/billingcalc"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func signedCompositionAmount(itemType string, amount, quantity float64) float64 {
	return billingcalc.SignedItemAmount(itemType, amount, quantity)
}

func itemActiveInCycle(cycleStart, start, end *time.Time) bool {
	return billingcalc.ActiveDays(cycleStart, start, end) > 0
}

const lineLuxusBillingAmountSQL = `
COALESCE(
	NULLIF((
		SELECT SUM(
			CASE bp."ItemType"
				WHEN 'discount' THEN -bp."Amount" * COALESCE(bp."Quantity", 1)
				ELSE bp."Amount" * COALESCE(bp."Quantity", 1)
			END
		)
		FROM "LineBillingProcessings" pr
		JOIN "LineBillingCompositionItems" bp ON bp."ProcessingId" = pr."Id" AND bp."Active" = true
		WHERE pr."PhoneLineCustomerLinkId" = l."Id"
			AND pr."Perspective" = 'luxus_customer'::billing_processing_perspective
			AND pr."Active" = true
	), 0),
	CASE
		WHEN COALESCE(l."MonthlyAmount", 0) > 0 THEN l."MonthlyAmount"
		ELSE COALESCE(pl."CostWithConsumption", pl."BaseCost", 0)
	END
)`

type BillingProcessingRow struct {
	ID                      string
	PhoneLineCustomerLinkID string
	Perspective             string
	Label                   *string
	MirrorFromPrimary       bool
	OrganizationalUnit      *string
	Department              *string
	CostCenterLabel         *string
	Active                  bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type BillingCompositionItemRow struct {
	ID                    string
	ProcessingID          string
	ItemType              string
	Description           string
	Amount                float64
	Quantity              float64
	InstallmentCount      *int
	InstallmentCurrent    *int
	StartDate             *time.Time
	EndDate               *time.Time
	Active                bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ServiceType           *string
	ProviderPlanServiceID *string
	Proportional          bool
}

func (s *Store) ListActiveCustomerLinkIDs(ctx context.Context, customerID string) ([]string, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id" FROM "PhoneLineCustomerLinks"
		WHERE "CustomerId" = $1 AND "EndDate" IS NULL`, customerID)
	if err != nil {
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

func (s *Store) CustomerIsReseller(ctx context.Context, customerID string) (bool, error) {
	var isReseller bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE("IsReseller", false) FROM "Customers" WHERE "Id" = $1`, customerID).Scan(&isReseller)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isReseller, err
}

func (s *Store) UpdateCustomerIsReseller(ctx context.Context, orgID, customerID string, isReseller bool) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Customers" SET "IsReseller" = $3
		WHERE "Id" = $1 AND "OrganizationId" = $2`, customerID, orgID, isReseller)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetActiveLinkIDForPhoneLine(ctx context.Context, orgID, phoneLineID string) (string, error) {
	var linkID string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT l."Id"
		FROM "PhoneLineCustomerLinks" l
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE l."PhoneLineId" = $1 AND l."EndDate" IS NULL AND p."OrganizationId" = $2`,
		phoneLineID, orgID).Scan(&linkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return linkID, err
}

func (s *Store) GetBillingProcessing(ctx context.Context, orgID, processingID string) (*BillingProcessingRow, error) {
	var row BillingProcessingRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT pr."Id", pr."PhoneLineCustomerLinkId", pr."Perspective"::text, pr."Label",
			pr."MirrorFromPrimary", pr."OrganizationalUnit", pr."Department", pr."CostCenterLabel",
			pr."Active", pr."CreatedAt", pr."UpdatedAt"
		FROM "LineBillingProcessings" pr
		JOIN "PhoneLineCustomerLinks" l ON l."Id" = pr."PhoneLineCustomerLinkId"
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE pr."Id" = $1 AND p."OrganizationId" = $2 AND pr."Active" = true`,
		processingID, orgID).Scan(
		&row.ID, &row.PhoneLineCustomerLinkID, &row.Perspective, &row.Label,
		&row.MirrorFromPrimary, &row.OrganizationalUnit, &row.Department, &row.CostCenterLabel,
		&row.Active, &row.CreatedAt, &row.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) ListBillingProcessingsForLink(ctx context.Context, linkID string) ([]BillingProcessingRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "PhoneLineCustomerLinkId", "Perspective"::text, "Label",
			"MirrorFromPrimary", "OrganizationalUnit", "Department", "CostCenterLabel",
			"Active", "CreatedAt", "UpdatedAt"
		FROM "LineBillingProcessings"
		WHERE "PhoneLineCustomerLinkId" = $1 AND "Active" = true
		ORDER BY CASE "Perspective" WHEN 'luxus_customer' THEN 0 ELSE 1 END`,
		linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillingProcessingRow
	for rows.Next() {
		var row BillingProcessingRow
		if err := rows.Scan(&row.ID, &row.PhoneLineCustomerLinkID, &row.Perspective, &row.Label,
			&row.MirrorFromPrimary, &row.OrganizationalUnit, &row.Department, &row.CostCenterLabel,
			&row.Active, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	if items == nil {
		items = []BillingProcessingRow{}
	}
	return items, rows.Err()
}

func (s *Store) CreateBillingProcessing(ctx context.Context, row BillingProcessingRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "LineBillingProcessings" (
			"Id", "PhoneLineCustomerLinkId", "Perspective", "Label",
			"MirrorFromPrimary", "OrganizationalUnit", "Department", "CostCenterLabel",
			"Active", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3::billing_processing_perspective, $4, $5, $6, $7, $8, $9, $10, $11)`,
		row.ID, row.PhoneLineCustomerLinkID, row.Perspective, row.Label,
		row.MirrorFromPrimary, row.OrganizationalUnit, row.Department, row.CostCenterLabel,
		row.Active, row.CreatedAt, row.UpdatedAt)
	return err
}

func (s *Store) UpdateBillingProcessing(ctx context.Context, id string, label *string, mirrorFromPrimary *bool, orgUnit, department, costCenter *string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "LineBillingProcessings"
		SET "Label" = COALESCE($2, "Label"),
			"MirrorFromPrimary" = COALESCE($3, "MirrorFromPrimary"),
			"OrganizationalUnit" = COALESCE($4, "OrganizationalUnit"),
			"Department" = COALESCE($5, "Department"),
			"CostCenterLabel" = COALESCE($6, "CostCenterLabel"),
			"UpdatedAt" = $7
		WHERE "Id" = $1 AND "Active" = true`, id, label, mirrorFromPrimary, orgUnit, department, costCenter, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) SumBillingProcessingTotal(ctx context.Context, processingID string) (float64, error) {
	return s.SumBillingProcessingTotalForCycle(ctx, processingID, nil)
}

func (s *Store) SumBillingProcessingTotalForCycle(ctx context.Context, processingID string, cycleStart *time.Time) (float64, error) {
	items, err := s.ListBillingCompositionItems(ctx, processingID)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, it := range items {
		signed := signedCompositionAmount(it.ItemType, it.Amount, it.Quantity)
		if it.Proportional && cycleStart != nil {
			if signed < 0 {
				signed = -billingcalc.ProRataAmount(-signed, cycleStart, it.StartDate, it.EndDate)
			} else {
				signed = billingcalc.ProRataAmount(signed, cycleStart, it.StartDate, it.EndDate)
			}
		} else if cycleStart != nil && !itemActiveInCycle(cycleStart, it.StartDate, it.EndDate) {
			signed = 0
		}
		total += signed
	}
	return total, nil
}

func (s *Store) ListBillingCompositionItems(ctx context.Context, processingID string) ([]BillingCompositionItemRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "ProcessingId", "ItemType"::text, "Description", "Amount", "Quantity",
			"InstallmentCount", "InstallmentCurrent", "StartDate", "EndDate", "Active", "CreatedAt", "UpdatedAt",
			"ServiceType"::text, "ProviderPlanServiceId", COALESCE("Proportional", true)
		FROM "LineBillingCompositionItems"
		WHERE "ProcessingId" = $1 AND "Active" = true
		ORDER BY "CreatedAt"`, processingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillingCompositionItemRow
	for rows.Next() {
		var row BillingCompositionItemRow
		if err := rows.Scan(&row.ID, &row.ProcessingID, &row.ItemType, &row.Description, &row.Amount, &row.Quantity,
			&row.InstallmentCount, &row.InstallmentCurrent, &row.StartDate, &row.EndDate, &row.Active, &row.CreatedAt, &row.UpdatedAt,
			&row.ServiceType, &row.ProviderPlanServiceID, &row.Proportional); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	if items == nil {
		items = []BillingCompositionItemRow{}
	}
	return items, rows.Err()
}

func (s *Store) GetBillingCompositionItem(ctx context.Context, orgID, itemID string) (*BillingCompositionItemRow, error) {
	var row BillingCompositionItemRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT ci."Id", ci."ProcessingId", ci."ItemType"::text, ci."Description", ci."Amount", ci."Quantity",
			ci."InstallmentCount", ci."InstallmentCurrent", ci."StartDate", ci."EndDate", ci."Active", ci."CreatedAt", ci."UpdatedAt",
			ci."ServiceType"::text, ci."ProviderPlanServiceId", COALESCE(ci."Proportional", true)
		FROM "LineBillingCompositionItems" ci
		JOIN "LineBillingProcessings" pr ON pr."Id" = ci."ProcessingId"
		JOIN "PhoneLineCustomerLinks" l ON l."Id" = pr."PhoneLineCustomerLinkId"
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE ci."Id" = $1 AND p."OrganizationId" = $2 AND ci."Active" = true`,
		itemID, orgID).Scan(&row.ID, &row.ProcessingID, &row.ItemType, &row.Description, &row.Amount, &row.Quantity,
		&row.InstallmentCount, &row.InstallmentCurrent, &row.StartDate, &row.EndDate, &row.Active, &row.CreatedAt, &row.UpdatedAt,
		&row.ServiceType, &row.ProviderPlanServiceID, &row.Proportional)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) CreateBillingCompositionItem(ctx context.Context, row BillingCompositionItemRow) error {
	if !row.Proportional && row.CreatedAt.IsZero() {
		row.Proportional = true
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "LineBillingCompositionItems" (
			"Id", "ProcessingId", "ItemType", "Description", "Amount", "Quantity",
			"InstallmentCount", "InstallmentCurrent", "StartDate", "EndDate", "Active", "CreatedAt", "UpdatedAt",
			"ServiceType", "ProviderPlanServiceId", "Proportional"
		) VALUES ($1, $2, $3::billing_composition_item_type, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16)`,
		row.ID, row.ProcessingID, row.ItemType, row.Description, row.Amount, row.Quantity,
		row.InstallmentCount, row.InstallmentCurrent, row.StartDate, row.EndDate, row.Active, row.CreatedAt, row.UpdatedAt,
		nullableServiceType(row.ServiceType), row.ProviderPlanServiceID, row.Proportional)
	return err
}

func nullableServiceType(v *string) any {
	if v == nil || *v == "" {
		return nil
	}
	return *v
}

func (s *Store) UpdateBillingCompositionItem(ctx context.Context, id string, description *string, amount, quantity *float64,
	installmentCount, installmentCurrent *int, startDate, endDate *time.Time, now time.Time, serviceType *string, proportional *bool) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "LineBillingCompositionItems"
		SET "Description" = COALESCE($2, "Description"),
			"Amount" = COALESCE($3, "Amount"),
			"Quantity" = COALESCE($4, "Quantity"),
			"InstallmentCount" = COALESCE($5, "InstallmentCount"),
			"InstallmentCurrent" = COALESCE($6, "InstallmentCurrent"),
			"StartDate" = COALESCE($7, "StartDate"),
			"EndDate" = COALESCE($8, "EndDate"),
			"UpdatedAt" = $9,
			"ServiceType" = COALESCE($10, "ServiceType"),
			"Proportional" = COALESCE($11, "Proportional")
		WHERE "Id" = $1 AND "Active" = true`,
		id, description, amount, quantity, installmentCount, installmentCurrent, startDate, endDate, now,
		nullableServiceType(serviceType), proportional)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeactivateBillingCompositionItem(ctx context.Context, id string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "LineBillingCompositionItems"
		SET "Active" = false, "UpdatedAt" = $2
		WHERE "Id" = $1 AND "Active" = true`, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeactivateBillingCompositionItemsForProcessing(ctx context.Context, processingID string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "LineBillingCompositionItems"
		SET "Active" = false, "UpdatedAt" = $2
		WHERE "ProcessingId" = $1 AND "Active" = true`, processingID, now)
	return err
}

func (s *Store) ToBillingProcessingResponse(ctx context.Context, row BillingProcessingRow) (models.LineBillingProcessingResponse, error) {
	total, err := s.SumBillingProcessingTotal(ctx, row.ID)
	if err != nil {
		return models.LineBillingProcessingResponse{}, err
	}
	items, err := s.ListBillingCompositionItems(ctx, row.ID)
	if err != nil {
		return models.LineBillingProcessingResponse{}, err
	}
	resp := models.LineBillingProcessingResponse{
		ID:                 row.ID,
		Perspective:        row.Perspective,
		Label:              row.Label,
		MirrorFromPrimary:  row.MirrorFromPrimary,
		OrganizationalUnit: row.OrganizationalUnit,
		Department:         row.Department,
		CostCenterLabel:    row.CostCenterLabel,
		TotalAmount:        total,
		Items:              make([]models.LineBillingCompositionItemResponse, 0, len(items)),
	}
	for _, it := range items {
		resp.Items = append(resp.Items, CompositionItemToModel(it))
	}
	return resp, nil
}

func CompositionItemToModel(row BillingCompositionItemRow) models.LineBillingCompositionItemResponse {
	return models.LineBillingCompositionItemResponse{
		ID:                    row.ID,
		ItemType:              row.ItemType,
		Description:           row.Description,
		Amount:                row.Amount,
		Quantity:              row.Quantity,
		InstallmentCount:      row.InstallmentCount,
		InstallmentCurrent:    row.InstallmentCurrent,
		StartDate:             row.StartDate,
		EndDate:               row.EndDate,
		ServiceType:           row.ServiceType,
		ProviderPlanServiceID: row.ProviderPlanServiceID,
		Proportional:          row.Proportional,
	}
}

func (s *Store) ActiveCompositionServiceTypeExists(ctx context.Context, processingID, serviceType, excludeItemID string) (bool, error) {
	if serviceType == "" {
		return false, nil
	}
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "LineBillingCompositionItems"
			WHERE "ProcessingId" = $1 AND "Active" = true
				AND "ItemType" = 'service'::billing_composition_item_type
				AND "ServiceType" = $2::service_type
				AND ($3 = '' OR "Id" != $3)
		)`, processingID, serviceType, excludeItemID).Scan(&exists)
	return exists, err
}
