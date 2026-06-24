package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type CustomerBillingItemRow struct {
	Description string
	ItemType    string
	Amount      float64
}

func (s *Store) ListCustomerDeviceLinks(ctx context.Context, orgID, customerID string, page httputil.PageSearch) ([]models.CustomerDeviceLinkResponse, int64, error) {
	base := `
		FROM "CustomerDeviceLinks" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE c."OrganizationId" = $1 AND d."CustomerId" = $2`
	args := []any{orgID, customerID}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offsetParam := len(args) + 1
	limitParam := len(args) + 2
	rows, err := s.q(ctx).Query(ctx, `
		SELECT d."Id", d."CustomerId", d."DeviceStockItemId", d."Description", d."Brand", d."Model",
			d."MonthlyAmount", d."StartDate", d."EndDate"
		`+base+`
		ORDER BY d."StartDate" DESC
		OFFSET $`+itoa(offsetParam)+` LIMIT $`+itoa(limitParam),
		append(args, page.Offset(), page.Limit())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanCustomerDeviceLinks(rows, total)
}

func scanCustomerDeviceLinks(rows pgx.Rows, total int64) ([]models.CustomerDeviceLinkResponse, int64, error) {
	var items []models.CustomerDeviceLinkResponse
	for rows.Next() {
		var item models.CustomerDeviceLinkResponse
		var endDate *time.Time
		if err := rows.Scan(
			&item.ID, &item.CustomerID, &item.DeviceStockItemID, &item.Description, &item.Brand, &item.Model,
			&item.MonthlyAmount, &item.StartDate, &endDate,
		); err != nil {
			return nil, 0, err
		}
		item.EndDate = endDate
		item.IsActive = endDate == nil
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetCustomerDeviceLink(ctx context.Context, orgID, customerID, linkID string) (*models.CustomerDeviceLinkResponse, error) {
	var item models.CustomerDeviceLinkResponse
	var endDate *time.Time
	err := s.q(ctx).QueryRow(ctx, `
		SELECT d."Id", d."CustomerId", d."DeviceStockItemId", d."Description", d."Brand", d."Model",
			d."MonthlyAmount", d."StartDate", d."EndDate"
		FROM "CustomerDeviceLinks" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE c."OrganizationId" = $1 AND d."CustomerId" = $2 AND d."Id" = $3`,
		orgID, customerID, linkID).
		Scan(
			&item.ID, &item.CustomerID, &item.DeviceStockItemID, &item.Description, &item.Brand, &item.Model,
			&item.MonthlyAmount, &item.StartDate, &endDate,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.EndDate = endDate
	item.IsActive = endDate == nil
	return &item, nil
}

func (s *Store) DeviceStockItemActiveCustomerLink(ctx context.Context, deviceStockItemID string) (string, error) {
	var customerID string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "CustomerId" FROM "CustomerDeviceLinks"
		WHERE "DeviceStockItemId" = $1 AND "EndDate" IS NULL`, deviceStockItemID).Scan(&customerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return customerID, err
}

func (s *Store) CreateCustomerDeviceLink(ctx context.Context, id, customerID string, deviceStockItemID *string, description, brand, model string, monthlyAmount float64, startDate time.Time, createdAt time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerDeviceLinks" (
			"Id", "CustomerId", "DeviceStockItemId", "Description", "Brand", "Model",
			"MonthlyAmount", "StartDate", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, $9)`,
		id, customerID, deviceStockItemID, description, brand, model, monthlyAmount, startDate, createdAt)
	return err
}

func (s *Store) UpdateCustomerDeviceLink(ctx context.Context, orgID, customerID, linkID string, description *string, monthlyAmount *float64) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerDeviceLinks" d
		SET "Description" = COALESCE($4, d."Description"),
			"MonthlyAmount" = COALESCE($5, d."MonthlyAmount")
		FROM "Customers" c
		WHERE c."Id" = d."CustomerId" AND c."OrganizationId" = $1
			AND d."CustomerId" = $2 AND d."Id" = $3 AND d."EndDate" IS NULL`,
		orgID, customerID, linkID, description, monthlyAmount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) EndCustomerDeviceLink(ctx context.Context, orgID, customerID, linkID string, endDate time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerDeviceLinks" d
		SET "EndDate" = $4::date
		FROM "Customers" c
		WHERE c."Id" = d."CustomerId" AND c."OrganizationId" = $1
			AND d."CustomerId" = $2 AND d."Id" = $3 AND d."EndDate" IS NULL`,
		orgID, customerID, linkID, endDate)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListCustomerBillingItems(ctx context.Context, customerID string) ([]CustomerBillingItemRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT
			COALESCE(pp."Name", pl."Number") AS description,
			'Mensal' AS item_type,
			COALESCE(l."MonthlyAmount", 0) AS amount
		FROM "PhoneLineCustomerLinks" l
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		LEFT JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
		WHERE l."CustomerId" = $1 AND l."EndDate" IS NULL
		UNION ALL
		SELECT d."Description", 'Aparelho', d."MonthlyAmount"
		FROM "CustomerDeviceLinks" d
		WHERE d."CustomerId" = $1 AND d."EndDate" IS NULL
		ORDER BY 1`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CustomerBillingItemRow
	for rows.Next() {
		var item CustomerBillingItemRow
		if err := rows.Scan(&item.Description, &item.ItemType, &item.Amount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
