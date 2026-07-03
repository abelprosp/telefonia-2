CREATE TABLE "CustomerDeviceLinks" (
    "Id" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "DeviceStockItemId" character varying(36),
    "Description" character varying(512) NOT NULL,
    "Brand" character varying(128) NOT NULL,
    "Model" character varying(256) NOT NULL,
    "MonthlyAmount" numeric(18,2) NOT NULL,
    "StartDate" date NOT NULL,
    "EndDate" date,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CustomerDeviceLinks" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerDeviceLinks_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_CustomerDeviceLinks_DeviceStockItems_DeviceStockItemId" FOREIGN KEY ("DeviceStockItemId") REFERENCES "DeviceStockItems" ("Id") ON DELETE SET NULL
);

CREATE INDEX "IX_CustomerDeviceLinks_CustomerId" ON "CustomerDeviceLinks" ("CustomerId");
CREATE INDEX "IX_CustomerDeviceLinks_CustomerId_EndDate" ON "CustomerDeviceLinks" ("CustomerId", "EndDate");
