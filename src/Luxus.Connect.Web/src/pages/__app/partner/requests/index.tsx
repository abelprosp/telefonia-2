import { useMemo } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatPhoneNumber } from '@/lib/format';
import {
  type PhoneLineOperationRequest,
  usePartnerLineRequests
} from '@/lib/partner-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

const SKELETON_COLUMNS = [
  { header: 'Linha', cell: 'text' as const },
  { header: 'Cliente', cell: 'text' as const },
  { header: 'Operação', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  { header: 'Enviada em', cell: 'text' as const },
  { header: 'Resposta', cell: 'text' as const }
];

export const Route = createFileRoute('/__app/partner/requests/')({
  validateSearch: searchSchema,
  component: PartnerRequestsPage
});

function PartnerRequestsPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();

  const listQuery = usePartnerLineRequests({
    page_index: page - 1,
    page_size: pageSize
  });

  const columns = useMemo<ColumnDef<PhoneLineOperationRequest>[]>(
    () => [
      {
        accessorKey: 'phone_line_number',
        header: 'Linha',
        cell: ({ row }) =>
          formatPhoneNumber(row.original.phone_line_number) ??
          row.original.phone_line_number
      },
      {
        accessorKey: 'customer_name',
        header: 'Cliente'
      },
      {
        accessorKey: 'operation_type',
        header: 'Operação',
        cell: ({ row }) =>
          row.original.operation_type === 'activation' ? 'Ativação' : 'Desativação'
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatStatus(row.original.status)
      },
      {
        accessorKey: 'created_at',
        header: 'Enviada em',
        cell: ({ row }) =>
          new Date(row.original.created_at).toLocaleString('pt-BR')
      },
      {
        accessorKey: 'admin_notes',
        header: 'Resposta da empresa',
        cell: ({ row }) => row.original.admin_notes?.trim() || '—'
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
        title="Solicitações de linha"
        description="Acompanhe pedidos de ativação e desativação enviados à empresa"
      />

      <DataTable
        columns={columns}
        data={items}
        getRowId={(row) => row.id}
      />
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
    </div>
  );
}

function formatStatus(status: string) {
  switch (status) {
    case 'pending':
      return 'Pendente';
    case 'approved':
      return 'Aprovada';
    case 'rejected':
      return 'Rejeitada';
    case 'cancelled':
      return 'Cancelada';
    default:
      return status;
  }
}
