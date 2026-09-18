ALTER TABLE "ProviderInvoices"
  ADD COLUMN IF NOT EXISTS "DigitableLine" character varying(128),
  ADD COLUMN IF NOT EXISTS "PixQrCode" text,
  ADD COLUMN IF NOT EXISTS "Barcode" character varying(64);
