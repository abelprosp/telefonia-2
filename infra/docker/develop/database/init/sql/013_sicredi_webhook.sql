CREATE TABLE "SicrediWebhookEvents" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36),
    "NossoNumero" character varying(32),
    "SeuNumero" character varying(64),
    "IdTituloEmpresa" character varying(64),
    "EventType" character varying(64) NOT NULL,
    "Payload" jsonb NOT NULL,
    "Processed" boolean NOT NULL DEFAULT false,
    "ProcessError" character varying(2000),
    "CreatedAt" timestamp with time zone NOT NULL,
    "ProcessedAt" timestamp with time zone,
    CONSTRAINT "PK_SicrediWebhookEvents" PRIMARY KEY ("Id")
);

CREATE UNIQUE INDEX "UX_SicrediWebhookEvents_Dedup"
    ON "SicrediWebhookEvents" ("NossoNumero", "EventType", "CreatedAt")
    WHERE "NossoNumero" IS NOT NULL;

CREATE INDEX "IX_SicrediWebhookEvents_OrganizationId" ON "SicrediWebhookEvents" ("OrganizationId");
CREATE INDEX "IX_SicrediWebhookEvents_NossoNumero" ON "SicrediWebhookEvents" ("NossoNumero");
CREATE INDEX "IX_SicrediWebhookEvents_Processed" ON "SicrediWebhookEvents" ("Processed") WHERE "Processed" = false;
