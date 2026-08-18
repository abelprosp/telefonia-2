-- ==============================================================================
-- LUXUS CONNECT - SEED DE DADOS MOCK PARA DEMONSTRAÇÃO VISUAL
-- Todos os registros criados usam prefixo '[MOCK]' ou IDs 'mock-...'
-- Para remover tudo a qualquer momento, execute 'clean_mock_data.sql'
-- ==============================================================================

BEGIN;

DO $$
DECLARE
    v_org_id VARCHAR(36);
    v_provider_id VARCHAR(36);
    v_plan_id VARCHAR(36);
    v_month_id VARCHAR(36);
    v_cust1_id VARCHAR(36) := 'mock-cust-00000000-0000-0000-000001';
    v_cust2_id VARCHAR(36) := 'mock-cust-00000000-0000-0000-000002';
    v_cust3_id VARCHAR(36) := 'mock-cust-00000000-0000-0000-000003';
    v_sale1_id VARCHAR(36) := 'mock-sale-00000000-0000-0000-000001';
    v_sale2_id VARCHAR(36) := 'mock-sale-00000000-0000-0000-000002';
    v_line_rec RECORD;
    v_link_id VARCHAR(36);
    v_proc_id VARCHAR(36);
    v_idx INT := 0;
BEGIN
    -- 1. Obter organização ativa
    SELECT "OrganizationId" INTO v_org_id FROM "OrganizationSettings" LIMIT 1;
    IF v_org_id IS NULL THEN
        v_org_id := '00000000-0000-0000-0000-000000000001';
    END IF;

    -- 2. Obter provedor e plano padrão
    SELECT "Id" INTO v_provider_id FROM "Providers" WHERE "OrganizationId" = v_org_id LIMIT 1;
    SELECT "Id" INTO v_plan_id FROM "ProviderPlans" WHERE "ProviderId" = v_provider_id LIMIT 1;

    -- 3. Obter ou identificar Mês de Processamento atual
    SELECT "Id" INTO v_month_id FROM "ProcessingMonths" WHERE "OrganizationId" = v_org_id ORDER BY "Year" DESC, "Month" DESC LIMIT 1;

    -- 4. Criar Clientes Mock
    INSERT INTO "Customers" (
        "Id", "OrganizationId", "Type", "Name", "LegalName", "BirthOrOpeningDate", "Active", "BillingEmail", "IsReseller"
    ) VALUES 
    (v_cust1_id, v_org_id, 'pj', '[MOCK] NexaTech Soluções Digitais', '[MOCK] NexaTech Soluções Digitais Ltda', '2021-03-15', true, 'financeiro@nexatech.mock.br', false),
    (v_cust2_id, v_org_id, 'pj', '[MOCK] Logística Brasil Express', '[MOCK] Logística Brasil Express S/A', '2019-07-20', true, 'contas@logexpress.mock.br', false),
    (v_cust3_id, v_org_id, 'pj', '[MOCK] Vanguarda Distribuidora', '[MOCK] Vanguarda Distribuidora de Bebidas Ltda', '2018-11-10', true, 'faturamento@vanguarda.mock.br', false)
    ON CONFLICT ("Id") DO NOTHING;

    -- Documentos dos clientes mock
    INSERT INTO "CustomerDocuments" ("Id", "CustomerId", "DocumentType", "Number") VALUES
    ('mock-doc-0001', v_cust1_id, 'cnpj', '45812934000192'),
    ('mock-doc-0002', v_cust2_id, 'cnpj', '19482015000164'),
    ('mock-doc-0003', v_cust3_id, 'cnpj', '82910475000130')
    ON CONFLICT ("Id") DO NOTHING;

    -- Endereços dos clientes mock
    INSERT INTO "CustomerAddresses" ("Id", "CustomerId", "ZipCode", "Street", "Number", "Neighborhood", "City", "State", "Country") VALUES
    ('mock-addr-0001', v_cust1_id, '01310-100', 'Avenida Paulista', '1000', 'Bela Vista', 'São Paulo', 'SP', 'Brasil'),
    ('mock-addr-0002', v_cust2_id, '80010-000', 'Rua XV de Novembro', '500', 'Centro', 'Curitiba', 'PR', 'Brasil'),
    ('mock-addr-0003', v_cust3_id, '30130-000', 'Avenida Afonso Pena', '1500', 'Funcionários', 'Belo Horizonte', 'MG', 'Brasil')
    ON CONFLICT ("Id") DO NOTHING;

    -- 5. Ativar e vincular 24 linhas do estoque aos clientes mock (8 para cada cliente)
    FOR v_line_rec IN (
        SELECT "Id", "Number" FROM "PhoneLines"
        WHERE "Status" = 'in_stock'
        LIMIT 24
    ) LOOP
        v_idx := v_idx + 1;
        v_link_id := 'mock-link-' || LPAD(v_idx::text, 6, '0');
        v_proc_id := 'mock-proc-' || LPAD(v_idx::text, 6, '0');

        -- Determinar qual cliente mock recebe a linha
        DECLARE
            v_target_cust VARCHAR(36);
            v_monthly_val NUMERIC(18,2) := 79.90;
            v_base_cost NUMERIC(18,2) := 39.90;
        BEGIN
            IF v_idx <= 8 THEN
                v_target_cust := v_cust1_id;
            ELSIF v_idx <= 16 THEN
                v_target_cust := v_cust2_id;
            ELSE
                v_target_cust := v_cust3_id;
            END IF;

            -- Atualizar linha para ativa com base_cost
            UPDATE "PhoneLines"
            SET "Status" = 'active', "BaseCost" = v_base_cost, "CostWithConsumption" = v_base_cost
            WHERE "Id" = v_line_rec."Id";

            -- Criar vínculo da linha com cliente
            INSERT INTO "PhoneLineCustomerLinks" (
                "Id", "PhoneLineId", "CustomerId", "StartDate", "MonthlyAmount"
            ) VALUES (
                v_link_id, v_line_rec."Id", v_target_cust, '2026-01-10', v_monthly_val
            ) ON CONFLICT ("Id") DO NOTHING;

            -- Criar processamento de faturamento mensal da linha
            INSERT INTO "LineBillingProcessings" (
                "Id", "PhoneLineCustomerLinkId", "Perspective", "Label", "MirrorFromPrimary", "Active", "CreatedAt", "UpdatedAt"
            ) VALUES (
                v_proc_id, v_link_id, 'luxus_customer', 'Faturamento Mensal Linha ' || v_line_rec."Number", false, true, now(), now()
            ) ON CONFLICT ("Id") DO NOTHING;

            -- Adicionar composição do faturamento (Plano Voz/Dados + SVA de Segurança)
            INSERT INTO "LineBillingCompositionItems" (
                "Id", "ProcessingId", "ItemType", "ServiceType", "Description", "Amount", "Quantity", "Active", "CreatedAt", "UpdatedAt", "Proportional"
            ) VALUES
            ('mock-comp-pln-' || v_idx::text, v_proc_id, 'service', 'subscription', 'Plano Smart Ilimitado 40GB', 69.90, 1, true, now(), now(), false),
            ('mock-comp-sva-' || v_idx::text, v_proc_id, 'service', 'other', 'Segurança Digital & Backup Móvel', 10.00, 1, true, now(), now(), false)
            ON CONFLICT ("Id") DO NOTHING;
        END;
    END LOOP;

    -- 6. Criar Vendas Mock (Comercial)
    INSERT INTO "Sales" (
        "Id", "OrganizationId", "CustomerId", "SalespersonUserId", "SaleNumber", "Status", "TotalAmount", "Notes", "SoldAt", "CreatedAt", "UpdatedAt"
    ) VALUES
    (v_sale1_id, v_org_id, v_cust1_id, 'admin', 'VEN-MOCK-001', 'confirmed', 639.20, 'Venda corporativa de 8 linhas Smart', CURRENT_DATE, now(), now()),
    (v_sale2_id, v_org_id, v_cust2_id, 'admin', 'VEN-MOCK-002', 'confirmed', 639.20, 'Venda corporativa frota 8 linhas', CURRENT_DATE, now(), now())
    ON CONFLICT ("Id") DO NOTHING;

    INSERT INTO "SaleLineItems" (
        "Id", "SaleId", "LineItemType", "Description", "Quantity", "UnitPrice", "TotalPrice", "SortOrder"
    ) VALUES
    ('mock-sli-0001', v_sale1_id, 'phone_line', 'Linhas Telefônicas Corporativas Smart 40GB (x8)', 8, 79.90, 639.20, 1),
    ('mock-sli-0002', v_sale2_id, 'phone_line', 'Linhas Telefônicas Frota Smart 40GB (x8)', 8, 79.90, 639.20, 1)
    ON CONFLICT ("Id") DO NOTHING;

    -- 7. Criar Faturas de Clientes Mock (CustomerBillingDocuments) com Contas a Receber
    INSERT INTO "AccountsReceivable" (
        "Id", "OrganizationId", "CustomerId", "ProcessingMonthId", "Description", "IssueDate", "DueDate", "Amount", "ReceivedAmount", "Status", "CreatedAt", "UpdatedAt"
    ) VALUES
    ('mock-ar-0001', v_org_id, v_cust1_id, v_month_id, 'Fatura Telefonia 08/2026 - NexaTech', '2026-08-01', '2026-08-15', 639.20, 639.20, 'settled', now(), now()),
    ('mock-ar-0002', v_org_id, v_cust2_id, v_month_id, 'Fatura Telefonia 08/2026 - LogExpress', '2026-08-01', '2026-08-20', 639.20, 0.00, 'open', now(), now()),
    ('mock-ar-0003', v_org_id, v_cust3_id, v_month_id, 'Fatura Telefonia 08/2026 - Vanguarda', '2026-08-01', '2026-08-25', 639.20, 0.00, 'open', now(), now())
    ON CONFLICT ("Id") DO NOTHING;

    INSERT INTO "CustomerBillingDocuments" (
        "Id", "OrganizationId", "CustomerId", "AccountsReceivableId", "ProcessingMonthId", "InvoiceNumber",
        "IssueDate", "DueDate", "Amount", "Status", "RecipientEmail", "EmailSubject", "EmailBodyHtml",
        "CreatedAt", "UpdatedAt"
    ) VALUES
    ('mock-doc-inv-001', v_org_id, v_cust1_id, 'mock-ar-0001', v_month_id, 'FAT-MOCK-101', '2026-08-01', '2026-08-15', 639.20, 'sent', 'financeiro@nexatech.mock.br', 'Sua Fatura de Telefonia - Agosto/2026', '<p>Fatura paga com sucesso.</p>', now(), now()),
    ('mock-doc-inv-002', v_org_id, v_cust2_id, 'mock-ar-0002', v_month_id, 'FAT-MOCK-102', '2026-08-01', '2026-08-20', 639.20, 'ready', 'contas@logexpress.mock.br', 'Sua Fatura de Telefonia - Agosto/2026', '<p>Fatura pronta para pagamento.</p>', now(), now()),
    ('mock-doc-inv-003', v_org_id, v_cust3_id, 'mock-ar-0003', v_month_id, 'FAT-MOCK-103', '2026-08-01', '2026-08-25', 639.20, 'ready', 'faturamento@vanguarda.mock.br', 'Sua Fatura de Telefonia - Agosto/2026', '<p>Fatura pronta para pagamento.</p>', now(), now())
    ON CONFLICT ("Id") DO NOTHING;

END $$;

COMMIT;
