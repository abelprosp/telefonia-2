import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Pencil } from 'lucide-react';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Badge } from '@/components/ui/badge';
import { formatMoney } from '@/lib/financial-api';
import {
  type CustomerBillingDocument,
  formatBillingStatus,
  useCustomerBillingDocuments
} from '@/lib/billing-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10),
  status: z.string().optional()
});

export const Route = createFileRoute('/__app/finance/customer-invoices/')({
  validateSearch: searchSchema,
  component: CustomerInvoicesPage
});

function CustomerInvoicesPage() {
  const { page, pageSize, status } = Route.useSearch();
  const navigate = Route.useNavigate();
  const listQuery = useCustomerBillingDocuments({
    page_index: page - 1,
    page_size: pageSize,
    status: status || undefined
  });

  const columns: ColumnDef<CustomerBillingDocument>[] = [
    { accessorKey: 'invoice_number', header: 'Número' },
    { accessorKey: 'customer_name', header: 'Cliente' },
    {
      accessorKey: 'amount',
      header: 'Valor',
      cell: ({ row }) => formatMoney(row.original.amount)
    },
    {
      accessorKey: 'due_date',
      header: 'Vencimento',
      cell: ({ row }) => new Date(row.original.due_date).toLocaleDateString('pt-BR')
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => (
        <Badge variant="outline">{formatBillingStatus(row.original.status)}</Badge>
      )
    },
    {
      accessorKey: 'recipient_email',
      header: 'E-mail'
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Link
          to="/finance/customer-invoices/$id"
          params={{ id: row.original.id }}
          className="text-primary inline-flex items-center text-sm font-medium hover:underline"
        >
          <Pencil className="mr-1 size-4" />
          Editar / enviar
        </Link>
      )
    }
  ];

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }, { header: 'C', cell: 'text' }]} />;
  }

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Faturas para envio"
        description="Prepare, edite e envie faturas por e-mail aos clientes"
      />
      <DataTable columns={columns} data={listQuery.data?.items ?? []} getRowId={(r) => r.id} />
      <DataTablePagination
        page={page}
        pageSize={pageSize}
        total={total}
        totalPages={totalPages}
        onPageChange={(p) => void navigate({ search: (s) => ({ ...s, page: p }) })}
        onPageSizeChange={(ps) =>
          void navigate({ search: (s) => ({ ...s, page: 1, pageSize: ps }) })
        }
      />
    </div>
  );
}
