-- Migration 024: Suporte a PDF Base, Coordenadas de Assinatura, ZapSign e Fluxo Manual para Contratos

-- Adiciona campos em ContractTemplates
ALTER TABLE "ContractTemplates"
    ADD COLUMN IF NOT EXISTS "PdfBaseUrl" text,
    ADD COLUMN IF NOT EXISTS "PdfStorageKey" character varying(256),
    ADD COLUMN IF NOT EXISTS "SignersConfig" jsonb DEFAULT '[]'::jsonb;

-- Adiciona campos em GeneratedContracts
ALTER TABLE "GeneratedContracts"
    ADD COLUMN IF NOT EXISTS "CustomerId" character varying(36),
    ADD COLUMN IF NOT EXISTS "PdfUrl" text,
    ADD COLUMN IF NOT EXISTS "PdfStorageKey" character varying(256),
    ADD COLUMN IF NOT EXISTS "SignatureMethod" character varying(32) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS "ZapSignDocToken" character varying(128),
    ADD COLUMN IF NOT EXISTS "ZapSignOpenID" bigint,
    ADD COLUMN IF NOT EXISTS "ZapSignSignURL" text,
    ADD COLUMN IF NOT EXISTS "ZapSignStatus" character varying(64),
    ADD COLUMN IF NOT EXISTS "SignedPdfUrl" text,
    ADD COLUMN IF NOT EXISTS "SignedPdfStorageKey" character varying(256),
    ADD COLUMN IF NOT EXISTS "SignedAt" timestamp with time zone,
    ADD COLUMN IF NOT EXISTS "SignedBy" text,
    ADD COLUMN IF NOT EXISTS "AttachedAt" timestamp with time zone;

-- Índice para busca rápida de contratos de clientes
CREATE INDEX IF NOT EXISTS "IX_GeneratedContracts_CustomerId" ON "GeneratedContracts" ("CustomerId");
CREATE INDEX IF NOT EXISTS "IX_GeneratedContracts_ZapSignDocToken" ON "GeneratedContracts" ("ZapSignDocToken");
