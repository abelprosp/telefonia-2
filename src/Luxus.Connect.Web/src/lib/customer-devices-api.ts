import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';

export type CustomerDeviceLink = {
  id: string;
  customer_id: string;
  device_stock_item_id?: string | null;
  description: string;
  brand: string;
  model: string;
  monthly_amount: number;
  start_date: string;
  end_date?: string | null;
  is_active: boolean;
};

export type AssignCustomerDeviceInput = {
  device_stock_item_id?: string | null;
  description?: string | null;
  brand?: string | null;
  model?: string | null;
  monthly_amount: number;
  start_date?: string | null;
};

export type UpdateCustomerDeviceInput = {
  description?: string;
  monthly_amount?: number;
};

export const customerDevicesQueryKey = (customerId: string) =>
  ['customer-devices', customerId] as const;

export function useCustomerDevices(customerId: string) {
  return useQuery({
    queryKey: customerDevicesQueryKey(customerId),
    queryFn: async () => {
      const { data } = await client<{ items: CustomerDeviceLink[]; total_count: number }>({
        url: `/v1/customers/${customerId}/devices`,
        method: 'GET',
        params: { page_index: 0, page_size: 100 }
      });
      return data;
    },
    enabled: Boolean(customerId)
  });
}

export function useAssignCustomerDevice(customerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: AssignCustomerDeviceInput) => {
      const { data } = await client<CustomerDeviceLink>({
        url: `/v1/customers/${customerId}/devices`,
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: customerDevicesQueryKey(customerId) });
    }
  });
}

export function useUpdateCustomerDevice(customerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      linkId,
      ...input
    }: UpdateCustomerDeviceInput & { linkId: string }) => {
      const { data } = await client<CustomerDeviceLink>({
        url: `/v1/customers/${customerId}/devices/${linkId}`,
        method: 'PATCH',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: customerDevicesQueryKey(customerId) });
    }
  });
}

export function useUnassignCustomerDevice(customerId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (linkId: string) => {
      await client({
        url: `/v1/customers/${customerId}/devices/${linkId}`,
        method: 'DELETE',
        data: {}
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: customerDevicesQueryKey(customerId) });
    }
  });
}
