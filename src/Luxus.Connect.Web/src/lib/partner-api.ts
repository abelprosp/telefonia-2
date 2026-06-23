import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { parseTotalCount } from '@/lib/query-utils';

import type { ListCustomerResponse } from '@/api';

export type PartnerDashboardStats = {
  customers_count: number;
  phone_lines_count: number;
  pending_operation_requests_count: number;
  total_base_cost: number;
  total_cost_with_consumption: number;
};

export type PartnerPhoneLine = {
  id: string;
  number: string;
  status: string;
  transition_sub_status?: string | null;
  customer_id?: string | null;
  customer_name?: string | null;
  provider_plan_name: string;
  base_cost?: number | null;
  cost_with_consumption?: number | null;
};

export type PhoneLineOperationRequest = {
  id: string;
  phone_line_id: string;
  phone_line_number: string;
  customer_id: string;
  customer_name: string;
  operation_type: 'activation' | 'deactivation';
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  justification?: string | null;
  admin_notes?: string | null;
  requested_by_user_id: string;
  reviewed_by_user_id?: string | null;
  reviewed_at?: string | null;
  created_at: string;
};

type Paged<T> = { items: T[]; total_count: number | string };

type PageParams = { page_index?: number; page_size?: number };

async function pagedGet<T>(url: string, params?: PageParams & Record<string, unknown>) {
  const { data } = await client<Paged<T>>({ url, method: 'GET', params });
  return {
    items: data.items ?? [],
    totalCount: parseTotalCount(data.total_count)
  };
}

export const partnerKeys = {
  stats: ['partner', 'stats'] as const,
  customers: (params?: PageParams & { provider_id?: string }) =>
    ['partner', 'customers', params] as const,
  customer: (id: string) => ['partner', 'customers', id] as const,
  customerPhoneLines: (id: string, params?: PageParams) =>
    ['partner', 'customers', id, 'phone-lines', params] as const,
  phoneLines: (params?: PageParams) => ['partner', 'phone-lines', params] as const,
  providers: ['partner', 'providers'] as const,
  requests: (params?: PageParams) =>
    ['partner', 'line-operation-requests', params] as const,
  adminRequests: (params?: PageParams) =>
    ['admin', 'line-operation-requests', params] as const
};

export function usePartnerDashboardStats() {
  return useQuery({
    queryKey: partnerKeys.stats,
    queryFn: async () => {
      const { data } = await client<PartnerDashboardStats>({
        url: '/v1/partner/stats/dashboard',
        method: 'GET'
      });
      return data;
    }
  });
}

export function usePartnerProviders() {
  return useQuery({
    queryKey: partnerKeys.providers,
    queryFn: () =>
      pagedGet<{ id: string; name: string; slug: string }>(
        '/v1/partner/providers',
        { page_index: 0, page_size: 500 }
      )
  });
}

export function usePartnerCustomers(params: PageParams & { provider_id?: string }) {
  return useQuery({
    queryKey: partnerKeys.customers(params),
    queryFn: () =>
      pagedGet<ListCustomerResponse>('/v1/partner/customers', params)
  });
}

export function usePartnerCustomer(id: string) {
  return useQuery({
    queryKey: partnerKeys.customer(id),
    queryFn: async () => {
      const { data } = await client<ListCustomerResponse>({
        url: `/v1/partner/customers/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: !!id
  });
}

export function usePartnerPhoneLines(params: PageParams) {
  return useQuery({
    queryKey: partnerKeys.phoneLines(params),
    queryFn: () => pagedGet<PartnerPhoneLine>('/v1/partner/phone-lines', params)
  });
}

export function usePartnerLineRequests(params: PageParams) {
  return useQuery({
    queryKey: partnerKeys.requests(params),
    queryFn: () =>
      pagedGet<PhoneLineOperationRequest>(
        '/v1/partner/phone-line-operation-requests',
        params
      )
  });
}

export function useAdminLineRequests(params: PageParams) {
  return useQuery({
    queryKey: partnerKeys.adminRequests(params),
    queryFn: () =>
      pagedGet<PhoneLineOperationRequest>(
        '/v1/phone-line-operation-requests',
        params
      )
  });
}

export function useCreatePartnerLineRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      phone_line_id: string;
      customer_id: string;
      operation_type: 'activation' | 'deactivation';
      justification?: string;
    }) => {
      const { data } = await client<PhoneLineOperationRequest>({
        url: '/v1/partner/phone-line-operation-requests',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['partner'] });
      void queryClient.invalidateQueries({ queryKey: ['admin', 'line-operation-requests'] });
    }
  });
}

export function useReviewLineRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      status,
      admin_notes
    }: {
      id: string;
      status: 'approved' | 'rejected';
      admin_notes?: string;
    }) => {
      const { data } = await client<PhoneLineOperationRequest>({
        url: `/v1/phone-line-operation-requests/${id}`,
        method: 'PATCH',
        data: { status, admin_notes }
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['partner'] });
      void queryClient.invalidateQueries({ queryKey: ['admin', 'line-operation-requests'] });
    }
  });
}

export function useCreatePartnerCustomer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const { data } = await client<{ id: string }>({
        url: '/v1/partner/customers',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: partnerKeys.stats });
      void queryClient.invalidateQueries({ queryKey: ['partner', 'customers'] });
    }
  });
}
