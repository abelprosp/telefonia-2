-- 031: differentiate physical SIM vs eSIM on phone lines; optional ICCID identity.

ALTER TABLE "PhoneLines"
    ADD COLUMN IF NOT EXISTS "SimType" character varying(16) NOT NULL DEFAULT 'UNKNOWN',
    ADD COLUMN IF NOT EXISTS "ICCID" character varying(64);

UPDATE "PhoneLines"
SET "SimType" = 'UNKNOWN'
WHERE "SimType" IS NULL OR "SimType" = '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'CK_PhoneLines_SimType'
    ) THEN
        ALTER TABLE "PhoneLines"
            ADD CONSTRAINT "CK_PhoneLines_SimType"
            CHECK ("SimType" IN ('PHYSICAL', 'ESIM', 'UNKNOWN'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS "IX_PhoneLines_SimType"
    ON "PhoneLines" ("SimType");

CREATE INDEX IF NOT EXISTS "IX_PhoneLines_ICCID"
    ON "PhoneLines" ("ICCID")
    WHERE "ICCID" IS NOT NULL AND "ICCID" <> '';
