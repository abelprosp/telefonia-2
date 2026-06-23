import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { parseTotalCount } from '@/lib/query-utils';

export type FinancialSummary = {
  total_payable_open: number;
  total_receivable_open: number;
  total_partner_commission_accrued: number;
  payable_overdue_count: number;
  receivable_overdue_count: number;
  provider_invoices_count: number;
  provider_invoices_total_amount: number;
  provider_invoices_without_payable_count: number;
  open_processing_months_count: number;
  billing_documents_draft_count: number;
  billing_documents_ready_count: number;
  billing_documents_sent_count: number;
};

export type AccountPayable = {
  id: string;
  description: string;
  vendor_name: string;
  provider_invoice_id?: string | null;
  partner_salesperson_user_id?: string | null;
  issue_date: string;
  due_date: string;
  amount: number;
  paid_amount: number;
  balance: number;
  status: string;
  notes?: string | null;
  created_at: string;
};

export type AccountReceivable = {
  id: string;
  customer_id: string;
  customer_name: string;
  description: string;
  processing_month_id?: string | null;
  issue_date: string;
  due_date: string;
  amount: number;
  received_amount: number;
  balance: number;
  status: string;
  notes?: string | null;
  created_at: string;
};

export type PartnerSale = {
  id: string;
  salesperson_user_id: string;
  customer_id: string;
  customer_name: string;
  phone_line_id: string;
  phone_line_number: string;
  reference_month: string;
  gross_amount: number;
  commission_percent: number;
  commission_amount: number;
  status: string;
  account_payable_id?: string | null;
  created_at: string;
};

export type PartnerFinancialSummary = {
  total_gross_sales: number;
  total_commission_accrued: number;
  total_commission_approved: number;
  total_commission_paid: number;
  total_receivable_from_sales: number;
  pending_sales_count: number;
};

type Paged<T> = { items: T[]; total_count: number | string };
type PageParams = { page_index?: number; page_size?: number };

async function pagedGet<T>(url: string, params?: PageParams & Record<string, unknown>) {
  const { data } = await client<Paged<T>>({ url, method: 'GET', params });
  return { items: data.items ?? [], totalCount: parseTotalCount(data.total_count) };
}

export const financialKeys = {
  summary: ['financial', 'summary'] as const,
  payables: (params?: PageParams & { status?: string }) =>
    ['financial', 'payables', params] as const,
  receivables: (params?: PageParams & { status?: string; customer_id?: string }) =>
    ['financial', 'receivables', params] as const,
  partnerSales: (params?: PageParams & { status?: string; salesperson_user_id?: string }) =>
    ['financial', 'partner-sales', params] as const,
  commissionSettings: ['financial', 'commission-settings'] as const,
  partnerSummary: ['partner', 'financial', 'summary'] as const,
  partnerSalesList: (params?: PageParams) => ['partner', 'sales', params] as const
};

export function useFinancialSummary() {
  return useQuery({
    queryKey: financialKeys.summary,
    queryFn: async () => {
      const { data } = await client<FinancialSummary>({
        url: '/v1/financial/summary',
        method: 'GET'
      });
      return data;
    }
  });
}

export function useAccountsPayable(params: PageParams & { status?: string }) {
  return useQuery({
    queryKey: financialKeys.payables(params),
    queryFn: () => pagedGet<AccountPayable>('/v1/accounts-payable', params)
  });
}

export function useAccountsReceivable(
  params: PageParams & { status?: string; customer_id?: string }
) {
  return useQuery({
    queryKey: financialKeys.receivables(params),
    queryFn: () => pagedGet<AccountReceivable>('/v1/accounts-receivable', params)
  });
}

export function usePartnerSalesAdmin(
  params: PageParams & { status?: string; salesperson_user_id?: string }
) {
  return useQuery({
    queryKey: financialKeys.partnerSales(params),
    queryFn: () => pagedGet<PartnerSale>('/v1/partner-sales', params)
  });
}

export function usePartnerCommissionSettings() {
  return useQuery({
    queryKey: financialKeys.commissionSettings,
    queryFn: async () => {
      const { data } = await client<{ default_commission_percent: number }>({
        url: '/v1/partner-commission-settings',
        method: 'GET'
      });
      return data;
    }
  });
}

export function usePartnerFinancialSummary() {
  return useQuery({
    queryKey: financialKeys.partnerSummary,
    queryFn: async () => {
      const { data } = await client<PartnerFinancialSummary>({
        url: '/v1/partner/financial/summary',
        method: 'GET'
      });
      return data;
    }
  });
}

export function usePartnerSales(params: PageParams) {
  return useQuery({
    queryKey: financialKeys.partnerSalesList(params),
    queryFn: () => pagedGet<PartnerSale>('/v1/partner/sales', params)
  });
}

export function useCreatePayableFromInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (invoiceId: string) => {
      const { data } = await client<{ id: string }>({
        url: `/v1/accounts-payable/from-provider-invoice/${invoiceId}`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
      void qc.invalidateQueries({ queryKey: [{ url: '/v1/provider-invoices' }] });
    }
  });
}

export function useCreatePayable() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const { data } = await client<AccountPayable>({
        url: '/v1/accounts-payable',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
    }
  });
}

export function useCreateReceivable() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const { data } = await client<AccountReceivable>({
        url: '/v1/accounts-receivable',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
    }
  });
}

export function useRegisterPayablePayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...input }: { id: string; amount: number; payment_date: string }) => {
      await client({ url: `/v1/accounts-payable/${id}/payments`, method: 'POST', data: input });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
    }
  });
}

export function useRegisterReceivablePayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...input }: { id: string; amount: number; payment_date: string }) => {
      await client({
        url: `/v1/accounts-receivable/${id}/payments`,
        method: 'POST',
        data: input
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
    }
  });
}

export function useSyncPartnerSales() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (reference_month: string) => {
      const { data } = await client<{ inserted_count: number }>({
        url: '/v1/partner-sales/sync',
        method: 'POST',
        data: { reference_month }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
      void qc.invalidateQueries({ queryKey: ['partner'] });
    }
  });
}

export function useUpdatePartnerSaleStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }) => {
      const { data } = await client<PartnerSale>({
        url: `/v1/partner-sales/${id}`,
        method: 'PATCH',
        data: { status }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['financial'] });
      void qc.invalidateQueries({ queryKey: ['partner'] });
    }
  });
}

export function useUpdateCommissionSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (default_commission_percent: number) => {
      const { data } = await client({
        url: '/v1/partner-commission-settings',
        method: 'PUT',
        data: { default_commission_percent }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: financialKeys.commissionSettings });
    }
  });
}

export const formatMoney = (value: number) =>
  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value ?? 0);

export const formatFinancialStatus = (status: string) => {
  switch (status) {
    case 'open':
      return 'Em aberto';
    case 'partially_settled':
      return 'Parcial';
    case 'settled':
      return 'Quitado';
    case 'overdue':
      return 'Vencido';
    case 'cancelled':
      return 'Cancelado';
    case 'accrued':
      return 'Provisionado';
    case 'approved':
      return 'Aprovado';
    case 'paid':
      return 'Pago';
    default:
      return status;
  }
};

export const todayISO = () => new Date().toISOString().slice(0, 10);

export const firstDayOfMonthISO = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`;
};
