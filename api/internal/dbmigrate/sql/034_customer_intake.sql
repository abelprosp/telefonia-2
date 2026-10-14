CREATE TABLE IF NOT EXISTS "CustomerIntakes" (
    "Id" varchar(36) PRIMARY KEY,
    "OrganizationId" varchar(36) NOT NULL,
    "CustomerId" varchar(36) NOT NULL REFERENCES "Customers" ("Id") ON DELETE RESTRICT,
    "SaleId" varchar(36) NOT NULL UNIQUE REFERENCES "Sales" ("Id") ON DELETE RESTRICT,
    "SourceSystem" varchar(128) NOT NULL,
    "ExternalSaleId" varchar(256) NOT NULL,
    "PayloadHash" varchar(64) NOT NULL,
    "SaleData" jsonb NOT NULL,
    "ClientData" jsonb NOT NULL,
    "AttachmentData" jsonb NOT NULL DEFAULT '[]',
    "ValueCents" bigint NOT NULL CHECK ("ValueCents" >= 0),
    "CommissionCents" bigint NOT NULL CHECK ("CommissionCents" >= 0),
    "ImportedAt" timestamptz NOT NULL DEFAULT now(),
    UNIQUE ("OrganizationId", "SourceSystem", "ExternalSaleId")
);
CREATE INDEX IF NOT EXISTS "IX_CustomerIntakes_Customer"
    ON "CustomerIntakes" ("OrganizationId", "CustomerId", "ImportedAt" DESC);
