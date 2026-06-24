package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListProviders(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.ListProvidersResponse, int64, error) {
	q := s.q(ctx)
	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM "Providers" WHERE "OrganizationId" = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.Query(ctx, `
		SELECT "Id", "Name", "Slug", "Active"
		FROM "Providers"
		WHERE "OrganizationId" = $1
		ORDER BY "Name"
		OFFSET $2 LIMIT $3`, orgID, page.Offset(), page.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListProvidersResponse
	for rows.Next() {
		var item models.ListProvidersResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Active); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetProvider(ctx context.Context, orgID, id string) (*models.GetProviderResponse, error) {
	q := s.q(ctx)
	var p models.GetProviderResponse
	err := q.QueryRow(ctx, `
		SELECT "Id", "OrganizationId", "Name", "Slug", "Active"
		FROM "Providers"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	planRows, err := q.Query(ctx, `
		SELECT "Id", "Name", "Code"
		FROM "ProviderPlans"
		WHERE "ProviderId" = $1
		ORDER BY "Name"`, id)
	if err != nil {
		return nil, err
	}
	defer planRows.Close()

	for planRows.Next() {
		var plan models.GetProviderPlanResponse
		if err := planRows.Scan(&plan.ID, &plan.Name, &plan.Code); err != nil {
			return nil, err
		}
		svcRows, err := q.Query(ctx, `
			SELECT "Id", "Name", "Active", "Recurring", "Price"
			FROM "ProviderPlanServices"
			WHERE "ProviderPlanId" = $1
			ORDER BY "Name"`, plan.ID)
		if err != nil {
			return nil, err
		}
		for svcRows.Next() {
			var svc models.GetProviderPlanServiceResponse
			if err := svcRows.Scan(&svc.ID, &svc.Name, &svc.Active, &svc.Recurring, &svc.Price); err != nil {
				svcRows.Close()
				return nil, err
			}
			plan.Services = append(plan.Services, svc)
		}
		svcRows.Close()
		if plan.Services == nil {
			plan.Services = []models.GetProviderPlanServiceResponse{}
		}
		p.Plans = append(p.Plans, plan)
	}
	if p.Plans == nil {
		p.Plans = []models.GetProviderPlanResponse{}
	}
	return &p, planRows.Err()
}

func (s *Store) ProviderSlugExists(ctx context.Context, orgID, slug, excludeID string) (bool, error) {
	q := s.q(ctx)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM "Providers" WHERE "OrganizationId" = $1 AND "Slug" = $2`
	args := []any{orgID, slug}
	if excludeID != "" {
		query += ` AND "Id" != $3`
		args = append(args, excludeID)
	}
	query += `)`
	if err := q.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) CreateProvider(ctx context.Context, orgID, id, name, slug string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "Providers" ("Id", "OrganizationId", "Name", "Slug", "Active")
		VALUES ($1, $2, $3, $4, true)`, id, orgID, name, slug)
	return err
}

func (s *Store) UpdateProvider(ctx context.Context, orgID, id, name, slug string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Providers" SET "Name" = $3, "Slug" = $4
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, name, slug)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) InactivateProvider(ctx context.Context, orgID, id string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Providers" SET "Active" = false
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ProviderExists(ctx context.Context, orgID, id string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM "Providers" WHERE "OrganizationId" = $1 AND "Id" = $2 AND "Active" = true)`,
		orgID, id).Scan(&exists)
	return exists, err
}

func (s *Store) GetProviderByID(ctx context.Context, id string) (orgID, name string, err error) {
	err = s.q(ctx).QueryRow(ctx, `SELECT "OrganizationId", "Name" FROM "Providers" WHERE "Id" = $1`, id).
		Scan(&orgID, &name)
	return
}

type ProviderPlanRow struct {
	ID         string
	ProviderID string
	Code       string
	Name       string
}

func (s *Store) GetPlanByProviderAndCode(ctx context.Context, providerID, code string) (*ProviderPlanRow, error) {
	var p ProviderPlanRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ProviderId", "Code", "Name"
		FROM "ProviderPlans"
		WHERE "ProviderId" = $1 AND "Code" = $2`, providerID, code).
		Scan(&p.ID, &p.ProviderID, &p.Code, &p.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) CreateProviderPlan(ctx context.Context, id, providerID, code, name string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderPlans" ("Id", "ProviderId", "Code", "Name")
		VALUES ($1, $2, $3, $4)`, id, providerID, code, name)
	return err
}

func (s *Store) ProviderPlanCodeExists(ctx context.Context, providerID, code, excludeID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "ProviderPlans" WHERE "ProviderId" = $1 AND "Code" = $2`
	args := []any{providerID, code}
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

func (s *Store) GetProviderPlan(ctx context.Context, orgID, providerID, planID string) (*models.GetProviderPlanResponse, error) {
	var exists bool
	if err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProviderPlans" pp
			JOIN "Providers" p ON p."Id" = pp."ProviderId"
			WHERE p."OrganizationId" = $1 AND pp."ProviderId" = $2 AND pp."Id" = $3
		)`, orgID, providerID, planID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	var plan models.GetProviderPlanResponse
	if err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code"
		FROM "ProviderPlans"
		WHERE "Id" = $1`, planID).
		Scan(&plan.ID, &plan.Name, &plan.Code); err != nil {
		return nil, err
	}

	svcRows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "Name", "Active", "Recurring", "Price"
		FROM "ProviderPlanServices"
		WHERE "ProviderPlanId" = $1
		ORDER BY "Name"`, planID)
	if err != nil {
		return nil, err
	}
	defer svcRows.Close()
	for svcRows.Next() {
		var svc models.GetProviderPlanServiceResponse
		if err := svcRows.Scan(&svc.ID, &svc.Name, &svc.Active, &svc.Recurring, &svc.Price); err != nil {
			return nil, err
		}
		plan.Services = append(plan.Services, svc)
	}
	if plan.Services == nil {
		plan.Services = []models.GetProviderPlanServiceResponse{}
	}
	return &plan, svcRows.Err()
}

func (s *Store) UpdateProviderPlan(ctx context.Context, providerID, planID, code, name string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProviderPlans" SET "Code" = $3, "Name" = $4
		WHERE "ProviderId" = $1 AND "Id" = $2`, providerID, planID, code, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetPlanServiceByPlanAndName(ctx context.Context, planID, name string) (id string, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "Id" FROM "ProviderPlanServices"
		WHERE "ProviderPlanId" = $1 AND "Name" = $2`, planID, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) CreatePlanService(ctx context.Context, id, planID, name string, recurring bool, price *float64) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderPlanServices" ("Id", "ProviderPlanId", "Name", "Active", "Recurring", "Price")
		VALUES ($1, $2, $3, true, $4, $5)`, id, planID, name, recurring, price)
	return err
}

func (s *Store) UpdatePlanServicePrice(ctx context.Context, serviceID string, price *float64) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProviderPlanServices"
		SET "Price" = $2, "Recurring" = true, "Active" = true
		WHERE "Id" = $1`, serviceID, price)
	return err
}

type ContractingCompanyRow struct {
	ID         string
	ProviderID string
	LegalName  string
	TaxID      string
}

func (s *Store) GetContractingCompanyByTaxID(ctx context.Context, providerID, taxID string) (*ContractingCompanyRow, error) {
	var c ContractingCompanyRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ProviderId", "LegalName", "TaxId"
		FROM "ContractingCompanies"
		WHERE "ProviderId" = $1 AND "TaxId" = $2`, providerID, taxID).
		Scan(&c.ID, &c.ProviderID, &c.LegalName, &c.TaxID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateContractingCompany(ctx context.Context, id, providerID, legalName, taxID string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ContractingCompanies" ("Id", "ProviderId", "LegalName", "TaxId")
		VALUES ($1, $2, $3, $4)`, id, providerID, legalName, taxID)
	return err
}

type ProviderAccountRow struct {
	ID                    string
	ContractingCompanyID  string
	AccountNumber         string
}

func (s *Store) GetProviderAccount(ctx context.Context, companyID, accountNumber string) (*ProviderAccountRow, error) {
	var a ProviderAccountRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ContractingCompanyId", "AccountNumber"
		FROM "ProviderAccounts"
		WHERE "ContractingCompanyId" = $1 AND "AccountNumber" = $2`, companyID, accountNumber).
		Scan(&a.ID, &a.ContractingCompanyID, &a.AccountNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) CreateProviderAccount(ctx context.Context, id, companyID, accountNumber string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderAccounts" ("Id", "ContractingCompanyId", "AccountNumber")
		VALUES ($1, $2, $3)`, id, companyID, accountNumber)
	return err
}

func (s *Store) CountDashboardProviders(ctx context.Context, orgID string) (int32, error) {
	var n int32
	err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*)::int FROM "Providers" WHERE "OrganizationId" = $1`, orgID).Scan(&n)
	return n, err
}
