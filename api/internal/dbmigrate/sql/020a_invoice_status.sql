-- Commit this enum change before 021 uses it in index predicates.
-- Status substituta: as duas faturas permanecem
DO $$ BEGIN
    ALTER TYPE provider_invoice_status ADD VALUE 'substituted';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

