ALTER TABLE "PhoneLineCustomerLinks"
    ADD COLUMN IF NOT EXISTS "MonthlyAmount" numeric(18,2);
