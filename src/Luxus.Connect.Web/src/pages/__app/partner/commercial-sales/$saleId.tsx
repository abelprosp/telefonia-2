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
  usePartnerCancelSale,
  usePartnerCommercialSale,
  usePartnerConfirmSale
} from '@/lib/sales-api';

export const Route = createFileRoute('/__app/partner/commercial-sales/$saleId')({
  component: PartnerSaleDetailPage
});

function PartnerSaleDetailPage() {
  const { saleId } = Route.useParams();
  const saleQuery = usePartnerCommercialSale(saleId);
  const confirmMutation = usePartnerConfirmSale();
  const cancelMutation = usePartnerCancelSale();

  const sale = saleQuery.data;

  if (saleQuery.isLoading || !sale) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Parceiro', to: '/partner' }, { label: 'Vendas', to: '/partner/commercial-sales' }]}>
        <p className="p-6">{saleQuery.isLoading ? 'Carregando…' : 'Venda não encontrada.'}</p>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Parceiro', to: '/partner' },
        { label: 'Vendas', to: '/partner/commercial-sales' },
        { label: sale.sale_number }
      ]}
    >
      <div className="flex flex-col gap-6 p-6">
        <div className="flex justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">{sale.sale_number}</h1>
            <p className="text-muted-foreground">
              {sale.customer_name} · {formatSaleStatus(sale.status)} · {formatMoney(sale.total_amount)}
            </p>
          </div>
          {sale.status === 'draft' && (
            <div className="flex gap-2">
              <Button
                onClick={() =>
                  confirmMutation.mutate(saleId, {
                    onSuccess: () => {
                      toast.success('Venda confirmada.');
                      void saleQuery.refetch();
                    },
                    onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
                disabled={confirmMutation.isPending || sale.items.length === 0}
              >
                Confirmar
              </Button>
              <Button
                variant="outline"
                onClick={() =>
                  cancelMutation.mutate(saleId, {
                    onSuccess: () => void saleQuery.refetch()
                  })
                }
              >
                Cancelar
              </Button>
            </div>
          )}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Itens</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            {sale.items.map((item) => (
              <div key={item.id} className="flex justify-between">
                <span>
                  {formatLineItemType(item.line_item_type)} — {item.description}
                </span>
                <span>{formatMoney(item.total_price)}</span>
              </div>
            ))}
          </CardContent>
        </Card>

        {sale.contract?.rendered_html && (
          <Card>
            <CardHeader>
              <CardTitle>Contrato</CardTitle>
            </CardHeader>
            <CardContent>
              <div
                className="prose prose-sm max-w-none rounded border bg-white p-4"
                dangerouslySetInnerHTML={{ __html: sale.contract.rendered_html }}
              />
            </CardContent>
          </Card>
        )}

        <Link
          to="/partner/commercial-sales"
          search={{ page: 1, pageSize: 10 }}
          className="text-primary text-sm underline-offset-4 hover:underline"
        >
          Voltar
        </Link>
      </div>
    </PageWrapper>
  );
}
