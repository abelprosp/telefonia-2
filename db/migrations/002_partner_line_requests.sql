CREATE TYPE phone_line_operation_type AS ENUM ('activation', 'deactivation');
CREATE TYPE phone_line_operation_status AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

CREATE TABLE "PhoneLineOperationRequests" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "RequestedByUserId" character varying(256) NOT NULL,
    "OperationType" phone_line_operation_type NOT NULL,
    "Status" phone_line_operation_status NOT NULL DEFAULT 'pending',
    "Justification" character varying(4000),
    "AdminNotes" character varying(4000),
    "ReviewedByUserId" character varying(256),
    "ReviewedAt" timestamp with time zone,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_PhoneLineOperationRequests" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_PhoneLineOperationRequests_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PhoneLineOperationRequests_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE
);

CREATE INDEX "IX_PhoneLineOperationRequests_OrganizationId" ON "PhoneLineOperationRequests" ("OrganizationId");
CREATE INDEX "IX_PhoneLineOperationRequests_RequestedByUserId" ON "PhoneLineOperationRequests" ("RequestedByUserId");
CREATE INDEX "IX_PhoneLineOperationRequests_PhoneLineId" ON "PhoneLineOperationRequests" ("PhoneLineId");
CREATE INDEX "IX_PhoneLineOperationRequests_Status" ON "PhoneLineOperationRequests" ("Status");
