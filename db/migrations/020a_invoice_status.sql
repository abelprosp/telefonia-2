-- Commit this enum change before 021 uses it in index predicates.
-- Status substituta: as duas faturas permanecem.
-- No-op when the base schema (db/migrations) has not been applied yet.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'provider_invoice_status') THEN
        ALTER TYPE provider_invoice_status ADD VALUE IF NOT EXISTS 'substituted';
    END IF;
EXCEPTION
    WHEN duplicate_object THEN NULL;
    WHEN undefined_object THEN NULL;
END $$;
