-- 032: manual in-person (cash) payment for sales and customer invoices.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_enum e
        JOIN pg_type t ON e.enumtypid = t.oid
        WHERE t.typname = 'sale_status' AND e.enumlabel = 'paid'
    ) THEN
        ALTER TYPE sale_status ADD VALUE 'paid';
    END IF;
END $$;

ALTER TABLE "Sales"
    ADD COLUMN IF NOT EXISTS "PaidAt" timestamp with time zone,
    ADD COLUMN IF NOT EXISTS "PaymentMethod" character varying(32),
    ADD COLUMN IF NOT EXISTS "PaymentNotes" text,
    ADD COLUMN IF NOT EXISTS "PaidByUserId" character varying(256);

ALTER TABLE "CustomerBillingDocuments"
    ADD COLUMN IF NOT EXISTS "PaymentMethod" character varying(32),
    ADD COLUMN IF NOT EXISTS "ManualPaymentNotes" text,
    ADD COLUMN IF NOT EXISTS "PaidByUserId" character varying(256);

CREATE INDEX IF NOT EXISTS "IX_Sales_PaidAt"
    ON "Sales" ("OrganizationId", "PaidAt" DESC)
    WHERE "PaidAt" IS NOT NULL;

CREATE INDEX IF NOT EXISTS "IX_CustomerBillingDocuments_PaymentMethod"
    ON "CustomerBillingDocuments" ("OrganizationId", "PaymentMethod")
    WHERE "PaymentMethod" IS NOT NULL;
