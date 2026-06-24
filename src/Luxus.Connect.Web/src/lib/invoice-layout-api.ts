import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { parseTotalCount } from '@/lib/query-utils';

import type { InvoiceLayoutConfig, InvoiceLayoutTemplate } from './invoice-layout/types';

type Paged<T> = { items: T[]; total_count: number | string };
type PageParams = { page_index?: number; page_size?: number };

export const invoiceLayoutKeys = {
  list: (params?: PageParams) => ['invoice-layout', 'list', params] as const,
  detail: (id: string) => ['invoice-layout', id] as const
};

async function pagedGet<T>(url: string, params?: PageParams) {
  const { data } = await client<Paged<T>>({ url, method: 'GET', params });
  return { items: data.items ?? [], totalCount: parseTotalCount(data.total_count) };
}

export function useInvoiceLayoutTemplates(params?: PageParams) {
  return useQuery({
    queryKey: invoiceLayoutKeys.list(params),
    queryFn: () =>
      pagedGet<InvoiceLayoutTemplate>('/v1/invoice-layout-templates', {
        page_index: params?.page_index ?? 0,
        page_size: params?.page_size ?? 50
      })
  });
}

export function useInvoiceLayoutTemplate(id: string) {
  return useQuery({
    queryKey: invoiceLayoutKeys.detail(id),
    queryFn: async () => {
      const { data } = await client<InvoiceLayoutTemplate>({
        url: `/v1/invoice-layout-templates/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateInvoiceLayoutTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      name: string;
      code: string;
      config_json: InvoiceLayoutConfig;
      active?: boolean;
    }) => {
      const { data } = await client<InvoiceLayoutTemplate>({
        url: '/v1/invoice-layout-templates',
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['invoice-layout'] })
  });
}

export function useUpdateInvoiceLayoutTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: string;
      name: string;
      config_json: InvoiceLayoutConfig;
      active?: boolean;
    }) => {
      const { data } = await client<InvoiceLayoutTemplate>({
        url: `/v1/invoice-layout-templates/${id}`,
        method: 'PATCH',
        data: body
      });
      return data;
    },
    onSuccess: (_, v) => {
      void qc.invalidateQueries({ queryKey: invoiceLayoutKeys.detail(v.id) });
      void qc.invalidateQueries({ queryKey: ['invoice-layout'] });
    }
  });
}

export function usePreviewInvoiceLayout() {
  return useMutation({
    mutationFn: async (config_json: InvoiceLayoutConfig) => {
      const { data } = await client<{ html: string }>({
        url: '/v1/invoice-layout-templates/preview',
        method: 'POST',
        data: { config_json }
      });
      return data.html;
    }
  });
}
