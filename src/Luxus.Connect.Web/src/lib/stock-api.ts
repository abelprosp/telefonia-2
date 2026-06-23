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
