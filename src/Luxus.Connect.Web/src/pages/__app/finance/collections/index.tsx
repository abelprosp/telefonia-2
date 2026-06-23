import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Mail } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  type OverdueReceivable,
  useOverdueReceivables,
  useSendCollectionReminder
} from '@/lib/billing-api';
import { formatMoney } from '@/lib/financial-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

export const Route = createFileRoute('/__app/finance/collections/')({
  validateSearch: searchSchema,
  component: CollectionsPage
});

function CollectionsPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();
  const listQuery = useOverdueReceivables({ page_index: page - 1, page_size: pageSize });
  const remindMutation = useSendCollectionReminder();

  const sendReminder = async (receivableId: string, level: number) => {
    try {
      const result = await remindMutation.mutateAsync({
        accounts_receivable_id: receivableId,
        reminder_level: level,
        template_code: 'default-collection-reminder'
      });
      toast.success(result.message);
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const columns: ColumnDef<OverdueReceivable>[] = [
    { accessorKey: 'customer_name', header: 'Cliente' },
    {
      accessorKey: 'balance',
      header: 'Em aberto',
      cell: ({ row }) => formatMoney(row.original.balance)
    },
    {
      accessorKey: 'due_date',
      header: 'Vencimento',
      cell: ({ row }) => new Date(row.original.due_date).toLocaleDateString('pt-BR')
    },
    { accessorKey: 'billing_email', header: 'E-mail cobrança' },
    {
      accessorKey: 'reminders_sent',
      header: 'Lembretes'
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="outline"
          disabled={!row.original.billing_email || remindMutation.isPending}
          onClick={() => void sendReminder(row.original.id, row.original.reminders_sent + 1)}
        >
          <Mail className="mr-1 size-4" />
          Enviar cobrança
        </Button>
      )
    }
  ];

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }]} />;
  }

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Inadimplentes"
        description="Clientes com contas a receber vencidas — envie lembretes por e-mail"
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
