import { useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatLineItemType,
  formatMoney,
  formatSaleStatus,
  useCancelSale,
  useConfirmSale,
  useMarkSalePaid,
  useSale
} from '@/lib/sales-api';

export const Route = createFileRoute('/__app/sales/$saleId')({
  component: SaleDetailPage
});

function SaleDetailPage() {
  const { saleId } = Route.useParams();
  const saleQuery = useSale(saleId);
  const confirmMutation = useConfirmSale();
  const cancelMutation = useCancelSale();
  const markPaidMutation = useMarkSalePaid();
  const [payOpen, setPayOpen] = useState(false);
  const [payMethod, setPayMethod] = useState('cash');
  const [payNotes, setPayNotes] = useState('');

  if (saleQuery.isLoading) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Vendas', to: '/sales' }, { label: '…' }]}>
        <p className="text-muted-foreground p-6">Carregando…</p>
      </PageWrapper>
    );
  }

  const sale = saleQuery.data;
  if (!sale) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Vendas', to: '/sales' }]}>
        <p className="p-6">Venda não encontrada.</p>
      </PageWrapper>
    );
  }

  const handleConfirm = () => {
    confirmMutation.mutate(saleId, {
      onSuccess: () => {
        toast.success('Venda confirmada.');
        void saleQuery.refetch();
      },
      onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
    });
  };

  const handleCancel = () => {
    cancelMutation.mutate(saleId, {
      onSuccess: () => {
        toast.success('Venda cancelada.');
        void saleQuery.refetch();
      },
      onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
    });
  };

  const canMarkPaid = sale.status === 'draft' || sale.status === 'confirmed';

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Vendas', to: '/sales' },
        { label: sale.sale_number }
      ]}
    >
      <div className="flex flex-col gap-6 p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">{sale.sale_number}</h1>
            <p className="text-muted-foreground">
              {sale.customer_name} · {formatSaleStatus(sale.status)} · {formatMoney(sale.total_amount)}
            </p>
            {sale.paid_at ? (
              <p className="text-muted-foreground mt-1 text-sm">
                Paga em {new Date(sale.paid_at).toLocaleDateString('pt-BR')}
                {sale.payment_method
                  ? ` · ${sale.payment_method === 'cash' ? 'Dinheiro' : sale.payment_method}`
                  : ''}
              </p>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-2">
            {sale.status === 'draft' && (
              <>
                <Button onClick={handleConfirm} disabled={confirmMutation.isPending || sale.items.length === 0}>
                  Confirmar venda
                </Button>
                <Button variant="outline" onClick={handleCancel} disabled={cancelMutation.isPending}>
                  Cancelar
                </Button>
              </>
            )}
            {canMarkPaid ? (
              <Button variant="secondary" onClick={() => setPayOpen(true)} disabled={markPaidMutation.isPending}>
                Dar baixa (pagamento presencial)
              </Button>
            ) : null}
            {sale.status === 'confirmed' && (
              <Button variant="outline" onClick={handleCancel} disabled={cancelMutation.isPending}>
                Cancelar venda
              </Button>
            )}
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Itens</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              {sale.items.map((item) => (
                <li key={item.id} className="flex justify-between gap-4 border-b pb-2">
                  <span>
                    <strong>{formatLineItemType(item.line_item_type)}</strong> — {item.description} (×
                    {item.quantity})
                  </span>
                  <span>{formatMoney(item.total_price)}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        {sale.contract?.rendered_html && (
          <Card>
            <CardHeader>
              <CardTitle>Contrato gerado</CardTitle>
            </CardHeader>
            <CardContent>
              <div
                className="prose prose-sm max-w-none rounded-md border bg-white p-6 dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: sale.contract.rendered_html }}
              />
            </CardContent>
          </Card>
        )}

        {sale.contract_template_name && !sale.contract?.rendered_html && sale.status === 'confirmed' && (
          <p className="text-muted-foreground text-sm">
            Contrato com status: {sale.contract?.status ?? 'pendente'}.
          </p>
        )}

        <Link
          to="/sales"
          search={{ page: 1, pageSize: 10 }}
          className="text-primary text-sm underline-offset-4 hover:underline"
        >
          Voltar para vendas
        </Link>
      </div>

      {payOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
          role="presentation"
          onClick={() => setPayOpen(false)}
        >
          <div
            className="bg-background border-border w-full max-w-md space-y-4 rounded-lg border p-5 shadow-lg"
            role="dialog"
            aria-modal="true"
            onClick={(e) => e.stopPropagation()}
          >
            <div>
              <h2 className="text-base font-semibold">Registrar pagamento presencial</h2>
              <p className="text-muted-foreground mt-2 text-sm">
                Confirma a baixa desta venda no valor de {formatMoney(sale.total_amount)}? Use quando o
                cliente pagar em dinheiro ou outra forma presencial.
              </p>
            </div>
            <div className="space-y-2">
              <Label>Forma de pagamento</Label>
              <Select value={payMethod} onValueChange={(v) => setPayMethod(v ?? 'cash')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="cash">Dinheiro</SelectItem>
                  <SelectItem value="pix_presencial">PIX presencial</SelectItem>
                  <SelectItem value="card_presencial">Cartão presencial</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Observação (opcional)</Label>
              <Textarea
                value={payNotes}
                onChange={(e) => setPayNotes(e.target.value)}
                placeholder="Ex.: recebido no balcão"
                className="min-h-[72px]"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setPayOpen(false)}>
                Cancelar
              </Button>
              <Button
                type="button"
                disabled={markPaidMutation.isPending}
                onClick={() =>
                  markPaidMutation.mutate(
                    {
                      id: saleId,
                      payment_method: payMethod as 'cash' | 'pix_presencial' | 'card_presencial',
                      ...(payNotes.trim() ? { notes: payNotes.trim() } : {})
                    },
                    {
                      onSuccess: () => {
                        setPayOpen(false);
                        toast.success('Pagamento presencial registrado. Venda marcada como paga.');
                        void saleQuery.refetch();
                      },
                      onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                    }
                  )
                }
              >
                Confirmar baixa
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </PageWrapper>
  );
}
