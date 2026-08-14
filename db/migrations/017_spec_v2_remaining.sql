-- V2 restante: dual processing (classificadores), excedentes automáticos,
-- fidelidade, contratos automáticos, PDF stub, boleto por grupo e exportação.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    ALTER TYPE exceedance_charge_type ADD VALUE 'tabulated';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TYPE exceedance_charge_type ADD VALUE 'mirrored';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE "LineBillingProcessings"
    ADD COLUMN IF NOT EXISTS "OrganizationalUnit" character varying(128),
    ADD COLUMN IF NOT EXISTS "Department" character varying(128),
    ADD COLUMN IF NOT EXISTS "CostCenterLabel" character varying(128);

ALTER TABLE "PhoneLines"
    ADD COLUMN IF NOT EXISTS "ChargeExceedances" boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS "ExceedanceChargeType" exceedance_charge_type NOT NULL DEFAULT 'mirroed'::exceedance_charge_type;

ALTER TABLE "CustomerBillingDocuments"
    ADD COLUMN IF NOT EXISTS "PhoneLineId" character varying(36),
    ADD COLUMN IF NOT EXISTS "BillingGroupType" character varying(32);

ALTER TABLE "GeneratedContracts"
    ALTER COLUMN "SaleId" DROP NOT NULL;

ALTER TABLE "GeneratedContracts"
    ADD COLUMN IF NOT EXISTS "CustomerId" character varying(36),
    ADD COLUMN IF NOT EXISTS "PhoneLineId" character varying(36),
    ADD COLUMN IF NOT EXISTS "Trigger" character varying(64);

DO $$ BEGIN
    ALTER TABLE "GeneratedContracts"
        ADD CONSTRAINT "FK_GeneratedContracts_Customers_CustomerId"
        FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE "GeneratedContracts"
        ADD CONSTRAINT "FK_GeneratedContracts_PhoneLines_PhoneLineId"
        FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS "IX_CustomerBillingDocuments_PhoneLineId"
    ON "CustomerBillingDocuments" ("PhoneLineId");

CREATE INDEX IF NOT EXISTS "IX_GeneratedContracts_CustomerId"
    ON "GeneratedContracts" ("CustomerId");

CREATE INDEX IF NOT EXISTS "IX_GeneratedContracts_PhoneLineId"
    ON "GeneratedContracts" ("PhoneLineId");

CREATE TABLE IF NOT EXISTS "ExceedanceTerms" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Term" character varying(256) NOT NULL,
    "ChargeType" exceedance_charge_type NOT NULL DEFAULT 'mirroed'::exceedance_charge_type,
    "TabulatedAmount" numeric(18,2),
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_ExceedanceTerms" PRIMARY KEY ("Id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_ExceedanceTerms_Org_Term"
    ON "ExceedanceTerms" ("OrganizationId", lower("Term"))
    WHERE "Active" = true;

CREATE TABLE IF NOT EXISTS "InvoiceDetectedExceedances" (
    "Id" character varying(36) NOT NULL,
    "InvoiceId" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36),
    "TermId" character varying(36),
    "Term" character varying(256) NOT NULL,
    "Description" character varying(512) NOT NULL,
    "InvoiceAmount" numeric(18,2) NOT NULL,
    "ChargedAmount" numeric(18,2) NOT NULL,
    "ChargeType" character varying(32) NOT NULL,
    "Applied" boolean NOT NULL DEFAULT false,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_InvoiceDetectedExceedances" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_InvoiceDetectedExceedances_Invoice" FOREIGN KEY ("InvoiceId")
        REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_InvoiceDetectedExceedances_Invoice_Line_Term"
    ON "InvoiceDetectedExceedances" ("InvoiceId", COALESCE("PhoneLineId", ''), lower("Term"));

CREATE INDEX IF NOT EXISTS "IX_InvoiceDetectedExceedances_InvoiceId"
    ON "InvoiceDetectedExceedances" ("InvoiceId");

CREATE TABLE IF NOT EXISTS "LineFidelities" (
    "Id" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "StartDate" date NOT NULL,
    "InitialMonths" integer NOT NULL,
    "PredictedEndDate" date NOT NULL,
    "AutoRenew" boolean NOT NULL DEFAULT false,
    "RenewalPeriodMonths" integer,
    "Status" character varying(32) NOT NULL DEFAULT 'active',
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_LineFidelities" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_LineFidelities_PhoneLines" FOREIGN KEY ("PhoneLineId")
        REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "CK_LineFidelities_InitialMonths" CHECK ("InitialMonths" > 0),
    CONSTRAINT "CK_LineFidelities_RenewalPeriod" CHECK ("RenewalPeriodMonths" IS NULL OR "RenewalPeriodMonths" > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_LineFidelities_PhoneLine"
    ON "LineFidelities" ("PhoneLineId");

CREATE TABLE IF NOT EXISTS "LineFidelityEvents" (
    "Id" character varying(36) NOT NULL,
    "FidelityId" character varying(36) NOT NULL,
    "EventType" character varying(32) NOT NULL,
    "OccurredAt" timestamp with time zone NOT NULL,
    "UserId" character varying(64),
    "Notes" character varying(1000),
    CONSTRAINT "PK_LineFidelityEvents" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_LineFidelityEvents_Fidelity" FOREIGN KEY ("FidelityId")
        REFERENCES "LineFidelities" ("Id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "IX_LineFidelityEvents_FidelityId"
    ON "LineFidelityEvents" ("FidelityId", "OccurredAt" DESC);

CREATE TABLE IF NOT EXISTS "FidelityRenewalTriggers" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "EventKey" character varying(64) NOT NULL,
    "Label" character varying(256) NOT NULL,
    "PromptEnabled" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_FidelityRenewalTriggers" PRIMARY KEY ("Id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_FidelityRenewalTriggers_Org_Event"
    ON "FidelityRenewalTriggers" ("OrganizationId", "EventKey");

INSERT INTO "ExceedanceTerms" (
    "Id", "OrganizationId", "Term", "ChargeType", "TabulatedAmount", "Active", "CreatedAt", "UpdatedAt"
)
SELECT gen_random_uuid()::text, src."OrganizationId", src."Term", 'mirroed'::exceedance_charge_type, NULL, true, NOW(), NOW()
FROM (
    SELECT DISTINCT p."OrganizationId", t."Term"
    FROM "Providers" p
    CROSS JOIN (VALUES
        ('Roaming'),
        ('Roaming Internacional'),
        ('Excedente')
    ) AS t("Term")
) src
WHERE NOT EXISTS (
    SELECT 1 FROM "ExceedanceTerms" e
    WHERE e."OrganizationId" = src."OrganizationId" AND lower(e."Term") = lower(src."Term")
);

INSERT INTO "FidelityRenewalTriggers" (
    "Id", "OrganizationId", "EventKey", "Label", "PromptEnabled", "CreatedAt", "UpdatedAt"
)
SELECT gen_random_uuid()::text, src."OrganizationId", src."EventKey", src."Label", true, NOW(), NOW()
FROM (
    SELECT DISTINCT p."OrganizationId", t."EventKey", t."Label"
    FROM "Providers" p
    CROSS JOIN (VALUES
        ('upgrade_data', 'Upgrade de dados'),
        ('downgrade_data', 'Downgrade de dados'),
        ('device', 'Aquisição de aparelho'),
        ('plan', 'Alteração de plano'),
        ('new_service', 'Inclusão de novo serviço')
    ) AS t("EventKey", "Label")
) src
WHERE NOT EXISTS (
    SELECT 1 FROM "FidelityRenewalTriggers" f
    WHERE f."OrganizationId" = src."OrganizationId" AND f."EventKey" = src."EventKey"
);
