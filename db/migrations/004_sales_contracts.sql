CREATE TYPE sale_status AS ENUM ('draft', 'confirmed', 'cancelled');
CREATE TYPE sale_line_item_type AS ENUM ('phone_line', 'device', 'other');
CREATE TYPE generated_contract_status AS ENUM ('pending', 'generated', 'failed');

CREATE TABLE "ContractTemplates" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Name" character varying(256) NOT NULL,
    "Code" character varying(64) NOT NULL,
    "BodyTemplate" text NOT NULL,
    "Active" boolean NOT NULL DEFAULT true,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_ContractTemplates" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_ContractTemplates_OrgCode" UNIQUE ("OrganizationId", "Code")
);

CREATE TABLE "Sales" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "SalespersonUserId" character varying(256) NOT NULL,
    "ContractTemplateId" character varying(36),
    "Status" sale_status NOT NULL DEFAULT 'draft',
    "SaleNumber" character varying(32) NOT NULL,
    "SoldAt" date,
    "Notes" character varying(4000),
    "TotalAmount" numeric(18,2) NOT NULL DEFAULT 0,
    "CreatedAt" timestamp with time zone NOT NULL,
    "UpdatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_Sales" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_Sales_Customers_CustomerId" FOREIGN KEY ("CustomerId") REFERENCES "Customers" ("Id") ON DELETE RESTRICT,
    CONSTRAINT "FK_Sales_ContractTemplates_ContractTemplateId" FOREIGN KEY ("ContractTemplateId") REFERENCES "ContractTemplates" ("Id") ON DELETE SET NULL
);

CREATE TABLE "SaleLineItems" (
    "Id" character varying(36) NOT NULL,
    "SaleId" character varying(36) NOT NULL,
    "LineItemType" sale_line_item_type NOT NULL,
    "Description" character varying(512) NOT NULL,
    "Quantity" numeric(18,4) NOT NULL DEFAULT 1,
    "UnitPrice" numeric(18,2) NOT NULL,
    "TotalPrice" numeric(18,2) NOT NULL,
    "PhoneLineId" character varying(36),
    "DeviceSku" character varying(128),
    "SortOrder" integer NOT NULL DEFAULT 0,
    CONSTRAINT "PK_SaleLineItems" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_SaleLineItems_Sales_SaleId" FOREIGN KEY ("SaleId") REFERENCES "Sales" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_SaleLineItems_PhoneLines_PhoneLineId" FOREIGN KEY ("PhoneLineId") REFERENCES "PhoneLines" ("Id") ON DELETE SET NULL
);

CREATE TABLE "GeneratedContracts" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "SaleId" character varying(36) NOT NULL,
    "ContractTemplateId" character varying(36) NOT NULL,
    "Status" generated_contract_status NOT NULL DEFAULT 'pending',
    "RenderedHtml" text,
    "GeneratedAt" timestamp with time zone,
    "CreatedAt" timestamp with time zone NOT NULL,
    CONSTRAINT "PK_GeneratedContracts" PRIMARY KEY ("Id"),
    CONSTRAINT "FK_GeneratedContracts_Sales_SaleId" FOREIGN KEY ("SaleId") REFERENCES "Sales" ("Id") ON DELETE CASCADE,
    CONSTRAINT "FK_GeneratedContracts_ContractTemplates_ContractTemplateId" FOREIGN KEY ("ContractTemplateId") REFERENCES "ContractTemplates" ("Id") ON DELETE RESTRICT
);

CREATE INDEX "IX_Sales_OrganizationId" ON "Sales" ("OrganizationId");
CREATE INDEX "IX_Sales_CustomerId" ON "Sales" ("CustomerId");
CREATE INDEX "IX_Sales_SalespersonUserId" ON "Sales" ("SalespersonUserId");
CREATE INDEX "IX_Sales_Status" ON "Sales" ("Status");
CREATE INDEX "IX_SaleLineItems_SaleId" ON "SaleLineItems" ("SaleId");
CREATE INDEX "IX_ContractTemplates_OrganizationId" ON "ContractTemplates" ("OrganizationId");
CREATE INDEX "IX_GeneratedContracts_SaleId" ON "GeneratedContracts" ("SaleId");

INSERT INTO "ContractTemplates" (
    "Id", "OrganizationId", "Name", "Code", "BodyTemplate", "Active", "CreatedAt", "UpdatedAt"
) VALUES (
    '00000000-0000-0000-0000-000000000101',
    '00000000-0000-0000-0000-000000000001',
    'Contrato padrão de prestação de serviços',
    'default_service',
    '<h1>Contrato de Prestação de Serviços</h1>
<p>Pelo presente instrumento, <strong>{{customer.name}}</strong>, inscrito(a) no documento <strong>{{customer.document}}</strong>, doravante CONTRATANTE, contrata os serviços abaixo descritos.</p>
<p><strong>Razão social:</strong> {{customer.legal_name}}</p>
<p><strong>Endereço:</strong> {{customer.address.full}}</p>
<p><strong>Data da venda:</strong> {{sale.sold_at}}</p>
<p><strong>Valor total:</strong> {{sale.total_amount}}</p>
<h2>Itens contratados</h2>
{{sale.items_table}}
<p><strong>Vendedor responsável:</strong> {{salesperson.name}}</p>
<p>As partes declaram estar de acordo com os termos acima.</p>',
    true,
    NOW(),
    NOW()
) ON CONFLICT ("Id") DO NOTHING;
