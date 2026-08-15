-- Migração 019: Configurações da Organização (Empresa, Whitelabel e Parâmetros do Sistema)

CREATE TABLE IF NOT EXISTS "OrganizationSettings" (
    "OrganizationId" character varying(36) NOT NULL,
    
    -- Dados da Empresa
    "CompanyName" character varying(256) NOT NULL DEFAULT 'Luxus Telefonia Ltda',
    "TradingName" character varying(256) NOT NULL DEFAULT 'Luxus Connect',
    "Cnpj" character varying(32) NOT NULL DEFAULT '11.309.896/0001-01',
    "StateRegistration" character varying(32) NOT NULL DEFAULT '',
    "Email" character varying(256) NOT NULL DEFAULT 'contato@luxusconnect.com.br',
    "Phone" character varying(32) NOT NULL DEFAULT '(11) 99999-9999',
    "Website" character varying(256) NOT NULL DEFAULT 'https://telefonia.redobrai.online',
    "ZipCode" character varying(16) NOT NULL DEFAULT '',
    "Street" character varying(256) NOT NULL DEFAULT '',
    "Number" character varying(32) NOT NULL DEFAULT '',
    "Complement" character varying(128) NOT NULL DEFAULT '',
    "Neighborhood" character varying(128) NOT NULL DEFAULT '',
    "City" character varying(128) NOT NULL DEFAULT '',
    "State" character varying(8) NOT NULL DEFAULT '',

    -- Configurações de Whitelabel
    "AppName" character varying(128) NOT NULL DEFAULT 'Luxus Connect',
    "AppSlogan" character varying(256) NOT NULL DEFAULT 'Gestão Inteligente de Telefonia',
    "LogoUrl" text NOT NULL DEFAULT '',
    "DarkLogoUrl" text NOT NULL DEFAULT '',
    "FaviconUrl" text NOT NULL DEFAULT '',
    "PrimaryColor" character varying(32) NOT NULL DEFAULT '#0f766e',
    "SupportEmail" character varying(256) NOT NULL DEFAULT 'suporte@luxusconnect.com.br',
    "SupportPhone" character varying(32) NOT NULL DEFAULT '(11) 99999-9999',
    "FooterText" text NOT NULL DEFAULT '© 2026 Luxus Connect. Todos os direitos reservados.',

    -- Parâmetros do Sistema
    "DefaultDueDay" integer NOT NULL DEFAULT 10,
    "LateFeePercentage" numeric(5,2) NOT NULL DEFAULT 2.00,
    "InterestRateMonthly" numeric(5,2) NOT NULL DEFAULT 1.00,
    "DaysBeforeDueReminder" integer NOT NULL DEFAULT 3,
    "DaysAfterDueReminder" integer NOT NULL DEFAULT 2,
    "AutoSendInvoiceEmail" boolean NOT NULL DEFAULT TRUE,
    "AutoSendCollectionReminder" boolean NOT NULL DEFAULT FALSE,

    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedBy" character varying(36),

    CONSTRAINT "PK_OrganizationSettings" PRIMARY KEY ("OrganizationId")
);
