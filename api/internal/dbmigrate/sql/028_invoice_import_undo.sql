ALTER TABLE "ProviderInvoices"
    ADD COLUMN IF NOT EXISTS "UndoTracked" boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS "UndoneAt" timestamptz,
    ADD COLUMN IF NOT EXISTS "UndoParentStatus" text;

CREATE TABLE IF NOT EXISTS "InvoiceImportChanges" (
    "Sequence" bigserial PRIMARY KEY,
    "InvoiceId" varchar(36) NOT NULL,
    "TableName" text NOT NULL,
    "RowId" text NOT NULL,
    "Before" jsonb,
    "After" jsonb NOT NULL
);
CREATE INDEX IF NOT EXISTS "IX_InvoiceImportChanges_Invoice" ON "InvoiceImportChanges" ("InvoiceId", "Sequence");

CREATE TABLE IF NOT EXISTS "InvoiceReceivableSources" (
    "InvoiceId" varchar(36) NOT NULL REFERENCES "ProviderInvoices" ("Id"),
    "ReceivableId" varchar(36) NOT NULL REFERENCES "AccountsReceivable" ("Id") ON DELETE CASCADE,
    PRIMARY KEY ("InvoiceId", "ReceivableId")
);

CREATE OR REPLACE FUNCTION record_invoice_import_change() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE invoice_id text := current_setting('telefonia.import_invoice_id', true);
BEGIN
    IF invoice_id IS NULL OR invoice_id = '' THEN RETURN NEW; END IF;
    IF TG_OP = 'UPDATE' AND to_jsonb(OLD) = to_jsonb(NEW) THEN RETURN NEW; END IF;
    INSERT INTO "InvoiceImportChanges" ("InvoiceId", "TableName", "RowId", "Before", "After")
    VALUES (invoice_id, TG_TABLE_NAME, NEW."Id",
        CASE WHEN TG_OP = 'UPDATE' THEN to_jsonb(OLD) ELSE NULL END, to_jsonb(NEW));
    RETURN NEW;
END $$;

DO $$ DECLARE tbl text; BEGIN
    FOREACH tbl IN ARRAY ARRAY['PhoneLines', 'PhoneLineCustomerLinks', 'CustomerProviderLinks',
        'Customers', 'LineBillingProcessings', 'LineBillingCompositionItems'] LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS invoice_import_journal ON %I', tbl);
        EXECUTE format('CREATE TRIGGER invoice_import_journal AFTER INSERT OR UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION record_invoice_import_change()', tbl);
    END LOOP;
END $$;
