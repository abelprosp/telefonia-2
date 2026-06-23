import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine } from 'lucide-react';

import type { ListProviderInvoiceResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatInvoiceStatus } from '@/lib/format';
import { formatFinancialStatus } from '@/lib/financial-api';

type InvoicesListSearch = {
  page: number;
  pageSize: number;
  processingMonthId: string | undefined;
};

export function createInvoicesColumns(opts: {
  page: number;
  pageSize: number;
  processingMonthId?: string;
  processingMonthLabelById: Map<string, string>;
}): ColumnDef<ListProviderInvoiceResponse>[] {
  const { page, pageSize, processingMonthId, processingMonthLabelById } = opts;
  const listSearch: InvoicesListSearch = {
    page,
    pageSize,
    processingMonthId
  };

  return [
    {
      accessorKey: 'provider_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Operadora" />
      ),
      cell: ({ row }) => <span>{row.original.provider_name}</span>
    },
    {
      accessorKey: 'provider_account_id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Conta" />
      ),
      cell: ({ row }) => (
        <span>
          {row.original.contracting_company_name}
          {' - '}
          {row.original.provider_account_number}
        </span>
      )
    },
    {
      accessorKey: 'issue_date',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Emissão" />
      ),
      cell: ({ row }) =>
        row.original.issue_date?.toDate()?.format('dd/MM/yyyy') ?? '—'
    },
    {
      accessorKey: 'due_date',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Vencimento" />
      ),
      cell: ({ row }) =>
        row.original.due_date?.toDate()?.format('dd/MM/yyyy') ?? '—'
    },
    {
      id: 'processing_month',
      header: () => <span>Mês proc.</span>,
      cell: ({ row }) => {
        const id = row.original.processing_month_id;
        if (!id) {
          return <span className="text-muted-foreground">—</span>;
        }
        return (
          <span>{processingMonthLabelById.get(id) ?? id.slice(0, 8)}</span>
        );
      }
    },
    {
      accessorKey: 'total_amount',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Valor" />
      ),
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.total_amount.toCurrency()}
        </span>
      )
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Situação" />
      ),
      cell: ({ row }) => formatInvoiceStatus(row.original.status)
    },
    {
      id: 'financial',
      header: () => <span>Financeiro</span>,
      cell: ({ row }) => {
        const inv = row.original as ListProviderInvoiceResponse & {
          account_payable_id?: string | null;
          account_payable_status?: string | null;
        };
        if (inv.account_payable_id) {
          return (
            <span className="text-sm">
              {formatFinancialStatus(inv.account_payable_status ?? 'open')}
            </span>
          );
        }
        return <span className="text-muted-foreground text-sm">Sem conta</span>;
      }
    },
    {
      id: 'actions',
      enableSorting: false,
      header: () => <div className="text-right">Ações</div>,
      cell: ({ row }) => {
        const inv = row.original;
        return (
          <div className="flex justify-end gap-2">
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
                        to="/invoices/$invoiceId"
                        params={{ invoiceId: inv.id }}
                        search={listSearch}
                        className="text-primary hover:underline"
                      >
                        <FilePenLine />
                      </Link>
                    }
                  />
                }
              />
              <TooltipContent>Abrir</TooltipContent>
            </Tooltip>
          </div>
        );
      }
    }
  ];
}
