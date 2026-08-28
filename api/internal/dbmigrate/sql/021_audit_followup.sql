-- 021: idempotência de importação, fatura substituta, pró-rata configurável,
-- aprovação em dois níveis, tickets, divergências persistentes, pipeline mensal,
-- vínculo portal CPF/CNPJ e métricas operacionais.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Pró-rata configurável (OrganizationSettings da 019)
ALTER TABLE "OrganizationSettings"
    ADD COLUMN IF NOT EXISTS "ProrataDivisor" integer NOT NULL DEFAULT 30;

-- Status substituta: as duas faturas permanecem
DO $$ BEGIN
    ALTER TYPE provider_invoice_status ADD VALUE 'substituted';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE "ProviderInvoices"
    ADD COLUMN IF NOT EXISTS "ContentSHA256" character varying(64),
    ADD COLUMN IF NOT EXISTS "SubstitutionImpact" numeric(18,4);

CREATE UNIQUE INDEX IF NOT EXISTS "UX_ProviderInvoices_ContentSHA256_Active"
    ON "ProviderInvoices" ("ContentSHA256")
    WHERE "ContentSHA256" IS NOT NULL AND "Status" <> 'substituted'::provider_invoice_status;

CREATE UNIQUE INDEX IF NOT EXISTS "UX_ProviderInvoices_BusinessKey_Active"
    ON "ProviderInvoices" ("ProviderAccountId", "DueDate", "ProcessingMonthId")
    WHERE "Status" <> 'substituted'::provider_invoice_status;

ALTER TABLE "ProviderInvoiceImportRequests"
    ADD COLUMN IF NOT EXISTS "AllowSubstitute" boolean NOT NULL DEFAULT false;

-- Aprovação em dois níveis (requester ≠ approvers; dois usuários distintos)
CREATE TABLE IF NOT EXISTS "ApprovalRequests" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ActionType" character varying(64) NOT NULL,
    "EntityType" character varying(64) NOT NULL,
    "EntityId" character varying(36) NOT NULL,
    "Status" character varying(32) NOT NULL DEFAULT 'pending_first',
    "RequesterUserId" character varying(256) NOT NULL,
    "FirstApproverUserId" character varying(256),
    "SecondApproverUserId" character varying(256),
    "Justification" text,
    "Payload" jsonb,
    "BeforeSnapshot" jsonb,
    "AfterSnapshot" jsonb,
    "RejectionReason" text,
    "RejectedBy" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "FirstApprovedAt" timestamp with time zone,
    "SecondApprovedAt" timestamp with time zone,
    "RejectedAt" timestamp with time zone,
    "ExecutedAt" timestamp with time zone,
    CONSTRAINT "PK_ApprovalRequests" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_ApprovalRequests_Org_Status"
    ON "ApprovalRequests" ("OrganizationId", "Status", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_ApprovalRequests_Entity"
    ON "ApprovalRequests" ("EntityType", "EntityId");

-- Tickets de suporte (M30)
CREATE TABLE IF NOT EXISTS "SupportTickets" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Number" integer NOT NULL,
    "Title" character varying(256) NOT NULL,
    "Category" character varying(64) NOT NULL DEFAULT 'geral',
    "Priority" character varying(16) NOT NULL DEFAULT 'media',
    "Status" character varying(32) NOT NULL DEFAULT 'aberto',
    "SlaDueAt" timestamp with time zone,
    "AssigneeUserId" character varying(256),
    "RequesterUserId" character varying(256),
    "CustomerId" character varying(36),
    "PhoneLineId" character varying(36),
    "ChargeRef" character varying(128),
    "InvoiceId" character varying(36),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "ResolvedAt" timestamp with time zone,
    "ClosedAt" timestamp with time zone,
    CONSTRAINT "PK_SupportTickets" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_SupportTickets_Org_Number" UNIQUE ("OrganizationId", "Number")
);

CREATE INDEX IF NOT EXISTS "IX_SupportTickets_Org_Status"
    ON "SupportTickets" ("OrganizationId", "Status", "UpdatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_SupportTickets_Customer"
    ON "SupportTickets" ("CustomerId");

CREATE TABLE IF NOT EXISTS "SupportTicketMessages" (
    "Id" character varying(36) NOT NULL,
    "TicketId" character varying(36) NOT NULL,
    "AuthorUserId" character varying(256),
    "AuthorName" character varying(256),
    "Visibility" character varying(16) NOT NULL DEFAULT 'public',
    "Body" text NOT NULL,
    "AttachmentKey" character varying(512),
    "AttachmentName" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_SupportTicketMessages" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_SupportTicketMessages_Ticket" FOREIGN KEY ("TicketId") REFERENCES "SupportTickets" ("Id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "SupportTicketHistory" (
    "Id" character varying(36) NOT NULL,
    "TicketId" character varying(36) NOT NULL,
    "ActorUserId" character varying(256),
    "EventType" character varying(64) NOT NULL,
    "FromValue" text,
    "ToValue" text,
    "Notes" text,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_SupportTicketHistory" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_SupportTicketHistory_Ticket" FOREIGN KEY ("TicketId") REFERENCES "SupportTickets" ("Id") ON DELETE CASCADE
);

-- Centro de divergências persistente
CREATE TABLE IF NOT EXISTS "BillingDivergences" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ProcessingMonthId" character varying(36),
    "DivergenceType" character varying(64) NOT NULL,
    "Severity" character varying(16) NOT NULL DEFAULT 'MEDIUM',
    "Competence" character varying(32),
    "OperatorName" character varying(128),
    "AccountNumber" character varying(64),
    "CustomerId" character varying(36),
    "PhoneLineId" character varying(36),
    "PhoneNumber" character varying(32),
    "OwnerUserId" character varying(256),
    "FinancialImpact" numeric(18,4) NOT NULL DEFAULT 0,
    "Status" character varying(32) NOT NULL DEFAULT 'open',
    "Cause" text,
    "RecommendedAction" text,
    "Evidence" text,
    "ResolutionNotes" text,
    "ResolvedBy" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "ResolvedAt" timestamp with time zone,
    CONSTRAINT "PK_BillingDivergences" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_BillingDivergences_Org_Status"
    ON "BillingDivergences" ("OrganizationId", "Status", "CreatedAt" DESC);

CREATE TABLE IF NOT EXISTS "BillingDivergenceComments" (
    "Id" character varying(36) NOT NULL,
    "DivergenceId" character varying(36) NOT NULL,
    "AuthorUserId" character varying(256),
    "Body" text NOT NULL,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_BillingDivergenceComments" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_BillingDivergenceComments_Div" FOREIGN KEY ("DivergenceId") REFERENCES "BillingDivergences" ("Id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "BillingDivergenceHistory" (
    "Id" character varying(36) NOT NULL,
    "DivergenceId" character varying(36) NOT NULL,
    "ActorUserId" character varying(256),
    "EventType" character varying(64) NOT NULL,
    "FromValue" text,
    "ToValue" text,
    "Notes" text,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_BillingDivergenceHistory" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_BillingDivergenceHistory_Div" FOREIGN KEY ("DivergenceId") REFERENCES "BillingDivergences" ("Id") ON DELETE CASCADE
);

-- Pipeline mensal orquestrado (versões comparáveis)
CREATE TABLE IF NOT EXISTS "ProcessingMonthRuns" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ProcessingMonthId" character varying(36) NOT NULL,
    "Version" integer NOT NULL,
    "Status" character varying(32) NOT NULL DEFAULT 'running',
    "TriggeredBy" character varying(256),
    "Summary" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "CompletedAt" timestamp with time zone,
    CONSTRAINT "PK_ProcessingMonthRuns" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_ProcessingMonthRuns_Month_Version" UNIQUE ("ProcessingMonthId", "Version")
);

CREATE TABLE IF NOT EXISTS "ProcessingMonthRunSteps" (
    "Id" character varying(36) NOT NULL,
    "RunId" character varying(36) NOT NULL,
    "StepKey" character varying(64) NOT NULL,
    "StepOrder" integer NOT NULL,
    "Label" character varying(128) NOT NULL,
    "Status" character varying(32) NOT NULL DEFAULT 'pending',
    "StartedAt" timestamp with time zone,
    "CompletedAt" timestamp with time zone,
    "DurationMs" integer,
    "Error" text,
    "Summary" jsonb,
    CONSTRAINT "PK_ProcessingMonthRunSteps" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProcessingMonthRunSteps_Run" FOREIGN KEY ("RunId") REFERENCES "ProcessingMonthRuns" ("Id") ON DELETE CASCADE,
    CONSTRAINT "UX_ProcessingMonthRunSteps_Run_Key" UNIQUE ("RunId", "StepKey")
);

-- Portal 1:1 por CPF/CNPJ
CREATE TABLE IF NOT EXISTS "PortalCustomerLinks" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "UserId" character varying(256) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "Document" character varying(32) NOT NULL,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_PortalCustomerLinks" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_User" UNIQUE ("OrganizationId", "UserId"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_Document" UNIQUE ("OrganizationId", "Document"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_Customer" UNIQUE ("OrganizationId", "CustomerId"),
    CONSTRAINT "FK_PortalCustomerLinks_Customer" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE
);

-- Métricas de duração (sem APM externo)
CREATE TABLE IF NOT EXISTS "OperationMetrics" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36),
    "Operation" character varying(128) NOT NULL,
    "DurationMs" integer NOT NULL,
    "Success" boolean NOT NULL DEFAULT true,
    "Metadata" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_OperationMetrics" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_OperationMetrics_Op_Created"
    ON "OperationMetrics" ("Operation", "CreatedAt" DESC);
