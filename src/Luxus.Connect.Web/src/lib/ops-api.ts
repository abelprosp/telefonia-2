import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { client } from '@/lib/client';

export type StateTransitionLog = {
  id: string;
  entity_type: string;
  entity_id: string;
  from_state: string;
  to_state: string;
  trigger_event: string;
  justification?: string | null;
  actor_user_id?: string | null;
  created_at: string;
};

export function useStateTransitions(entityType: string, entityId: string, enabled = true) {
  return useQuery({
    queryKey: ['state-transitions', entityType, entityId],
    queryFn: async () => {
      const { data } = await client<StateTransitionLog[]>({
        url: '/v1/state-transitions',
        method: 'GET',
        params: { entity_type: entityType, entity_id: entityId }
      });
      return data ?? [];
    },
    enabled: Boolean(entityType && entityId) && enabled
  });
}

export type SupportTicket = {
  id: string;
  number: number;
  title: string;
  category: string;
  priority: string;
  status: string;
  sla_due_at?: string | null;
  customer_id?: string | null;
  phone_line_id?: string | null;
  charge_ref?: string | null;
  invoice_id?: string | null;
  created_at: string;
  updated_at: string;
  messages?: {
    id: string;
    body: string;
    author_name?: string | null;
    visibility: string;
    created_at: string;
    attachment_name?: string | null;
    attachment_key?: string | null;
    attachment_bucket?: string | null;
  }[];
  history?: { id: string; event_type: string; from_value?: string | null; to_value?: string | null; created_at: string }[];
};

export function useSupportTickets(status = '') {
  return useQuery({
    queryKey: ['tickets', status],
    queryFn: async () => {
      const { data } = await client<{ items: SupportTicket[]; total_count: number }>({
        url: '/v1/tickets',
        method: 'GET',
        params: status ? { status, page_size: 100 } : { page_size: 100 }
      });
      return data;
    }
  });
}

export function useSupportTicket(id: string) {
  return useQuery({
    queryKey: ['ticket', id],
    queryFn: async () => {
      const { data } = await client<SupportTicket>({ url: `/v1/tickets/${id}`, method: 'GET' });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateSupportTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { title: string; category?: string; priority?: string; message?: string; customer_id?: string; phone_line_id?: string; invoice_id?: string }) => {
      const { data } = await client<SupportTicket>({ url: '/v1/tickets', method: 'POST', data: body });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['tickets'] })
  });
}

export function useUpdateSupportTicket(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { status?: string; priority?: string; assignee_user_id?: string }) => {
      const { data } = await client<SupportTicket>({ url: `/v1/tickets/${id}`, method: 'PATCH', data: body });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['tickets'] });
      void qc.invalidateQueries({ queryKey: ['ticket', id] });
    }
  });
}

export function useAddTicketMessage(id: string, portal = false) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      body: string;
      visibility?: string;
      attachment_name?: string;
      attachment_key?: string;
      attachment_bucket?: string;
      attachment_content_type?: string;
      attachment_size_bytes?: number;
    }) => {
      const url = portal ? `/v1/portal/tickets/${id}/messages` : `/v1/tickets/${id}/messages`;
      const { data } = await client<SupportTicket>({ url, method: 'POST', data: body });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['ticket', id] });
      void qc.invalidateQueries({ queryKey: ['portal-tickets'] });
    }
  });
}

export type DivergenceItem = {
  id: string;
  processing_month_id: string;
  divergence_type: string;
  severity: string;
  phone_number?: string | null;
  description: string;
  status: string;
  financial_impact?: number;
  age_hours?: number;
  owner_user_id?: string | null;
  recommended_action?: string | null;
  created_at: string;
};

export function useDivergences(monthId = '') {
  return useQuery({
    queryKey: ['divergences', monthId],
    queryFn: async () => {
      const { data } = await client<DivergenceItem[]>({
        url: '/v1/divergences',
        method: 'GET',
        params: monthId ? { processing_month_id: monthId } : undefined
      });
      return data ?? [];
    }
  });
}

export function useResolveDivergence() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, action, notes }: { id: string; action: string; notes?: string }) => {
      const { data } = await client({ url: `/v1/divergences/${id}/resolve`, method: 'POST', data: { action, notes } });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['divergences'] })
  });
}

export function useAssignDivergence() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, owner_user_id }: { id: string; owner_user_id: string }) => {
      const { data } = await client({
        url: `/v1/divergences/${id}/assign`,
        method: 'POST',
        data: { owner_user_id }
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['divergences'] })
  });
}

export function useCommentDivergence() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, body }: { id: string; body: string }) => {
      const { data } = await client({
        url: `/v1/divergences/${id}/comments`,
        method: 'POST',
        data: { body }
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['divergences'] })
  });
}

export type ApprovalRequest = {
  id: string;
  action_type: string;
  entity_id: string;
  status: string;
  requester_user_id: string;
  created_at: string;
};

export function useApprovals(status = '') {
  return useQuery({
    queryKey: ['approvals', status],
    queryFn: async () => {
      const { data } = await client<ApprovalRequest[]>({
        url: '/v1/approvals',
        method: 'GET',
        params: status ? { status } : undefined
      });
      return data ?? [];
    }
  });
}

export function useApproveRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client({ url: `/v1/approvals/${id}/approve`, method: 'POST' });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['approvals'] })
  });
}

export function useRejectRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason?: string }) => {
      const { data } = await client({
        url: `/v1/approvals/${id}/reject`,
        method: 'POST',
        data: { reason: reason ?? '' }
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['approvals'] })
  });
}

export type PipelineRun = {
  id: string;
  version: number;
  status: string;
  created_at: string;
  steps: {
    key: string;
    label: string;
    status: string;
    duration_ms?: number | null;
    error?: string | null;
    summary_json?: string | null;
  }[];
};

export function useProcessingMonthRuns(monthId: string) {
  return useQuery({
    queryKey: ['pipeline-runs', monthId],
    queryFn: async () => {
      const { data } = await client<PipelineRun[]>({
        url: `/v1/processing-months/${monthId}/pipeline`,
        method: 'GET'
      });
      return data ?? [];
    },
    enabled: Boolean(monthId)
  });
}

export function useRunProcessingMonthPipeline(monthId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<PipelineRun>({
        url: `/v1/processing-months/${monthId}/pipeline`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['pipeline-runs', monthId] })
  });
}

export type PortalMe = {
  customer: { id: string; name: string; cpf_cnpj: string; billing_email?: string | null };
  active_lines_count: number;
  total_monthly_amount: number;
};

export function usePortalMe() {
  return useQuery({
    queryKey: ['portal-me'],
    queryFn: async () => {
      const { data } = await client<PortalMe>({ url: '/v1/portal/me', method: 'GET' });
      return data;
    }
  });
}

export function usePortalLines() {
  return useQuery({
    queryKey: ['portal-lines'],
    queryFn: async () => {
      const { data } = await client<{ id: string; number: string; plan_name: string; monthly_amount: number; status: string }[]>({
        url: '/v1/portal/lines',
        method: 'GET'
      });
      return data ?? [];
    }
  });
}

export function usePortalInvoices() {
  return useQuery({
    queryKey: ['portal-invoices'],
    queryFn: async () => {
      const { data } = await client<{ id: string; invoice_number: string; due_date: string; total_amount: number; status: string }[]>({
        url: '/v1/portal/invoices',
        method: 'GET'
      });
      return data ?? [];
    }
  });
}

export function usePortalContracts() {
  return useQuery({
    queryKey: ['portal-contracts'],
    queryFn: async () => {
      const { data } = await client<{ id: string; trigger?: string | null; status: string; created_at: string }[]>({
        url: '/v1/portal/contracts',
        method: 'GET'
      });
      return data ?? [];
    }
  });
}

export function usePortalTickets() {
  return useQuery({
    queryKey: ['portal-tickets'],
    queryFn: async () => {
      const { data } = await client<{ items: SupportTicket[]; total_count: number }>({
        url: '/v1/portal/tickets',
        method: 'GET'
      });
      return data;
    }
  });
}

export function usePortalTicket(id: string) {
  return useQuery({
    queryKey: ['ticket', id, 'portal'],
    queryFn: async () => {
      const { data } = await client<SupportTicket>({ url: `/v1/portal/tickets/${id}`, method: 'GET' });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useUpdatePortalMe() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { billing_email: string }) => {
      const { data } = await client<PortalMe>({ url: '/v1/portal/me', method: 'PATCH', data: body });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['portal-me'] })
  });
}

export function useCreatePortalTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { title: string; category?: string; priority?: string; message?: string }) => {
      const { data } = await client<SupportTicket>({ url: '/v1/portal/tickets', method: 'POST', data: body });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['portal-tickets'] })
  });
}

export async function downloadPortalInvoice(documentId: string, filename?: string) {
  const { data } = await client<Blob>({
    url: `/v1/portal/invoices/${documentId}/download`,
    method: 'GET',
    responseType: 'blob'
  });
  const url = URL.createObjectURL(data);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename ?? `fatura-${documentId}.html`;
  link.click();
  URL.revokeObjectURL(url);
}

export type OperationalDashboard = {
  lines_summary: {
    total_lines: number;
    active_lines: number;
    in_transition_lines: number;
    canceled_lines: number;
    orphan_lines: number;
  };
  customers_summary: {
    total_customers: number;
    active_customers: number;
  };
  financial_summary: {
    projected_monthly_revenue: number;
    total_base_cost: number;
    projected_margin: number;
    margin_percentage: number;
  };
  current_month_status?: {
    processing_month_id: string;
    display_name: string;
    status: string;
    critical_alerts: number;
    warning_alerts: number;
  } | null;
  pending_divergences: number;
};

export function useOperationalDashboard() {
  return useQuery({
    queryKey: ['operational-dashboard'],
    queryFn: async () => {
      const { data } = await client<OperationalDashboard>({
        url: '/v1/stats/operational-dashboard',
        method: 'GET'
      });
      return data;
    }
  });
}

export type PhoneLine360 = {
  line: { id: string; number?: string | null; status?: string | null };
  active_customer_link?: {
    customer_id: string;
    customer_name?: string | null;
    monthly_amount?: number | null;
  } | null;
  active_fidelity?: {
    initial_months?: number | null;
    predicted_end_date?: string | null;
    status?: string | null;
  } | null;
  penalty_estimate?: { penalty_amount?: number | null } | null;
  billing_explanation?: BillingExplanation | null;
  recent_timeline?: { event_type?: string; created_at?: string; description?: string }[];
};

export type BillingExplanation = {
  phone_line_id: string;
  phone_number: string;
  customer_id: string;
  customer_name: string;
  processing_month_id: string;
  total_amount: number;
  formula_text: string;
  components: { type: string; description: string; amount: number; details?: string }[];
};

export function usePhoneLine360(phoneLineId: string, enabled = true) {
  return useQuery({
    queryKey: ['phone-line-360', phoneLineId],
    queryFn: async () => {
      const { data } = await client<PhoneLine360>({
        url: `/v1/phone-lines/${phoneLineId}/full-360`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(phoneLineId) && enabled
  });
}

export function useBillingExplanation(phoneLineId: string, enabled = true) {
  return useQuery({
    queryKey: ['billing-explanation', phoneLineId],
    queryFn: async () => {
      const { data } = await client<BillingExplanation>({
        url: `/v1/phone-lines/${phoneLineId}/billing-explanation`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(phoneLineId) && enabled
  });
}

export type Customer360 = {
  customer: { id: string; name: string; cpf_cnpj?: string };
  total_lines_count: number;
  active_lines_count: number;
  total_monthly_amount: number;
  phone_lines: { id: string; number?: string | null; status?: string | null }[];
  generated_contracts: { id: string; status?: string; created_at?: string }[];
  attachments_count: number;
};

export function useCustomer360(customerId: string, enabled = true) {
  return useQuery({
    queryKey: ['customer-360', customerId],
    queryFn: async () => {
      const { data } = await client<Customer360>({
        url: `/v1/customers/${customerId}/full-360`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(customerId) && enabled
  });
}

export type BillingImpactSimulation = {
  processing_month_id: string;
  display_name: string;
  projected_revenue: number;
  projected_cost: number;
  projected_margin: number;
  margin_percentage: number;
  previous_revenue: number;
  revenue_delta: number;
  revenue_delta_percentage: number;
  total_active_lines: number;
};

export function useSimulateBillingImpact(monthId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<BillingImpactSimulation>({
        url: `/v1/processing-months/${monthId}/simulate-impact`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['pipeline-runs', monthId] })
  });
}

export type LineReadiness = {
  processing_month_id: string;
  total_lines: number;
  ready_lines: number;
  blocked_lines: number;
  items: {
    phone_line_id: string;
    phone_number: string;
    customer_name?: string | null;
    monthly_amount: number;
    is_ready: boolean;
    blocking_rules: string[];
  }[];
};

export function useProcessingMonthLineReadiness(monthId: string, enabled = true) {
  return useQuery({
    queryKey: ['line-readiness', monthId],
    queryFn: async () => {
      const { data } = await client<LineReadiness>({
        url: `/v1/processing-months/${monthId}/line-readiness`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(monthId) && enabled
  });
}

export type CloseMonthWithHash = {
  processing_month_id: string;
  status: string;
  closed_at?: string;
  closed_by?: string;
  consolidation_hash: string;
  total_revenue?: number;
};

export function useCloseProcessingMonthWithHash(monthId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<CloseMonthWithHash>({
        url: `/v1/processing-months/${monthId}/close-with-hash`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['processing-months'] });
      void qc.invalidateQueries({ queryKey: processingMonthKeys(monthId) });
    }
  });
}

export function useReopenProcessingMonth(monthId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client({
        url: `/v1/processing-months/${monthId}/reopen`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['processing-months'] });
      void qc.invalidateQueries({ queryKey: processingMonthKeys(monthId) });
    }
  });
}

function processingMonthKeys(monthId: string) {
  return ['processing-month', monthId] as const;
}

export type ImportPreview = {
  summary: {
    invoice_number: string;
    account_number: string;
    issue_date: string;
    due_date: string;
    total_amount: number;
    lines_count: number;
    known_lines: number;
    unknown_lines: number;
  };
  line_numbers?: string[];
  warnings?: string[];
  is_valid: boolean;
  file_sha256?: string;
};

export async function previewProviderInvoiceImport(body: {
  provider_id: string;
  processing_month_id: string;
  storage_bucket: string;
  storage_object_key: string;
  original_file_name?: string | null;
  allow_substitute?: boolean;
}) {
  const { data } = await client<ImportPreview>({
    url: '/v1/provider-invoices/preview',
    method: 'POST',
    data: body
  });
  return data;
}

export type FinancialSummaryReport = {
  generated_at: string;
  total_lines: number;
  total_gross: number;
  total_cost: number;
  total_margin: number;
  items: {
    customer_id: string;
    customer_name: string;
    contracted_luxus_cnpj?: string | null;
    lines_count: number;
    total_gross_revenue: number;
    total_operator_cost: number;
    gross_margin: number;
    margin_percentage: number;
  }[];
};

export function useFinancialSummaryReport(enabled = true) {
  return useQuery({
    queryKey: ['reports', 'financial-summary'],
    queryFn: async () => {
      const { data } = await client<FinancialSummaryReport>({
        url: '/v1/reports/financial-summary',
        method: 'GET'
      });
      return data;
    },
    enabled
  });
}

export type CustomerProfitabilityReport = {
  generated_at: string;
  items: {
    customer_id: string;
    customer_name: string;
    responsible_sales?: string | null;
    active_lines: number;
    average_ticket: number;
    gross_revenue: number;
    cost: number;
    margin: number;
    margin_percentage: number;
  }[];
};

export function useCustomerProfitabilityReport(enabled = true) {
  return useQuery({
    queryKey: ['reports', 'customer-profitability'],
    queryFn: async () => {
      const { data } = await client<CustomerProfitabilityReport>({
        url: '/v1/reports/customer-profitability',
        method: 'GET'
      });
      return data;
    },
    enabled
  });
}

export const WEBHOOK_EVENT_OPTIONS = [
  { value: 'LINE_TRANSITION_ALERT', label: 'Transição de linha' },
  { value: 'FIDELITY_EXPIRING_ALERT', label: 'Fidelidade a vencer' },
  { value: 'DIVERGENCE_DETECTED', label: 'Divergência detectada' },
  { value: 'BILLING_CLOSED', label: 'Competência fechada' }
] as const;

export type WebhookSubscription = {
  id: string;
  url: string;
  events: string[];
  is_active: boolean;
  secret: string;
  created_at: string;
};

export function useWebhooks(enabled = true) {
  return useQuery({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const { data } = await client<WebhookSubscription[]>({
        url: '/v1/webhooks',
        method: 'GET'
      });
      return data ?? [];
    },
    enabled
  });
}

export function useCreateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { url: string; events: string[] }) => {
      const { data } = await client<WebhookSubscription>({
        url: '/v1/webhooks',
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['webhooks'] })
  });
}

export function useDeleteWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await client({ url: `/v1/webhooks/${id}`, method: 'DELETE' });
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['webhooks'] })
  });
}

export function useTestWebhook() {
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{ success: boolean; status_code: number; response_body: string }>({
        url: `/v1/webhooks/${id}/test`,
        method: 'POST'
      });
      return data;
    }
  });
}

export type OrganizationDataExport = {
  exported_at: string;
  exported_by: string;
  checksum_sha256?: string;
  payload_bytes?: number;
};

export function useExportOrganizationData() {
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<OrganizationDataExport>({
        url: '/v1/organization/data-export',
        method: 'POST'
      });
      return data;
    }
  });
}

export function useCustomerPersonalData(customerId: string, enabled = false) {
  return useQuery({
    queryKey: ['customer-personal-data', customerId],
    queryFn: async () => {
      const { data } = await client<Record<string, unknown>>({
        url: `/v1/customers/${customerId}/personal-data`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(customerId) && enabled
  });
}

export function useAnonymizeCustomer(customerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client({
        url: `/v1/customers/${customerId}/anonymize`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['customer-360', customerId] });
      void qc.invalidateQueries({ queryKey: ['customers'] });
    }
  });
}
