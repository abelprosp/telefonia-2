import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine, Link2 } from 'lucide-react';

import type { ListPhoneLineResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import {
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
        <span>{formatPhoneNumber(row.original.number) ?? '—'}</span>
      )
    },
    {
      accessorKey: 'line_classification',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Classificação" />
      ),
      cell: ({ row }) => (
        <span>
          {formatLineClassification(row.original.line_classification)}
        </span>
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
              <TooltipContent>Abrir</TooltipContent>
            </Tooltip>
          </div>
        );
      }
    }
  ];
}
