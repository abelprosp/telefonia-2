-- Migração 020: Máquinas de Estado Formais (Regras de Transição e Trilha de Execução de Estados)

CREATE TABLE IF NOT EXISTS "StateTransitionRules" (
    "Id" character varying(36) NOT NULL,
    "EntityType" character varying(64) NOT NULL,
    "FromState" character varying(64) NOT NULL,
    "ToState" character varying(64) NOT NULL,
    "AllowedRoles" character varying(256) NOT NULL DEFAULT 'operational,master',
    "Description" text NOT NULL DEFAULT '',
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),

    CONSTRAINT "PK_StateTransitionRules" PRIMARY KEY ("Id"),
    CONSTRAINT "UX_StateTransitionRules_Entity_From_To" UNIQUE ("EntityType", "FromState", "ToState")
);

CREATE TABLE IF NOT EXISTS "StateTransitionLogs" (
    "Id" character varying(36) NOT NULL,
    "OrganizationId" character varying(36) NOT NULL,
    "EntityType" character varying(64) NOT NULL,
    "EntityId" character varying(36) NOT NULL,
    "FromState" character varying(64) NOT NULL,
    "ToState" character varying(64) NOT NULL,
    "TriggerEvent" character varying(128) NOT NULL,
    "Justification" text,
    "ActorUserId" character varying(256),
    "Metadata" jsonb,
    "CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),

    CONSTRAINT "PK_StateTransitionLogs" PRIMARY KEY ("Id")
);

CREATE INDEX IF NOT EXISTS "IX_StateTransitionLogs_Entity" 
    ON "StateTransitionLogs" ("EntityType", "EntityId");

CREATE INDEX IF NOT EXISTS "IX_StateTransitionLogs_Org_Created" 
    ON "StateTransitionLogs" ("OrganizationId", "CreatedAt" DESC);

-- Seeds: Transições Válidas para phone_line
INSERT INTO "StateTransitionRules" ("Id", "EntityType", "FromState", "ToState", "AllowedRoles", "Description")
VALUES
    -- in_stock
    ('rule_pl_stock_trans', 'phone_line', 'in_stock', 'in_transition', 'operational,master', 'Vinculação antecipada com trâmite'),
    ('rule_pl_stock_active', 'phone_line', 'in_stock', 'active', 'operational,master', 'Ativação/vinculação direta ou importação'),
    ('rule_pl_stock_inactive', 'phone_line', 'in_stock', 'inactive', 'operational,master', 'Linha de estoque ausente em fatura'),
    ('rule_pl_stock_cancelled', 'phone_line', 'in_stock', 'cancelled', 'operational,master', 'Cancelamento ou descarte do número'),

    -- in_transition
    ('rule_pl_trans_active', 'phone_line', 'in_transition', 'active', 'operational,master', 'Conciliação automática via fatura ou ativação'),
    ('rule_pl_trans_stock', 'phone_line', 'in_transition', 'in_stock', 'operational,master', 'Desvinculação/cancelamento do trâmite de migração'),
    ('rule_pl_trans_cancelled', 'phone_line', 'in_transition', 'cancelled', 'operational,master', 'Cancelamento do trâmite de ativação'),

    -- active
    ('rule_pl_act_awaiting', 'phone_line', 'active', 'awaiting_invoice', 'operational,master', 'Linha sumiu da fatura atual'),
    ('rule_pl_act_suspended', 'phone_line', 'active', 'suspended', 'operational,master', 'Suspensão temporária a pedido ou inadimplência'),
    ('rule_pl_act_stock', 'phone_line', 'active', 'in_stock', 'operational,master', 'Desvinculação do cliente e retorno ao estoque'),
    ('rule_pl_act_cancelled', 'phone_line', 'active', 'cancelled', 'operational,master', 'Cancelamento definitivo da linha'),
    ('rule_pl_act_trans', 'phone_line', 'active', 'in_transition', 'operational,master', 'Início de processo de troca/migração de titularidade'),

    -- awaiting_invoice
    ('rule_pl_await_active', 'phone_line', 'awaiting_invoice', 'active', 'operational,master', 'Linha reapareceu em fatura recente'),
    ('rule_pl_await_stock', 'phone_line', 'awaiting_invoice', 'in_stock', 'operational,master', 'Desvinculação do cliente'),
    ('rule_pl_await_cancelled', 'phone_line', 'awaiting_invoice', 'cancelled', 'operational,master', 'Cancelamento definitivo confirmado'),

    -- suspended
    ('rule_pl_susp_active', 'phone_line', 'suspended', 'active', 'operational,master', 'Reativação da linha após suspensão'),
    ('rule_pl_susp_cancelled', 'phone_line', 'suspended', 'cancelled', 'operational,master', 'Cancelamento definitivo da linha suspensa'),
    ('rule_pl_susp_stock', 'phone_line', 'suspended', 'in_stock', 'operational,master', 'Desvinculação e devolução ao estoque'),

    -- inactive
    ('rule_pl_inact_stock', 'phone_line', 'inactive', 'in_stock', 'operational,master', 'Linha inativa reapareceu em fatura'),
    ('rule_pl_inact_cancelled', 'phone_line', 'inactive', 'cancelled', 'operational,master', 'Baixa permanente do número'),

    -- cancelled
    ('rule_pl_canc_stock', 'phone_line', 'cancelled', 'in_stock', 'master', 'Reativação excepcional de linha cancelada'),

    -- processing_month
    ('rule_pm_open_closed', 'processing_month', 'open', 'closed', 'operational,master', 'Fechamento regular ou contingencial do mês'),
    ('rule_pm_closed_open', 'processing_month', 'closed', 'open', 'master', 'Reabertura excepcional do mês para ajustes controlados'),

    -- provider_invoice
    ('rule_pi_draft_pending', 'provider_invoice', 'draft', 'pending', 'operational,master', 'Fatura importada e aguardando conciliação'),
    ('rule_pi_pend_paid', 'provider_invoice', 'pending', 'paid', 'financial,master', 'Liquidação da fatura da operadora'),
    ('rule_pi_pend_overdue', 'provider_invoice', 'pending', 'overdue', 'financial,master', 'Fatura vencida sem baixa'),
    ('rule_pi_pend_canc', 'provider_invoice', 'pending', 'cancelled', 'operational,master', 'Fatura descartada/cancelada'),
    ('rule_pi_over_paid', 'provider_invoice', 'overdue', 'paid', 'financial,master', 'Liquidação de fatura em atraso'),
    ('rule_pi_over_canc', 'provider_invoice', 'overdue', 'cancelled', 'operational,master', 'Cancelamento de fatura vencida'),

    -- customer_billing_document
    ('rule_cbd_draft_ready', 'customer_billing_document', 'draft', 'ready', 'financial,master', 'Fatura do cliente validada e pronta para envio'),
    ('rule_cbd_draft_canc', 'customer_billing_document', 'draft', 'cancelled', 'financial,master', 'Fatura rascunho cancelada'),
    ('rule_cbd_ready_sent', 'customer_billing_document', 'ready', 'sent', 'financial,master', 'Fatura e boleto despachados ao cliente'),
    ('rule_cbd_ready_canc', 'customer_billing_document', 'ready', 'cancelled', 'financial,master', 'Fatura pronta cancelada antes do envio'),
    ('rule_cbd_sent_canc', 'customer_billing_document', 'sent', 'cancelled', 'financial,master', 'Fatura enviada estornada/cancelada')
ON CONFLICT ("EntityType", "FromState", "ToState") DO NOTHING;
