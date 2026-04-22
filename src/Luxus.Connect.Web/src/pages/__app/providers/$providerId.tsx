import { useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';

import { useProvidersControllerGetById } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';

import { ProviderDetailView } from './-components/provider-detail-view';

export const Route = createFileRoute('/__app/providers/$providerId')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    return { page, pageSize };
  },
  component: ProviderDetailRoute
});

function ProviderDetailRoute() {
  const { providerId } = Route.useParams();
  const listSearch = Route.useSearch();
  const [openPlanId, setOpenPlanId] = useState<string | null>(null);

  const detailQuery = useProvidersControllerGetById(providerId);

  const op = detailQuery.data;
  const breadcrumbLabel = op?.name ?? providerId;

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        {
          label: 'Operadoras',
          to: '/providers',
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

        {detailQuery.isSuccess && op && (
          <ProviderDetailView
            provider={op}
            providerId={providerId}
            listSearch={listSearch}
            openPlanId={openPlanId}
            onTogglePlan={setOpenPlanId}
          />
        )}
      </div>
    </PageWrapper>
  );
}
