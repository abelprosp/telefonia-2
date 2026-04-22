import { createFileRoute, Link } from '@tanstack/react-router';

import { useBillingCyclesControllerGetById } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';

import { BillingCycleDetailView } from './-components/billing-cycle-detail-view';

export const Route = createFileRoute('/__app/billing-cycles/$cycleId')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    return { page, pageSize };
  },
  component: BillingCycleDetailRoute
});

function BillingCycleDetailRoute() {
  const { cycleId } = Route.useParams();
  const listSearch = Route.useSearch();

  const detailQuery = useBillingCyclesControllerGetById(cycleId);

  const c = detailQuery.data;
  const breadcrumbLabel = c?.name ?? c?.code ?? cycleId;

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Faturamento' },
        {
          label: 'Ciclos de faturamento',
          to: '/billing-cycles',
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

        {detailQuery.isSuccess && c && (
          <BillingCycleDetailView cycle={c} listSearch={listSearch} />
        )}

        {detailQuery.isSuccess && !c && (
          <div className="flex flex-col items-start">
            <p className="text-muted-foreground text-sm">
              Ciclo de faturamento não encontrado.
            </p>
            <Button
              nativeButton={false}
              variant="outline"
              render={<Link to="/billing-cycles" search={listSearch} />}
            >
              Voltar à lista
            </Button>
          </div>
        )}
      </div>
    </PageWrapper>
  );
}
