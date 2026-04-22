import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine } from 'lucide-react';

import type { ListBillingCycleResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatBillingCycleStatus } from '@/lib/format';

export function createBillingCyclesColumns(opts: {
  listSearch: { page: number; pageSize: number };
}): ColumnDef<ListBillingCycleResponse>[] {
  const { listSearch } = opts;

  return [
    {
      accessorKey: 'code',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Código" />
      ),
      cell: ({ row }) => <span>{row.original.code}</span>
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Nome" />
      ),
      cell: ({ row }) => <span>{row.original.name}</span>
    },
    {
      accessorKey: 'start_date',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Início" />
      ),
      cell: ({ row }) => (
        <span>{row.original.start_date.formatAsDate('dd/MM/yyyy')}</span>
      )
    },
    {
      accessorKey: 'end_date',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Fim" />
      ),
      cell: ({ row }) => (
        <span>{row.original.end_date.formatAsDate('dd/MM/yyyy')}</span>
      )
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Situação" />
      ),
      cell: ({ row }) => (
        <span>{formatBillingCycleStatus(row.original.status)}</span>
      )
    },
    {
      id: 'actions',
      enableSorting: false,
      header: () => <div className="text-right">Ações</div>,
      cell: ({ row }) => {
        const c = row.original;
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
                        to="/billing-cycles/$cycleId"
                        params={{ cycleId: c.id }}
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
