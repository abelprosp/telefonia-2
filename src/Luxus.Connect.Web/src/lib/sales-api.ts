import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { parseTotalCount } from '@/lib/query-utils';

export type ContractTemplate = {
  id: string;
  name: string;
  code: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type ContractTemplateDetail = ContractTemplate & {
  body_template: string;
};

export type SaleLineItem = {
  id: string;
  line_item_type: 'phone_line' | 'device' | 'other';
  description: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  phone_line_id?: string | null;
  device_sku?: string | null;
  sort_order: number;
};

export type GeneratedContract = {
  id: string;
  contract_template_id: string;
  status: string;
  rendered_html?: string | null;
  generated_at?: string | null;
};

export type Sale = {
  id: string;
  sale_number: string;
  customer_id: string;
  customer_name: string;
  salesperson_user_id: string;
  contract_template_id?: string | null;
  contract_template_name?: string | null;
  status: 'draft' | 'confirmed' | 'cancelled';
  sold_at?: string | null;
  total_amount: number;
  notes?: string | null;
  created_at: string;
  updated_at: string;
};

export type SaleDetail = Sale & {
  items: SaleLineItem[];
  contract?: GeneratedContract | null;
};

export type CreateSaleLineItemInput = {
  line_item_type: 'phone_line' | 'device' | 'other';
  description: string;
  quantity: number;
  unit_price: number;
  phone_line_id?: string;
  device_sku?: string;
};

export type CreateSaleInput = {
  customer_id: string;
  contract_template_id?: string;
  notes?: string;
  items?: CreateSaleLineItemInput[];
};

type Paged<T> = { items: T[]; total_count: number | string };
type PageParams = { page_index?: number; page_size?: number };

async function pagedGet<T>(url: string, params?: PageParams & Record<string, unknown>) {
  const { data } = await client<Paged<T>>({ url, method: 'GET', params });
  return { items: data.items ?? [], totalCount: parseTotalCount(data.total_count) };
}

export const salesKeys = {
  all: ['sales'] as const,
  list: (params: Record<string, unknown>) => ['sales', 'list', params] as const,
  detail: (id: string) => ['sales', 'detail', id] as const,
  templates: ['contract-templates'] as const,
  template: (id: string) => ['contract-templates', id] as const,
  partnerList: (params: Record<string, unknown>) => ['partner-commercial-sales', params] as const,
  partnerDetail: (id: string) => ['partner-commercial-sales', id] as const
};

export function formatSaleStatus(status: string) {
  const map: Record<string, string> = {
    draft: 'Rascunho',
    confirmed: 'Confirmada',
    cancelled: 'Cancelada'
  };
  return map[status] ?? status;
}

export function formatLineItemType(type: string) {
  const map: Record<string, string> = {
    phone_line: 'Linha telefônica',
    device: 'Aparelho',
    other: 'Outro'
  };
  return map[type] ?? type;
}

export function formatMoney(value: number) {
  return value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

export const CONTRACT_PLACEHOLDERS = [
  '{{customer.name}}',
  '{{customer.legal_name}}',
  '{{customer.document}}',
  '{{customer.type}}',
  '{{customer.address.full}}',
  '{{sale.sale_number}}',
  '{{sale.sold_at}}',
  '{{sale.total_amount}}',
  '{{sale.items_table}}',
  '{{salesperson.name}}'
];

export function useSales(params: PageParams & { status?: string }) {
  return useQuery({
    queryKey: salesKeys.list(params),
    queryFn: () => pagedGet<Sale>('/v1/sales', params)
  });
}

export function useSale(id: string) {
  return useQuery({
    queryKey: salesKeys.detail(id),
    queryFn: async () => {
      const { data } = await client<SaleDetail>({ url: `/v1/sales/${id}`, method: 'GET' });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateSaleInput) => {
      const { data } = await client<SaleDetail>({ url: '/v1/sales', method: 'POST', data: input });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: salesKeys.all });
    }
  });
}

export function useConfirmSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<SaleDetail>({ url: `/v1/sales/${id}/confirm`, method: 'POST' });
      return data;
    },
    onSuccess: (_data, id) => {
      void qc.invalidateQueries({ queryKey: salesKeys.all });
      void qc.invalidateQueries({ queryKey: salesKeys.detail(id) });
    }
  });
}

export function useCancelSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<SaleDetail>({ url: `/v1/sales/${id}/cancel`, method: 'POST' });
      return data;
    },
    onSuccess: (_data, id) => {
      void qc.invalidateQueries({ queryKey: salesKeys.all });
      void qc.invalidateQueries({ queryKey: salesKeys.detail(id) });
    }
  });
}

export function useContractTemplates(activeOnly = false) {
  return useQuery({
    queryKey: [...salesKeys.templates, { activeOnly }],
    queryFn: () =>
      pagedGet<ContractTemplate>('/v1/contract-templates', {
        page_index: 0,
        page_size: 200,
        active_only: activeOnly
      })
  });
}

export function useContractTemplate(id: string) {
  return useQuery({
    queryKey: salesKeys.template(id),
    queryFn: async () => {
      const { data } = await client<ContractTemplateDetail>({
        url: `/v1/contract-templates/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateContractTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: {
      name: string;
      code: string;
      body_template: string;
      active?: boolean;
    }) => {
      const { data } = await client<ContractTemplateDetail>({
        url: '/v1/contract-templates',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: salesKeys.templates });
    }
  });
}

export function useUpdateContractTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: {
      id: string;
      name?: string;
      code?: string;
      body_template?: string;
      active?: boolean;
    }) => {
      const { data } = await client<ContractTemplateDetail>({
        url: `/v1/contract-templates/${id}`,
        method: 'PATCH',
        data: input
      });
      return data;
    },
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: salesKeys.templates });
      void qc.invalidateQueries({ queryKey: salesKeys.template(vars.id) });
    }
  });
}

export function usePartnerCommercialSales(params: PageParams & { status?: string }) {
  return useQuery({
    queryKey: salesKeys.partnerList(params),
    queryFn: () => pagedGet<Sale>('/v1/partner/commercial-sales', params)
  });
}

export function usePartnerCommercialSale(id: string) {
  return useQuery({
    queryKey: salesKeys.partnerDetail(id),
    queryFn: async () => {
      const { data } = await client<SaleDetail>({
        url: `/v1/partner/commercial-sales/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function usePartnerCreateSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateSaleInput) => {
      const { data } = await client<SaleDetail>({
        url: '/v1/partner/commercial-sales',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['partner-commercial-sales'] });
    }
  });
}

export function usePartnerConfirmSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<SaleDetail>({
        url: `/v1/partner/commercial-sales/${id}/confirm`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_data, id) => {
      void qc.invalidateQueries({ queryKey: ['partner-commercial-sales'] });
      void qc.invalidateQueries({ queryKey: salesKeys.partnerDetail(id) });
    }
  });
}

export function usePartnerCancelSale() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<SaleDetail>({
        url: `/v1/partner/commercial-sales/${id}/cancel`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: (_data, id) => {
      void qc.invalidateQueries({ queryKey: ['partner-commercial-sales'] });
      void qc.invalidateQueries({ queryKey: salesKeys.partnerDetail(id) });
    }
  });
}

export function usePartnerContractTemplates() {
  return useQuery({
    queryKey: ['partner-contract-templates'],
    queryFn: () =>
      pagedGet<ContractTemplate>('/v1/partner/contract-templates', {
        page_index: 0,
        page_size: 200
      })
  });
}
