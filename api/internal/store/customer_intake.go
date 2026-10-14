package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

const intakeSelect = `SELECT "Id", "SourceSystem", "CustomerId", "SaleId", "SaleData", "ClientData", "AttachmentData", "ImportedAt", "PayloadHash" FROM "CustomerIntakes" `

func scanIntake(row pgx.Row) (*models.CustomerIntakeRecord, error) {
	var r models.CustomerIntakeRecord
	var sale, client, attachments []byte
	err := row.Scan(&r.ID, &r.SourceSystem, &r.CustomerID, &r.SaleID, &sale, &client, &attachments, &r.ImportedAt, &r.PayloadHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(sale, &r.Sale); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(client, &r.Client); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(attachments, &r.Attachments); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) GetCustomerIntake(ctx context.Context, org, source, externalID string) (*models.CustomerIntakeRecord, error) {
	return scanIntake(s.q(ctx).QueryRow(ctx, intakeSelect+`WHERE "OrganizationId"=$1 AND "SourceSystem"=$2 AND "ExternalSaleId"=$3`, org, source, externalID))
}

func (s *Store) GetSaleIntake(ctx context.Context, org, saleID string) (*models.CustomerIntakeRecord, error) {
	return scanIntake(s.q(ctx).QueryRow(ctx, intakeSelect+`WHERE "OrganizationId"=$1 AND "SaleId"=$2`, org, saleID))
}

func (s *Store) ListCustomerIntakes(ctx context.Context, org, customerID string) ([]models.CustomerIntakeRecord, error) {
	rows, err := s.q(ctx).Query(ctx, intakeSelect+`WHERE "OrganizationId"=$1 AND "CustomerId"=$2 ORDER BY "ImportedAt" DESC LIMIT 100`, org, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.CustomerIntakeRecord{}
	for rows.Next() {
		r, err := scanIntake(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *r)
	}
	return items, rows.Err()
}

// Includes inactive customers, so an import cannot silently recreate an inactivated CPF/CNPJ.
func (s *Store) IntakeCustomersByDocument(ctx context.Context, org, document string) ([]string, error) {
	rows, err := s.q(ctx).Query(ctx, `SELECT DISTINCT c."Id" FROM "Customers" c JOIN "CustomerDocuments" d ON d."CustomerId"=c."Id" WHERE c."OrganizationId"=$1 AND d."DocumentType" IN ('cpf'::customer_document_type,'cnpj'::customer_document_type) AND regexp_replace(d."Number", '[^0-9]', '', 'g')=$2`, org, document)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) SaveCustomerIntake(ctx context.Context, org string, r models.CustomerIntakeRecord) error {
	sale, err := json.Marshal(r.Sale)
	if err != nil {
		return err
	}
	attachments, err := json.Marshal(r.Attachments)
	if err != nil {
		return err
	}
	client, err := json.Marshal(r.Client)
	if err != nil {
		return err
	}
	_, err = s.q(ctx).Exec(ctx, `INSERT INTO "CustomerIntakes" ("Id","OrganizationId","CustomerId","SaleId","SourceSystem","ExternalSaleId","PayloadHash","SaleData","ClientData","AttachmentData","ValueCents","CommissionCents","ImportedAt") VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb,$10::jsonb,$11,$12,$13)`, r.ID, org, r.CustomerID, r.SaleID, r.SourceSystem, r.Sale.ID, r.PayloadHash, sale, client, attachments, int64(*r.Sale.Value), int64(*r.Sale.CommissionValue), r.ImportedAt)
	return err
}

// Imported sales are confirmed historical records. This does not issue invoices,
// settle payments, allocate stock or send a new signature request.
func (s *Store) CreateIntakeSale(ctx context.Context, org, customerID, saleID, number, actor, itemID string, sale models.CustomerIntakeSale, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `INSERT INTO "Sales" ("Id","OrganizationId","CustomerId","SalespersonUserId","Status","SaleNumber","SoldAt","Notes","TotalAmount","CreatedAt","UpdatedAt") VALUES ($1,$2,$3,$4,'confirmed'::sale_status,$5,$6,$7,$8::numeric,$9,$9)`, saleID, org, customerID, actor, number, sale.SoldAt, sale.Notes, sale.Value.String(), now)
	if err != nil {
		return err
	}
	_, err = s.q(ctx).Exec(ctx, `INSERT INTO "SaleLineItems" ("Id","SaleId","LineItemType","Description","Quantity","UnitPrice","TotalPrice","SortOrder") VALUES ($1,$2,'other'::sale_line_item_type,$3,1,$4::numeric,$4::numeric,0)`, itemID, saleID, "Venda importada: "+sale.PlanName, sale.Value.String())
	return err
}
