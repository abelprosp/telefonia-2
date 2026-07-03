-- Processamento duplo (Luxus→Cliente / Cliente→Usuário final) com composição financeira por perspectiva.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE billing_processing_perspective AS ENUM ('luxus_customer', 'customer_end_user');
CREATE TYPE billing_composition_item_type AS ENUM ('service', 'discount', 'extra_charge', 'installment');

ALTER TABLE "Customers"
    ADD COLUMN IF NOT EXISTS "IsReseller" boolean NOT NULL DEFAULT false;

CREATE TABLE "LineBillingProcessings" (
    "Id" character varying(36) NOT NULL,
    "PhoneLineCustomerLinkId" character varying(36) NOT NULL,
    "Perspective" billing_processing_perspective NOT NULL,
    "Label" character varying(256),
    "MirrorFromPrimary" boolean NOT NULL DEFAULT false,
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_LineBillingProcessings" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_LineBillingProcessings_Link" FOREIGN KEY ("PhoneLineCustomerLinkId")
        REFERENCES "PhoneLineCustomerLinks" ("Id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX "UX_LineBillingProcessings_Link_Perspective"
    ON "LineBillingProcessings" ("PhoneLineCustomerLinkId", "Perspective")
    WHERE "Active" = true;

CREATE INDEX "IX_LineBillingProcessings_LinkId" ON "LineBillingProcessings" ("PhoneLineCustomerLinkId");

CREATE TABLE "LineBillingCompositionItems" (
    "Id" character varying(36) NOT NULL,
    "ProcessingId" character varying(36) NOT NULL,
    "ItemType" billing_composition_item_type NOT NULL,
    "Description" character varying(512) NOT NULL,
    "Amount" numeric(18,2) NOT NULL,
    "Quantity" numeric(18,4) NOT NULL DEFAULT 1,
    "InstallmentCount" integer,
    "InstallmentCurrent" integer,
    "StartDate" date,
    "EndDate" date,
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_LineBillingCompositionItems" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_LineBillingCompositionItems_Processing" FOREIGN KEY ("ProcessingId")
        REFERENCES "LineBillingProcessings" ("Id") ON DELETE CASCADE
);

CREATE INDEX "IX_LineBillingCompositionItems_ProcessingId" ON "LineBillingCompositionItems" ("ProcessingId");

-- Backfill: processamento 1 para vínculos ativos existentes
INSERT INTO "LineBillingProcessings" (
    "Id", "PhoneLineCustomerLinkId", "Perspective", "Label", "MirrorFromPrimary", "Active", "CreatedAt", "UpdatedAt"
)
SELECT
    gen_random_uuid()::text,
    l."Id",
    'luxus_customer'::billing_processing_perspective,
    NULL,
    false,
    true,
    NOW(),
    NOW()
FROM "PhoneLineCustomerLinks" l
WHERE l."EndDate" IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM "LineBillingProcessings" p
    WHERE p."PhoneLineCustomerLinkId" = l."Id"
      AND p."Perspective" = 'luxus_customer'
      AND p."Active" = true
  );

INSERT INTO "LineBillingCompositionItems" (
    "Id", "ProcessingId", "ItemType", "Description", "Amount", "Quantity", "Active", "CreatedAt", "UpdatedAt"
)
SELECT
    gen_random_uuid()::text,
    p."Id",
    'service'::billing_composition_item_type,
    'Mensalidade',
    COALESCE(NULLIF(l."MonthlyAmount", 0), COALESCE(pl."CostWithConsumption", pl."BaseCost", 0)),
    1,
    true,
    NOW(),
    NOW()
FROM "LineBillingProcessings" p
JOIN "PhoneLineCustomerLinks" l ON l."Id" = p."PhoneLineCustomerLinkId"
JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
WHERE p."Perspective" = 'luxus_customer'
  AND NOT EXISTS (
    SELECT 1 FROM "LineBillingCompositionItems" ci
    WHERE ci."ProcessingId" = p."Id" AND ci."Active" = true
  );
