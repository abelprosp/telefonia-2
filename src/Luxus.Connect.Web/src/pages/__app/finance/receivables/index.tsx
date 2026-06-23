import { useMemo, useState } from 'react';

import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { FileText, Plus } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';

import { useGetV1Customers, useGetV1ProcessingMonths } from '@/api';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { useCreateBillingDocumentFromReceivable } from '@/lib/billing-api';
import {
  type AccountReceivable,
  formatFinancialStatus,
  formatMoney,
  todayISO,
  useAccountsReceivable,
  useCreateReceivable,
  useRegisterReceivablePayment
} from '@/lib/financial-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10),
  status: z.string().optional()
});

export const Route = createFileRoute('/__app/finance/receivables/')({
  validateSearch: searchSchema,
  component: ReceivablesPage
});

function ReceivablesPage() {
  const { page, pageSize, status } = Route.useSearch();
  const navigate = Route.useNavigate();
  const routerNavigate = useNavigate();
  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState({
    customer_id: '',
    processing_month_id: '',
    description: '',
    issue_date: todayISO(),
    due_date: todayISO(),
    amount: ''
  });

  const customersQuery = useGetV1Customers({ page_index: 0, page_size: 500 });
  const processingMonthsQuery = useGetV1ProcessingMonths({
    page_index: 0,
    page_size: 500
  });
  const listQuery = useAccountsReceivable({
    page_index: page - 1,
    page_size: pageSize,
    ...(status ? { status } : {})
  });

  const createMutation = useCreateReceivable();
  const receiveMutation = useRegisterReceivablePayment();
  const billingDocMutation = useCreateBillingDocumentFromReceivable();

  const columns = useMemo<ColumnDef<AccountReceivable>[]>(
    () => [
      { accessorKey: 'customer_name', header: 'Cliente' },
      { accessorKey: 'description', header: 'Descrição' },
      {
        id: 'processing_month',
        header: 'Mês proc.',
        cell: ({ row }) =>
          row.original.processing_month_id ? (
            <span className="text-sm">{row.original.processing_month_id.slice(0, 8)}…</span>
          ) : (
            '—'
          )
      },
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
        id: 'billing',
        header: 'Fatura e-mail',
        cell: ({ row }) => (
          <Button
            size="sm"
            variant="ghost"
            disabled={billingDocMutation.isPending}
            onClick={() => {
              billingDocMutation.mutate(
                { receivableId: row.original.id },
                {
                  onSuccess: (data) => {
                    toast.success('Fatura criada para envio.');
                    void routerNavigate({
                      to: '/finance/customer-invoices/$id',
                      params: { id: data.id }
                    });
                  },
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              );
            }}
          >
            <FileText className="mr-1 size-4" />
            Gerar
          </Button>
        )
      },
      {
        id: 'receive',
        header: 'Recebimento',
        cell: ({ row }) =>
          row.original.balance > 0 ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                receiveMutation.mutate(
                  {
                    id: row.original.id,
                    amount: row.original.balance,
                    payment_date: todayISO()
                  },
                  {
                    onSuccess: () => toast.success('Recebimento registrado.'),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  }
                );
              }}
            >
              Receber
            </Button>
          ) : (
            '—'
          )
      }
    ],
    [receiveMutation, billingDocMutation, routerNavigate]
  );

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={[{ header: 'A', cell: 'text' }, { header: 'B', cell: 'text' }, { header: 'C', cell: 'text' }]} />;
  }

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Contas a receber"
        description="Cobranças de clientes e receitas da operação"
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
            <SheetTitle>Nova conta a receber</SheetTitle>
          </SheetHeader>
          <div className="space-y-3 px-4">
            <div>
              <Label>Cliente</Label>
              <Select value={form.customer_id} onValueChange={(v) => setForm({ ...form, customer_id: v ?? '' })}>
                <SelectTrigger><SelectValue placeholder="Selecione" /></SelectTrigger>
                <SelectContent>
                  {(customersQuery.data?.items ?? []).map((c) => (
                    <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Mês de processamento (refaturamento)</Label>
              <Select
                value={form.processing_month_id}
                onValueChange={(v) =>
                  setForm({ ...form, processing_month_id: v ?? '' })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Opcional" />
                </SelectTrigger>
                <SelectContent>
                  {(processingMonthsQuery.data?.items ?? []).map((pm) => (
                    <SelectItem key={pm.id} value={pm.id}>
                      {pm.display_name} — {pm.status}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
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
                    customer_id: form.customer_id,
                    description: form.description,
                    issue_date: form.issue_date,
                    due_date: form.due_date,
                    amount: Number(form.amount),
                    ...(form.processing_month_id.trim()
                      ? { processing_month_id: form.processing_month_id }
                      : {})
                  },
                  {
                    onSuccess: () => {
                      toast.success('Conta a receber criada.');
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
