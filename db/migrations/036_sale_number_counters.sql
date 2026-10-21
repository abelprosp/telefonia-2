-- Numeração de vendas atômica por organização/dia.
-- O contador substitui COUNT(*)+1 e nunca reaproveita números apagados.

CREATE TABLE IF NOT EXISTS "SaleNumberCounters" (
    "OrganizationId" character varying(36) NOT NULL,
    "Day" character(8) NOT NULL,
    "LastValue" integer NOT NULL DEFAULT 0,
    CONSTRAINT "PK_SaleNumberCounters" PRIMARY KEY ("OrganizationId", "Day")
);

-- Semeia o contador com o maior número já emitido em cada dia. Sem isto, o
-- primeiro pedido após a migração poderia reutilizar um número existente.
INSERT INTO "SaleNumberCounters" ("OrganizationId", "Day", "LastValue")
SELECT "OrganizationId",
       substring("SaleNumber" from '^VND-([0-9]{8})-'),
       MAX((regexp_match("SaleNumber", '^VND-[0-9]{8}-([0-9]+)$'))[1]::int)
FROM "Sales"
WHERE "SaleNumber" ~ '^VND-[0-9]{8}-[0-9]+$'
GROUP BY 1, 2
ON CONFLICT ("OrganizationId", "Day") DO NOTHING;

-- Garante unicidade sem impedir a subida caso existam duplicados históricos.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM "Sales"
        GROUP BY "OrganizationId", "SaleNumber"
        HAVING COUNT(*) > 1
    ) THEN
        CREATE UNIQUE INDEX IF NOT EXISTS "UX_Sales_Org_SaleNumber"
            ON "Sales" ("OrganizationId", "SaleNumber");
    ELSE
        RAISE NOTICE 'UX_Sales_Org_SaleNumber não criado: existem números de venda duplicados';
    END IF;
END $$;
