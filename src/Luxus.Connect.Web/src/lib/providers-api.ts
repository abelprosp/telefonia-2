import { useMutation, useQueryClient } from '@tanstack/react-query';

import { providersControllerGetByIdQueryKey, type GetProviderPlanResponse } from '@/api';
import client from '@/lib/client';

export type CreateProviderPlanInput = {
  code: string;
  name: string;
  monthly_price?: number | null;
};

export type UpdateProviderPlanInput = CreateProviderPlanInput;

export function useCreateProviderPlan(providerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateProviderPlanInput) => {
      const { data } = await client<GetProviderPlanResponse>({
        url: `/v1/providers/${providerId}/plans`,
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: providersControllerGetByIdQueryKey(providerId) });
    }
  });
}

export type CreateProviderPlanServiceInput = {
  name: string;
  invoice_name?: string | null;
  service_type: string;
  recurring: boolean;
  price?: number | null;
  application_type?: string | null;
  availability_rule?: string | null;
  exclusive_customer_id?: string | null;
};

export type UpdateProviderPlanServiceInput = {
  name?: string | null;
  invoice_name?: string | null;
  service_type?: string | null;
  recurring?: boolean | null;
  price?: number | null;
  active?: boolean | null;
  application_type?: string | null;
  availability_rule?: string | null;
  exclusive_customer_id?: string | null;
};

export function useCreateProviderPlanService(providerId: string, planId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateProviderPlanServiceInput) => {
      const { data } = await client<{ id: string }>({
        url: `/v1/providers/${providerId}/plans/${planId}/services`,
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: providersControllerGetByIdQueryKey(providerId) });
    }
  });
}

export function useUpdateProviderPlanService(providerId: string, planId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ serviceId, ...input }: UpdateProviderPlanServiceInput & { serviceId: string }) => {
      const { data } = await client<{ id: string }>({
        url: `/v1/providers/${providerId}/plans/${planId}/services/${serviceId}`,
        method: 'PATCH',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: providersControllerGetByIdQueryKey(providerId) });
    }
  });
}

export function useDeleteProviderPlanService(providerId: string, planId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (serviceId: string) => {
      await client<void>({
        url: `/v1/providers/${providerId}/plans/${planId}/services/${serviceId}`,
        method: 'DELETE'
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: providersControllerGetByIdQueryKey(providerId) });
    }
  });
}

