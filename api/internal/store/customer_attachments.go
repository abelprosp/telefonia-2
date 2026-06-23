package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListCustomerAttachments(ctx context.Context, orgID, customerID string) ([]models.CustomerAttachmentResponse, error) {
	ok, err := s.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil || !ok {
		return nil, err
	}
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "Title", "OriginalFileName", "StorageBucket", "StorageObjectKey",
			"ContentType", "SizeBytes", "UploadedAtUtc"
		FROM "CustomerAttachments"
		WHERE "OrganizationId" = $1 AND "CustomerId" = $2
		ORDER BY "UploadedAtUtc"`, orgID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.CustomerAttachmentResponse
	for rows.Next() {
		var item models.CustomerAttachmentResponse
		if err := rows.Scan(&item.ID, &item.Title, &item.OriginalFileName, &item.StorageBucket,
			&item.StorageObjectKey, &item.ContentType, &item.SizeBytes, &item.UploadedAtUTC); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.CustomerAttachmentResponse{}
	}
	return items, rows.Err()
}

func (s *Store) CreateCustomerAttachment(ctx context.Context, orgID, customerID, id string, input models.RegisterCustomerAttachmentInput) (*models.CustomerAttachmentResponse, error) {
	now := time.Now().UTC()
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerAttachments" ("Id", "OrganizationId", "CustomerId", "Title",
			"OriginalFileName", "StorageBucket", "StorageObjectKey", "ContentType", "SizeBytes", "UploadedAtUtc")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id, orgID, customerID, input.Title, input.OriginalFileName, input.StorageBucket,
		input.StorageObjectKey, input.ContentType, input.SizeBytes, now)
	if err != nil {
		return nil, err
	}
	return &models.CustomerAttachmentResponse{
		ID: id, Title: input.Title, OriginalFileName: input.OriginalFileName,
		StorageBucket: input.StorageBucket, StorageObjectKey: input.StorageObjectKey,
		ContentType: input.ContentType, SizeBytes: input.SizeBytes, UploadedAtUTC: now,
	}, nil
}

func (s *Store) DeleteCustomerAttachment(ctx context.Context, orgID, customerID, attachmentID string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		DELETE FROM "CustomerAttachments"
		WHERE "OrganizationId" = $1 AND "CustomerId" = $2 AND "Id" = $3`, orgID, customerID, attachmentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetBillingReadiness(ctx context.Context, orgID, customerID, processingMonthID string) (*models.GetCustomerBillingReadinessResponse, error) {
	ok, err := s.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil || !ok {
		return nil, nil
	}

	var providerID string
	var monthExists bool
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "ProviderId" FROM "ProcessingMonths"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, processingMonthID).Scan(&providerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	monthExists = true
	_ = monthExists

	hasProvider, err := s.CustomerHasActiveProvider(ctx, customerID, providerID)
	if err != nil || !hasProvider {
		return nil, nil
	}

	cnpj, _ := s.GetCustomerCNPJ(ctx, customerID)
	cnpjDigits := normalizeDigits(cnpj)
	usesCnpj := len(cnpjDigits) == 14

	var accountsExpected, accountsWithInvoice int
	if usesCnpj {
		rows, err := s.q(ctx).Query(ctx, `
			SELECT pa."Id" FROM "ProviderAccounts" pa
			JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			WHERE cc."ProviderId" = $1 AND cc."TaxId" = $2`, providerID, cnpjDigits)
		if err != nil {
			return nil, err
		}
		var accountIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			accountIDs = append(accountIDs, id)
		}
		rows.Close()
		accountsExpected = len(accountIDs)
		if accountsExpected > 0 {
			err = s.q(ctx).QueryRow(ctx, `
				SELECT COUNT(DISTINCT "ProviderAccountId")::int
				FROM "ProviderInvoices"
				WHERE "ProcessingMonthId" = $1 AND "ProviderAccountId" = ANY($2)`,
				processingMonthID, accountIDs).Scan(&accountsWithInvoice)
			if err != nil {
				return nil, err
			}
		}
	}

	var manualJustification, manualUser string
	var manualAt time.Time
	var manualFound bool
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "Justification", "ReleasedAt", "ReleasedByUserId"
		FROM "CustomerProcessingMonthManualReleases"
		WHERE "OrganizationId" = $1 AND "CustomerId" = $2 AND "ProcessingMonthId" = $3`,
		orgID, customerID, processingMonthID).Scan(&manualJustification, &manualAt, &manualUser)
	if err == nil {
		manualFound = true
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	automaticComplete := usesCnpj && accountsExpected > 0 && accountsWithInvoice == accountsExpected
	isManuallyReleased := manualFound
	isReleased := isManuallyReleased || automaticComplete

	statusName := "Pendente"
	if isReleased {
		statusName = "Liberado para faturamento"
	}

	resp := &models.GetCustomerBillingReadinessResponse{
		CustomerID:       customerID,
		ProcessingMonthID: processingMonthID,
		StatusDisplayName: statusName,
		IsReleasedForBilling: isReleased,
		IsAutomaticallyComplete: automaticComplete,
		IsManuallyReleased: isManuallyReleased,
		AutomaticEvaluationUsesCnpjContractingCompanies: usesCnpj,
		AccountsExpectedForAutomaticRule: accountsExpected,
		AccountsWithInvoiceInProcessingMonth: accountsWithInvoice,
	}
	if manualFound {
		resp.ManualRelease = &models.BillingReadinessManualReleaseDto{
			Justification: manualJustification,
			ReleasedAt: manualAt,
			ReleasedByUserID: manualUser,
		}
	}
	return resp, nil
}

func (s *Store) ManualReleaseExists(ctx context.Context, orgID, customerID, processingMonthID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "CustomerProcessingMonthManualReleases"
			WHERE "OrganizationId" = $1 AND "CustomerId" = $2 AND "ProcessingMonthId" = $3)`,
		orgID, customerID, processingMonthID).Scan(&exists)
	return exists, err
}

func (s *Store) CreateManualRelease(ctx context.Context, orgID, customerID, processingMonthID, id, justification, userID string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerProcessingMonthManualReleases"
		("Id", "OrganizationId", "CustomerId", "ProcessingMonthId", "Justification", "ReleasedAt", "ReleasedByUserId")
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, orgID, customerID, processingMonthID, justification, time.Now().UTC(), userID)
	return err
}

func normalizeDigits(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b = append(b, s[i])
		}
	}
	return string(b)
}
