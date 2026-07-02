CREATE TYPE financial_entry_status AS ENUM ('open', 'partially_settled', 'settled', 'overdue', 'cancelled');
CREATE TYPE partner_sale_status AS ENUM ('accrued', 'approved', 'paid', 'cancelled');

CREATE TABLE "PartnerCommissionSettings" (
    "OrganizationId" character varying(36) NOT NULL,
    "DefaultCommissionPercent" numeric(5,2) NOT NULL DEFAULT 10.00,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_PartnerCommissionSettings" PRIMARY KEY ("OrganizationId")
);

CREATE TABLE "AccountsPayable" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Description" character varying(512) NOT NULL,
    "VendorName" character varying(256) NOT NULL,
    "ProviderInvoiceId" character varying(36),
    "PartnerSalespersonUserId" character varying(256),
    "IssueDate" date NOT NULL,
    "DueDate" date NOT NULL,
    "Amount" numeric(18,2) NOT NULL,
    "PaidAmount" numeric(18,2) NOT NULL DEFAULT 0,
    "Status" financial_entry_status NOT NULL DEFAULT 'open',
    "Notes" character varying(4000),
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_AccountsPayable" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_AccountsPayable_ProviderInvoices_ProviderInvoiceId" FOREIGN KEY ("ProviderInvoiceId") REFERENCES "ProviderInvoices" ("Id") ON DELETE SET NULL
);

CREATE TABLE "AccountsReceivable" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "Description" character varying(512) NOT NULL,
    "ProcessingMonthId" character varying(36),
    "IssueDate" date NOT NULL,
    "DueDate" date NOT NULL,
    "Amount" numeric(18,2) NOT NULL,
    "ReceivedAmount" numeric(18,2) NOT NULL DEFAULT 0,
    "Status" financial_entry_status NOT NULL DEFAULT 'open',
    "Notes" character varying(4000),
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_AccountsReceivable" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_AccountsReceivable_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_AccountsReceivable_ProcessingMonths_ProcessingMonthId" FOREIGN KEY ("ProcessingMonthId") REFERENCES "ProcessingMonths" ("Id") ON DELETE SET NULL
);

CREATE TABLE "FinancialPayments" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "AccountType" character varying(16) NOT NULL,
    "AccountId" character varying(36) NOT NULL,
    "Amount" numeric(18,2) NOT NULL,
    "PaymentDate" date NOT NULL,
    "Reference" character varying(256),
    "Notes" character varying(4000),
    "CreatedByUserId" character varying(256) NOT NULL,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_FinancialPayments" PRIMARY KEY ("Id")
);

CREATE TABLE "PartnerSalesRecords" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "SalespersonUserId" character varying(256) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "ReferenceMonth" date NOT NULL,
    "GrossAmount" numeric(18,2) NOT NULL,
    "CommissionPercent" numeric(5,2) NOT NULL,
    "CommissionAmount" numeric(18,2) NOT NULL,
    "Status" partner_sale_status NOT NULL DEFAULT 'accrued',
    "AccountPayableId" character varying(36),
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_PartnerSalesRecords" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_PartnerSalesRecords_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PartnerSalesRecords_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PartnerSalesRecords_AccountsPayable_AccountPayableId" FOREIGN KEY ("AccountPayableId") REFERENCES "AccountsPayable" ("Id") ON DELETE SET NULL
);

CREATE INDEX "IX_AccountsPayable_OrganizationId" ON "AccountsPayable" ("OrganizationId");
CREATE INDEX "IX_AccountsPayable_Status" ON "AccountsPayable" ("Status");
CREATE INDEX "IX_AccountsPayable_DueDate" ON "AccountsPayable" ("DueDate");
CREATE INDEX "IX_AccountsReceivable_OrganizationId" ON "AccountsReceivable" ("OrganizationId");
CREATE INDEX "IX_AccountsReceivable_CustomerId" ON "AccountsReceivable" ("CustomerId");
CREATE INDEX "IX_AccountsReceivable_Status" ON "AccountsReceivable" ("Status");
CREATE INDEX "IX_PartnerSalesRecords_OrganizationId" ON "PartnerSalesRecords" ("OrganizationId");
CREATE INDEX "IX_PartnerSalesRecords_SalespersonUserId" ON "PartnerSalesRecords" ("SalespersonUserId");
CREATE INDEX "IX_PartnerSalesRecords_ReferenceMonth" ON "PartnerSalesRecords" ("ReferenceMonth");
CREATE UNIQUE INDEX "UX_PartnerSalesRecords_LineMonth" ON "PartnerSalesRecords" ("PhoneLineId", "ReferenceMonth");
