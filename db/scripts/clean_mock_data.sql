-- ==============================================================================
-- LUXUS CONNECT - LIMPEZA COMPLETA DE DADOS MOCK
-- Remove todos os registros identificados com prefixo [MOCK] ou IDs mock-...
-- Restaura as linhas vinculadas para o status 'in_stock'.
-- Não afeta nenhum dado real de clientes, vendas ou faturas do usuário.
-- ==============================================================================

BEGIN;

-- 1. Restaurar status das linhas mockadas de volta para 'in_stock'
UPDATE "PhoneLines"
SET "Status" = 'in_stock', "BaseCost" = NULL, "CostWithConsumption" = NULL
WHERE "Id" IN (
    SELECT "PhoneLineId" FROM "PhoneLineCustomerLinks"
    WHERE "Id" LIKE 'mock-link-%' OR "CustomerId" LIKE 'mock-cust-%'
);

-- 2. Deletar Itens de Composição de Faturamento Mock
DELETE FROM "LineBillingCompositionItems"
WHERE "Id" LIKE 'mock-comp-%'
   OR "LineBillingProcessingId" LIKE 'mock-proc-%';

-- 3. Deletar Processamentos de Faturamento Mock
DELETE FROM "LineBillingProcessings"
WHERE "Id" LIKE 'mock-proc-%'
   OR "PhoneLineCustomerLinkId" LIKE 'mock-link-%';

-- 4. Deletar Vínculos de Linhas com Clientes Mock
DELETE FROM "PhoneLineCustomerLinks"
WHERE "Id" LIKE 'mock-link-%'
   OR "CustomerId" LIKE 'mock-cust-%';

-- 5. Deletar Faturas de Clientes Mock
DELETE FROM "CustomerBillingDocuments"
WHERE "Id" LIKE 'mock-doc-inv-%'
   OR "InvoiceNumber" LIKE 'FAT-MOCK-%'
   OR "CustomerId" LIKE 'mock-cust-%';

-- 6. Deletar Contas a Receber Mock
DELETE FROM "AccountsReceivable"
WHERE "Id" LIKE 'mock-ar-%'
   OR "DocumentNumber" LIKE 'FAT-MOCK-%'
   OR "CustomerId" LIKE 'mock-cust-%';

-- 7. Deletar Contas a Pagar Mock
DELETE FROM "AccountsPayable"
WHERE "Id" LIKE 'mock-ap-%'
   OR "DocumentNumber" LIKE 'VIVO-MOCK-%';

-- 8. Deletar Vendas e Itens de Venda Mock
DELETE FROM "SaleLineItems"
WHERE "SaleId" IN (SELECT "Id" FROM "Sales" WHERE "Id" LIKE 'mock-sale-%' OR "CustomerId" LIKE 'mock-cust-%');

DELETE FROM "Sales"
WHERE "Id" LIKE 'mock-sale-%'
   OR "SaleNumber" LIKE 'VEN-MOCK-%'
   OR "CustomerId" LIKE 'mock-cust-%';

-- 9. Deletar Endereços e Documentos dos Clientes Mock
DELETE FROM "CustomerAddresses"
WHERE "Id" LIKE 'mock-addr-%'
   OR "CustomerId" LIKE 'mock-cust-%';

DELETE FROM "CustomerDocuments"
WHERE "Id" LIKE 'mock-doc-%'
   OR "CustomerId" LIKE 'mock-cust-%';

-- 10. Deletar os Clientes Mock
DELETE FROM "Customers"
WHERE "Id" LIKE 'mock-cust-%'
   OR "Name" LIKE '[MOCK]%';

COMMIT;
