package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (s *Service) ListCustomerDevices(ctx context.Context, customerID string, page httputil.PageSearch) ([]models.CustomerDeviceLinkResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, 0, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, 0, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	return s.Store.ListCustomerDeviceLinks(ctx, orgID, customerID, page)
}

func (s *Service) AssignCustomerDevice(ctx context.Context, customerID string, input models.AssignCustomerDeviceInput) (*models.CustomerDeviceLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	if input.MonthlyAmount < 0 {
		return nil, httputil.ValidationError(notifications.CustomerDeviceMonthlyAmountInvalid)
	}

	var deviceStockID *string
	brand := strings.TrimSpace(ptrStr(input.Brand))
	model := strings.TrimSpace(ptrStr(input.Model))
	description := strings.TrimSpace(ptrStr(input.Description))

	if input.DeviceStockItemID != nil && strings.TrimSpace(*input.DeviceStockItemID) != "" {
		stockID := strings.TrimSpace(*input.DeviceStockItemID)
		device, err := s.Store.GetDeviceStockItem(ctx, orgID, stockID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if device == nil {
			return nil, httputil.NotFoundError(notifications.DeviceStockNotFound)
		}
		if device.Status != "in_stock" {
			return nil, httputil.BusinessError(notifications.N("DEVICE_STOCK_NOT_AVAILABLE", "Device is not available in stock."))
		}
		linkedCustomer, err := s.Store.DeviceStockItemActiveCustomerLink(ctx, stockID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if linkedCustomer != "" && linkedCustomer != customerID {
			return nil, httputil.BusinessError(notifications.CustomerDeviceAlreadyLinked)
		}
		deviceStockID = &stockID
		brand = device.Brand
		model = device.Model
		if description == "" {
			description = fmt.Sprintf("Aparelho %s %s", brand, model)
			if device.StorageCapacity != nil && *device.StorageCapacity != "" {
				description += " " + *device.StorageCapacity
			}
		}
		if input.MonthlyAmount == 0 && device.SalePrice != nil && *device.SalePrice > 0 {
			input.MonthlyAmount = *device.SalePrice
		}
	} else {
		if brand == "" || model == "" {
			return nil, httputil.ValidationError(notifications.DeviceStockBrandRequired)
		}
		if description == "" {
			description = fmt.Sprintf("Aparelho %s %s", brand, model)
		}
	}

	if description == "" {
		return nil, httputil.ValidationError(notifications.CustomerDeviceDescriptionRequired)
	}

	start := time.Now().UTC()
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		parsed, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		start = parsed
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateCustomerDeviceLink(ctx, id, customerID, deviceStockID, description, brand, model, input.MonthlyAmount, start, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.Store.GetCustomerDeviceLink(ctx, orgID, customerID, id)
}

func (s *Service) UpdateCustomerDevice(ctx context.Context, customerID, linkID string, input models.UpdateCustomerDeviceInput) (*models.CustomerDeviceLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.MonthlyAmount != nil && *input.MonthlyAmount < 0 {
		return nil, httputil.ValidationError(notifications.CustomerDeviceMonthlyAmountInvalid)
	}
	var description *string
	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		if trimmed == "" {
			return nil, httputil.ValidationError(notifications.CustomerDeviceDescriptionRequired)
		}
		description = &trimmed
	}
	if err := s.Store.UpdateCustomerDeviceLink(ctx, orgID, customerID, linkID, description, input.MonthlyAmount); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.CustomerDeviceNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.Store.GetCustomerDeviceLink(ctx, orgID, customerID, linkID)
}

func (s *Service) UnassignCustomerDevice(ctx context.Context, customerID, linkID string, input models.UnassignCustomerDeviceInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	end := time.Now().UTC()
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		parsed, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return err
		}
		end = parsed
	}
	if err := s.Store.EndCustomerDeviceLink(ctx, orgID, customerID, linkID, end); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.CustomerDeviceNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func ptrStr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
