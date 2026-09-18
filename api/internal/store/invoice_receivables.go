package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateBillingReceivable records the exact source invoices while generating the
// receivable, even when rendering/issuing the customer document fails later.
func (s *Store) CreateBillingReceivable(ctx context.Context, id, orgID, customerID, description string,
	monthID, lineID *string, issue, due time.Time, amount float64, now time.Time) error {
	return s.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = CtxWithTx(ctx, tx)
		if err := s.CreateAccountReceivable(ctx, id, orgID, customerID, description, monthID, issue, due, amount, nil, now); err != nil {
			return err
		}
		if monthID == nil {
			return nil
		}
		tag, err := tx.Exec(ctx, `INSERT INTO "InvoiceReceivableSources" ("InvoiceId", "ReceivableId")
   SELECT DISTINCT i."Id", $1 FROM "ProviderInvoices" i
   JOIN "ProviderInvoicePhoneLines" j ON j."ProviderInvoicesId" = i."Id"
   JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = j."PhoneLinesId"
   JOIN "Customers" c ON c."Id" = l."CustomerId"
   WHERE c."OrganizationId" = $2 AND c."Id" = $3 AND i."ProcessingMonthId" = $4
    AND ($5::varchar IS NULL OR l."PhoneLineId" = $5) AND i."Status" NOT IN ('cancelled','substituted')
   ON CONFLICT DO NOTHING`, id, orgID, customerID, monthID, lineID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return undoBlocked("Não há fatura importada ativa para esta cobrança. Atualize a prévia de faturamento.")
		}
		return nil
	})
}
