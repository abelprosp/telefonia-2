import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getV1PhoneLinesQueryKey } from '@/api';
import client from '@/lib/client';

export type CreateStockPhoneLineInput = {
  number: string;
  provider_id: string;
  provider_account_number: string;
  provider_plan_id: string;
};

export function useCreateStockPhoneLine() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateStockPhoneLineInput) => {
      const { data } = await client({ url: '/v1/phone-lines/stock', method: 'POST', data: input });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: getV1PhoneLinesQueryKey() });
    }
  });
}

export type BulkCreateStockPhoneLinesInput = {
  provider_id: string;
  provider_account_number: string;
  provider_plan_id: string;
  lines: { number: string; provider_id?: string; provider_account_number?: string; provider_plan_id?: string }[];
};

export type BulkCreateStockPhoneLinesResponse = {
  created: number;
  ignored: number;
  rejected: number;
  items: { index: number; number: string; status: string; reason?: string }[];
};

export function useBulkCreateStockPhoneLines() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: BulkCreateStockPhoneLinesInput) => {
      const { data } = await client({
        url: '/v1/phone-lines/stock/bulk',
        method: 'POST',
        data: input
      });
      return data as BulkCreateStockPhoneLinesResponse;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: getV1PhoneLinesQueryKey() });
    }
  });
}
