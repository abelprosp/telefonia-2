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

export function useUpdateProviderPlan(providerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ planId, ...input }: UpdateProviderPlanInput & { planId: string }) => {
      const { data } = await client<GetProviderPlanResponse>({
        url: `/v1/providers/${providerId}/plans/${planId}`,
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
