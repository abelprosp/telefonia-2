import { createFileRoute, Link } from '@tanstack/react-router';

import { useProviderInvoicesControllerGetById } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';

import { InvoiceDetailView } from './-components/invoice-detail-view';

export const Route = createFileRoute('/__app/invoices/$invoiceId')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    const rawPm = search.processingMonthId;
    const processingMonthId: string | undefined =
      typeof rawPm === 'string' && rawPm.trim() !== ''
        ? rawPm.trim()
        : undefined;
    return { page, pageSize, processingMonthId };
  },
  component: InvoiceDetailRoute
});

function InvoiceDetailRoute() {
  const { invoiceId } = Route.useParams();
  const listSearch = Route.useSearch();

  const detailQuery = useProviderInvoicesControllerGetById(invoiceId);

  const shortBreadcrumb = detailQuery.data?.number ?? '';

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Faturamento' },
        {
          label: 'Faturas',
          to: '/invoices',
          search: {
            page: listSearch.page,
            pageSize: listSearch.pageSize,
            processingMonthId: listSearch.processingMonthId
          }
        },
        { label: shortBreadcrumb }
      ]}
    >
      <div className="mx-auto flex w-full flex-col gap-6">
        {detailQuery.isPending && (
          <div className="space-y-3">
            <Skeleton className="h-10 w-full max-w-md" />
            <Skeleton className="h-48 rounded-xl" />
            <Skeleton className="h-48 rounded-xl" />
          </div>
        )}

        {detailQuery.isError && (
          <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-lg border px-4 py-3 text-sm">
            {isApiHttpError(detailQuery.error)
              ? detailQuery.error.message
              : getErrorMessage(detailQuery.error)}
          </div>
        )}

        {detailQuery.isSuccess && detailQuery.data && (
          <InvoiceDetailView
            invoice={detailQuery.data}
            listSearch={listSearch}
          />
        )}

        {detailQuery.isSuccess && !detailQuery.data && (
          <div className="flex flex-col items-start gap-4">
            <p className="text-muted-foreground text-sm">
              Registro não encontrado.
            </p>
            <Button
              nativeButton={false}
              variant="outline"
              size="sm"
              render={<Link to="/invoices" search={listSearch} />}
            >
              Voltar à lista
            </Button>
          </div>
        )}
      </div>
    </PageWrapper>
  );
}
