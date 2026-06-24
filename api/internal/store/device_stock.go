package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type DeviceStockRow struct {
	ID              string
	OrganizationID  string
	Sku             string
	Brand           string
	Model           string
	Imei            *string
	Color           *string
	StorageCapacity *string
	UnitCost        *float64
	SalePrice       *float64
	Status          string
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (s *Store) ListDeviceStockItems(ctx context.Context, orgID string, status *string, page httputil.PageSearch) ([]models.ListDeviceStockItemResponse, int64, error) {
	base := ` FROM "DeviceStockItems" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if status != nil && *status != "" {
		base += ` AND "Status" = $2::device_stock_status`
		args = append(args, *status)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offsetParam := len(args) + 1
	limitParam := len(args) + 2
	selectQ := `
		SELECT "Id", "Sku", "Brand", "Model", "Imei", "Color", "StorageCapacity",
			"UnitCost", "SalePrice", "Status"::text, "Notes", "CreatedAt", "UpdatedAt"
		` + base + `
		ORDER BY "CreatedAt" DESC
		OFFSET $` + itoa(offsetParam) + ` LIMIT $` + itoa(limitParam)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListDeviceStockItemResponse
	for rows.Next() {
		var item models.ListDeviceStockItemResponse
		if err := rows.Scan(
			&item.ID, &item.Sku, &item.Brand, &item.Model, &item.Imei, &item.Color, &item.StorageCapacity,
			&item.UnitCost, &item.SalePrice, &item.Status, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetDeviceStockItem(ctx context.Context, orgID, id string) (*models.GetDeviceStockItemResponse, error) {
	var item models.GetDeviceStockItemResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Sku", "Brand", "Model", "Imei", "Color", "StorageCapacity",
			"UnitCost", "SalePrice", "Status"::text, "Notes", "CreatedAt", "UpdatedAt"
		FROM "DeviceStockItems"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(
			&item.ID, &item.Sku, &item.Brand, &item.Model, &item.Imei, &item.Color, &item.StorageCapacity,
			&item.UnitCost, &item.SalePrice, &item.Status, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) DeviceStockSkuExists(ctx context.Context, orgID, sku, excludeID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "DeviceStockItems" WHERE "OrganizationId" = $1 AND "Sku" = $2`
	args := []any{orgID, sku}
	if excludeID != "" {
		query += ` AND "Id" != $3`
		args = append(args, excludeID)
	}
	query += `)`
	var exists bool
	if err := s.q(ctx).QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) DeviceStockImeiExists(ctx context.Context, orgID, imei, excludeID string) (bool, error) {
	if imei == "" {
		return false, nil
	}
	query := `SELECT EXISTS(SELECT 1 FROM "DeviceStockItems" WHERE "OrganizationId" = $1 AND "Imei" = $2`
	args := []any{orgID, imei}
	if excludeID != "" {
		query += ` AND "Id" != $3`
		args = append(args, excludeID)
	}
	query += `)`
	var exists bool
	if err := s.q(ctx).QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) CreateDeviceStockItem(ctx context.Context, row DeviceStockRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "DeviceStockItems" (
			"Id", "OrganizationId", "Sku", "Brand", "Model", "Imei", "Color", "StorageCapacity",
			"UnitCost", "SalePrice", "Status", "Notes", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::device_stock_status, $12, $13, $14)`,
		row.ID, row.OrganizationID, row.Sku, row.Brand, row.Model, row.Imei, row.Color, row.StorageCapacity,
		row.UnitCost, row.SalePrice, row.Status, row.Notes, row.CreatedAt, row.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateDeviceStockItem(ctx context.Context, orgID, id string, row DeviceStockRow) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "DeviceStockItems"
		SET "Sku" = $3, "Brand" = $4, "Model" = $5, "Imei" = $6, "Color" = $7, "StorageCapacity" = $8,
			"UnitCost" = $9, "SalePrice" = $10, "Status" = $11::device_stock_status, "Notes" = $12, "UpdatedAt" = $13
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, row.Sku, row.Brand, row.Model, row.Imei, row.Color, row.StorageCapacity,
		row.UnitCost, row.SalePrice, row.Status, row.Notes, row.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
