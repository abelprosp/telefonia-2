-- 018_spec_v2_portfolio.sql
-- Portfólio de Serviços: Nome em fatura, Definição de aplicação e Regra de disponibilidade (§6.1)

ALTER TABLE "ProviderPlanServices"
    ADD COLUMN IF NOT EXISTS "InvoiceName" character varying(256),
    ADD COLUMN IF NOT EXISTS "ApplicationType" character varying(32) NOT NULL DEFAULT 'both',
    ADD COLUMN IF NOT EXISTS "AvailabilityRule" character varying(32) NOT NULL DEFAULT 'global',
    ADD COLUMN IF NOT EXISTS "ExclusiveCustomerId" character varying(36);

DO $$ BEGIN
    ALTER TABLE "ProviderPlanServices"
        ADD CONSTRAINT "FK_ProviderPlanServices_Customers_ExclusiveCustomerId"
        FOREIGN KEY ("ExclusiveCustomerId") REFERENCES "Customers" ("Id") ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS "IX_ProviderPlanServices_ExclusiveCustomerId"
    ON "ProviderPlanServices" ("ExclusiveCustomerId");
