import { useMemo, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { CalendarDays, Plus } from 'lucide-react';

import { useGetV1ProcessingMonths, useGetV1Providers } from '@/api';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { parseTotalCount } from '@/lib/query-utils';

import { createProcessingMonthsColumns } from './columns';
import { ProcessingMonthCreateSheet } from './processing-month-create-sheet';

const routeApi = getRouteApi('/__app/processing-months/');

const PROVIDER_LIST_PAGE_SIZE = 500;

const SKELETON_COLUMNS = [
  { header: 'Competência', cell: 'text' as const },
  { header: 'Operadora', cell: 'text' as const },
  { header: 'Situação', cell: 'text' as const },
  { header: 'Fechamento', cell: 'text' as const },
  { header: 'Contingência', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

export function ProcessingMonthsList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);

  const pageIndex = page - 1;

  const listQuery = useGetV1ProcessingMonths({
    page_index: pageIndex,
    page_size: pageSize
  });

  const providersQuery = useGetV1Providers({
    page_index: 0,
    page_size: PROVIDER_LIST_PAGE_SIZE
  });

  const providerNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const p of providersQuery.data?.items ?? []) {
      map.set(p.id, p.name);
    }
    return map;
  }, [providersQuery.data?.items]);

  const total = parseTotalCount(listQuery.data?.total_count);
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const setPage = (next: number) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: Math.min(Math.max(1, next), totalPages)
      })
    });
  };

  const setPageSize = (next: number) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        pageSize: next
      })
    });
  };

  const columns = useMemo(
    () =>
      createProcessingMonthsColumns({
        listSearch: { page, pageSize },
        providerNameById
      }),
    [page, pageSize, providerNameById]
  );

  if (listQuery.isPending || providersQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={SKELETON_COLUMNS} />;
  }

  if (listQuery.isError) {
    const err = listQuery.error;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-lg border px-4 py-3 text-sm">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const items = listQuery.data?.items ?? [];

  return (
    <div className="flex flex-col gap-6">
      <ListPageHeader
        title="Meses de processamento"
        description="Abertura, fechamento operacional e fecho em contingência por competência na operadora"
        action={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <Plus />
            Novo mês
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={items}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <CalendarDays />
              </EmptyMedia>
              <EmptyTitle>Nenhum mês cadastrado</EmptyTitle>
              <EmptyDescription>
                Use &quot;Novo mês&quot; para registrar a primeira competência
                de processamento.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        }
        getRowId={(row) => row.id}
      />

      <DataTablePagination
        page={page}
        totalPages={totalPages}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <ProcessingMonthCreateSheet
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  );
}
