import { createFileRoute, Link } from '@tanstack/react-router';

import { useGetV1Providers, useProcessingMonthsControllerGetById } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';

import { ProcessingMonthDetailView } from './-components/processing-month-detail-view';

export const Route = createFileRoute(
  '/__app/processing-months/$processingMonthId'
)({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    return { page, pageSize };
  },
  component: ProcessingMonthDetailRoute
});

function ProcessingMonthDetailRoute() {
  const { processingMonthId } = Route.useParams();
  const listSearch = Route.useSearch();

  const detailQuery = useProcessingMonthsControllerGetById(processingMonthId);

  const providersQuery = useGetV1Providers(
    { page_index: 0, page_size: 500 },
    { query: { enabled: detailQuery.isSuccess } }
  );

  const m = detailQuery.data;
  const providerName =
    m &&
    (providersQuery.data?.items ?? []).find((p) => p.id === m.provider_id)
      ?.name;

  const breadcrumbLabel = m?.display_name ?? processingMonthId;

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Faturamento' },
        {
          label: 'Meses de processamento',
          to: '/processing-months',
          search: {
            page: listSearch.page,
            pageSize: listSearch.pageSize
          }
        },
        { label: breadcrumbLabel }
      ]}
    >
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
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

        {detailQuery.isSuccess && m && (
          <ProcessingMonthDetailView
            month={m}
            listSearch={listSearch}
            providerName={providerName}
          />
        )}

        {detailQuery.isSuccess && !m && (
          <div className="flex flex-col items-start gap-4">
            <p className="text-muted-foreground text-sm">
              Mês de processamento não encontrado.
            </p>
            <Button
              nativeButton={false}
              variant="outline"
              render={<Link to="/processing-months" search={listSearch} />}
            >
              Voltar à lista
            </Button>
          </div>
        )}
      </div>
    </PageWrapper>
  );
}
