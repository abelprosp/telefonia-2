import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { client } from '@/lib/client';

export type BillingCompositionItem = {
  id: string;
  item_type: 'service' | 'discount' | 'extra_charge' | 'installment';
  description: string;
  amount: number;
  quantity: number;
  installment_count?: number | null;
  installment_current?: number | null;
  start_date?: string | null;
  end_date?: string | null;
};

export type LineBillingProcessing = {
  id: string;
  perspective: 'luxus_customer' | 'customer_end_user';
  label?: string | null;
  mirror_from_primary: boolean;
  total_amount: number;
  items: BillingCompositionItem[];
};

export type ListLineBillingProcessingsResponse = {
  link_id: string;
  processings: LineBillingProcessing[];
};

export type AuditLogEntry = {
  id: string;
  change_type: string;
  entity_name: string;
  key_values: string;
  changed_by?: string | null;
  old_values?: string | null;
  new_values?: string | null;
  timestamp: string;
};

const keys = {
  all: (phoneLineId: string) => ['billing-processings', phoneLineId] as const
};

export function useLineBillingProcessings(phoneLineId: string, enabled = true) {
  return useQuery({
    queryKey: keys.all(phoneLineId),
    queryFn: async () => {
      const { data } = await client<ListLineBillingProcessingsResponse>({
        url: `/v1/phone-lines/${phoneLineId}/billing-processings`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(phoneLineId) && enabled
  });
}

export function useEnableEndUserProcessing(phoneLineId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<LineBillingProcessing>({
        url: `/v1/phone-lines/${phoneLineId}/billing-processings/end-user`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all(phoneLineId) })
  });
}

export function useMirrorBillingProcessing(phoneLineId: string, processingId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<LineBillingProcessing>({
        url: `/v1/phone-lines/${phoneLineId}/billing-processings/${processingId}/mirror-from-primary`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all(phoneLineId) })
  });
}

export function useCreateBillingCompositionItem(phoneLineId: string, processingId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      item_type: string;
      description: string;
      amount: number;
      quantity?: number;
      installment_count?: number;
      installment_current?: number;
    }) => {
      const { data } = await client<BillingCompositionItem>({
        url: `/v1/phone-lines/${phoneLineId}/billing-processings/${processingId}/items`,
        method: 'POST',
        data: body
      });
      return data;
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all(phoneLineId) })
  });
}

export function useDeleteBillingCompositionItem(
  phoneLineId: string,
  processingId: string
) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (itemId: string) => {
      await client({
        url: `/v1/phone-lines/${phoneLineId}/billing-processings/${processingId}/items/${itemId}`,
        method: 'DELETE'
      });
    },
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all(phoneLineId) })
  });
}

export function perspectiveLabel(perspective: string) {
  if (perspective === 'customer_end_user') return 'Cliente → Usuário final';
  return 'Luxus → Cliente';
}

export function itemTypeLabel(type: string) {
  switch (type) {
    case 'discount':
      return 'Desconto';
    case 'extra_charge':
      return 'Cobrança extra';
    case 'installment':
      return 'Parcelamento';
    default:
      return 'Serviço';
  }
}
