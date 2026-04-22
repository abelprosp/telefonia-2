import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { format, parseISO } from 'date-fns';
import { FilePenLine } from 'lucide-react';

import type { ListProcessingMonthResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatProcessingMonthStatus } from '@/lib/format';

export function createProcessingMonthsColumns(opts: {
  listSearch: { page: number; pageSize: number };
  providerNameById: Map<string, string>;
}): ColumnDef<ListProcessingMonthResponse>[] {
  const { listSearch, providerNameById } = opts;

  return [
    {
      accessorKey: 'display_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Competência" />
      ),
      cell: ({ row }) => <span>{row.original.display_name}</span>
    },
    {
      id: 'provider',
      header: () => <span>Operadora</span>,
      cell: ({ row }) => (
        <span>
          {providerNameById.get(row.original.provider_id) ??
            row.original.provider_id}
        </span>
      )
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Situação" />
      ),
      cell: ({ row }) => (
        <span>{formatProcessingMonthStatus(row.original.status)}</span>
      )
    },
    {
      accessorKey: 'closed_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Fechamento" />
      ),
      cell: ({ row }) => {
        const raw = row.original.closed_at;
        if (!raw) {
          return <span className="text-muted-foreground text-sm">{'—'}</span>;
        }
        try {
          return (
            <span className="text-muted-foreground text-sm">
              {format(parseISO(raw), 'dd/MM/yyyy HH:mm')}
            </span>
          );
        } catch {
          return <span className="text-muted-foreground text-sm">{raw}</span>;
        }
      }
    },
    {
      accessorKey: 'closed_in_contingency',
      header: () => <span>Contingência</span>,
      cell: ({ row }) => (
        <span>{row.original.closed_in_contingency ? 'Sim' : 'Não'}</span>
      )
    },
    {
      id: 'actions',
      enableSorting: false,
      header: () => <div className="text-right">Ações</div>,
      cell: ({ row }) => {
        const pm = row.original;
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
                        to="/processing-months/$processingMonthId"
                        params={{ processingMonthId: pm.id }}
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
