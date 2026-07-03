CREATE TYPE customer_billing_document_status AS ENUM ('draft', 'ready', 'sent', 'cancelled');
CREATE TYPE invoice_email_template_kind AS ENUM ('billing_invoice', 'collection_reminder', 'collection_final');
CREATE TYPE collection_reminder_status AS ENUM ('pending', 'sent', 'failed', 'cancelled');

ALTER TABLE "Customers"
    ADD COLUMN IF NOT EXISTS "BillingEmail" character varying(256);

CREATE TABLE "InvoiceEmailTemplates" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Code" character varying(64) NOT NULL,
    "Kind" invoice_email_template_kind NOT NULL DEFAULT 'billing_invoice',
    "SubjectTemplate" character varying(512) NOT NULL,
    "BodyTemplateHtml" text NOT NULL,
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_InvoiceEmailTemplates" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_InvoiceEmailTemplates_OrgCode" UNIQUE ("OrganizationId", "Code")
);

CREATE TABLE "CustomerBillingDocuments" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "AccountsReceivableId" character varying(36),
    "ProcessingMonthId" character varying(36),
    "InvoiceNumber" character varying(64) NOT NULL,
    "IssueDate" date NOT NULL,
    "DueDate" date NOT NULL,
    "Amount" numeric(18,2) NOT NULL,
    "Status" customer_billing_document_status NOT NULL DEFAULT 'draft',
    "RecipientEmail" character varying(256) NOT NULL,
    "EmailSubject" character varying(512) NOT NULL,
    "EmailBodyHtml" text NOT NULL,
    "SentAt" timestamp with time zone,
    "LastSentAt" timestamp with time zone,
    "SendCount" integer NOT NULL DEFAULT 0,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CustomerBillingDocuments" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerBillingDocuments_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_CustomerBillingDocuments_AccountsReceivable_AccountsReceivableId" FOREIGN KEY ("AccountsReceivableId") REFERENCES "AccountsReceivable" ("Id") ON DELETE SET NULL,
    CONSTRAINT "FK_CustomerBillingDocuments_ProcessingMonths_ProcessingMonthId" FOREIGN KEY ("ProcessingMonthId") REFERENCES "ProcessingMonths" ("Id") ON DELETE SET NULL
);

CREATE TABLE "CustomerBillingSendLog" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "DocumentId" character varying(36) NOT NULL,
    "RecipientEmail" character varying(256) NOT NULL,
    "Subject" character varying(512) NOT NULL,
    "Success" boolean NOT NULL,
    "ErrorMessage" character varying(4000),
    "SentByUserId" character varying(256) NOT NULL,
    "SentAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CustomerBillingSendLog" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CustomerBillingSendLog_Documents_DocumentId" FOREIGN KEY ("DocumentId") REFERENCES "CustomerBillingDocuments" ("Id") ON DELETE CASCADE
);

CREATE TABLE "CollectionReminders" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "AccountsReceivableId" character varying(36) NOT NULL,
    "ReminderLevel" integer NOT NULL DEFAULT 1,
    "RecipientEmail" character varying(256) NOT NULL,
    "Subject" character varying(512) NOT NULL,
    "BodyHtml" text NOT NULL,
    "Status" collection_reminder_status NOT NULL DEFAULT 'pending',
    "ErrorMessage" character varying(4000),
    "SentByUserId" character varying(256),
    "SentAt" timestamp with time zone,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_CollectionReminders" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_CollectionReminders_AccountsReceivable_AccountsReceivableId" FOREIGN KEY ("AccountsReceivableId") REFERENCES "AccountsReceivable" ("Id") ON DELETE CASCADE
);

CREATE INDEX "IX_CustomerBillingDocuments_OrganizationId" ON "CustomerBillingDocuments" ("OrganizationId");
CREATE INDEX "IX_CustomerBillingDocuments_CustomerId" ON "CustomerBillingDocuments" ("CustomerId");
CREATE INDEX "IX_CustomerBillingDocuments_Status" ON "CustomerBillingDocuments" ("Status");
CREATE INDEX "IX_InvoiceEmailTemplates_OrganizationId" ON "InvoiceEmailTemplates" ("OrganizationId");
CREATE INDEX "IX_CollectionReminders_OrganizationId" ON "CollectionReminders" ("OrganizationId");
CREATE INDEX "IX_CollectionReminders_AccountsReceivableId" ON "CollectionReminders" ("AccountsReceivableId");

INSERT INTO "InvoiceEmailTemplates" (
    "Id", "OrganizationId", "Name", "Code", "Kind", "SubjectTemplate", "BodyTemplateHtml", "Active", "CreatedAt", "UpdatedAt"
) VALUES (
    '00000000-0000-0000-0000-000000000201',
    '00000000-0000-0000-0000-000000000001',
    'Fatura mensal padrão',
    'default-billing-invoice',
    'billing_invoice',
    'Fatura {{invoice.number}} — {{customer.name}}',
    '<h2>Fatura {{invoice.number}}</h2>
<p>Olá, <strong>{{customer.name}}</strong>.</p>
<p>Segue a fatura referente ao período, no valor de <strong>{{invoice.amount}}</strong>, com vencimento em <strong>{{invoice.due_date}}</strong>.</p>
<p>{{invoice.description}}</p>
<p>Em caso de dúvidas, entre em contato conosco.</p>
<p>Atenciosamente,<br/>Luxus Connect</p>',
    true,
    NOW(),
    NOW()
),
(
    '00000000-0000-0000-0000-000000000202',
    '00000000-0000-0000-0000-000000000001',
    'Cobrança de inadimplência',
    'default-collection-reminder',
    'collection_reminder',
    'Lembrete de pagamento — {{customer.name}}',
    '<h2>Lembrete de pagamento</h2>
<p>Olá, <strong>{{customer.name}}</strong>.</p>
<p>Identificamos pendência no valor de <strong>{{invoice.amount}}</strong>, com vencimento em <strong>{{invoice.due_date}}</strong>.</p>
<p>Por favor, regularize o pagamento o quanto antes.</p>
<p>Atenciosamente,<br/>Luxus Connect</p>',
    true,
    NOW(),
    NOW()
);
