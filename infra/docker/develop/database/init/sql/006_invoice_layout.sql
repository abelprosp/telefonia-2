CREATE TABLE "InvoiceLayoutTemplates" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Code" character varying(64) NOT NULL,
    "ConfigJson" jsonb NOT NULL,
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_InvoiceLayoutTemplates" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_InvoiceLayoutTemplates_OrgCode" UNIQUE ("OrganizationId", "Code")
);

CREATE INDEX "IX_InvoiceLayoutTemplates_OrganizationId" ON "InvoiceLayoutTemplates" ("OrganizationId");

INSERT INTO "InvoiceLayoutTemplates" (
    "Id", "OrganizationId", "Name", "Code", "ConfigJson", "Active", "CreatedAt", "UpdatedAt"
) VALUES (
    '00000000-0000-0000-0000-000000000301',
    '00000000-0000-0000-0000-000000000001',
    'Detalhamento padrão Luxus',
    'default-invoice-layout',
    '{
      "theme": {
        "primaryColor": "#4a4a4a",
        "accentColor": "#00a0c6",
        "borderColor": "#222222",
        "headerBackground": "#ffffff",
        "titleColor": "#1a1a1a",
        "textColor": "#333333",
        "tableHeaderBackground": "#f7f7f7",
        "borderRadius": 12
      },
      "branding": {
        "logoDataUrl": "",
        "companyName": "LUXUS",
        "tagline": "SOLUÇÃO EM TELEFONIA",
        "documentTitle": "Detalhamento da Fatura"
      },
      "sections": {
        "userData": { "enabled": true, "title": "Dados do Usuário" },
        "accountValue": { "enabled": true, "title": "VALOR DA SUA CONTA" },
        "billingDates": { "enabled": true },
        "accountSummary": { "enabled": true, "title": "Resumo da Conta" },
        "detailedConsumption": { "enabled": true, "title": "Consumo Detalhado" }
      },
      "labels": {
        "name": "Nome:",
        "address": "Endereço:",
        "phone": "Número do telefone:",
        "totalServices": "Total Serviços:",
        "discounts": "Descontos:",
        "billingPeriod": "Período de faturamento:",
        "referenceMonth": "Mês de referência:",
        "dueDate": "Data de Vencimento:",
        "description": "Descrição",
        "quantity": "Quantidade",
        "type": "Tipo",
        "unitPrice": "Preço Unitário",
        "total": "Total",
        "totalLabel": "Total:"
      }
    }'::jsonb,
    true,
    NOW(),
    NOW()
);
