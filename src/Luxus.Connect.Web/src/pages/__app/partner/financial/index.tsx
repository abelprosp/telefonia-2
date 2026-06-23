import { useMemo } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Handshake, TrendingUp, Wallet } from 'lucide-react';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatPhoneNumber } from '@/lib/format';
import {
  type PartnerSale,
  formatFinancialStatus,
  formatMoney,
  usePartnerFinancialSummary,
  usePartnerSales
} from '@/lib/financial-api';

import { DashboardMetricCard } from '../../-components/dashboard/dashboard-metric-card';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

export const Route = createFileRoute('/__app/partner/financial/')({
  validateSearch: searchSchema,
  component: PartnerFinancialPage
});

function PartnerFinancialPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();

  const summaryQuery = usePartnerFinancialSummary();
  const salesQuery = usePartnerSales({ page_index: page - 1, page_size: pageSize });

  const columns = useMemo<ColumnDef<PartnerSale>[]>(
    () => [
      { accessorKey: 'customer_name', header: 'Cliente' },
      {
        accessorKey: 'phone_line_number',
        header: 'Linha',
        cell: ({ row }) => formatPhoneNumber(row.original.phone_line_number) ?? row.original.phone_line_number
      },
      {
        accessorKey: 'reference_month',
        header: 'Mês',
        cell: ({ row }) => new Date(row.original.reference_month).toLocaleDateString('pt-BR', { month: 'short', year: 'numeric' })
      },
      {
        accessorKey: 'gross_amount',
        header: 'Venda',
        cell: ({ row }) => formatMoney(row.original.gross_amount)
      },
      {
        accessorKey: 'commission_amount',
        header: 'Comissão',
        cell: ({ row }) => formatMoney(row.original.commission_amount)
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatFinancialStatus(row.original.status)
      }
    ],
    []
  );

  const total = salesQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (summaryQuery.isPending || salesQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }]} />;
  }

  if (summaryQuery.error || salesQuery.error) {
    const err = summaryQuery.error ?? salesQuery.error!;
    return (
      <div className="p-6 text-sm text-destructive">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const s = summaryQuery.data!;

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      <div>
        <h1 className="text-2xl font-semibold">Meu financeiro</h1>
        <p className="text-muted-foreground text-sm">Vendas, comissões e recebíveis da sua carteira</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <DashboardMetricCard title="Vendas brutas" value={formatMoney(s.total_gross_sales)} icon={Wallet} />
        <DashboardMetricCard title="Comissão provisionada" value={formatMoney(s.total_commission_accrued)} icon={Handshake} />
        <DashboardMetricCard title="Comissão aprovada" value={formatMoney(s.total_commission_approved)} icon={TrendingUp} />
        <DashboardMetricCard title="Comissão paga" value={formatMoney(s.total_commission_paid)} icon={Wallet} />
      </div>

      {summaryQuery.isPending ? (
        <Skeleton className="h-20 rounded-xl" />
      ) : (
        <div className="dashboard-card p-4 text-sm">
          Recebíveis em aberto dos seus clientes:{' '}
          <strong>{formatMoney(s.total_receivable_from_sales)}</strong>
        </div>
      )}

      <ListPageHeader
        title="Histórico de vendas"
        description="Comissões calculadas sobre linhas da sua carteira"
      />
      <DataTable columns={columns} data={salesQuery.data?.items ?? []} getRowId={(r) => r.id} />
      <DataTablePagination
        page={page}
        pageSize={pageSize}
        total={total}
        totalPages={totalPages}
        onPageChange={(next) => navigate({ search: (p) => ({ ...p, page: next }) })}
        onPageSizeChange={(next) => navigate({ search: (p) => ({ ...p, page: 1, pageSize: next }) })}
      />
    </div>
  );
}
