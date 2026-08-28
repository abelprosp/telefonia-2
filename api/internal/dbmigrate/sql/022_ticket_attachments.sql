-- 022: metadados de anexos binários em mensagens de ticket (upload via URL pré-assinada).
-- O conteúdo do arquivo permanece no object storage; a API só guarda chave, bucket e metadados.

ALTER TABLE "SupportTicketMessages"
    ADD COLUMN IF NOT EXISTS "AttachmentBucket" character varying(256),
    ADD COLUMN IF NOT EXISTS "AttachmentContentType" character varying(128),
    ADD COLUMN IF NOT EXISTS "AttachmentSizeBytes" bigint;

CREATE INDEX IF NOT EXISTS "IX_SupportTicketMessages_AttachmentKey"
    ON "SupportTicketMessages" ("AttachmentKey")
    WHERE "AttachmentKey" IS NOT NULL;
