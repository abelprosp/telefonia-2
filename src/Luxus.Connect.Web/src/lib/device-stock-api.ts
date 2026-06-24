import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';

export type DeviceStockItem = {
  id: string;
  sku: string;
  brand: string;
  model: string;
  imei?: string | null;
  color?: string | null;
  storage_capacity?: string | null;
  unit_cost?: number | null;
  sale_price?: number | null;
  status: 'in_stock' | 'sold' | 'inactive';
  notes?: string | null;
  created_at: string;
  updated_at: string;
};

export type CreateDeviceStockItemInput = {
  sku?: string | null;
  brand: string;
  model: string;
  imei?: string | null;
  color?: string | null;
  storage_capacity?: string | null;
  unit_cost?: number | null;
  sale_price?: number | null;
  notes?: string | null;
};

export type UpdateDeviceStockItemInput = Partial<CreateDeviceStockItemInput> & {
  status?: 'in_stock' | 'sold' | 'inactive';
};

type PagedDeviceStock = {
  items: DeviceStockItem[];
  total_count: number;
};

export const deviceStockQueryKey = (params?: {
  page_index?: number;
  page_size?: number;
  status?: string;
}) => ['device-stock', params] as const;

export function useDeviceStockList(params: {
  page_index: number;
  page_size: number;
  status?: string;
}) {
  return useQuery({
    queryKey: deviceStockQueryKey(params),
    queryFn: async () => {
      const { data } = await client<PagedDeviceStock>({
        url: '/v1/device-stock',
        method: 'GET',
        params
      });
      return data;
    }
  });
}

export function useDeviceStockItem(id: string) {
  return useQuery({
    queryKey: ['device-stock', id],
    queryFn: async () => {
      const { data } = await client<DeviceStockItem>({
        url: `/v1/device-stock/${id}`,
        method: 'GET'
      });
      return data;
    },
    enabled: Boolean(id)
  });
}

export function useCreateDeviceStockItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateDeviceStockItemInput) => {
      const { data } = await client<DeviceStockItem>({
        url: '/v1/device-stock',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['device-stock'] });
    }
  });
}

export function useUpdateDeviceStockItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...input }: UpdateDeviceStockItemInput & { id: string }) => {
      const { data } = await client<DeviceStockItem>({
        url: `/v1/device-stock/${id}`,
        method: 'PATCH',
        data: input
      });
      return data;
    },
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: ['device-stock'] });
      void qc.invalidateQueries({ queryKey: ['device-stock', vars.id] });
    }
  });
}

export function formatDeviceStockStatus(status?: string | null) {
  switch (status) {
    case 'in_stock':
      return 'Em estoque';
    case 'sold':
      return 'Vendido';
    case 'inactive':
      return 'Inativo';
    default:
      return status ?? '—';
  }
}
