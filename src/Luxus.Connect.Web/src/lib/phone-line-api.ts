import { useMutation, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import { getV1PhoneLinesIdCustomerLinksQueryKey } from '@/api';

export function useUpdatePhoneLineMonthlyAmount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      phoneLineId,
      monthly_amount
    }: {
      phoneLineId: string;
      monthly_amount: number | null;
    }) => {
      const { data } = await client({
        url: `/v1/phone-lines/${phoneLineId}/customer-links/active`,
        method: 'PATCH',
        data: { monthly_amount }
      });
      return data;
    },
    onSuccess: (_, vars) => {
      void qc.invalidateQueries({
        queryKey: getV1PhoneLinesIdCustomerLinksQueryKey(vars.phoneLineId)
      });
      void qc.invalidateQueries({ queryKey: ['v1', 'customers'] });
      void qc.invalidateQueries({ queryKey: ['getV1CustomersIdPhoneLines'] });
    }
  });
}

export function parseMoneyInput(value: string): number | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  const normalized = trimmed.replace(/\./g, '').replace(',', '.');
  const num = Number(normalized);
  return Number.isFinite(num) ? num : null;
}

export function formatMoneyInput(value: number | null | undefined): string {
  if (value === null || value === undefined) return '';
  return value.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
