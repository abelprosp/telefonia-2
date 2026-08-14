-- P0 spec V1: composição utilizável, proporcionalidade, hierarquia e cadastro comercial.

DO $$ BEGIN
    ALTER TYPE billing_composition_item_type ADD VALUE 'exceedance';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE "ProviderPlanServices"
    ADD COLUMN IF NOT EXISTS "ServiceType" service_type NOT NULL DEFAULT 'other'::service_type;

ALTER TABLE "PhoneLineServices"
    ADD COLUMN IF NOT EXISTS "StartDate" date,
    ADD COLUMN IF NOT EXISTS "EndDate" date,
    ADD COLUMN IF NOT EXISTS "ServiceType" service_type;

ALTER TABLE "LineBillingCompositionItems"
    ADD COLUMN IF NOT EXISTS "ServiceType" service_type,
    ADD COLUMN IF NOT EXISTS "ProviderPlanServiceId" character varying(36),
    ADD COLUMN IF NOT EXISTS "Proportional" boolean NOT NULL DEFAULT true;

ALTER TABLE "Customers"
    ADD COLUMN IF NOT EXISTS "CommercialActivationDate" date,
    ADD COLUMN IF NOT EXISTS "ContractedLuxusCnpj" character varying(14);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_PhoneLineServices_Line_ServiceType_Active"
    ON "PhoneLineServices" ("PhoneLineId", "ServiceType")
    WHERE "Active" = true AND "ServiceType" IS NOT NULL;

CREATE INDEX IF NOT EXISTS "IX_LineBillingCompositionItems_ServiceType"
    ON "LineBillingCompositionItems" ("ProcessingId", "ServiceType")
    WHERE "Active" = true AND "ServiceType" IS NOT NULL;
