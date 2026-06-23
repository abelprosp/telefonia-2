import { useMemo, useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  type AccountPayable,
  formatFinancialStatus,
  formatMoney,
  todayISO,
  useAccountsPayable,
  useCreatePayable,
  useRegisterPayablePayment
} from '@/lib/financial-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10),
  status: z.string().optional()
});

export const Route = createFileRoute('/__app/finance/payables/')({
  validateSearch: searchSchema,
  component: PayablesPage
});

function PayablesPage() {
  const { page, pageSize, status } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState({
    description: '',
    vendor_name: '',
    issue_date: todayISO(),
    due_date: todayISO(),
    amount: ''
  });

  const listQuery = useAccountsPayable({
    page_index: page - 1,
    page_size: pageSize,
    ...(status ? { status } : {})
  });

  const createMutation = useCreatePayable();
  const payMutation = useRegisterPayablePayment();

  const columns = useMemo<ColumnDef<AccountPayable>[]>(
    () => [
      { accessorKey: 'vendor_name', header: 'Fornecedor' },
      { accessorKey: 'description', header: 'Descrição' },
      {
        accessorKey: 'due_date',
        header: 'Vencimento',
        cell: ({ row }) => new Date(row.original.due_date).toLocaleDateString('pt-BR')
      },
      {
        accessorKey: 'amount',
        header: 'Valor',
        cell: ({ row }) => formatMoney(row.original.amount)
      },
      {
        accessorKey: 'balance',
        header: 'Saldo',
        cell: ({ row }) => formatMoney(row.original.balance)
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatFinancialStatus(row.original.status)
      },
      {
        id: 'invoice',
        header: 'Fatura',
        cell: ({ row }) =>
          row.original.provider_invoice_id ? (
            <Link
              to="/invoices/$invoiceId"
              params={{ invoiceId: row.original.provider_invoice_id }}
              search={{ page: 1, pageSize: 10, processingMonthId: undefined }}
              className="text-primary text-sm hover:underline"
            >
              Ver fatura
            </Link>
          ) : (
            '—'
          )
      },
      {
        id: 'pay',
        header: 'Baixa',
        cell: ({ row }) =>
          row.original.balance > 0 ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                payMutation.mutate(
                  {
                    id: row.original.id,
                    amount: row.original.balance,
                    payment_date: todayISO()
                  },
                  {
                    onSuccess: () => toast.success('Pagamento registrado.'),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  }
                );
              }}
            >
              Quitar
            </Button>
          ) : (
            '—'
          )
      }
    ],
    [payMutation]
  );

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }, { header: 'C', cell: 'text' }, { header: 'D', cell: 'text' }]} />;
  }

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Contas a pagar"
        description="Obrigações com operadoras, parceiros e demais fornecedores"
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            Nova conta
          </Button>
        }
      />
      <DataTable columns={columns} data={listQuery.data?.items ?? []} getRowId={(r) => r.id} />
      <DataTablePagination
        page={page}
        pageSize={pageSize}
        total={total}
        totalPages={totalPages}
        onPageChange={(next) => navigate({ search: (p) => ({ ...p, page: next }) })}
        onPageSizeChange={(next) => navigate({ search: (p) => ({ ...p, page: 1, pageSize: next }) })}
      />

      <Sheet open={createOpen} onOpenChange={setCreateOpen}>
        <SheetContent className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Nova conta a pagar</SheetTitle>
          </SheetHeader>
          <div className="space-y-3 px-4">
            <div>
              <Label>Fornecedor</Label>
              <Input value={form.vendor_name} onChange={(e) => setForm({ ...form, vendor_name: e.target.value })} />
            </div>
            <div>
              <Label>Descrição</Label>
              <Input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>Emissão</Label>
                <Input type="date" value={form.issue_date} onChange={(e) => setForm({ ...form, issue_date: e.target.value })} />
              </div>
              <div>
                <Label>Vencimento</Label>
                <Input type="date" value={form.due_date} onChange={(e) => setForm({ ...form, due_date: e.target.value })} />
              </div>
            </div>
            <div>
              <Label>Valor</Label>
              <Input type="number" step="0.01" value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} />
            </div>
          </div>
          <SheetFooter>
            <Button
              onClick={() =>
                createMutation.mutate(
                  {
                    vendor_name: form.vendor_name,
                    description: form.description,
                    issue_date: form.issue_date,
                    due_date: form.due_date,
                    amount: Number(form.amount)
                  },
                  {
                    onSuccess: () => {
                      toast.success('Conta a pagar criada.');
                      setCreateOpen(false);
                    },
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  }
                )
              }
            >
              Salvar
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
