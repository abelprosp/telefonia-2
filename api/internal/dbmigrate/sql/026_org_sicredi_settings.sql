<<<<<<< HEAD
-- Credenciais Sicredi por organização (multi-tenant).

=======
>>>>>>> 6b82d54 (fix tenant isolation and invoice processing reliability)
ALTER TABLE "OrganizationSettings"
    ADD COLUMN IF NOT EXISTS "SicrediEnabled" boolean NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "SicrediSandbox" boolean NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS "SicrediAPIKey" text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediUsername" character varying(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediPassword" text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediCooperativa" character varying(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediPosto" character varying(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediCodigoBeneficiario" character varying(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediAccountNumber" character varying(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediWebhookToken" text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "SicrediPublicAPIURL" text NOT NULL DEFAULT '';
