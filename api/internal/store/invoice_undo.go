package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func undoBlocked(message string) error {
	return httputil.BusinessError(notifications.N("INVOICE_UNDO_BLOCKED", message))
}

func (s *Store) BeginInvoiceUndoTracking(ctx context.Context, id string) error {
	if _, ok := ctx.Value(txCtxKey{}).(pgx.Tx); !ok {
		return errors.New("import tracking requires a transaction")
	}
	if _, err := s.q(ctx).Exec(ctx, `SELECT set_config('telefonia.import_invoice_id', $1, true)`, id); err != nil {
		return err
	}
	_, err := s.q(ctx).Exec(ctx, `UPDATE "ProviderInvoices" SET "UndoTracked" = true WHERE "Id" = $1`, id)
	return err
}

func (s *Store) EndInvoiceUndoTracking(ctx context.Context) error {
	_, err := s.q(ctx).Exec(ctx, `SELECT set_config('telefonia.import_invoice_id', '', true)`)
	return err
}

// CancelProviderInvoice reverses the import and its uncollected financial outputs
// in one transaction. The invoice and change journal remain as an audit trail.
func (s *Store) CancelProviderInvoice(ctx context.Context, orgID, id, actorID string) error {
	err := s.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = CtxWithTx(ctx, tx)
		var providerID string
		err := tx.QueryRow(ctx, `SELECT p."Id" FROM "ProviderInvoices" i
   JOIN "ContractingCompanies" cc ON cc."Id" = i."ContractingCompanyId"
   JOIN "Providers" p ON p."Id" = cc."ProviderId"
   WHERE i."Id" = $1 AND p."OrganizationId" = $2`, id, orgID).Scan(&providerID)
		if errors.Is(err, pgx.ErrNoRows) {
			return httputil.NotFoundError(notifications.InvoiceNotFound)
		}
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "invoice-import:"+providerID); err != nil {
			return err
		}
		// Also serialize downstream financial writes while checking and reverting dependencies.
		if _, err = tx.Exec(ctx, `LOCK TABLE "ProviderInvoices", "PhoneLines", "PhoneLineCustomerLinks",
   "LineBillingProcessings", "LineBillingCompositionItems", "AccountsPayable", "AccountsReceivable",
   "CustomerBillingDocuments", "FinancialPayments", "PartnerSalesRecords", "SaleLineItems",
   "CollectionReminders", "InvoiceReceivableSources" IN SHARE ROW EXCLUSIVE MODE`); err != nil {
			return err
		}
		var status, reference string
		var monthID, parentID, parentStatus *string
		var due time.Time
		var undone *time.Time
		var tracked bool
		if err = tx.QueryRow(ctx, `SELECT "Status"::text, "Number", "DueDate", "ProcessingMonthId",
   "ParentInvoiceId", "UndoTracked", "UndoneAt", "UndoParentStatus" FROM "ProviderInvoices" WHERE "Id" = $1 FOR UPDATE`, id).
			Scan(&status, &reference, &due, &monthID, &parentID, &tracked, &undone, &parentStatus); err != nil {
			return err
		}
		if undone != nil {
			return nil
		}
		if status == "substituted" {
			return undoBlocked("Desfaça primeiro a fatura que substituiu esta importação.")
		}
		// Restoring a substituted invoice requires its previous status, which older imports did not record.
		if parentID != nil && parentStatus == nil {
			return undoBlocked("Esta importação substituiu outra fatura. A restauração da anterior precisa ser conciliada antes de desfazer.")
		}

		changes, err := s.invoiceChanges(ctx, id)
		if err != nil {
			return err
		}
		created := []string{}
		touched := []string{}
		for _, c := range changes {
			if c.Table == "PhoneLines" {
				touched = append(touched, c.ID)
				if len(c.Before) == 0 {
					created = append(created, c.ID)
				}
			}
		}
		if !tracked {
			// Legacy imports have only a creation audit. Never infer ownership just from LastInvoiceId.
			rows, err := tx.Query(ctx, `SELECT pl."Id", EXISTS (
    SELECT 1 FROM "AuditLogs" a WHERE a."EntityName" = 'PhoneLine' AND a."KeyValues" = pl."Id"
     AND a."ChangedBy" = 'import' AND a."ChangeType" = 'Create'
     AND a."NewValues"::jsonb->>'message' = $2)
    FROM "ProviderInvoicePhoneLines" j JOIN "PhoneLines" pl ON pl."Id" = j."PhoneLinesId"
    WHERE j."ProviderInvoicesId" = $1`, id,
				fmt.Sprintf("Linha criada em estoque automaticamente a partir da fatura %s / %s.", reference, due.Format("02/01/2006")))
			if err != nil {
				return err
			}
			for rows.Next() {
				var lineID string
				var owned bool
				if err := rows.Scan(&lineID, &owned); err != nil {
					rows.Close()
					return err
				}
				if !owned {
					rows.Close()
					return undoBlocked("Esta importação antiga alterou linhas preexistentes sem guardar os valores anteriores. Não é possível desfazer tudo automaticamente sem perder dados.")
				}
				created = append(created, lineID)
				touched = append(touched, lineID)
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return err
			}
			var applied bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "InvoiceDetectedExceedances" WHERE "InvoiceId" = $1 AND "Applied")`, id).Scan(&applied); err != nil {
				return err
			}
			if applied {
				return undoBlocked("Esta importação antiga aplicou excedentes sem registrar os itens criados. Concilie esses excedentes antes de desfazer.")
			}
			var unknownChanges bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "PhoneLines" p JOIN "ProviderInvoices" i ON i."ProviderAccountId"=p."ProviderAccountId"
 WHERE i."Id"=$1 AND NOT(p."Id"=ANY($2::varchar[])))`, id, created).Scan(&unknownChanges); err != nil {
				return err
			}
			if unknownChanges {
				return undoBlocked("Esta importação antiga compartilha a conta com outras linhas e não registra o estado anterior delas. É necessário conciliar essas linhas antes de desfazer tudo.")
			}
		}

		var dependent bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "ProviderInvoicePhoneLines" j
   JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
   WHERE j."PhoneLinesId" = ANY($3::varchar[]) AND i."Id" <> $1 AND i."UndoneAt" IS NULL
   UNION ALL SELECT 1 FROM "PhoneLines" p JOIN "ProviderInvoicePhoneLines" j ON j."PhoneLinesId"=p."Id"
    WHERE j."ProviderInvoicesId"=$1 AND p."LastInvoiceId" IS NOT NULL AND p."LastInvoiceId"<>$1
   UNION ALL SELECT 1 FROM "InvoiceImportChanges" c JOIN "ProviderInvoices" i ON i."Id"=c."InvoiceId"
    WHERE c."TableName"='PhoneLines' AND c."RowId"=ANY($2::varchar[]) AND c."InvoiceId"<>$1 AND i."UndoneAt" IS NULL
     AND c."Sequence">(SELECT min("Sequence") FROM "InvoiceImportChanges" WHERE "InvoiceId"=$1)
   UNION ALL SELECT 1 FROM "SaleLineItems" WHERE "PhoneLineId" = ANY($3::varchar[])
   UNION ALL SELECT 1 FROM "PhoneLines" WHERE "TitularLineId" = ANY($3::varchar[]) AND NOT ("Id" = ANY($3::varchar[])))`,
			id, touched, created).Scan(&dependent); err != nil {
			return err
		}
		if dependent {
			return undoBlocked("Há linhas utilizadas por outra fatura, venda ou linha dependente. Desfaça esses vínculos antes desta importação.")
		}
		if err = s.cancelInvoiceFinancialOutputs(ctx, orgID, id, monthID, created); err != nil {
			return err
		}

		if tracked {
			for _, change := range changes {
				if err = s.reverseInvoiceChange(ctx, change); err != nil {
					return err
				}
			}
		} else {
			if _, err = tx.Exec(ctx, `DELETE FROM "PhoneLines" WHERE "Id" = ANY($1::varchar[])`, created); err != nil {
				return err
			}
		}
		for _, table := range []string{"ProviderInvoiceItems", "ProviderInvoiceServices", "ProviderInvoiceQuotaSharing", "InvoiceDetectedExceedances"} {
			if _, err = tx.Exec(ctx, `DELETE FROM `+pgx.Identifier{table}.Sanitize()+` WHERE "InvoiceId" = $1`, id); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(ctx, `DELETE FROM "ProviderInvoicePhoneLines" WHERE "ProviderInvoicesId" = $1`, id); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE "ProviderInvoices" SET "Status" = 'cancelled', "UndoneAt" = now() WHERE "Id" = $1`, id); err != nil {
			return err
		}
		if parentID != nil {
			tag, err := tx.Exec(ctx, `UPDATE "ProviderInvoices" SET "Status" = $2::provider_invoice_status WHERE "Id" = $1 AND "Status"='substituted' AND "UndoneAt" IS NULL`, parentID, parentStatus)
			if err != nil {
				return err
			}
			if tag.RowsAffected() != 1 {
				return undoBlocked("A fatura anterior mudou após a substituição. Concilie a sequência de importações antes de desfazer.")
			}
		}
		payload := fmt.Sprintf(`{"invoice_id":%q,"removed_lines":%d,"financial_outputs_cancelled":true}`, id, len(created))
		return s.InsertAuditLog(ctx, uuid.NewString(), "UndoImport", "ProviderInvoice", id, &actorID, nil, &payload, time.Now().UTC(), "")
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return undoBlocked("Existem registros posteriores que dependem desta importação. Nada foi desfeito; desvincule esses registros primeiro.")
	}
	return err
}

type invoiceChange struct {
	Table, ID     string
	Before, After json.RawMessage
}

func (s *Store) invoiceChanges(ctx context.Context, id string) ([]invoiceChange, error) {
	rows, err := s.q(ctx).Query(ctx, `SELECT "TableName", "RowId", "Before", "After" FROM "InvoiceImportChanges" WHERE "InvoiceId" = $1 ORDER BY "Sequence" DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []invoiceChange
	for rows.Next() {
		var c invoiceChange
		if err := rows.Scan(&c.Table, &c.ID, &c.Before, &c.After); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (s *Store) reverseInvoiceChange(ctx context.Context, c invoiceChange) error {
	switch c.Table {
	case "PhoneLines", "PhoneLineCustomerLinks", "CustomerProviderLinks", "Customers", "LineBillingProcessings", "LineBillingCompositionItems":
	default:
		return fmt.Errorf("unexpected import journal table: %s", c.Table)
	}
	table := pgx.Identifier{c.Table}.Sanitize()
	var tag pgconn.CommandTag
	var err error
	if len(c.Before) == 0 {
		tag, err = s.q(ctx).Exec(ctx, `DELETE FROM `+table+` t WHERE "Id" = $1 AND to_jsonb(t) @> $2::jsonb`, c.ID, c.After)
	} else {
		// Restore only columns the import actually changed; preserve unrelated later edits.
		var before, after map[string]json.RawMessage
		if err := json.Unmarshal(c.Before, &before); err != nil {
			return err
		}
		if err := json.Unmarshal(c.After, &after); err != nil {
			return err
		}
		columns := []string{}
		expected := map[string]json.RawMessage{}
		for key, value := range before {
			if string(value) != string(after[key]) {
				columns = append(columns, key)
				expected[key] = after[key]
			}
		}
		if len(columns) == 0 {
			return nil
		}
		sort.Strings(columns)
		sets := []string{}
		for _, column := range columns {
			name := pgx.Identifier{column}.Sanitize()
			sets = append(sets, name+" = old."+name)
		}
		expectedJSON, err := json.Marshal(expected)
		if err != nil {
			return err
		}
		tag, err = s.q(ctx).Exec(ctx, `UPDATE `+table+` t SET `+strings.Join(sets, ", ")+`
   FROM jsonb_populate_record(NULL::`+table+`, $2::jsonb) old
   WHERE t."Id" = $1 AND to_jsonb(t) @> $3::jsonb`, c.ID, c.Before, expectedJSON)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return undoBlocked("Um cadastro afetado pela importação foi alterado depois. Nada foi desfeito para preservar essas alterações.")
	}
	return nil
}

func (s *Store) cancelInvoiceFinancialOutputs(ctx context.Context, orgID, id string, monthID *string, created []string) error {
	q := s.q(ctx)
	// Materialize dependencies before deleting line/customer links. No month-wide cancellation.
	if _, err := q.Exec(ctx, `CREATE TEMP TABLE undo_documents ON COMMIT DROP AS
  SELECT d."Id", d."AccountsReceivableId" FROM "CustomerBillingDocuments" d
  WHERE d."OrganizationId" = $1 AND (EXISTS (
   SELECT 1 FROM "InvoiceReceivableSources" s WHERE s."ReceivableId" = d."AccountsReceivableId" AND s."InvoiceId" = $2)
   OR (d."ProcessingMonthId" = $3 AND NOT EXISTS (SELECT 1 FROM "InvoiceReceivableSources" s WHERE s."ReceivableId" = d."AccountsReceivableId") AND EXISTS (
   SELECT 1 FROM "ProviderInvoicePhoneLines" j JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = j."PhoneLinesId"
   WHERE j."ProviderInvoicesId" = $2 AND l."CustomerId" = d."CustomerId"
    AND (d."PhoneLineId" IS NULL OR d."PhoneLineId" = j."PhoneLinesId"))))`, orgID, id, monthID); err != nil {
		return err
	}
	if _, err := q.Exec(ctx, `CREATE TEMP TABLE undo_receivables ON COMMIT DROP AS
 SELECT "AccountsReceivableId" AS "Id" FROM undo_documents WHERE "AccountsReceivableId" IS NOT NULL
 UNION SELECT s."ReceivableId" FROM "InvoiceReceivableSources" s JOIN "AccountsReceivable" r ON r."Id"=s."ReceivableId"
 WHERE s."InvoiceId"=$1 AND r."OrganizationId"=$2`, id, orgID); err != nil {
		return err
	}
	var blocked bool
	if err := q.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM "InvoiceReceivableSources" s JOIN "ProviderInvoices" i ON i."Id"=s."InvoiceId"
 JOIN "AccountsReceivable" r ON r."Id"=s."ReceivableId" WHERE s."ReceivableId" IN (SELECT "Id" FROM undo_receivables)
 AND s."InvoiceId" <> $1 AND i."Status" NOT IN ('cancelled','substituted') AND r."Status" <> 'cancelled'
 UNION ALL SELECT 1 FROM "CustomerBillingDocuments" d WHERE d."AccountsReceivableId" IN (SELECT "Id" FROM undo_receivables)
 AND d."Id" NOT IN (SELECT "Id" FROM undo_documents) AND d."Status" <> 'cancelled')`, id).Scan(&blocked); err != nil {
		return err
	}
	if blocked {
		return undoBlocked("Uma conta a receber reúne cobranças de outras importações. Separe ou cancele a cobrança compartilhada antes de desfazer.")
	}
	if err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "CustomerBillingDocuments" d JOIN undo_documents u ON u."Id" = d."Id"
  JOIN "PhoneLineCustomerLinks" l ON l."CustomerId" = d."CustomerId"
  JOIN "ProviderInvoicePhoneLines" j ON j."PhoneLinesId" = l."PhoneLineId"
  JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
  WHERE i."Id" <> $1 AND i."ProcessingMonthId" = d."ProcessingMonthId" AND i."Status" NOT IN ('cancelled', 'substituted')
   AND (d."PhoneLineId" IS NULL OR d."PhoneLineId" = l."PhoneLineId") AND d."Status" <> 'cancelled')`, id).Scan(&blocked); err != nil {
		return err
	}
	if blocked {
		return undoBlocked("Uma cobrança reúne esta fatura e outras importações. Cancele ou separe essa cobrança antes de desfazer.")
	}
	if _, err := q.Exec(ctx, `CREATE TEMP TABLE undo_payables ON COMMIT DROP AS
  SELECT "Id" FROM "AccountsPayable" WHERE "OrganizationId" = $1 AND "ProviderInvoiceId" = $2
  UNION SELECT p."AccountPayableId" FROM "PartnerSalesRecords" p WHERE p."OrganizationId" = $1 AND p."PhoneLineId" = ANY($3::varchar[]) AND p."AccountPayableId" IS NOT NULL`, orgID, id, created); err != nil {
		return err
	}
	if err := q.QueryRow(ctx, `SELECT EXISTS(
  SELECT 1 FROM "AccountsPayable" WHERE "Id" IN (SELECT "Id" FROM undo_payables) AND ("PaidAmount" <> 0 OR "Status" IN ('settled', 'partially_settled'))
  UNION ALL SELECT 1 FROM "AccountsReceivable" WHERE "Id" IN (SELECT "Id" FROM undo_receivables) AND ("ReceivedAmount" <> 0 OR "Status" IN ('settled', 'partially_settled'))
  UNION ALL SELECT 1 FROM "FinancialPayments" WHERE "OrganizationId" = $1 AND ("AccountId" IN (SELECT "Id" FROM undo_payables) OR "AccountId" IN (SELECT "Id" FROM undo_receivables))
  UNION ALL SELECT 1 FROM "PartnerSalesRecords" WHERE "OrganizationId" = $1 AND "PhoneLineId" = ANY($2::varchar[]) AND "Status" = 'paid'
  UNION ALL SELECT 1 FROM "CustomerBillingDocuments" WHERE "Id" IN (SELECT "Id" FROM undo_documents) AND ("SicrediPaidAt" IS NOT NULL OR
   (COALESCE("SicrediNossoNumero", '') <> '' AND COALESCE("SicrediBoletoStatus", '') NOT IN ('cancelled', 'canceled', 'baixado'))))`, orgID, created).Scan(&blocked); err != nil {
		return err
	}
	if blocked {
		return undoBlocked("Há pagamento, recebimento ou boleto bancário ativo vinculado. Estorne os pagamentos e confirme a baixa do boleto no banco antes de desfazer. Nenhum registro foi alterado.")
	}
	if err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM "PartnerSalesRecords" WHERE "AccountPayableId" IN (SELECT "Id" FROM undo_payables) AND NOT ("PhoneLineId" = ANY($1::varchar[])))`, created).Scan(&blocked); err != nil {
		return err
	}
	if blocked {
		return undoBlocked("A conta de comissão também inclui outras linhas. Separe essa conta antes de desfazer a importação.")
	}
	for _, sql := range []string{
		`UPDATE "AccountsPayable" SET "Status" = 'cancelled', "UpdatedAt" = now() WHERE "Id" IN (SELECT "Id" FROM undo_payables)`,
		`UPDATE "AccountsReceivable" SET "Status" = 'cancelled', "UpdatedAt" = now() WHERE "Id" IN (SELECT "Id" FROM undo_receivables)`,
		`UPDATE "CustomerBillingDocuments" SET "Status" = 'cancelled', "UpdatedAt" = now() WHERE "Id" IN (SELECT "Id" FROM undo_documents)`,
		`UPDATE "CollectionReminders" SET "Status" = 'cancelled' WHERE "AccountsReceivableId" IN (SELECT "Id" FROM undo_receivables) AND "Status" IN ('pending', 'failed')`,
	} {
		if _, err := q.Exec(ctx, sql); err != nil {
			return err
		}
	}
	_, err := q.Exec(ctx, `UPDATE "PartnerSalesRecords" SET "Status" = 'cancelled', "UpdatedAt" = now() WHERE "OrganizationId" = $1 AND "PhoneLineId" = ANY($2::varchar[])`, orgID, created)
	return err
}
