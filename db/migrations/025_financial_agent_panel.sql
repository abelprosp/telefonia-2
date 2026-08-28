-- 025: painel do agente financeiro (Evolution/WhatsApp + histórico de atendimento)

CREATE TABLE IF NOT EXISTS "FinancialAgentSettings" (
    "OrganizationId" character varying(36) NOT NULL,
    "Enabled" boolean NOT NULL DEFAULT true,
    "EvolutionApiUrl" text NOT NULL DEFAULT '',
    "EvolutionApiKey" text NOT NULL DEFAULT '',
    "EvolutionInstance" character varying(128) NOT NULL DEFAULT 'luxus',
    "N8nWebhookUrl" text NOT NULL DEFAULT '',
    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedBy" character varying(256),
    CONSTRAINT "PK_FinancialAgentSettings" PRIMARY KEY ("OrganizationId")
);

CREATE TABLE IF NOT EXISTS "FinancialAgentEvents" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "EventType" character varying(64) NOT NULL,
    "WhatsAppNumber" character varying(32),
    "CustomerId" character varying(36),
    "CustomerName" character varying(256),
    "InvoiceId" character varying(36),
    "InvoiceNumber" character varying(64),
    "Success" boolean NOT NULL DEFAULT true,
    "Summary" text NOT NULL DEFAULT '',
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_FinancialAgentEvents" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_FinancialAgentEvents_Org_Created"
    ON "FinancialAgentEvents" ("OrganizationId", "CreatedAt" DESC);
