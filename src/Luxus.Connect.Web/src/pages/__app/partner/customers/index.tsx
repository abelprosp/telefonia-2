import { useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Plus, Users } from 'lucide-react';
import { z } from 'zod';

import { type ListCustomerResponse } from '@/api';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatCpfCnpj } from '@/lib/format';
import { usePartnerCustomers } from '@/lib/partner-api';

import { PartnerCustomerCreateSheet } from '../-components/partner-customer-create-sheet';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10),
  providerId: z.string().optional()
});

const SKELETON_COLUMNS = [
  { header: 'Nome', cell: 'text' as const },
  { header: 'Documento', cell: 'text' as const },
  { header: 'Tipo', cell: 'text' as const },
  { header: 'Situação', cell: 'text' as const }
];

export const Route = createFileRoute('/__app/partner/customers/')({
  validateSearch: searchSchema,
  component: PartnerCustomersPage
});

function PartnerCustomersPage() {
  const { page, pageSize, providerId } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);

  const listQuery = usePartnerCustomers({
    page_index: page - 1,
    page_size: pageSize,
    ...(providerId ? { provider_id: providerId } : {})
  });

  const columns = useMemo<ColumnDef<ListCustomerResponse>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Nome',
        cell: ({ row }) => <span className="font-medium">{row.original.name}</span>
      },
      {
        accessorKey: 'cpf_cnpj',
        header: 'Documento',
        cell: ({ row }) => formatCpfCnpj(row.original.cpf_cnpj)
      },
      {
        accessorKey: 'type',
        header: 'Tipo'
      },
      {
        accessorKey: 'active',
        header: 'Situação',
        cell: ({ row }) => (row.original.active ? 'Ativo' : 'Inativo')
      }
    ],
    []
  );

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton pageSize={pageSize} columns={SKELETON_COLUMNS} />
    );
  }

  if (listQuery.error) {
    return (
      <div className="p-6 text-sm text-destructive">
        {isApiHttpError(listQuery.error)
          ? listQuery.error.message
          : getErrorMessage(listQuery.error)}
      </div>
    );
  }

  const items = listQuery.data?.items ?? [];

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Meus clientes"
        description="Clientes cadastrados na sua carteira de vendas"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            Novo cliente
          </Button>
        }
      />

      {items.length === 0 ? (
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Users />
            </EmptyMedia>
            <EmptyTitle>Nenhum cliente na sua carteira</EmptyTitle>
            <EmptyDescription>
              Cadastre o primeiro cliente para acompanhar linhas e financeiro.
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Button onClick={() => setCreateOpen(true)}>
              <Plus />
              Cadastrar cliente
            </Button>
          </EmptyContent>
        </Empty>
      ) : (
        <>
          <DataTable columns={columns} data={items} getRowId={(row) => row.id} />
          <DataTablePagination
            page={page}
            pageSize={pageSize}
            total={total}
            totalPages={totalPages}
            onPageChange={(next) =>
              navigate({ search: (prev) => ({ ...prev, page: next }) })
            }
            onPageSizeChange={(next) =>
              navigate({ search: (prev) => ({ ...prev, page: 1, pageSize: next }) })
            }
          />
        </>
      )}

      <PartnerCustomerCreateSheet
        open={createOpen}
        onOpenChange={setCreateOpen}
        preferredProviderId={providerId}
      />
    </div>
  );
}
