import { useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { RefreshCw } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatPhoneNumber } from '@/lib/format';
import {
  type PartnerSale,
  firstDayOfMonthISO,
  formatFinancialStatus,
  formatMoney,
  usePartnerCommissionSettings,
  usePartnerSalesAdmin,
  useSyncPartnerSales,
  useUpdateCommissionSettings,
  useUpdatePartnerSaleStatus
} from '@/lib/financial-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10),
  status: z.string().optional()
});

export const Route = createFileRoute('/__app/finance/partner-sales/')({
  validateSearch: searchSchema,
  component: PartnerSalesAdminPage
});

function PartnerSalesAdminPage() {
  const { page, pageSize, status } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [refMonth, setRefMonth] = useState(firstDayOfMonthISO());
  const [commissionPct, setCommissionPct] = useState('10');

  const settingsQuery = usePartnerCommissionSettings();
  const listQuery = usePartnerSalesAdmin({
    page_index: page - 1,
    page_size: pageSize,
    ...(status ? { status } : {})
  });

  const syncMutation = useSyncPartnerSales();
  const statusMutation = useUpdatePartnerSaleStatus();
  const settingsMutation = useUpdateCommissionSettings();

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
        header: 'Mês ref.',
        cell: ({ row }) => new Date(row.original.reference_month).toLocaleDateString('pt-BR', { month: 'short', year: 'numeric' })
      },
      {
        accessorKey: 'gross_amount',
        header: 'Venda bruta',
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
      },
      {
        id: 'actions',
        header: 'Ações',
        cell: ({ row }) => {
          const s = row.original.status;
          return (
            <div className="flex gap-2">
              {s === 'accrued' && (
                <Button
                  size="sm"
                  onClick={() =>
                    statusMutation.mutate(
                      { id: row.original.id, status: 'approved' },
                      {
                        onSuccess: () => toast.success('Venda aprovada. Conta a pagar gerada.'),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  Aprovar
                </Button>
              )}
              {s === 'approved' && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    statusMutation.mutate(
                      { id: row.original.id, status: 'paid' },
                      {
                        onSuccess: () => toast.success('Comissão paga.'),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  Pagar
                </Button>
              )}
            </div>
          );
        }
      }
    ],
    [statusMutation]
  );

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }, { header: 'C', cell: 'text' }]} />;
  }

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Vendas de parceiros"
        description="Controle de comissões sobre linhas vendidas"
        action={
          <Button
            variant="outline"
            onClick={() =>
              syncMutation.mutate(refMonth, {
                onSuccess: (r) =>
                  toast.success(`${r.inserted_count} registro(s) sincronizado(s).`),
                onError: (e) =>
                  toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              })
            }
          >
            <RefreshCw />
            Sincronizar linhas
          </Button>
        }
      />

      <div className="dashboard-card grid gap-4 p-4 md:grid-cols-3">
        <div>
          <Label>Mês de referência (sync)</Label>
          <Input type="date" value={refMonth} onChange={(e) => setRefMonth(e.target.value)} />
        </div>
        <div>
          <Label>Comissão padrão (%)</Label>
          <Input
            type="number"
            step="0.1"
            value={commissionPct}
            onChange={(e) => setCommissionPct(e.target.value)}
            placeholder={String(settingsQuery.data?.default_commission_percent ?? 10)}
          />
        </div>
        <div className="flex items-end">
          <Button
            variant="secondary"
            onClick={() =>
              settingsMutation.mutate(Number(commissionPct), {
                onSuccess: () => toast.success('Comissão padrão atualizada.'),
                onError: (e) =>
                  toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              })
            }
          >
            Salvar comissão
          </Button>
        </div>
      </div>

      <DataTable columns={columns} data={listQuery.data?.items ?? []} getRowId={(r) => r.id} />
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
