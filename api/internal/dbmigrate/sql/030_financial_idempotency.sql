CREATE UNIQUE INDEX IF NOT EXISTS "UX_AccountsPayable_ProviderInvoice_Active"
    ON "AccountsPayable" ("ProviderInvoiceId")
    WHERE "ProviderInvoiceId" IS NOT NULL AND "Status" <> 'cancelled';
