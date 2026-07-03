CREATE TYPE device_stock_status AS ENUM ('in_stock', 'sold', 'inactive');

CREATE TABLE "DeviceStockItems" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Sku" character varying(64) NOT NULL,
    "Brand" character varying(128) NOT NULL,
    "Model" character varying(256) NOT NULL,
    "Imei" character varying(20),
    "Color" character varying(64),
    "StorageCapacity" character varying(32),
    "UnitCost" numeric(18,2),
    "SalePrice" numeric(18,2),
    "Status" device_stock_status NOT NULL DEFAULT 'in_stock',
    "Notes" character varying(4000),
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_DeviceStockItems" PRIMARY KEY ("Id")
);

CREATE UNIQUE INDEX "IX_DeviceStockItems_OrganizationId_Sku" ON "DeviceStockItems" ("OrganizationId", "Sku");
CREATE UNIQUE INDEX "IX_DeviceStockItems_OrganizationId_Imei" ON "DeviceStockItems" ("OrganizationId", "Imei") WHERE "Imei" IS NOT NULL;
CREATE INDEX "IX_DeviceStockItems_OrganizationId_Status" ON "DeviceStockItems" ("OrganizationId", "Status");
