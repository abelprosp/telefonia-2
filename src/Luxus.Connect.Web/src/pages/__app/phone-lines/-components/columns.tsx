import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine } from 'lucide-react';

import type { ListPhoneLineResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatPhoneLineStatus, formatPhoneNumber } from '@/lib/format';

export function createPhoneLinesColumns(opts: {
  listSearch: { page: number; pageSize: number };
}): ColumnDef<ListPhoneLineResponse>[] {
  const { listSearch } = opts;

  return [
    {
      accessorKey: 'number',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Número" />
      ),
      cell: ({ row }) => (
        <span>{formatPhoneNumber(row.original.number) ?? '—'}</span>
      )
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => (
        <span>{formatPhoneLineStatus(row.original.status)}</span>
      )
    },
    {
      accessorKey: 'line_classification',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Classificação" />
      ),
      cell: ({ row }) => <span>{row.original.line_classification}</span>
    },
    {
      accessorKey: 'provider_account_id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Conta operadora" />
      ),
      cell: ({ row }) => (
        <span className="truncate">{row.original.provider_account_number}</span>
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
                        to="/phone-lines/$phoneLineId"
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
              <TooltipContent>Abrir</TooltipContent>
            </Tooltip>
          </div>
        );
      }
    }
  ];
}
