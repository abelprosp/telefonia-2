DROP INDEX IF EXISTS "UX_SicrediWebhookEvents_Dedup";

CREATE UNIQUE INDEX IF NOT EXISTS "UX_SicrediWebhookEvents_NossoEvent"
    ON "SicrediWebhookEvents" ("NossoNumero", "EventType")
    WHERE "NossoNumero" IS NOT NULL AND "EventType" <> '';
