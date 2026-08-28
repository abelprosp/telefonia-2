-- Schema mínimo aplicado no arranque da API quando a VPS não correu db/migrations.

ALTER TABLE IF EXISTS "OrganizationSettings"
    ADD COLUMN IF NOT EXISTS "ProrataDivisor" integer NOT NULL DEFAULT 30;

CREATE TABLE IF NOT EXISTS "PortalCustomerLinks" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "UserId" character varying(256) NOT NULL,
    "CustomerId" character varying(36) NOT NULL,
    "Document" character varying(32) NOT NULL,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_PortalCustomerLinks" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_User" UNIQUE ("OrganizationId", "UserId"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_Document" UNIQUE ("OrganizationId", "Document"),
    CONSTRAINT "UX_PortalCustomerLinks_Org_Customer" UNIQUE ("OrganizationId", "CustomerId")
);

CREATE TABLE IF NOT EXISTS "SupportTickets" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "Number" integer NOT NULL,
    "Title" character varying(256) NOT NULL,
    "Category" character varying(64) NOT NULL DEFAULT 'geral',
    "Priority" character varying(16) NOT NULL DEFAULT 'media',
    "Status" character varying(32) NOT NULL DEFAULT 'aberto',
    "SlaDueAt" timestamp with time zone,
    "AssigneeUserId" character varying(256),
    "RequesterUserId" character varying(256),
    "CustomerId" character varying(36),
    "PhoneLineId" character varying(36),
    "ChargeRef" character varying(128),
    "InvoiceId" character varying(36),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "ResolvedAt" timestamp with time zone,
    "ClosedAt" timestamp with time zone,
    CONSTRAINT "PK_SupportTickets" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_SupportTickets_Org_Number" UNIQUE ("OrganizationId", "Number")
);

CREATE TABLE IF NOT EXISTS "SupportTicketMessages" (
    "Id" character varying(36) NOT NULL,
    "TicketId" character varying(36) NOT NULL,
    "AuthorUserId" character varying(256),
    "AuthorName" character varying(256),
    "Visibility" character varying(16) NOT NULL DEFAULT 'public',
    "Body" text NOT NULL,
    "AttachmentKey" character varying(512),
    "AttachmentName" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_SupportTicketMessages" PRIMARY KEY ("Id")
);

CREATE TABLE IF NOT EXISTS "SupportTicketHistory" (
    "Id" character varying(36) NOT NULL,
    "TicketId" character varying(36) NOT NULL,
    "ActorUserId" character varying(256),
    "EventType" character varying(64) NOT NULL,
    "FromValue" text,
    "ToValue" text,
    "Notes" text,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_SupportTicketHistory" PRIMARY KEY ("Id")
);

CREATE TABLE IF NOT EXISTS "ApprovalRequests" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "ActionType" character varying(64) NOT NULL,
    "EntityType" character varying(64) NOT NULL,
    "EntityId" character varying(36) NOT NULL,
    "Status" character varying(32) NOT NULL DEFAULT 'pending_first',
    "RequesterUserId" character varying(256) NOT NULL,
    "FirstApproverUserId" character varying(256),
    "SecondApproverUserId" character varying(256),
    "Justification" text,
    "Payload" jsonb,
    "BeforeSnapshot" jsonb,
    "AfterSnapshot" jsonb,
    "RejectionReason" text,
    "RejectedBy" character varying(256),
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    "FirstApprovedAt" timestamp with time zone,
    "SecondApprovedAt" timestamp with time zone,
    "RejectedAt" timestamp with time zone,
    "ExecutedAt" timestamp with time zone,
    CONSTRAINT "PK_ApprovalRequests" PRIMARY KEY ("Id")
);

CREATE TABLE IF NOT EXISTS "OperationMetrics" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36),
    "Operation" character varying(128) NOT NULL,
    "DurationMs" integer NOT NULL,
    "Success" boolean NOT NULL DEFAULT true,
    "Metadata" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT "PK_OperationMetrics" PRIMARY KEY ("Id")
);
