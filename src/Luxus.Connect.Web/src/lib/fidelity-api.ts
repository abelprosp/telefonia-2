import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { client } from '@/lib/client';

export type LineFidelityEvent = {
  id: string;
  event_type: string;
  occurred_at: string;
  user_id?: string | null;
  notes?: string | null;
};

export type LineFidelity = {
  id: string;
  phone_line_id: string;
  start_date: string;
  initial_months: number;
  predicted_end_date: string;
  auto_renew: boolean;
  renewal_period_months?: number | null;
  status: string;
  history: LineFidelityEvent[];
};

export type FidelityRenewalTrigger = {
  id: string;
  event_key: string;
  label: string;
  prompt_enabled: boolean;
};

export function useLineFidelity(phoneLineId: string, enabled = true) {
  return useQuery({
    queryKey: ['line-fidelity', phoneLineId],
    queryFn: async () => {
      const { data } = await client<LineFidelity>({
        url: `/v1/phone-lines/${phoneLineId}/fidelity`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(phoneLineId) && enabled,
    retry: false
  });
}

export function useUpsertLineFidelity(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      start_date: string;
      initial_months: number;
      auto_renew?: boolean;
      renewal_period_months?: number;
    }) => {
      const { data } = await client<LineFidelity>({
        url: `/v1/phone-lines/${phoneLineId}/fidelity`,
        method: 'PUT',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['line-fidelity', phoneLineId] })
  });
}

export function useFidelityRenewalDecision(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { renew: boolean; trigger?: string; notes?: string }) => {
      const { data } = await client<LineFidelity>({
        url: `/v1/phone-lines/${phoneLineId}/fidelity/renewal-decision`,
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['line-fidelity', phoneLineId] })
  });
}

export function useFidelityRenewalTriggers() {
  return useQuery({
    queryKey: ['fidelity-renewal-triggers'],
    queryFn: async () => {
      const { data } = await client<FidelityRenewalTrigger[]>({
        url: '/v1/fidelity-renewal-triggers',
        method: 'GET'
      });
      return data;
    }
  });
}

export function useUpdatePhoneLineExceedance(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      charge_exceedances?: boolean;
      exceedance_charge_type?: string;
    }) => {
      const { data } = await client({
        url: `/v1/phone-lines/${phoneLineId}/exceedance-settings`,
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

export type ExceedanceTerm = {
  id: string;
  term: string;
  charge_type: string;
  tabulated_amount?: number | null;
  active: boolean;
};

export function useExceedanceTerms() {
  return useQuery({
    queryKey: ['exceedance-terms'],
    queryFn: async () => {
      const { data } = await client<ExceedanceTerm[]>({
        url: '/v1/exceedance-terms',
        method: 'GET'
      });
      return data;
    }
  });
}

export function useCreateExceedanceTerm() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      term: string;
      charge_type: string;
      tabulated_amount?: number;
      active?: boolean;
    }) => {
      const { data } = await client<ExceedanceTerm>({
        url: '/v1/exceedance-terms',
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['exceedance-terms'] })
  });
}

export function useUpdateExceedanceTerm() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: string;
      term?: string;
      charge_type?: string;
      tabulated_amount?: number;
      active?: boolean;
    }) => {
      const { data } = await client<ExceedanceTerm>({
        url: `/v1/exceedance-terms/${id}`,
        method: 'PATCH',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['exceedance-terms'] })
  });
}

export type GeneratedContract = {
  id: string;
  trigger?: string | null;
  status: string;
  contract_template_name?: string | null;
  generated_at?: string | null;
  created_at: string;
};

export function usePhoneLineGeneratedContracts(phoneLineId: string, enabled = true) {
  return useQuery({
    queryKey: ['generated-contracts', 'phone-line', phoneLineId],
    queryFn: async () => {
      const { data } = await client<GeneratedContract[]>({
        url: `/v1/phone-lines/${phoneLineId}/generated-contracts`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(phoneLineId) && enabled
  });
}

export async function downloadFinancialExport(
  processingMonthId: string,
  format: 'json' | 'csv'
) {
  if (format === 'csv') {
    const { data } = await client<Blob>({
      url: `/v1/processing-months/${processingMonthId}/financial-export`,
      method: 'GET',
      params: { format: 'csv' },
      responseType: 'blob'
    });
    const url = URL.createObjectURL(data);
    const link = document.createElement('a');
    link.href = url;
    link.download = `export-financeiro-${processingMonthId}.csv`;
    link.click();
    URL.revokeObjectURL(url);
    return;
  }
  const { data } = await client({
    url: `/v1/processing-months/${processingMonthId}/financial-export`,
    method: 'GET',
    params: { format: 'json' }
  });
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `export-financeiro-${processingMonthId}.json`;
  link.click();
  URL.revokeObjectURL(url);
}
