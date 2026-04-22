import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { FilePenLine, Trash2 } from 'lucide-react';

import type { ListCustomerResponse } from '@/api';
import { DataTableColumnHeader } from '@/components/data-table/data-table-column-header';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { formatCpfCnpj } from '@/lib/format';

export function createCustomersColumns(opts: {
  onDeleteClick: (id: string) => void;
  deletePending: boolean;
  listSearch: {
    page: number;
    pageSize: number;
    providerId: string | undefined;
  };
}): ColumnDef<ListCustomerResponse>[] {
  const { onDeleteClick, deletePending, listSearch } = opts;

  return [
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Nome" />
      ),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.name}</span>
      )
    },
    {
      accessorKey: 'responsible_salesperson_user_id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Vendedor" />
      ),
      cell: ({ row }) => {
        const id = row.original.responsible_salesperson_user_id?.trim();
        if (!id) {
          return <span className="text-muted-foreground">—</span>;
        }
        const short = id.length > 14 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id;
        return <span title={id}>{short}</span>;
      }
    },
    {
      accessorKey: 'cpf_cnpj',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Documento" />
      ),
      cell: ({ row }) => <span>{formatCpfCnpj(row.original.cpf_cnpj)}</span>
    },
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Tipo" />
      ),
      cell: ({ row }) => row.original.type
    },
    {
      accessorKey: 'active',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Situação" />
      ),
      cell: ({ row }) => (row.original.active ? 'Ativo' : 'Inativo'),
      sortingFn: (rowA, rowB) => {
        const a = rowA.original.active ? 1 : 0;
        const b = rowB.original.active ? 1 : 0;
        return a - b;
      }
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
                        to="/customers/$customerId"
                        params={{ customerId: c.id }}
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

            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-destructive hover:text-destructive"
                    disabled={deletePending}
                    onClick={() => onDeleteClick(c.id)}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                }
              />
              <TooltipContent>Excluir</TooltipContent>
            </Tooltip>
          </div>
        );
      }
    }
  ];
}
