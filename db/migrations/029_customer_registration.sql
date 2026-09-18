ALTER TABLE "Customers" ADD COLUMN IF NOT EXISTS "RegistrationProfile" jsonb NOT NULL DEFAULT '{}'::jsonb;
