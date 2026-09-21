-- 030: foundation for multi-req delivery — phone identity, link audit fields,
-- domain audit events, external line import jobs (Axon/etc.), consumption report indexes.

-- Phone line presentation vs comparison identity (do not invent DDI stripping).
ALTER TABLE "PhoneLines"
    ADD COLUMN IF NOT EXISTS "NumberOriginal" character varying(64),
    ADD COLUMN IF NOT EXISTS "NormalizedNumber" character varying(32);

UPDATE "PhoneLines"
SET "NormalizedNumber" = regexp_replace(COALESCE("Number", ''), '[^0-9]', '', 'g')
WHERE "NormalizedNumber" IS NULL OR "NormalizedNumber" = '';

UPDATE "PhoneLines"
SET "NumberOriginal" = "Number"
WHERE "NumberOriginal" IS NULL;

CREATE INDEX IF NOT EXISTS "IX_PhoneLines_NormalizedNumber"
    ON "PhoneLines" ("NormalizedNumber");

-- Prefer unique normalized number when present (keeps legacy Number unique as well).
CREATE UNIQUE INDEX IF NOT EXISTS "UX_PhoneLines_NormalizedNumber"
    ON "PhoneLines" ("NormalizedNumber")
    WHERE "NormalizedNumber" IS NOT NULL AND "NormalizedNumber" <> '';

-- Temporal link enrichment (Start/End already exist; exclusivity via partial unique).
ALTER TABLE "PhoneLineCustomerLinks"
    ADD COLUMN IF NOT EXISTS "EndReason" character varying(128),
    ADD COLUMN IF NOT EXISTS "ChangedByUserId" character varying(256),
    ADD COLUMN IF NOT EXISTS "Status" character varying(32) NOT NULL DEFAULT 'active';

-- Domain-level audit trail (complements AuditLogs / StateTransitionLogs).
CREATE TABLE IF NOT EXISTS "DomainAuditEvents" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "EntityType" character varying(64) NOT NULL,
    "EntityId" character varying(64) NOT NULL,
    "Action" character varying(64) NOT NULL,
    "ActorUserId" character varying(256),
    "ActorKind" character varying(32) NOT NULL DEFAULT 'user',
    "Source" character varying(64) NOT NULL DEFAULT 'api',
    "CorrelationId" character varying(64),
    "BeforeJson" jsonb,
    "AfterJson" jsonb,
    "MetadataJson" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_DomainAuditEvents" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_DomainAuditEvents_Org_Created"
    ON "DomainAuditEvents" ("OrganizationId", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_DomainAuditEvents_Entity"
    ON "DomainAuditEvents" ("OrganizationId", "EntityType", "EntityId", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_DomainAuditEvents_Action"
    ON "DomainAuditEvents" ("OrganizationId", "Action", "CreatedAt" DESC);
CREATE INDEX IF NOT EXISTS "IX_DomainAuditEvents_Actor"
    ON "DomainAuditEvents" ("OrganizationId", "ActorUserId", "CreatedAt" DESC);

-- External line inventory imports (Axon and other spreadsheets). Layout mapping is
-- operator-specific and must be configured; no invented column names are assumed.
CREATE TABLE IF NOT EXISTS "ExternalLineImportJobs" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Source" character varying(64) NOT NULL,
    "ProviderId" character varying(36),
    "Status" character varying(32) NOT NULL DEFAULT 'pending',
    "StorageBucket" character varying(256),
    "StorageObjectKey" character varying(2048),
    "FileName" character varying(512),
    "ContentSHA256" character varying(64),
    "LayoutCode" character varying(64),
    "ReferencePeriod" character varying(32),
    "IsPartialSource" boolean NOT NULL DEFAULT false,
    "TotalRows" integer NOT NULL DEFAULT 0,
    "AcceptedRows" integer NOT NULL DEFAULT 0,
    "RejectedRows" integer NOT NULL DEFAULT 0,
    "ErrorMessage" text,
    "CreatedByUserId" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "CompletedAt" timestamp with time zone,
    CONSTRAINT "PK_ExternalLineImportJobs" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_ExternalLineImportJobs_Org_Created"
    ON "ExternalLineImportJobs" ("OrganizationId", "CreatedAt" DESC);

CREATE TABLE IF NOT EXISTS "ExternalLineSnapshots" (
    "Id" character varying(36) NOT NULL,
    "ImportJobId" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Source" character varying(64) NOT NULL,
    "ExternalId" character varying(128),
    "PhoneNumberOriginal" character varying(64),
    "NormalizedNumber" character varying(32),
    "LineType" character varying(32),
    "StatusExternal" character varying(64),
    "ICCID" character varying(64),
    "EID" character varying(64),
    "CustomerExternalRef" character varying(256),
    "RawRowJson" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_ExternalLineSnapshots" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ExternalLineSnapshots_Job" FOREIGN KEY ("ImportJobId")
        REFERENCES "ExternalLineImportJobs" ("Id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "IX_ExternalLineSnapshots_Job"
    ON "ExternalLineSnapshots" ("ImportJobId");
CREATE INDEX IF NOT EXISTS "IX_ExternalLineSnapshots_Norm"
    ON "ExternalLineSnapshots" ("OrganizationId", "Source", "NormalizedNumber");

CREATE TABLE IF NOT EXISTS "LineReconciliationFindings" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ImportJobId" character varying(36),
    "FindingType" character varying(64) NOT NULL,
    "Severity" character varying(16) NOT NULL DEFAULT 'warning',
    "PhoneLineId" character varying(36),
    "NormalizedNumber" character varying(32),
    "InternalStatus" character varying(32),
    "ExternalStatus" character varying(64),
    "CustomerId" character varying(36),
    "IsInconclusive" boolean NOT NULL DEFAULT false,
    "Evidence" text,
    "Status" character varying(32) NOT NULL DEFAULT 'open',
    "ResolvedByUserId" character varying(256),
    "ResolvedAt" timestamp with time zone,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_LineReconciliationFindings" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_LineReconciliationFindings_Org_Type"
    ON "LineReconciliationFindings" ("OrganizationId", "FindingType", "Status");

-- Strengthen invoice content hash uniqueness if older environments missed 021.
CREATE UNIQUE INDEX IF NOT EXISTS "UX_ProviderInvoices_ContentSHA256_Active"
    ON "ProviderInvoices" ("ContentSHA256")
    WHERE "ContentSHA256" IS NOT NULL AND "Status" <> 'substituted'::provider_invoice_status;
