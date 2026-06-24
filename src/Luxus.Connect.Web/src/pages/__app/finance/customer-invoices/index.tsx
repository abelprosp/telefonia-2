import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { FileStack, Pencil, Plug, RefreshCw } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  type CustomerBillingDocument,
  formatBillingStatus,
  formatSicrediBoletoStatus,
  useCustomerBillingDocuments,
  useRegisterSicrediWebhook,
  useSicrediStatus,
  useSyncSicrediPayments,
  useTestSicrediConnection
} from '@/lib/billing-api';
import { formatMoney } from '@/lib/financial-api';

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
  const syncPaymentsMutation = useSyncSicrediPayments();
  const sicrediStatusQuery = useSicrediStatus();
  const testConnectionMutation = useTestSicrediConnection();
  const registerWebhookMutation = useRegisterSicrediWebhook();
  const [publicApiUrl, setPublicApiUrl] = useState('');

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
      id: 'payment',
      header: 'Pagamento',
      cell: ({ row }) => {
        const paid = Boolean(row.original.sicredi_paid_at) || row.original.sicredi_boleto_status === 'paid';
        return paid ? (
          <Badge className="bg-green-600 text-white hover:bg-green-600">
            Pago
            {row.original.sicredi_paid_at
              ? ` · ${new Date(row.original.sicredi_paid_at).toLocaleDateString('pt-BR')}`
              : ''}
          </Badge>
        ) : row.original.sicredi_nosso_numero ? (
          <Badge variant="secondary">{formatSicrediBoletoStatus(row.original.sicredi_boleto_status)}</Badge>
        ) : (
          '—'
        );
      }
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
        action={
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              disabled={syncPaymentsMutation.isPending}
              onClick={() =>
                syncPaymentsMutation.mutate(7, {
                  onSuccess: (data) =>
                    toast.success(`${data.paid} pagamento(s) sincronizado(s) de ${data.checked} fatura(s).`),
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                })
              }
            >
              <RefreshCw className="mr-2 size-4" />
              Sincronizar pagamentos
            </Button>
            <Button render={<Link to="/finance/customer-invoices/bulk-generate" />}>
              <FileStack className="mr-2 size-4" />
              Gerar em massa
            </Button>
          </div>
        }
      />
      {sicrediStatusQuery.data?.enabled && (
        <div className="dashboard-card space-y-3 p-4 text-sm">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="font-medium">Integração Sicredi</p>
              <p className="text-muted-foreground">
                {sicrediStatusQuery.data.sandbox ? 'Sandbox' : 'Produção'} · Ag.{' '}
                {sicrediStatusQuery.data.cooperativa} · Convênio{' '}
                {sicrediStatusQuery.data.codigo_beneficiario}
              </p>
              <div className="mt-2 flex flex-wrap gap-2">
                <Badge variant={sicrediStatusQuery.data.connected ? 'default' : 'destructive'}>
                  {sicrediStatusQuery.data.connected ? 'Conectado' : 'Desconectado'}
                </Badge>
                <Badge variant={sicrediStatusQuery.data.webhook_registered ? 'default' : 'outline'}>
                  {sicrediStatusQuery.data.webhook_registered ? 'Webhook ativo' : 'Webhook pendente'}
                </Badge>
              </div>
              {sicrediStatusQuery.data.connection_error && (
                <p className="text-destructive mt-2 text-xs">{sicrediStatusQuery.data.connection_error}</p>
              )}
              {sicrediStatusQuery.data.webhook_url && (
                <p className="text-muted-foreground mt-2 break-all text-xs">
                  Webhook: {sicrediStatusQuery.data.webhook_url}
                </p>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                size="sm"
                variant="outline"
                disabled={testConnectionMutation.isPending}
                onClick={() =>
                  testConnectionMutation.mutate(undefined, {
                    onSuccess: (data) =>
                      data.success ? toast.success(data.message) : toast.error(data.message),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
              >
                <Plug className="mr-2 size-4" />
                {testConnectionMutation.isPending ? 'Testando…' : 'Testar conexão'}
              </Button>
            </div>
          </div>
          {!sicrediStatusQuery.data.webhook_registered && (
            <div className="flex flex-wrap items-end gap-2 border-t pt-3">
              <div className="min-w-[280px] flex-1">
                <label className="text-muted-foreground mb-1 block text-xs">
                  URL pública da API (HTTPS, ex. ngrok) para registrar webhook no Sicredi
                </label>
                <Input
                  placeholder="https://xxxx.ngrok-free.app"
                  value={publicApiUrl || sicrediStatusQuery.data.public_api_url || ''}
                  onChange={(e) => setPublicApiUrl(e.target.value)}
                />
              </div>
              <Button
                size="sm"
                variant="outline"
                disabled={registerWebhookMutation.isPending || !sicrediStatusQuery.data.connected}
                onClick={() => {
                  const url = (publicApiUrl || sicrediStatusQuery.data.public_api_url || '').trim();
                  registerWebhookMutation.mutate(url || undefined, {
                    onSuccess: (data) => toast.success(data.message),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  });
                }}
              >
                {registerWebhookMutation.isPending ? 'Registrando…' : 'Registrar webhook'}
              </Button>
            </div>
          )}
        </div>
      )}
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
