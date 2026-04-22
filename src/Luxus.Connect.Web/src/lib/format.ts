/** Formata CPF ou CNPJ para exibição. */
export const formatCpfCnpj = (doc: string) => {
  const cleaned = doc.replace(/\D/g, '');

  if (cleaned.length === 11) {
    return cleaned.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, '$1.$2.$3-$4');
  }

  if (cleaned.length === 14) {
    return cleaned.replace(
      /(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/,
      '$1.$2.$3/$4-$5'
    );
  }

  return doc;
};

/**
 * Formata número de telefone BR: celular `(11) 98765-4321` (11 dígitos: DDD + 9);
 * fixo `(11) 3456-7890` (10 dígitos). Outros formatos são devolvidos sem alteração.
 */
export function formatPhoneNumber(value: string): string | undefined {
  const cleaned = value.replace(/\D/g, '');

  if (!cleaned) {
    return;
  }

  if (cleaned.length === 11) {
    return cleaned.replace(/(\d{2})(\d{5})(\d{4})/, '($1) $2-$3');
  }

  if (cleaned.length === 10) {
    return cleaned.replace(/(\d{2})(\d{4})(\d{4})/, '($1) $2-$3');
  }

  return value.trim();
}

/**
 * Rótulos em português para enums do domínio (API: strings de GetDescription(),
 * nomes de enum em snake_case via JSON, ou valores ordinais).
 */

const EMPTY = '—';

function normalizeEnumKey(raw: string): string {
  const t = raw.trim();
  if (!t) {
    return t;
  }
  const withUnderscores = t.replace(/([a-z\d])([A-Z])/g, '$1_$2');
  return withUnderscores.replace(/[\s-]+/g, '_').toLowerCase();
}

function formatEnum(
  wireToLabel: Record<string, string>,
  ordinalToLabel: Record<number, string> | undefined,
  value: string | number | null | undefined
): string {
  if (value === null || value === undefined) {
    return EMPTY;
  }
  if (typeof value === 'number') {
    if (ordinalToLabel && Object.hasOwn(ordinalToLabel, value)) {
      return ordinalToLabel[value]!;
    }
    return String(value);
  }
  const s = String(value).trim();
  if (s === '') {
    return EMPTY;
  }
  const k = normalizeEnumKey(s);
  if (wireToLabel[k]) {
    return wireToLabel[k]!;
  }
  const lower = s.toLowerCase();
  if (wireToLabel[lower]) {
    return wireToLabel[lower]!;
  }
  return s;
}

// --- InvoiceStatus (domínio + wire) ---

const invoiceStatusWire: Record<string, string> = {
  draft: 'Rascunho',
  pending: 'Pendente',
  paid: 'Paga',
  overdue: 'Vencida',
  cancelled: 'Cancelada'
};

const invoiceStatusOrdinal: Record<number, string> = {
  0: 'Rascunho',
  1: 'Pendente',
  2: 'Paga',
  3: 'Vencida',
  4: 'Cancelada'
};

export function formatInvoiceStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(invoiceStatusWire, invoiceStatusOrdinal, status);
}

// --- PhoneLineStatus ---

const phoneLineStatusWire: Record<string, string> = {
  inactive: 'Inativa',
  active: 'Ativa',
  cancelled: 'Cancelada',
  suspended: 'Suspensa',
  in_stock: 'Em estoque',
  awaiting_invoice: 'Aguardando fatura',
  in_transition: 'Em transição'
};

const phoneLineStatusOrdinal: Record<number, string> = {
  0: 'Inativa',
  1: 'Ativa',
  2: 'Cancelada',
  3: 'Suspensa',
  4: 'Em estoque',
  5: 'Aguardando fatura',
  6: 'Em transição'
};

export function formatPhoneLineStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(phoneLineStatusWire, phoneLineStatusOrdinal, status);
}

// --- TransitionSubStatus ---

const transitionSubStatusWire: Record<string, string> = {
  none: 'Nenhuma',
  pending_activation: 'Ativação pendente',
  pending_cancellation: 'Cancelamento pendente',
  pending_portability: 'Portabilidade pendente',
  pending_tt: 'Transf. de titularidade (TT) pendente',
  pending_pp: 'Portabilidade parcial (PP) pendente'
};

const transitionSubStatusOrdinal: Record<number, string> = {
  0: 'Nenhuma',
  1: 'Ativação pendente',
  2: 'Cancelamento pendente',
  3: 'Portabilidade pendente',
  4: 'Transf. de titularidade (TT) pendente',
  5: 'Portabilidade parcial (PP) pendente'
};

export function formatTransitionSubStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(
    transitionSubStatusWire,
    transitionSubStatusOrdinal,
    status
  );
}

// --- LineClassification ---

const lineClassificationWire: Record<string, string> = {
  normal: 'Normal',
  titular: 'Titular',
  dependent: 'Dependente',
  other: 'Outro'
};

const lineClassificationOrdinal: Record<number, string> = {
  0: 'Normal',
  1: 'Titular',
  2: 'Dependente',
  3: 'Outro'
};

export function formatLineClassification(
  value: string | number | null | undefined
): string {
  return formatEnum(lineClassificationWire, lineClassificationOrdinal, value);
}

// --- BillingCycleStatus ---

const billingCycleStatusWire: Record<string, string> = {
  open: 'Aberto',
  closed: 'Fechado'
};

const billingCycleStatusOrdinal: Record<number, string> = {
  0: 'Aberto',
  1: 'Fechado'
};

export function formatBillingCycleStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(billingCycleStatusWire, billingCycleStatusOrdinal, status);
}

// --- ProcessingMonthStatus ---

const processingMonthStatusWire: Record<string, string> = {
  open: 'Aberto',
  closed: 'Fechado'
};

const processingMonthStatusOrdinal: Record<number, string> = {
  0: 'Aberto',
  1: 'Fechado'
};

export function formatProcessingMonthStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(
    processingMonthStatusWire,
    processingMonthStatusOrdinal,
    status
  );
}

// --- InvoiceImportRequestStatus ---

const invoiceImportRequestStatusWire: Record<string, string> = {
  pending: 'Pendente',
  processing: 'Processando',
  completed: 'Concluída',
  failed: 'Falhou'
};

const invoiceImportRequestStatusOrdinal: Record<number, string> = {
  0: 'Pendente',
  1: 'Processando',
  2: 'Concluída',
  3: 'Falhou'
};

export function formatInvoiceImportRequestStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(
    invoiceImportRequestStatusWire,
    invoiceImportRequestStatusOrdinal,
    status
  );
}

// --- InvoiceItemType ---

const invoiceItemTypeWire: Record<string, string> = {
  usage: 'Uso',
  plan: 'Plano',
  discount: 'Desconto',
  other: 'Outro',
  extra_header: 'Cabeçalho extra',
  extra_location: 'Localização extra',
  extra_detail: 'Detalhe extra',
  service: 'Serviço'
};

const invoiceItemTypeOrdinal: Record<number, string> = {
  0: 'Uso',
  1: 'Plano',
  2: 'Desconto',
  3: 'Outro',
  4: 'Cabeçalho extra',
  5: 'Localização extra',
  6: 'Detalhe extra',
  7: 'Serviço'
};

export function formatInvoiceItemType(
  type: string | number | null | undefined
): string {
  return formatEnum(invoiceItemTypeWire, invoiceItemTypeOrdinal, type);
}

// --- InvoiceItemUnit ---

const invoiceItemUnitWire: Record<string, string> = {
  min: 'Minutos',
  sms: 'SMS',
  kb: 'KB',
  mb: 'MB',
  gb: 'GB',
  tb: 'TB'
};

const invoiceItemUnitOrdinal: Record<number, string> = {
  0: 'Minutos',
  1: 'SMS',
  2: 'KB',
  3: 'MB',
  4: 'GB',
  5: 'TB'
};

export function formatInvoiceItemUnit(
  unit: string | number | null | undefined
): string {
  return formatEnum(invoiceItemUnitWire, invoiceItemUnitOrdinal, unit);
}

// --- CustomerBillingReadinessStatus ---

const customerBillingReadinessWire: Record<string, string> = {
  pending: 'Pendente',
  released_for_billing: 'Liberado para cobrança',
  manually_released: 'Liberado manualmente'
};

const customerBillingReadinessOrdinal: Record<number, string> = {
  0: 'Pendente',
  1: 'Liberado para cobrança',
  2: 'Liberado manualmente'
};

export function formatCustomerBillingReadinessStatus(
  status: string | number | null | undefined
): string {
  return formatEnum(
    customerBillingReadinessWire,
    customerBillingReadinessOrdinal,
    status
  );
}

// --- ProcessingMonthEntradaKind ---

const processingMonthEntradaKindWire: Record<string, string> = {
  first_appearance: 'Primeira aparição',
  returned_after_absence: 'Retorno após ausência'
};

const processingMonthEntradaKindOrdinal: Record<number, string> = {
  0: 'Primeira aparição',
  1: 'Retorno após ausência'
};

export function formatProcessingMonthEntradaKind(
  kind: string | number | null | undefined
): string {
  return formatEnum(
    processingMonthEntradaKindWire,
    processingMonthEntradaKindOrdinal,
    kind
  );
}

// --- CustomerType ---

const customerTypeWire: Record<string, string> = {
  pf: 'Pessoa Física',
  pj: 'Pessoa Jurídica'
};

const customerTypeOrdinal: Record<number, string> = {
  0: 'Pessoa Física',
  1: 'Pessoa Jurídica'
};

export function formatCustomerType(
  type: string | number | null | undefined
): string {
  return formatEnum(customerTypeWire, customerTypeOrdinal, type);
}

// --- CustomerDocumentType ---

const customerDocumentTypeWire: Record<string, string> = {
  other: 'Outro',
  rg: 'RG',
  cnh: 'CNH',
  cpf: 'CPF',
  cnpj: 'CNPJ',
  municipal_registration: 'Inscrição municipal',
  state_registration: 'Inscrição estadual'
};

const customerDocumentTypeOrdinal: Record<number, string> = {
  0: 'Outro',
  1: 'RG',
  2: 'CNH',
  3: 'CPF',
  4: 'CNPJ',
  5: 'Inscrição municipal',
  6: 'Inscrição estadual'
};

export function formatCustomerDocumentType(
  type: string | number | null | undefined
): string {
  return formatEnum(
    customerDocumentTypeWire,
    customerDocumentTypeOrdinal,
    type
  );
}

// --- ServiceType ---

const serviceTypeWire: Record<string, string> = {
  data: 'Dados',
  other: 'Outro',
  roaming: 'Roaming',
  sms: 'SMS',
  subscription: 'Assinatura'
};

const serviceTypeOrdinal: Record<number, string> = {
  0: 'Dados',
  1: 'Outro',
  2: 'Roaming',
  3: 'SMS',
  4: 'Assinatura'
};

export function formatServiceType(
  type: string | number | null | undefined
): string {
  return formatEnum(serviceTypeWire, serviceTypeOrdinal, type);
}

// --- ServiceApplicationType ---

const serviceApplicationTypeWire: Record<string, string> = {
  plan: 'Plano',
  addon: 'Adicional',
  service: 'Serviço'
};

const serviceApplicationTypeOrdinal: Record<number, string> = {
  0: 'Plano',
  1: 'Adicional',
  2: 'Serviço'
};

export function formatServiceApplicationType(
  type: string | number | null | undefined
): string {
  return formatEnum(
    serviceApplicationTypeWire,
    serviceApplicationTypeOrdinal,
    type
  );
}

// --- ServiceAvailabilityRule ---

const serviceAvailabilityRuleWire: Record<string, string> = {
  always: 'Sempre',
  cycle_only: 'Somente no ciclo',
  custom: 'Personalizado'
};

const serviceAvailabilityRuleOrdinal: Record<number, string> = {
  0: 'Sempre',
  1: 'Somente no ciclo',
  2: 'Personalizado'
};

export function formatServiceAvailabilityRule(
  rule: string | number | null | undefined
): string {
  return formatEnum(
    serviceAvailabilityRuleWire,
    serviceAvailabilityRuleOrdinal,
    rule
  );
}

// --- ExceedanceChargeType ---

const exceedanceChargeTypeWire: Record<string, string> = {
  espelhado: 'Espelhado'
};

const exceedanceChargeTypeOrdinal: Record<number, string> = {
  0: 'Espelhado'
};

export function formatExceedanceChargeType(
  type: string | number | null | undefined
): string {
  return formatEnum(
    exceedanceChargeTypeWire,
    exceedanceChargeTypeOrdinal,
    type
  );
}

/** Identificadores dos enums de domínio com formatação centralizada. */
export type DomainEnumKind =
  | 'invoice_status'
  | 'phone_line_status'
  | 'transition_sub_status'
  | 'line_classification'
  | 'billing_cycle_status'
  | 'processing_month_status'
  | 'invoice_import_request_status'
  | 'invoice_item_type'
  | 'invoice_item_unit'
  | 'customer_billing_readiness_status'
  | 'processing_month_entrada_kind'
  | 'customer_type'
  | 'customer_document_type'
  | 'service_type'
  | 'service_application_type'
  | 'service_availability_rule'
  | 'exceedance_charge_type';
