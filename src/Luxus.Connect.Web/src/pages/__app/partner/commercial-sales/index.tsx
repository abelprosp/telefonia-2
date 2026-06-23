import { useMemo } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Plus } from 'lucide-react';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import {
  formatMoney,
  formatSaleStatus,
  usePartnerCommercialSales
} from '@/lib/sales-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

const SKELETON_COLUMNS = [
  { header: 'Nº', cell: 'text' as const },
  { header: 'Cliente', cell: 'text' as const },
  { header: 'Total', cell: 'text' as const }
];

export const Route = createFileRoute('/__app/partner/commercial-sales/')({
  validateSearch: searchSchema,
  component: PartnerCommercialSalesPage
});

function PartnerCommercialSalesPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();

  const listQuery = usePartnerCommercialSales({
    page_index: page - 1,
    page_size: pageSize
  });

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const columns = useMemo<ColumnDef<import('@/lib/sales-api').Sale>[]>(
    () => [
      { accessorKey: 'sale_number', header: 'Nº venda' },
      { accessorKey: 'customer_name', header: 'Cliente' },
      {
        accessorKey: 'total_amount',
        header: 'Total',
        cell: ({ row }) => formatMoney(row.original.total_amount)
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatSaleStatus(row.original.status)
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <Link
            to="/partner/commercial-sales/$saleId"
            params={{ saleId: row.original.id }}
            className="text-primary text-sm underline-offset-4 hover:underline"
          >
            Ver
          </Link>
        )
      }
    ],
    []
  );

  if (listQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Parceiro', to: '/partner' }, { label: 'Vendas' }]}>
        <ListPageSkeleton pageSize={pageSize} columns={SKELETON_COLUMNS} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper breadcrumbs={[{ label: 'Parceiro', to: '/partner' }, { label: 'Vendas' }]}>
      <div className="flex flex-col gap-6 p-6">
        <ListPageHeader
          title="Minhas vendas"
          description="Registre e acompanhe vendas dos seus clientes."
          action={
            <Link to="/partner/commercial-sales/new">
              <Button>
                <Plus />
                Nova venda
              </Button>
            </Link>
          }
        />
        <DataTable columns={columns} data={listQuery.data?.items ?? []} getRowId={(r) => r.id} />
        <DataTablePagination
          page={page}
          pageSize={pageSize}
          total={total}
          totalPages={totalPages}
          onPageChange={(next) => navigate({ search: (prev) => ({ ...prev, page: next }) })}
          onPageSizeChange={(next) => navigate({ search: (prev) => ({ ...prev, pageSize: next, page: 1 }) })}
        />
      </div>
    </PageWrapper>
  );
}
