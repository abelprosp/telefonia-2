ALTER TABLE "CustomerBillingDocuments"
    ADD COLUMN IF NOT EXISTS "SicrediNossoNumero" character varying(32),
    ADD COLUMN IF NOT EXISTS "SicrediLinhaDigitavel" character varying(64),
    ADD COLUMN IF NOT EXISTS "SicrediCodigoBarras" character varying(64),
    ADD COLUMN IF NOT EXISTS "SicrediPixQrCode" text,
    ADD COLUMN IF NOT EXISTS "SicrediPixTxId" character varying(64),
    ADD COLUMN IF NOT EXISTS "SicrediBoletoStatus" character varying(32),
    ADD COLUMN IF NOT EXISTS "SicrediBoletoError" character varying(2000);

CREATE INDEX IF NOT EXISTS "IX_CustomerBillingDocuments_SicrediNossoNumero"
    ON "CustomerBillingDocuments" ("SicrediNossoNumero")
    WHERE "SicrediNossoNumero" IS NOT NULL;
