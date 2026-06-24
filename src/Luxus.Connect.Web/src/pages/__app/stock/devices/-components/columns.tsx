import { type ColumnDef } from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { type DeviceStockItem, formatDeviceStockStatus } from '@/lib/device-stock-api';
import { formatMoney } from '@/lib/financial-api';

type DeviceStockColumnsOptions = {
  onMarkSold: (item: DeviceStockItem) => void;
};

export function createDeviceStockColumns({
  onMarkSold
}: DeviceStockColumnsOptions): ColumnDef<DeviceStockItem>[] {
  return [
    { accessorKey: 'sku', header: 'SKU' },
    {
      id: 'device',
      header: 'Aparelho',
      cell: ({ row }) => (
        <div>
          <div className="font-medium">
            {row.original.brand} {row.original.model}
          </div>
          {row.original.storage_capacity ? (
            <div className="text-muted-foreground text-xs">{row.original.storage_capacity}</div>
          ) : null}
        </div>
      )
    },
    {
      accessorKey: 'imei',
      header: 'IMEI',
      cell: ({ row }) => row.original.imei ?? '—'
    },
    {
      accessorKey: 'color',
      header: 'Cor',
      cell: ({ row }) => row.original.color ?? '—'
    },
    {
      accessorKey: 'sale_price',
      header: 'Preço venda',
      cell: ({ row }) =>
        row.original.sale_price != null ? formatMoney(row.original.sale_price) : '—'
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => (
        <Badge variant="outline">{formatDeviceStockStatus(row.original.status)}</Badge>
      )
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) =>
        row.original.status === 'in_stock' ? (
          <Button size="sm" variant="outline" onClick={() => onMarkSold(row.original)}>
            Marcar vendido
          </Button>
        ) : null
    }
  ];
}
