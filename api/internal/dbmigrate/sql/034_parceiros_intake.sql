-- Entrada das vendas concluídas do Luxus Parceiros.
-- O cadastro manual continua com as regras atuais.

ALTER TABLE "Customers"
    ADD COLUMN IF NOT EXISTS "Origin" character varying(32) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS "RawAddress" text,
    ADD COLUMN IF NOT EXISTS "RawDocument" text,
    ADD COLUMN IF NOT EXISTS "ReviewPending" boolean NOT NULL DEFAULT false;

ALTER TABLE "PhoneLines"
    ADD COLUMN IF NOT EXISTS "ExternalSource" character varying(32),
    ADD COLUMN IF NOT EXISTS "ExternalOperatorName" character varying(128),
    ADD COLUMN IF NOT EXISTS "ExternalPlanName" character varying(128),
    ADD COLUMN IF NOT EXISTS "DeviceImei" character varying(32),
    ADD COLUMN IF NOT EXISTS "DeviceEid" character varying(64),
    ADD COLUMN IF NOT EXISTS "DevicePlatform" character varying(16),
    ADD COLUMN IF NOT EXISTS "IsVirginChip" boolean,
    ADD COLUMN IF NOT EXISTS "IsPortability" boolean,
    ADD COLUMN IF NOT EXISTS "PortabilityNumber" character varying(32),
    ADD COLUMN IF NOT EXISTS "DonorOperator" character varying(32);

CREATE TABLE IF NOT EXISTS "ParceirosSaleImports" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ExternalSaleId" character varying(64) NOT NULL,
    "Protocol" character varying(64) NOT NULL,
    "TaskProtocol" character varying(64),
    "CustomerId" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36),
    "CustomerCreated" boolean NOT NULL DEFAULT false,
    "PhoneLineCreated" boolean NOT NULL DEFAULT false,
    "AttachmentsAccepted" integer NOT NULL DEFAULT 0,
    "AttachmentsSkipped" integer NOT NULL DEFAULT 0,
    "PartnerName" character varying(256),
    "BranchName" character varying(256),
    "OperatorName" character varying(128),
    "PlanName" character varying(128),
    "LineNumber" character varying(32),
    "Value" numeric(18,2),
    "CommissionValue" numeric(18,2),
    "SoldAt" timestamp with time zone,
    "ConcludedAt" timestamp with time zone,
    "SimType" character varying(16),
    "Iccid" character varying(64),
    "ContractFormat" character varying(32),
    "IsUpgrade" boolean NOT NULL DEFAULT false,
    "UpgradeOfProtocol" character varying(64),
    "IsPortability" boolean NOT NULL DEFAULT false,
    "PortabilityNumber" character varying(32),
    "DonorOperator" character varying(32),
    "Warnings" jsonb NOT NULL DEFAULT '[]'::jsonb,
    "Payload" jsonb NOT NULL DEFAULT '{}'::jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_ParceirosSaleImports" PRIMARY KEY ("Id"),
    CONSTRAINT "UQ_ParceirosSaleImports_Org_Sale" UNIQUE ("OrganizationId", "ExternalSaleId")
);

CREATE TABLE IF NOT EXISTS "ParceirosSaleAttachments" (
    "Id" character varying(36) NOT NULL,
    "ImportId" character varying(36) NOT NULL,
    "ExternalDocumentId" character varying(64) NOT NULL,
    "AttachmentId" character varying(36),
    CONSTRAINT "PK_ParceirosSaleAttachments" PRIMARY KEY ("Id"),
    CONSTRAINT "UQ_ParceirosSaleAttachments_Import_Document" UNIQUE ("ImportId", "ExternalDocumentId"),
    CONSTRAINT "FK_ParceirosSaleAttachments_Import" FOREIGN KEY ("ImportId") REFERENCES "ParceirosSaleImports" ("Id") ON DELETE CASCADE
);
