import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { parseTotalCount } from '@/lib/query-utils';

export type InvoiceEmailTemplate = {
  id: string;
  name: string;
  code: string;
  kind: string;
  subject_template: string;
  body_template_html?: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type CustomerBillingDocument = {
  id: string;
  customer_id: string;
  customer_name: string;
  accounts_receivable_id?: string | null;
  processing_month_id?: string | null;
  invoice_number: string;
  issue_date: string;
  due_date: string;
  amount: number;
  status: string;
  recipient_email: string;
  email_subject: string;
  email_body_html?: string;
  send_count: number;
  sent_at?: string | null;
  last_sent_at?: string | null;
  created_at: string;
  updated_at?: string;
  sicredi_nosso_numero?: string | null;
  sicredi_linha_digitavel?: string | null;
  sicredi_codigo_barras?: string | null;
  sicredi_pix_qr_code?: string | null;
  sicredi_pix_tx_id?: string | null;
  sicredi_boleto_status?: string | null;
  sicredi_boleto_error?: string | null;
  sicredi_paid_at?: string | null;
};

export type BillingSendLog = {
  id: string;
  recipient_email: string;
  subject: string;
  success: boolean;
  error_message?: string | null;
  sent_by_user_id: string;
  sent_at: string;
};

export type OverdueReceivable = {
  id: string;
  customer_id: string;
  customer_name: string;
  billing_email: string;
  description: string;
  due_date: string;
  balance: number;
  reminders_sent: number;
};

type Paged<T> = { items: T[]; total_count: number | string };
type PageParams = { page_index?: number; page_size?: number };

async function pagedGet<T>(url: string, params?: PageParams & Record<string, unknown>) {
  const { data } = await client<Paged<T>>({ url, method: 'GET', params });
  return { items: data.items ?? [], totalCount: parseTotalCount(data.total_count) };
}

export const billingKeys = {
  templates: (params?: PageParams & { kind?: string }) => ['billing', 'templates', params] as const,
  template: (id: string) => ['billing', 'template', id] as const,
  documents: (params?: PageParams & { status?: string; customer_id?: string }) =>
    ['billing', 'documents', params] as const,
  document: (id: string) => ['billing', 'document', id] as const,
  sendLog: (id: string) => ['billing', 'send-log', id] as const,
  overdue: (params?: PageParams) => ['billing', 'overdue', params] as const,
  bulkPreview: (processingMonthId: string) => ['billing', 'bulk-preview', processingMonthId] as const,
  manualPreview: (customerIds?: string[]) => ['billing', 'manual-preview', customerIds ?? []] as const
};

export type BulkBillingPreviewItem = {
  customer_id: string;
  customer_name: string;
  customer_document: string;
  billing_email: string;
  line_count: number;
  device_count?: number;
  monthly_amount: number;
  provider_cost?: number;
  already_billed: boolean;
  eligible: boolean;
  skip_reason?: string;
};

export type BulkBillingPreview = {
  processing_month_id?: string;
  processing_month_name?: string;
  provider_invoices_count?: number;
  items: BulkBillingPreviewItem[];
  eligible_count: number;
};

export type BulkGenerateResultItem = {
  customer_id: string;
  customer_name: string;
  status: 'created' | 'skipped' | 'failed' | string;
  message?: string;
  document_id?: string | null;
  receivable_id?: string | null;
  amount?: number;
};

export type BulkGenerateResult = {
  created: number;
  skipped: number;
  failed: number;
  items: BulkGenerateResultItem[];
};

export function useInvoiceEmailTemplates(params?: PageParams & { kind?: string }) {
  return useQuery({
    queryKey: billingKeys.templates(params),
    queryFn: () =>
      pagedGet<InvoiceEmailTemplate>('/v1/invoice-email-templates', {
        page_index: params?.page_index ?? 0,
        page_size: params?.page_size ?? 50,
        kind: params?.kind
      })
  });
}

export function useInvoiceEmailTemplate(id: string) {
  return useQuery({
    queryKey: billingKeys.template(id),
    queryFn: async () => {
      const { data } = await client<InvoiceEmailTemplate>({
        url: `/v1/invoice-email-templates/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateInvoiceEmailTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      name: string;
      code: string;
      kind: string;
      subject_template: string;
      body_template_html: string;
      active?: boolean;
    }) => {
      const { data } = await client<InvoiceEmailTemplate>({
        url: '/v1/invoice-email-templates',
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['billing', 'templates'] })
  });
}

export function useUpdateInvoiceEmailTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: string;
      name: string;
      subject_template: string;
      body_template_html: string;
      active?: boolean;
    }) => {
      const { data } = await client<InvoiceEmailTemplate>({
        url: `/v1/invoice-email-templates/${id}`,
        method: 'PATCH',
        data: body
      });
      return data;
    },
    onSuccess: (_, v) => {
      void qc.invalidateQueries({ queryKey: billingKeys.template(v.id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'templates'] });
    }
  });
}

export function useCustomerBillingDocuments(
  params?: PageParams & { status?: string; customer_id?: string }
) {
  return useQuery({
    queryKey: billingKeys.documents(params),
    queryFn: () =>
      pagedGet<CustomerBillingDocument>('/v1/customer-billing-documents', {
        page_index: params?.page_index ?? 0,
        page_size: params?.page_size ?? 10,
        status: params?.status,
        customer_id: params?.customer_id
      })
  });
}

export function useCustomerBillingDocument(id: string) {
  return useQuery({
    queryKey: billingKeys.document(id),
    queryFn: async () => {
      const { data } = await client<CustomerBillingDocument>({
        url: `/v1/customer-billing-documents/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateBillingDocumentFromReceivable() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      receivableId,
      template_code,
      layout_template_code
    }: {
      receivableId: string;
      template_code?: string;
      layout_template_code?: string;
    }) => {
      const { data } = await client<{ id: string }>({
        url: `/v1/customer-billing-documents/from-receivable/${receivableId}`,
        method: 'POST',
        data: {
          template_code: template_code ?? 'default-billing-invoice',
          layout_template_code: layout_template_code ?? 'default-invoice-layout'
        }
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['billing', 'documents'] })
  });
}

export function useBulkBillingPreview(processingMonthId: string) {
  return useQuery({
    queryKey: billingKeys.bulkPreview(processingMonthId),
    queryFn: async () => {
      const { data } = await client<BulkBillingPreview>({
        url: '/v1/customer-billing-documents/bulk-preview',
        method: 'GET',
        params: { processing_month_id: processingMonthId }
      });
      return data;
    },
    enabled: Boolean(processingMonthId)
  });
}

export function useManualBillingPreview(customerIds?: string[], enabled = true) {
  const idsKey = customerIds?.join(',') ?? '';
  return useQuery({
    queryKey: billingKeys.manualPreview(customerIds),
    queryFn: async () => {
      const { data } = await client<BulkBillingPreview>({
        url: '/v1/customer-billing-documents/manual-preview',
        method: 'GET',
        params: customerIds?.length ? { customer_ids: idsKey } : undefined
      });
      return data;
    },
    enabled
  });
}

export function useManualGenerateBillingDocuments() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      issue_date: string;
      due_date: string;
      description?: string;
      customer_ids?: string[];
    }) => {
      const { data } = await client<BulkGenerateResult>({
        url: '/v1/customer-billing-documents/manual-generate',
        method: 'POST',
        data: {
          template_code: 'default-billing-invoice',
          layout_template_code: 'default-invoice-layout',
          ...body
        }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: ['billing', 'manual-preview'] });
      void qc.invalidateQueries({ queryKey: ['financial', 'receivables'] });
    }
  });
}

export function useGenerateCustomerBillingDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      customerId,
      ...body
    }: {
      customerId: string;
      issue_date: string;
      due_date: string;
      description?: string;
      amount?: number;
    }) => {
      const { data } = await client<{
        id: string;
        receivable_id: string;
        amount: number;
        message: string;
      }>({
        url: `/v1/customers/${customerId}/generate-billing-document`,
        method: 'POST',
        data: {
          template_code: 'default-billing-invoice',
          layout_template_code: 'default-invoice-layout',
          ...body
        }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: ['billing', 'manual-preview'] });
      void qc.invalidateQueries({ queryKey: ['financial', 'receivables'] });
    }
  });
}

export function useBulkGenerateBillingDocuments() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      processing_month_id: string;
      issue_date: string;
      due_date: string;
      description?: string;
      template_code?: string;
      layout_template_code?: string;
      customer_ids?: string[];
    }) => {
      const { data } = await client<BulkGenerateResult>({
        url: '/v1/customer-billing-documents/bulk-generate',
        method: 'POST',
        data: {
          template_code: 'default-billing-invoice',
          layout_template_code: 'default-invoice-layout',
          ...body
        }
      });
      return data;
    },
    onSuccess: (_, vars) => {
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: billingKeys.bulkPreview(vars.processing_month_id) });
      void qc.invalidateQueries({ queryKey: ['financial', 'receivables'] });
    }
  });
}

export function useUpdateCustomerBillingDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: string;
      recipient_email: string;
      email_subject: string;
      email_body_html: string;
      status: string;
    }) => {
      const { data } = await client<CustomerBillingDocument>({
        url: `/v1/customer-billing-documents/${id}`,
        method: 'PATCH',
        data: body
      });
      return data;
    },
    onSuccess: (_, v) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(v.id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
    }
  });
}

export function useSendCustomerBillingDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{ success: boolean; message: string }>({
        url: `/v1/customer-billing-documents/${id}/send`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_, id) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: billingKeys.sendLog(id) });
    }
  });
}

export function useIssueSicrediBoleto() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{
        success: boolean;
        message: string;
        sicredi_nosso_numero?: string;
        sicredi_linha_digitavel?: string;
      }>({
        url: `/v1/customer-billing-documents/${id}/issue-boleto`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_, id) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
    }
  });
}

export function useSyncSicrediPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{
        checked: number;
        paid: number;
        items: Array<{
          document_id: string;
          status: string;
          message?: string;
          paid_at?: string;
        }>;
      }>({
        url: `/v1/customer-billing-documents/${id}/sync-payment`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_, id) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: ['financial', 'receivables'] });
    }
  });
}

export function useSyncSicrediPayments() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (daysBack = 7) => {
      const { data } = await client<{ checked: number; paid: number }>({
        url: '/v1/collections/sync-sicredi-payments',
        method: 'POST',
        params: { days_back: daysBack }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
      void qc.invalidateQueries({ queryKey: ['financial', 'receivables'] });
    }
  });
}

export async function downloadSicrediBoletoPDF(documentId: string, filename?: string) {
  const { data } = await client<Blob>({
    url: `/v1/customer-billing-documents/${documentId}/boleto-pdf`,
    method: 'GET',
    responseType: 'blob'
  });
  const url = URL.createObjectURL(data);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename ?? `boleto-${documentId}.pdf`;
  link.click();
  URL.revokeObjectURL(url);
}

export async function downloadCustomerBillingInvoice(documentId: string, filename?: string) {
  const { data } = await client<Blob>({
    url: `/v1/customer-billing-documents/${documentId}/download`,
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

export function useCancelSicrediBoleto() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{ success: boolean; message: string }>({
        url: `/v1/customer-billing-documents/${id}/cancel-boleto`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_, id) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
    }
  });
}

export function useAlterSicrediBoletoDueDate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, due_date }: { id: string; due_date: string }) => {
      const { data } = await client<{ success: boolean; message: string }>({
        url: `/v1/customer-billing-documents/${id}/boleto-due-date`,
        method: 'PATCH',
        data: { due_date }
      });
      return data;
    },
    onSuccess: (_, { id }) => {
      void qc.invalidateQueries({ queryKey: billingKeys.document(id) });
      void qc.invalidateQueries({ queryKey: ['billing', 'documents'] });
    }
  });
}

export type SicrediIntegrationStatus = {
  enabled: boolean;
  sandbox: boolean;
  production: boolean;
  connected: boolean;
  connection_error?: string;
  cooperativa?: string;
  posto?: string;
  codigo_beneficiario?: string;
  webhook_configured: boolean;
  webhook_registered: boolean;
  webhook_url?: string;
  public_api_url?: string;
  webhook_token_set: boolean;
  ready_for_production: boolean;
};

export type SicrediSetupStep = {
  name: string;
  ok: boolean;
  message?: string;
};

export type SicrediProductionSetupResponse = {
  success: boolean;
  message: string;
  steps: SicrediSetupStep[];
};

export function useSicrediStatus() {
  return useQuery({
    queryKey: ['sicredi', 'status'],
    queryFn: async () => {
      const { data } = await client<SicrediIntegrationStatus>({
        url: '/v1/collections/sicredi/status',
        method: 'GET'
      });
      return data;
    }
  });
}

export function useTestSicrediConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<{ success: boolean; message: string; sandbox: boolean }>({
        url: '/v1/collections/sicredi/test-connection',
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['sicredi', 'status'] });
    }
  });
}

export function useRegisterSicrediWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (publicApiUrl?: string) => {
      const { data } = await client<{ success: boolean; message: string }>({
        url: '/v1/collections/sicredi/register-webhook',
        method: 'POST',
        data: publicApiUrl ? { public_api_url: publicApiUrl } : undefined
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['sicredi', 'status'] });
    }
  });
}

export function useSetupSicrediProduction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (publicApiUrl?: string) => {
      const { data } = await client<SicrediProductionSetupResponse>({
        url: '/v1/collections/sicredi/setup-production',
        method: 'POST',
        data: publicApiUrl ? { public_api_url: publicApiUrl } : undefined
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['sicredi', 'status'] });
    }
  });
}

export function useBillingSendLog(documentId: string) {
  return useQuery({
    queryKey: billingKeys.sendLog(documentId),
    queryFn: async () => {
      const { data } = await client<BillingSendLog[]>({
        url: `/v1/customer-billing-documents/${documentId}/send-log`,
        method: 'GET'
      });
      return data ?? [];
    },
    enabled: Boolean(documentId)
  });
}

export function useOverdueReceivables(params?: PageParams) {
  return useQuery({
    queryKey: billingKeys.overdue(params),
    queryFn: () =>
      pagedGet<OverdueReceivable>('/v1/collections/overdue', {
        page_index: params?.page_index ?? 0,
        page_size: params?.page_size ?? 10
      })
  });
}

export function useSendCollectionReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      accounts_receivable_id: string;
      reminder_level?: number;
      template_code?: string;
    }) => {
      const { data } = await client<{ success: boolean; message: string }>({
        url: '/v1/collections/remind',
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['billing', 'overdue'] })
  });
}

export const BILLING_PLACEHOLDERS = [
  '{{customer.name}}',
  '{{customer.document}}',
  '{{customer.billing_email}}',
  '{{invoice.number}}',
  '{{invoice.amount}}',
  '{{invoice.due_date}}',
  '{{invoice.issue_date}}',
  '{{invoice.description}}'
];

export function formatBillingStatus(status: string) {
  const map: Record<string, string> = {
    draft: 'Rascunho',
    ready: 'Pronta',
    sent: 'Enviada',
    cancelled: 'Cancelada'
  };
  return map[status] ?? status;
}

export function formatSicrediBoletoStatus(status?: string | null, paidAt?: string | null) {
  if (paidAt || status === 'paid') return 'Pago';
  if (status === 'issued') return 'Aguardando pagamento';
  if (status === 'cancelled') return 'Baixado / cancelado';
  if (status === 'failed') return 'Falha no boleto';
  return status ?? '—';
}
