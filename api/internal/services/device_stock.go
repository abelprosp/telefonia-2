package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

var deviceStockStatuses = map[string]struct{}{
	"in_stock": {},
	"sold":     {},
	"inactive": {},
}

var skuSanitizeRe = regexp.MustCompile(`[^A-Za-z0-9]+`)

func (s *Service) ListDeviceStockItems(ctx context.Context, status *string, page httputil.PageSearch) ([]models.ListDeviceStockItemResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		normalized := strings.TrimSpace(strings.ToLower(*status))
		if _, ok := deviceStockStatuses[normalized]; !ok {
			return nil, 0, httputil.ValidationError(notifications.DeviceStockStatusInvalid)
		}
		status = &normalized
	}
	return s.Store.ListDeviceStockItems(ctx, orgID, status, page)
}

func (s *Service) GetDeviceStockItem(ctx context.Context, id string) (*models.GetDeviceStockItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	item, err := s.Store.GetDeviceStockItem(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if item == nil {
		return nil, httputil.NotFoundError(notifications.DeviceStockNotFound)
	}
	return item, nil
}

func (s *Service) CreateDeviceStockItem(ctx context.Context, input models.CreateDeviceStockItemInput) (*models.GetDeviceStockItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	brand, model, sku, imei, err := validateDeviceStockCreateInput(input)
	if err != nil {
		return nil, err
	}

	if sku == "" {
		sku = generateDeviceSku(brand, model)
	}

	dupSku, err := s.Store.DeviceStockSkuExists(ctx, orgID, sku, "")
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dupSku {
		return nil, httputil.BusinessError(notifications.DeviceStockSkuDuplicated)
	}

	if imei != "" {
		dupImei, err := s.Store.DeviceStockImeiExists(ctx, orgID, imei, "")
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if dupImei {
			return nil, httputil.BusinessError(notifications.DeviceStockImeiDuplicated)
		}
	}

	now := time.Now().UTC()
	id := uuid.New().String()
	row := store.DeviceStockRow{
		ID:              id,
		OrganizationID:  orgID,
		Sku:             sku,
		Brand:           brand,
		Model:           model,
		Imei:            nullableString(imei),
		Color:           trimOptionalString(input.Color),
		StorageCapacity: trimOptionalString(input.StorageCapacity),
		UnitCost:        input.UnitCost,
		SalePrice:       input.SalePrice,
		Status:          "in_stock",
		Notes:           trimOptionalString(input.Notes),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.Store.CreateDeviceStockItem(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetDeviceStockItem(ctx, id)
}

func (s *Service) UpdateDeviceStockItem(ctx context.Context, id string, input models.UpdateDeviceStockItemInput) (*models.GetDeviceStockItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.Store.GetDeviceStockItem(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if existing == nil {
		return nil, httputil.NotFoundError(notifications.DeviceStockNotFound)
	}

	brand := existing.Brand
	if input.Brand != nil {
		brand = strings.TrimSpace(*input.Brand)
		if brand == "" {
			return nil, httputil.ValidationError(notifications.DeviceStockBrandRequired)
		}
		if utf8.RuneCountInString(brand) > 128 {
			return nil, httputil.ValidationError(notifications.DeviceStockBrandMaxLength)
		}
	}

	model := existing.Model
	if input.Model != nil {
		model = strings.TrimSpace(*input.Model)
		if model == "" {
			return nil, httputil.ValidationError(notifications.DeviceStockModelRequired)
		}
		if utf8.RuneCountInString(model) > 256 {
			return nil, httputil.ValidationError(notifications.DeviceStockModelMaxLength)
		}
	}

	sku := existing.Sku
	if input.Sku != nil {
		sku = strings.TrimSpace(*input.Sku)
		if sku == "" {
			sku = generateDeviceSku(brand, model)
		}
		if utf8.RuneCountInString(sku) > 64 {
			return nil, httputil.ValidationError(notifications.DeviceStockSkuMaxLength)
		}
		dupSku, err := s.Store.DeviceStockSkuExists(ctx, orgID, sku, id)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if dupSku {
			return nil, httputil.BusinessError(notifications.DeviceStockSkuDuplicated)
		}
	}

	imei := existing.Imei
	if input.Imei != nil {
		normalized := httputil.NormalizeDigits(strings.TrimSpace(*input.Imei))
		if normalized != "" {
			if len(normalized) != 15 {
				return nil, httputil.ValidationError(notifications.DeviceStockImeiInvalid)
			}
			dupImei, err := s.Store.DeviceStockImeiExists(ctx, orgID, normalized, id)
			if err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			if dupImei {
				return nil, httputil.BusinessError(notifications.DeviceStockImeiDuplicated)
			}
		}
		imei = nullableString(normalized)
	}

	status := existing.Status
	if input.Status != nil {
		normalized := strings.TrimSpace(strings.ToLower(*input.Status))
		if _, ok := deviceStockStatuses[normalized]; !ok {
			return nil, httputil.ValidationError(notifications.DeviceStockStatusInvalid)
		}
		status = normalized
	}

	color := existing.Color
	if input.Color != nil {
		color = trimOptionalString(input.Color)
	}
	storage := existing.StorageCapacity
	if input.StorageCapacity != nil {
		storage = trimOptionalString(input.StorageCapacity)
	}
	unitCost := existing.UnitCost
	if input.UnitCost != nil {
		unitCost = input.UnitCost
	}
	salePrice := existing.SalePrice
	if input.SalePrice != nil {
		salePrice = input.SalePrice
	}
	notes := existing.Notes
	if input.Notes != nil {
		notes = trimOptionalString(input.Notes)
	}

	row := store.DeviceStockRow{
		Sku:             sku,
		Brand:           brand,
		Model:           model,
		Imei:            imei,
		Color:           color,
		StorageCapacity: storage,
		UnitCost:        unitCost,
		SalePrice:       salePrice,
		Status:          status,
		Notes:           notes,
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.Store.UpdateDeviceStockItem(ctx, orgID, id, row); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.DeviceStockNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetDeviceStockItem(ctx, id)
}

func validateDeviceStockCreateInput(input models.CreateDeviceStockItemInput) (brand, model, sku, imei string, err error) {
	brand = strings.TrimSpace(input.Brand)
	if brand == "" {
		return "", "", "", "", httputil.ValidationError(notifications.DeviceStockBrandRequired)
	}
	if utf8.RuneCountInString(brand) > 128 {
		return "", "", "", "", httputil.ValidationError(notifications.DeviceStockBrandMaxLength)
	}

	model = strings.TrimSpace(input.Model)
	if model == "" {
		return "", "", "", "", httputil.ValidationError(notifications.DeviceStockModelRequired)
	}
	if utf8.RuneCountInString(model) > 256 {
		return "", "", "", "", httputil.ValidationError(notifications.DeviceStockModelMaxLength)
	}

	if input.Sku != nil {
		sku = strings.TrimSpace(*input.Sku)
		if utf8.RuneCountInString(sku) > 64 {
			return "", "", "", "", httputil.ValidationError(notifications.DeviceStockSkuMaxLength)
		}
	}

	if input.Imei != nil {
		imei = httputil.NormalizeDigits(strings.TrimSpace(*input.Imei))
		if imei != "" && len(imei) != 15 {
			return "", "", "", "", httputil.ValidationError(notifications.DeviceStockImeiInvalid)
		}
	}
	return brand, model, sku, imei, nil
}

func generateDeviceSku(brand, model string) string {
	base := strings.ToUpper(skuSanitizeRe.ReplaceAllString(fmt.Sprintf("%s-%s", brand, model), "-"))
	base = strings.Trim(base, "-")
	if base == "" {
		base = "APARELHO"
	}
	if len(base) > 48 {
		base = base[:48]
	}
	return fmt.Sprintf("%s-%s", base, strings.ToUpper(uuid.New().String()[:6]))
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
