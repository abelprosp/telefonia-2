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
  overdue: (params?: PageParams) => ['billing', 'overdue', params] as const
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
      template_code
    }: {
      receivableId: string;
      template_code?: string;
    }) => {
      const { data } = await client<{ id: string }>({
        url: `/v1/customer-billing-documents/from-receivable/${receivableId}`,
        method: 'POST',
        data: { template_code: template_code ?? 'default-billing-invoice' }
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['billing', 'documents'] })
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
