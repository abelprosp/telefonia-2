-- 023: webhooks persistentes, correlation id na auditoria, hash de fechamento,
-- total de parcelamento (resíduo na última parcela) e dump organizacional.

ALTER TABLE "AuditLogs"
    ADD COLUMN IF NOT EXISTS "CorrelationId" character varying(64);

CREATE INDEX IF NOT EXISTS "IX_AuditLogs_CorrelationId"
    ON "AuditLogs" ("CorrelationId")
    WHERE "CorrelationId" IS NOT NULL;

ALTER TABLE "ProcessingMonths"
    ADD COLUMN IF NOT EXISTS "ConsolidationHash" character varying(64);

ALTER TABLE "LineBillingCompositionItems"
    ADD COLUMN IF NOT EXISTS "InstallmentTotal" numeric(18,4);

CREATE TABLE IF NOT EXISTS "WebhookSubscriptions" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Url" character varying(2048) NOT NULL,
    "Events" jsonb NOT NULL DEFAULT '[]'::jsonb,
    "Secret" character varying(128) NOT NULL,
    "IsActive" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_WebhookSubscriptions" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_WebhookSubscriptions_Org_Active"
    ON "WebhookSubscriptions" ("OrganizationId", "IsActive");

CREATE TABLE IF NOT EXISTS "WebhookDeliveries" (
    "Id" character varying(36) NOT NULL,
    "SubscriptionId" character varying(36) NOT NULL,
    "EventType" character varying(64) NOT NULL,
    "StatusCode" integer,
    "Success" boolean NOT NULL DEFAULT false,
    "Error" text,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_WebhookDeliveries" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_WebhookDeliveries_Subscription" FOREIGN KEY ("SubscriptionId")
        REFERENCES "WebhookSubscriptions" ("Id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "OrganizationDataExports" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ExportedBy" character varying(256) NOT NULL,
    "ChecksumSHA256" character varying(64) NOT NULL,
    "PayloadBytes" integer NOT NULL DEFAULT 0,
    "Summary" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_OrganizationDataExports" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_OrganizationDataExports_Org_Created"
    ON "OrganizationDataExports" ("OrganizationId", "CreatedAt" DESC);
