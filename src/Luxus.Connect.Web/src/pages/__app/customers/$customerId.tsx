import { createFileRoute, Link } from '@tanstack/react-router';

import { useCustomersControllerGetById } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';

import { CustomerDetailView } from './-components/customer-detail-view';

export const Route = createFileRoute('/__app/customers/$customerId')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    const providerId: string | undefined =
      typeof search.providerId === 'string' && search.providerId.length > 0
        ? search.providerId
        : undefined;
    return { page, pageSize, providerId };
  },
  component: CustomerDetailRoute
});

function CustomerDetailRoute() {
  const { customerId } = Route.useParams();
  const listSearch = Route.useSearch();

  const detailQuery = useCustomersControllerGetById(customerId);

  const c = detailQuery.data;
  const breadcrumbLabel = c?.name ?? customerId;

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        {
          label: 'Clientes',
          to: '/customers',
          search: {
            page: listSearch.page,
            pageSize: listSearch.pageSize,
            providerId: listSearch.providerId
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
            {String(detailQuery.error)}
          </div>
        )}

        {detailQuery.isSuccess && c && (
          <CustomerDetailView customer={c} listSearch={listSearch} />
        )}

        {detailQuery.isSuccess && !c && (
          <div className="flex flex-col items-start gap-4">
            <p className="text-muted-foreground text-sm">
              Cliente não encontrado.
            </p>
            <Button
              nativeButton={false}
              variant="outline"
              size="sm"
              render={
                <Link
                  to="/customers"
                  search={{
                    page: listSearch.page,
                    pageSize: listSearch.pageSize,
                    providerId: listSearch.providerId
                  }}
                />
              }
            >
              Voltar à lista
            </Button>
          </div>
        )}
      </div>
    </PageWrapper>
  );
}
