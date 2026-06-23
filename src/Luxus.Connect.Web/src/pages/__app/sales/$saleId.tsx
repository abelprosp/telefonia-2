import { createFileRoute, Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatLineItemType,
  formatMoney,
  formatSaleStatus,
  useCancelSale,
  useConfirmSale,
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
          </div>
          <div className="flex gap-2">
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
    </PageWrapper>
  );
}
