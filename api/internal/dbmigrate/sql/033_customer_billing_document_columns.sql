-- Ensure columns required by customer invoice generation exist on older DBs
-- that only received partial migrations.

ALTER TABLE "CustomerBillingDocuments"
    ADD COLUMN IF NOT EXISTS "PhoneLineId" character varying(36),
    ADD COLUMN IF NOT EXISTS "BillingGroupType" character varying(32),
    ADD COLUMN IF NOT EXISTS "PaymentMethod" character varying(32),
    ADD COLUMN IF NOT EXISTS "ManualPaymentNotes" text,
    ADD COLUMN IF NOT EXISTS "PaidByUserId" character varying(256);

CREATE INDEX IF NOT EXISTS "IX_CustomerBillingDocuments_PhoneLineId"
    ON "CustomerBillingDocuments" ("PhoneLineId");

CREATE INDEX IF NOT EXISTS "IX_CustomerBillingDocuments_PaymentMethod"
    ON "CustomerBillingDocuments" ("OrganizationId", "PaymentMethod")
    WHERE "PaymentMethod" IS NOT NULL;
