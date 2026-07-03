-- Schema consolidado das migrações EF (00–11) para bases novas sem .NET SDK.
-- Colunas em PascalCase, compatível com a API Go.

CREATE TYPE billing_cycle_status AS ENUM ('closed', 'open');
CREATE TYPE customer_document_type AS ENUM ('cnh', 'cnpj', 'cpf', 'municipal_registration', 'other', 'rg', 'state_registration');
CREATE TYPE customer_type AS ENUM ('pf', 'pj');
CREATE TYPE exceedance_charge_type AS ENUM ('mirroed');
CREATE TYPE invoice_item_unit AS ENUM ('gb', 'kb', 'mb', 'min', 'sms', 'tb');
CREATE TYPE line_classification AS ENUM ('dependent', 'normal', 'other', 'titular');
CREATE TYPE phone_line_status AS ENUM ('active', 'awaiting_invoice', 'cancelled', 'in_stock', 'in_transition', 'inactive', 'suspended');
CREATE TYPE processing_month_status AS ENUM ('closed', 'open');
CREATE TYPE provider_invoice_item_type AS ENUM ('discount', 'extra_detail', 'extra_header', 'extra_location', 'other', 'plan', 'service', 'usage');
CREATE TYPE provider_invoice_status AS ENUM ('cancelled', 'draft', 'overdue', 'paid', 'pending');
CREATE TYPE service_application_type AS ENUM ('addon', 'plan', 'service');
CREATE TYPE service_availability_rule AS ENUM ('always', 'custom', 'cycle_only');
CREATE TYPE service_type AS ENUM ('data', 'other', 'roaming', 'sms', 'subscription');
CREATE TYPE transition_sub_status AS ENUM ('none', 'pending_activation', 'pending_cancellation', 'pending_portability', 'pending_pp', 'pending_tt');

CREATE TABLE "AuditLogs" (
    "Id" character varying(36) NOT NULL,
    "ChangeType" character varying(16) NOT NULL,
    "EntityName" character varying(64) NOT NULL,
    "KeyValues" text NOT NULL,
    "ChangedBy" character varying(36),
    "OldValues" text,
    "NewValues" text,
    "Timestamp" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_AuditLogs" PRIMARY KEY ("Id")
);

CREATE TABLE "CostCenters" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Name" character varying(128) NOT NULL,
    "Description" character varying(256) NOT NULL,
    CONSTRAINT "PK_CostCenters" PRIMARY KEY ("Id")
);

CREATE TABLE "Providers" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Slug" character varying(256) NOT NULL,
    "Active" boolean NOT NULL DEFAULT TRUE,
    CONSTRAINT "PK_Providers" PRIMARY KEY ("Id")
);

CREATE TABLE "BillingCycles" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "Code" character varying(20) NOT NULL,
    "Name" character varying(100) NOT NULL,
    "StartDate" date NOT NULL,
    "EndDate" date NOT NULL,
    "Status" billing_cycle_status NOT NULL,
    "ClosedAt" timestamp with time zone,
    "ClosedBy" character varying(36),
    CONSTRAINT "PK_BillingCycles" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_BillingCycles_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ContractingCompanies" (
    "Id" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "LegalName" character varying(512) NOT NULL,
    "TaxId" character varying(14) NOT NULL,
    CONSTRAINT "PK_ContractingCompanies" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ContractingCompanies_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE RESTRICT
);

CREATE TABLE "Customers" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Type" customer_type NOT NULL,
    "Name" character varying(256) NOT NULL,
    "LegalName" character varying(512),
    "BirthOrOpeningDate" date,
    "Active" boolean NOT NULL,
    "ResponsibleSalespersonUserId" character varying(256),
    CONSTRAINT "PK_Customers" PRIMARY KEY ("Id")
);

CREATE TABLE "ProviderPlans" (
    "Id" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Code" character varying(64) NOT NULL,
    CONSTRAINT "PK_ProviderPlans" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderPlans_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProcessingMonths" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "Year" integer NOT NULL,
    "Month" integer NOT NULL,
    "DisplayName" character varying(128) NOT NULL,
    "Status" processing_month_status NOT NULL,
    "ClosedAt" timestamp with time zone,
    "ClosedBy" character varying(36),
    "ClosedInContingency" boolean NOT NULL,
    "ContingencyJustification" character varying(4000),
    CONSTRAINT "PK_ProcessingMonths" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProcessingMonths_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE RESTRICT
);

CREATE TABLE "ProviderAccounts" (
    "Id" character varying(36) NOT NULL,
    "ContractingCompanyId" character varying(36) NOT NULL,
    "AccountNumber" character varying(64) NOT NULL,
    CONSTRAINT "PK_ProviderAccounts" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderAccounts_ContractingCompanies_ContractingCompanyId" FOREIGN KEY ("ContractingCompanyId") REFERENCES "ContractingCompanies" ("Id") ON DELETE RESTRICT
);

CREATE TABLE "CustomerAddresses" (
    "Id" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "Street" character varying(256) NOT NULL,
    "Number" character varying(16) NOT NULL,
    "Neighborhood" character varying(128) NOT NULL,
    "Complement" character varying(128),
    "City" character varying(128) NOT NULL,
    "State" character varying(64) NOT NULL,
    "ZipCode" character varying(10) NOT NULL,
    "Country" character varying(32) NOT NULL,
    CONSTRAINT "PK_CustomerAddresses" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerAddresses_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "CustomerDocuments" (
    "Id" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "DocumentType" customer_document_type NOT NULL,
    "Number" character varying(128) NOT NULL,
    CONSTRAINT "PK_CustomerDocuments" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerDocuments_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "CustomerProviderLinks" (
    "Id" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "StartDate" date NOT NULL,
    "EndDate" date,
    CONSTRAINT "PK_CustomerProviderLinks" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerProviderLinks_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_CustomerProviderLinks_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "CustomerAttachments" (
    "Id" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Title" character varying(256),
    "OriginalFileName" character varying(512) NOT NULL,
    "StorageBucket" character varying(256) NOT NULL,
    "StorageObjectKey" character varying(2048) NOT NULL,
    "ContentType" character varying(128),
    "SizeBytes" bigint,
    "UploadedAtUtc" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CustomerAttachments" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerAttachments_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE
);

CREATE TABLE "CustomerProcessingMonthManualReleases" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "ProcessingMonthId" character varying(36) NOT NULL,
    "Justification" character varying(4000) NOT NULL,
    "ReleasedByUserId" character varying(64) NOT NULL,
    "ReleasedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CustomerProcessingMonthManualReleases" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerProcessingMonthManualReleases_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_CustomerProcessingMonthManualReleases_ProcessingMonths_Proc~" FOREIGN KEY ("ProcessingMonthId") REFERENCES "ProcessingMonths" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderPlanServices" (
    "Id" character varying(36) NOT NULL,
    "ProviderPlanId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Active" boolean NOT NULL DEFAULT TRUE,
    "Recurring" boolean NOT NULL,
    "Price" numeric(18,2),
    CONSTRAINT "PK_ProviderPlanServices" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderPlanServices_ProviderPlans_ProviderPlanId" FOREIGN KEY ("ProviderPlanId") REFERENCES "ProviderPlans" ("Id") ON DELETE CASCADE
);

CREATE TABLE "BillingCycleProviderAccounts" (
    "BillingCyclesId" character varying(36) NOT NULL,
    "ProviderAccountsId" character varying(36) NOT NULL,
    CONSTRAINT "PK_BillingCycleProviderAccounts" PRIMARY KEY ("BillingCyclesId", "ProviderAccountsId"),
    CONSTRAINT "FK_BillingCycleProviderAccounts_BillingCycles_BillingCyclesId" FOREIGN KEY ("BillingCyclesId") REFERENCES "BillingCycles" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_BillingCycleProviderAccounts_ProviderAccounts_ProviderAccou~" FOREIGN KEY ("ProviderAccountsId") REFERENCES "ProviderAccounts" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderInvoices" (
    "Id" character varying(36) NOT NULL,
    "ProviderAccountId" character varying(36) NOT NULL,
    "ContractingCompanyId" character varying(36) NOT NULL,
    "BillingCycleId" character varying(36) NOT NULL,
    "CostCenterId" character varying(36),
    "ParentInvoiceId" character varying(36),
    "ProcessingMonthId" character varying(36) NOT NULL,
    "Number" text NOT NULL DEFAULT '',
    "IssueDate" date NOT NULL,
    "DueDate" date NOT NULL,
    "TotalAmount" numeric(18,2) NOT NULL,
    "Status" provider_invoice_status NOT NULL,
    "SubtotalServices" numeric(18,2) NOT NULL,
    "SubtotalUsage" numeric(18,2) NOT NULL,
    "SubtotalTaxes" numeric(18,2) NOT NULL,
    "SubtotalDiscounts" numeric(18,2) NOT NULL,
    "SubtotalInstallments" numeric(18,2) NOT NULL,
    CONSTRAINT "PK_ProviderInvoices" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderInvoices_BillingCycles_BillingCycleId" FOREIGN KEY ("BillingCycleId") REFERENCES "BillingCycles" ("Id") ON DELETE RESTRICT,
    CONSTRAINT "FK_ProviderInvoices_ContractingCompanies_ContractingCompanyId" FOREIGN KEY ("ContractingCompanyId") REFERENCES "ContractingCompanies" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoices_CostCenters_CostCenterId" FOREIGN KEY ("CostCenterId") REFERENCES "CostCenters" ("Id") ON DELETE SET NULL,
    CONSTRAINT "FK_ProviderInvoices_ProviderAccounts_ProviderAccountId" FOREIGN KEY ("ProviderAccountId") REFERENCES "ProviderAccounts" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId" FOREIGN KEY ("ParentInvoiceId") REFERENCES "ProviderInvoices" ("Id"),
    CONSTRAINT "FK_ProviderInvoices_ProcessingMonths_ProcessingMonthId" FOREIGN KEY ("ProcessingMonthId") REFERENCES "ProcessingMonths" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderInvoiceImportRequests" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ProviderId" character varying(36) NOT NULL,
    "ProcessingMonthId" character varying(36) NOT NULL,
    "StorageBucket" character varying(256) NOT NULL,
    "StorageObjectKey" character varying(2048) NOT NULL,
    "OriginalFileName" character varying(512),
    "Status" integer NOT NULL,
    "Error" character varying(8000),
    "CompletedAt" timestamp with time zone,
    "CreatedBy" character varying(64) NOT NULL,
    CONSTRAINT "PK_ProviderInvoiceImportRequests" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderInvoiceImportRequests_Providers_ProviderId" FOREIGN KEY ("ProviderId") REFERENCES "Providers" ("Id") ON DELETE RESTRICT,
    CONSTRAINT "FK_ProviderInvoiceImportRequests_ProcessingMonths_ProcessingMo~" FOREIGN KEY ("ProcessingMonthId") REFERENCES "ProcessingMonths" ("Id") ON DELETE CASCADE
);

CREATE TABLE "PhoneLines" (
    "Id" character varying(36) NOT NULL,
    "ProviderPlanId" character varying(36) NOT NULL,
    "ProviderAccountId" character varying(36) NOT NULL,
    "CostCenterId" character varying(36),
    "LastInvoiceId" character varying(36),
    "TitularLineId" character varying(36),
    "Number" character varying(20) NOT NULL,
    "LineClassification" line_classification NOT NULL,
    "Status" phone_line_status NOT NULL,
    "TransitionSubStatus" transition_sub_status,
    "TransitionStartedAt" timestamp with time zone,
    "ActivationDate" date,
    "CancellationDate" date,
    "BaseCost" numeric(18,2),
    "CostWithConsumption" numeric(18,2),
    CONSTRAINT "PK_PhoneLines" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_PhoneLines_CostCenters_CostCenterId" FOREIGN KEY ("CostCenterId") REFERENCES "CostCenters" ("Id") ON DELETE SET NULL,
    CONSTRAINT "FK_PhoneLines_PhoneLines_TitularLineId" FOREIGN KEY ("TitularLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PhoneLines_ProviderAccounts_ProviderAccountId" FOREIGN KEY ("ProviderAccountId") REFERENCES "ProviderAccounts" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PhoneLines_ProviderInvoices_LastInvoiceId" FOREIGN KEY ("LastInvoiceId") REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PhoneLines_ProviderPlans_ProviderPlanId" FOREIGN KEY ("ProviderPlanId") REFERENCES "ProviderPlans" ("Id") ON DELETE SET NULL
);

CREATE TABLE "ProviderInvoiceItems" (
    "Id" character varying(36) NOT NULL,
    "InvoiceId" character varying(36) NOT NULL,
    "ParentId" character varying(36),
    "Description" character varying(512) NOT NULL,
    "Quantity" numeric(18,4) NOT NULL,
    "TotalPrice" numeric(18,2) NOT NULL,
    "ItemType" provider_invoice_item_type NOT NULL,
    "QuotaAmount" numeric(18,4),
    "ConsumedAmount" numeric(18,4),
    "Unit" invoice_item_unit,
    CONSTRAINT "PK_ProviderInvoiceItems" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderInvoiceItems_ProviderInvoiceItems_ParentId" FOREIGN KEY ("ParentId") REFERENCES "ProviderInvoiceItems" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoiceItems_ProviderInvoices_InvoiceId" FOREIGN KEY ("InvoiceId") REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderInvoiceServices" (
    "Id" character varying(36) NOT NULL,
    "InvoiceId" character varying(36) NOT NULL,
    "PlanId" character varying(36) NOT NULL,
    "Description" character varying(512) NOT NULL,
    "Quantity" numeric(18,4) NOT NULL,
    "TotalPrice" numeric(18,2) NOT NULL,
    "QuotaAmount" numeric(18,4),
    "ConsumedAmount" numeric(18,4),
    "Unit" invoice_item_unit,
    CONSTRAINT "PK_ProviderInvoiceServices" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderInvoiceServices_ProviderInvoices_InvoiceId" FOREIGN KEY ("InvoiceId") REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoiceServices_ProviderPlans_PlanId" FOREIGN KEY ("PlanId") REFERENCES "ProviderPlans" ("Id") ON DELETE CASCADE
);

CREATE TABLE "PhoneLineServices" (
    "Id" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "ProviderPlanServiceId" character varying(36) NOT NULL,
    "Name" text NOT NULL,
    "Code" text NOT NULL,
    "Recurring" boolean NOT NULL,
    "Price" numeric(18,2),
    "Active" boolean NOT NULL,
    CONSTRAINT "PK_PhoneLineServices" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_PhoneLineServices_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_PhoneLineServices_ProviderPlanServices_ProviderPlanServiceId" FOREIGN KEY ("ProviderPlanServiceId") REFERENCES "ProviderPlanServices" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderInvoicePhoneLines" (
    "PhoneLinesId" character varying(36) NOT NULL,
    "ProviderInvoicesId" character varying(36) NOT NULL,
    CONSTRAINT "PK_ProviderInvoicePhoneLines" PRIMARY KEY ("PhoneLinesId", "ProviderInvoicesId"),
    CONSTRAINT "FK_ProviderInvoicePhoneLines_PhoneLines_PhoneLinesId" FOREIGN KEY ("PhoneLinesId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoicePhoneLines_ProviderInvoices_ProviderInvoices~" FOREIGN KEY ("ProviderInvoicesId") REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE
);

CREATE TABLE "ProviderInvoiceQuotaSharing" (
    "Id" character varying(36) NOT NULL,
    "InvoiceId" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "Description" character varying(512) NOT NULL,
    "ConsumedAmount" numeric(18,4),
    CONSTRAINT "PK_ProviderInvoiceQuotaSharing" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_ProviderInvoiceQuotaSharing_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_ProviderInvoiceQuotaSharing_ProviderInvoices_InvoiceId" FOREIGN KEY ("InvoiceId") REFERENCES "ProviderInvoices" ("Id") ON DELETE CASCADE
);

CREATE TABLE "PhoneLineCustomerLinks" (
    "Id" character varying(36) NOT NULL,
    "PhoneLineId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "StartDate" date NOT NULL,
    "EndDate" date,
    CONSTRAINT "PK_PhoneLineCustomerLinks" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_PhoneLineCustomerLinks_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE RESTRICT,
    CONSTRAINT "FK_PhoneLineCustomerLinks_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE CASCADE
);

CREATE INDEX "IX_AuditLogs_EntityName" ON "AuditLogs" ("EntityName");
CREATE INDEX "IX_AuditLogs_EntityName_ChangedBy" ON "AuditLogs" ("EntityName", "ChangedBy");
CREATE INDEX "IX_BillingCycleProviderAccounts_ProviderAccountsId" ON "BillingCycleProviderAccounts" ("ProviderAccountsId");
CREATE INDEX "IX_BillingCycles_OrganizationId" ON "BillingCycles" ("OrganizationId");
CREATE UNIQUE INDEX "IX_BillingCycles_OrganizationId_Code" ON "BillingCycles" ("OrganizationId", "Code");
CREATE UNIQUE INDEX "IX_BillingCycles_OrganizationId_Name" ON "BillingCycles" ("OrganizationId", "Name");
CREATE INDEX "IX_BillingCycles_ProviderId" ON "BillingCycles" ("ProviderId");
CREATE UNIQUE INDEX "IX_ContractingCompanies_ProviderId_TaxId" ON "ContractingCompanies" ("ProviderId", "TaxId");
CREATE INDEX "IX_CostCenters_OrganizationId" ON "CostCenters" ("OrganizationId");
CREATE UNIQUE INDEX "IX_CostCenters_OrganizationId_Name" ON "CostCenters" ("OrganizationId", "Name");
CREATE INDEX "IX_CustomerAddresses_CustomerId" ON "CustomerAddresses" ("CustomerId");
CREATE INDEX "IX_CustomerAttachments_CustomerId" ON "CustomerAttachments" ("CustomerId");
CREATE INDEX "IX_CustomerAttachments_OrganizationId_CustomerId" ON "CustomerAttachments" ("OrganizationId", "CustomerId");
CREATE INDEX "IX_CustomerDocuments_CustomerId" ON "CustomerDocuments" ("CustomerId");
CREATE INDEX "IX_CustomerDocuments_DocumentType" ON "CustomerDocuments" ("DocumentType");
CREATE UNIQUE INDEX "IX_CustomerProcessingMonthManualReleases_CustomerId_Processing~" ON "CustomerProcessingMonthManualReleases" ("CustomerId", "ProcessingMonthId");
CREATE INDEX "IX_CustomerProcessingMonthManualReleases_OrganizationId" ON "CustomerProcessingMonthManualReleases" ("OrganizationId");
CREATE INDEX "IX_CustomerProcessingMonthManualReleases_ProcessingMonthId" ON "CustomerProcessingMonthManualReleases" ("ProcessingMonthId");
CREATE INDEX "IX_CustomerProviderLinks_ProviderId" ON "CustomerProviderLinks" ("ProviderId");
CREATE UNIQUE INDEX "IX_CustomerProviderLinks_CustomerId_ProviderId" ON "CustomerProviderLinks" ("CustomerId", "ProviderId") WHERE "EndDate" IS NULL;
CREATE INDEX "IX_CustomerProviderLinks_CustomerId_ProviderId_EndDate" ON "CustomerProviderLinks" ("CustomerId", "ProviderId", "EndDate");
CREATE INDEX "IX_CustomerProviderLinks_CustomerId_ProviderId_StartDate" ON "CustomerProviderLinks" ("CustomerId", "ProviderId", "StartDate");
CREATE INDEX "IX_Customers_OrganizationId" ON "Customers" ("OrganizationId");
CREATE UNIQUE INDEX "IX_Customers_OrganizationId_LegalName" ON "Customers" ("OrganizationId", "LegalName");
CREATE INDEX "IX_Customers_OrganizationId_Name" ON "Customers" ("OrganizationId", "Name");
CREATE INDEX "IX_PhoneLineCustomerLinks_CustomerId" ON "PhoneLineCustomerLinks" ("CustomerId");
CREATE UNIQUE INDEX "IX_PhoneLineCustomerLinks_PhoneLineId" ON "PhoneLineCustomerLinks" ("PhoneLineId") WHERE "EndDate" IS NULL;
CREATE INDEX "IX_PhoneLineCustomerLinks_PhoneLineId_StartDate" ON "PhoneLineCustomerLinks" ("PhoneLineId", "StartDate");
CREATE INDEX "IX_PhoneLines_CostCenterId" ON "PhoneLines" ("CostCenterId");
CREATE INDEX "IX_PhoneLines_LastInvoiceId" ON "PhoneLines" ("LastInvoiceId");
CREATE UNIQUE INDEX "IX_PhoneLines_Number" ON "PhoneLines" ("Number");
CREATE INDEX "IX_PhoneLines_ProviderAccountId" ON "PhoneLines" ("ProviderAccountId");
CREATE INDEX "IX_PhoneLines_ProviderPlanId" ON "PhoneLines" ("ProviderPlanId");
CREATE INDEX "IX_PhoneLines_Status_TransitionStartedAt" ON "PhoneLines" ("Status", "TransitionStartedAt");
CREATE INDEX "IX_PhoneLines_TitularLineId" ON "PhoneLines" ("TitularLineId");
CREATE UNIQUE INDEX "IX_PhoneLineServices_PhoneLineId_ProviderPlanServiceId" ON "PhoneLineServices" ("PhoneLineId", "ProviderPlanServiceId");
CREATE INDEX "IX_PhoneLineServices_ProviderPlanServiceId" ON "PhoneLineServices" ("ProviderPlanServiceId");
CREATE INDEX "IX_ProcessingMonths_OrganizationId" ON "ProcessingMonths" ("OrganizationId");
CREATE UNIQUE INDEX "IX_ProcessingMonths_OrganizationId_ProviderId_Year_Month" ON "ProcessingMonths" ("OrganizationId", "ProviderId", "Year", "Month");
CREATE INDEX "IX_ProcessingMonths_ProviderId" ON "ProcessingMonths" ("ProviderId");
CREATE INDEX "IX_ProviderAccounts_AccountNumber" ON "ProviderAccounts" ("AccountNumber");
CREATE UNIQUE INDEX "IX_ProviderAccounts_ContractingCompanyId_AccountNumber" ON "ProviderAccounts" ("ContractingCompanyId", "AccountNumber");
CREATE INDEX "IX_ProviderInvoiceImportRequests_OrganizationId" ON "ProviderInvoiceImportRequests" ("OrganizationId");
CREATE INDEX "IX_ProviderInvoiceImportRequests_OrganizationId_ProviderId" ON "ProviderInvoiceImportRequests" ("OrganizationId", "ProviderId");
CREATE INDEX "IX_ProviderInvoiceImportRequests_ProcessingMonthId" ON "ProviderInvoiceImportRequests" ("ProcessingMonthId");
CREATE INDEX "IX_ProviderInvoiceImportRequests_ProviderId" ON "ProviderInvoiceImportRequests" ("ProviderId");
CREATE INDEX "IX_ProviderInvoiceItems_InvoiceId" ON "ProviderInvoiceItems" ("InvoiceId");
CREATE INDEX "IX_ProviderInvoiceItems_ParentId" ON "ProviderInvoiceItems" ("ParentId");
CREATE INDEX "IX_ProviderInvoicePhoneLines_ProviderInvoicesId" ON "ProviderInvoicePhoneLines" ("ProviderInvoicesId");
CREATE INDEX "IX_ProviderInvoiceQuotaSharing_InvoiceId" ON "ProviderInvoiceQuotaSharing" ("InvoiceId");
CREATE INDEX "IX_ProviderInvoiceQuotaSharing_PhoneLineId" ON "ProviderInvoiceQuotaSharing" ("PhoneLineId");
CREATE INDEX "IX_ProviderInvoices_BillingCycleId" ON "ProviderInvoices" ("BillingCycleId");
CREATE INDEX "IX_ProviderInvoices_ContractingCompanyId" ON "ProviderInvoices" ("ContractingCompanyId");
CREATE INDEX "IX_ProviderInvoices_CostCenterId" ON "ProviderInvoices" ("CostCenterId");
CREATE INDEX "IX_ProviderInvoices_ParentInvoiceId" ON "ProviderInvoices" ("ParentInvoiceId");
CREATE INDEX "IX_ProviderInvoices_ProcessingMonthId" ON "ProviderInvoices" ("ProcessingMonthId");
CREATE UNIQUE INDEX "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Pro~" ON "ProviderInvoices" ("ProviderAccountId", "ContractingCompanyId", "ProcessingMonthId", "DueDate");
CREATE INDEX "IX_ProviderInvoiceServices_InvoiceId" ON "ProviderInvoiceServices" ("InvoiceId");
CREATE INDEX "IX_ProviderInvoiceServices_PlanId" ON "ProviderInvoiceServices" ("PlanId");
CREATE UNIQUE INDEX "IX_ProviderPlans_ProviderId_Code" ON "ProviderPlans" ("ProviderId", "Code");
CREATE INDEX "IX_ProviderPlanServices_ProviderPlanId" ON "ProviderPlanServices" ("ProviderPlanId");
CREATE INDEX "IX_Providers_OrganizationId" ON "Providers" ("OrganizationId");
CREATE INDEX "IX_Providers_OrganizationId_Active" ON "Providers" ("OrganizationId", "Active");
CREATE UNIQUE INDEX "IX_Providers_OrganizationId_Slug" ON "Providers" ("OrganizationId", "Slug");
