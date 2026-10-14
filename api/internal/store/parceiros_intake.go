package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ParceirosImportRow struct {
	ID                   string
	CustomerID           string
	PhoneLineID          *string
	CustomerCreated      bool
	PhoneLineCreated     bool
	AttachmentsAccepted  int
	AttachmentsSkipped   int
	Warnings             []string
}

type ParceirosShell struct {
	ProviderID string
	AccountID  string
	PlanID     string
}

func (s *Store) FindParceirosImport(ctx context.Context, orgID, externalSaleID string) (*ParceirosImportRow, error) {
	var row ParceirosImportRow
	var warnings []byte
	var lineID *string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "CustomerId", "PhoneLineId", "CustomerCreated", "PhoneLineCreated",
			"AttachmentsAccepted", "AttachmentsSkipped", "Warnings"
		FROM "ParceirosSaleImports"
		WHERE "OrganizationId" = $1 AND "ExternalSaleId" = $2`, orgID, externalSaleID).
		Scan(&row.ID, &row.CustomerID, &lineID, &row.CustomerCreated, &row.PhoneLineCreated,
			&row.AttachmentsAccepted, &row.AttachmentsSkipped, &warnings)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	row.PhoneLineID = lineID
	_ = json.Unmarshal(warnings, &row.Warnings)
	if row.Warnings == nil {
		row.Warnings = []string{}
	}
	return &row, nil
}

func (s *Store) FindCustomerIDByDocument(ctx context.Context, orgID, document string) (string, error) {
	var id string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT c."Id"
		FROM "Customers" c
		JOIN "CustomerDocuments" cd ON cd."CustomerId" = c."Id"
		WHERE c."OrganizationId" = $1 AND cd."Number" = $2
		ORDER BY c."Active" DESC, c."Id"
		LIMIT 1`, orgID, document).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) InsertParceirosCustomer(ctx context.Context, orgID, id, customerType, name, legalName, rawAddress, rawDocument string, reviewPending bool) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "Customers" (
			"Id", "OrganizationId", "Name", "LegalName", "Type", "Active", "Origin", "RawAddress", "RawDocument", "ReviewPending"
		) VALUES ($1, $2, $3, NULLIF($4, ''), $5::customer_type, true, 'luxus_parceiros', NULLIF($6, ''), NULLIF($7, ''), $8)`,
		id, orgID, name, legalName, customerType, rawAddress, rawDocument, reviewPending)
	return err
}

func (s *Store) InsertParceirosDocument(ctx context.Context, customerID, documentType, number string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerDocuments" ("Id", "CustomerId", "DocumentType", "Number")
		VALUES ($1, $2, $3::customer_document_type, $4)`,
		newUUID(), customerID, documentType, number)
	return err
}

func (s *Store) EnsureParceirosShell(ctx context.Context, orgID string) (ParceirosShell, error) {
	var shell ParceirosShell
	err := s.q(ctx).QueryRow(ctx, `SELECT "Id" FROM "Providers" WHERE "OrganizationId" = $1 AND "Slug" = 'luxus-parceiros'`, orgID).Scan(&shell.ProviderID)
	if errors.Is(err, pgx.ErrNoRows) {
		shell.ProviderID = newUUID()
		if _, err = s.q(ctx).Exec(ctx, `
			INSERT INTO "Providers" ("Id", "OrganizationId", "Name", "Slug", "Active")
			VALUES ($1, $2, 'Luxus Parceiros', 'luxus-parceiros', true)`, shell.ProviderID, orgID); err != nil {
			return shell, err
		}
	} else if err != nil {
		return shell, err
	}

	var companyID string
	err = s.q(ctx).QueryRow(ctx, `SELECT "Id" FROM "ContractingCompanies" WHERE "ProviderId" = $1 AND "TaxId" = '00000000000000'`, shell.ProviderID).Scan(&companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		companyID = newUUID()
		if _, err = s.q(ctx).Exec(ctx, `
			INSERT INTO "ContractingCompanies" ("Id", "ProviderId", "LegalName", "TaxId")
			VALUES ($1, $2, 'Luxus Parceiros', '00000000000000')`, companyID, shell.ProviderID); err != nil {
			return shell, err
		}
	} else if err != nil {
		return shell, err
	}

	err = s.q(ctx).QueryRow(ctx, `SELECT "Id" FROM "ProviderAccounts" WHERE "ContractingCompanyId" = $1 AND "AccountNumber" = 'parceiros'`, companyID).Scan(&shell.AccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		shell.AccountID = newUUID()
		if _, err = s.q(ctx).Exec(ctx, `
			INSERT INTO "ProviderAccounts" ("Id", "ContractingCompanyId", "AccountNumber")
			VALUES ($1, $2, 'parceiros')`, shell.AccountID, companyID); err != nil {
			return shell, err
		}
	} else if err != nil {
		return shell, err
	}

	err = s.q(ctx).QueryRow(ctx, `SELECT "Id" FROM "ProviderPlans" WHERE "ProviderId" = $1 AND "Code" = 'PARCEIROS'`, shell.ProviderID).Scan(&shell.PlanID)
	if errors.Is(err, pgx.ErrNoRows) {
		shell.PlanID = newUUID()
		_, err = s.q(ctx).Exec(ctx, `
			INSERT INTO "ProviderPlans" ("Id", "ProviderId", "Name", "Code")
			VALUES ($1, $2, 'Plano informado pelo Parceiros', 'PARCEIROS')`, shell.PlanID, shell.ProviderID)
	}
	return shell, err
}

type ParceirosPhoneLine struct {
	ID             string
	OrganizationID string
}

func (s *Store) FindParceirosPhoneLine(ctx context.Context, normalized string) (*ParceirosPhoneLine, error) {
	var row ParceirosPhoneLine
	err := s.q(ctx).QueryRow(ctx, `
		SELECT pl."Id", p."OrganizationId"
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE pl."NormalizedNumber" = $1
		LIMIT 1`, normalized).Scan(&row.ID, &row.OrganizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &row, err
}

func (s *Store) ActivePhoneLineCustomer(ctx context.Context, phoneLineID string) (string, error) {
	var customerID string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "CustomerId" FROM "PhoneLineCustomerLinks"
		WHERE "PhoneLineId" = $1 AND "EndDate" IS NULL
		LIMIT 1`, phoneLineID).Scan(&customerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return customerID, err
}

func (s *Store) LinkParceirosPhoneLine(ctx context.Context, phoneLineID, customerID string, start time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PhoneLineCustomerLinks" ("Id", "PhoneLineId", "CustomerId", "StartDate")
		VALUES ($1, $2, $3, $4)`, newUUID(), phoneLineID, customerID, start)
	return err
}

func (s *Store) UpdateParceirosPhoneLine(ctx context.Context, id, operatorName, planName, simType, iccid, imei, eid, platform, portabilityNumber, donor string, virgin, portability bool) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLines"
		SET "Status" = 'active'::phone_line_status,
			"ExternalSource" = 'luxus_parceiros',
			"ExternalOperatorName" = NULLIF($2, ''),
			"ExternalPlanName" = NULLIF($3, ''),
			"SimType" = $4,
			"ICCID" = NULLIF($5, ''),
			"DeviceImei" = NULLIF($6, ''),
			"DeviceEid" = NULLIF($7, ''),
			"DevicePlatform" = NULLIF($8, ''),
			"IsVirginChip" = $9,
			"IsPortability" = $10,
			"PortabilityNumber" = NULLIF($11, ''),
			"DonorOperator" = NULLIF($12, '')
		WHERE "Id" = $1`,
		id, operatorName, planName, simType, iccid, imei, eid, platform, virgin, portability, portabilityNumber, donor)
	return err
}

type ParceirosImportInsert struct {
	ID                  string
	OrganizationID      string
	ExternalSaleID      string
	Protocol            string
	TaskProtocol        string
	CustomerID          string
	PhoneLineID         *string
	CustomerCreated     bool
	PhoneLineCreated    bool
	AttachmentsAccepted int
	AttachmentsSkipped  int
	PartnerName         string
	BranchName          string
	OperatorName        string
	PlanName            string
	LineNumber          string
	Value               *float64
	CommissionValue     *float64
	SoldAt              *time.Time
	ConcludedAt         *time.Time
	SimType             string
	Iccid               string
	ContractFormat      string
	IsUpgrade           bool
	UpgradeOfProtocol   string
	IsPortability       bool
	PortabilityNumber   string
	DonorOperator       string
	Warnings            []string
	Payload             []byte
}

func (s *Store) InsertParceirosImport(ctx context.Context, row ParceirosImportInsert) error {
	warnings, err := json.Marshal(row.Warnings)
	if err != nil {
		return err
	}
	if row.Payload == nil {
		row.Payload = []byte("{}")
	}
	_, err = s.q(ctx).Exec(ctx, `
		INSERT INTO "ParceirosSaleImports" (
			"Id", "OrganizationId", "ExternalSaleId", "Protocol", "TaskProtocol", "CustomerId", "PhoneLineId",
			"CustomerCreated", "PhoneLineCreated", "AttachmentsAccepted", "AttachmentsSkipped",
			"PartnerName", "BranchName", "OperatorName", "PlanName", "LineNumber", "Value", "CommissionValue",
			"SoldAt", "ConcludedAt", "SimType", "Iccid", "ContractFormat", "IsUpgrade", "UpgradeOfProtocol",
			"IsPortability", "PortabilityNumber", "DonorOperator", "Warnings", "Payload"
		) VALUES (
			$1,$2,$3,$4,NULLIF($5,''),$6,$7,
			$8,$9,$10,$11,
			NULLIF($12,''),NULLIF($13,''),NULLIF($14,''),NULLIF($15,''),NULLIF($16,''),$17,$18,
			$19,$20,NULLIF($21,''),NULLIF($22,''),NULLIF($23,''),$24,NULLIF($25,''),
			$26,NULLIF($27,''),NULLIF($28,''),$29::jsonb,$30::jsonb
		)`,
		row.ID, row.OrganizationID, row.ExternalSaleID, row.Protocol, row.TaskProtocol, row.CustomerID, row.PhoneLineID,
		row.CustomerCreated, row.PhoneLineCreated, row.AttachmentsAccepted, row.AttachmentsSkipped,
		row.PartnerName, row.BranchName, row.OperatorName, row.PlanName, row.LineNumber, row.Value, row.CommissionValue,
		row.SoldAt, row.ConcludedAt, row.SimType, row.Iccid, row.ContractFormat, row.IsUpgrade, row.UpgradeOfProtocol,
		row.IsPortability, row.PortabilityNumber, row.DonorOperator, string(warnings), string(row.Payload))
	return err
}

func (s *Store) ParceirosAttachmentExists(ctx context.Context, importID, externalDocumentID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ParceirosSaleAttachments"
			WHERE "ImportId" = $1 AND "ExternalDocumentId" = $2
		)`, importID, externalDocumentID).Scan(&exists)
	return exists, err
}

func (s *Store) InsertParceirosAttachmentLink(ctx context.Context, importID, externalDocumentID, attachmentID string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ParceirosSaleAttachments" ("Id", "ImportId", "ExternalDocumentId", "AttachmentId")
		VALUES ($1, $2, $3, NULLIF($4, ''))`, newUUID(), importID, externalDocumentID, attachmentID)
	return err
}
