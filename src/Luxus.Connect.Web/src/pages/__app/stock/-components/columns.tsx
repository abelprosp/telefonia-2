import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine, Link2 } from 'lucide-react';

import type { ListPhoneLineResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatMoney } from '@/lib/financial-api';
import {
  formatCpfCnpj,
  formatLineClassification,
  formatPhoneLineStatus,
  formatPhoneNumber
} from '@/lib/format';

export function createStockLinesColumns(opts: {
  listSearch: { page: number; pageSize: number };
  onLinkCustomer?: (line: ListPhoneLineResponse) => void;
}): ColumnDef<ListPhoneLineResponse>[] {
  const { listSearch, onLinkCustomer } = opts;

  return [
    {
      accessorKey: 'number',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Número" />
      ),
      cell: ({ row }) => (
        <span className="font-medium text-foreground">
          {formatPhoneNumber(row.original.number) ?? '—'}
        </span>
      )
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const s = row.original.status?.toLowerCase();
        const isStock = s === 'in_stock' || s === '0';
        return (
          <Badge variant={isStock ? 'success-light' : 'secondary'} size="sm">
            {formatPhoneLineStatus(row.original.status)}
          </Badge>
        );
      }
    },
    {
      id: 'provider_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Operadora" />
      ),
      cell: ({ row }) => (
        <span>{row.original.provider_name || '—'}</span>
      )
    },
    {
      id: 'contracted_luxus_cnpj',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="CNPJ Luxus" />
      ),
      cell: ({ row }) => (
        <span className="text-xs text-muted-foreground">
          {row.original.contracted_luxus_cnpj
            ? formatCpfCnpj(row.original.contracted_luxus_cnpj)
            : '—'}
        </span>
      )
    },
    {
      accessorKey: 'provider_account_number',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Conta" />
      ),
      cell: ({ row }) => (
        <span>{row.original.provider_account_number || '—'}</span>
      )
    },
    {
      accessorKey: 'base_cost',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Custo Base" />
      ),
      cell: ({ row }) => (
        <span>{formatMoney(row.original.base_cost)}</span>
      )
    },
    {
      id: 'has_consumption',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Com Consumo" />
      ),
      cell: ({ row }) => {
        const base = row.original.base_cost ?? 0;
        const totalWith = row.original.cost_with_consumption ?? 0;
        const hasConsumption = totalWith > base;
        return (
          <Badge variant={hasConsumption ? 'info-light' : 'outline'} size="sm">
            {hasConsumption ? 'Sim' : 'Não'}
          </Badge>
        );
      }
    },
    {
      accessorKey: 'last_invoice_number',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Última Fatura" />
      ),
      cell: ({ row }) => (
        <span className="text-xs text-muted-foreground">
          {row.original.last_invoice_number ?? '—'}
        </span>
      )
    },
    {
      accessorKey: 'line_classification',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Classificação" />
      ),
      cell: ({ row }) => (
        <span className="text-xs">
          {formatLineClassification(row.original.line_classification)}
        </span>
      )
    },
    {
      id: 'actions',
      enableSorting: false,
      header: () => <div className="text-right">Ações</div>,
      cell: ({ row }) => {
        const line = row.original;
        return (
          <div className="flex justify-end gap-2">
            {onLinkCustomer ? (
              <Tooltip>
                <TooltipTrigger
                  render={
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="text-primary hover:text-primary"
                      onClick={() => onLinkCustomer(line)}
                    >
                      <Link2 />
                    </Button>
                  }
                />
                <TooltipContent>Vincular cliente</TooltipContent>
              </Tooltip>
            ) : null}
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    nativeButton={false}
                    variant="ghost"
                    size="sm"
                    className="text-primary hover:text-primary"
                    render={
                      <Link
                        to="/stock/$phoneLineId"
                        params={{ phoneLineId: line.id }}
                        search={listSearch}
                        className="text-primary hover:underline"
                      >
                        <FilePenLine />
                      </Link>
                    }
                  />
                }
              />
              <TooltipContent>Detalhes / Editar</TooltipContent>
            </Tooltip>
          </div>
        );
      }
    }
  ];
}
