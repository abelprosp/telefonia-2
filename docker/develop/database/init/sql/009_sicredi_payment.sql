ALTER TABLE "CustomerBillingDocuments"
    ADD COLUMN IF NOT EXISTS "SicrediPaidAt" timestamp with time zone;
