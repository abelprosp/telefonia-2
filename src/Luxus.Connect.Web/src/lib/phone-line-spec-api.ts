import { useMutation, useQueryClient } from '@tanstack/react-query';

import { client } from '@/lib/client';

export function useUpdatePhoneLineClassification(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      line_classification: string;
      titular_line_id?: string | null;
    }) => {
      const { data } = await client({
        url: `/v1/phone-lines/${phoneLineId}/classification`,
        method: 'PATCH',
        data: body
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['v1', 'phone-lines'] });
      void qc.invalidateQueries({ queryKey: ['phoneLinesController'] });
    }
  });
}

export function usePutPhoneLineTransition(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      transition_sub_status: string;
      start_date?: string;
    }) => {
      const { data } = await client({
        url: `/v1/phone-lines/${phoneLineId}/transition`,
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['v1', 'phone-lines'] });
      void qc.invalidateQueries({ queryKey: ['phoneLinesController'] });
    }
  });
}

export function useCreatePhoneLineService(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      provider_plan_service_id: string;
      name?: string;
      price?: number;
      service_type?: string;
      start_date?: string;
      end_date?: string;
    }) => {
      const { data } = await client({
        url: `/v1/phone-lines/${phoneLineId}/services`,
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['v1', 'phone-lines'] });
      void qc.invalidateQueries({ queryKey: ['phoneLinesController'] });
    }
  });
}

export function useDeletePhoneLineService(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (serviceId: string) => {
      await client({
        url: `/v1/phone-lines/${phoneLineId}/services/${serviceId}`,
        method: 'DELETE'
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['v1', 'phone-lines'] });
      void qc.invalidateQueries({ queryKey: ['phoneLinesController'] });
    }
  });
}

export function useManualReleaseCustomer(customerId: string, processingMonthId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (justification: string) => {
      const { data } = await client({
        url: `/v1/customers/${customerId}/processing-months/${processingMonthId}/manual-release`,
        method: 'POST',
        data: { justification }
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({
        queryKey: ['billing-readiness', customerId, processingMonthId]
      });
    }
  });
}
