import { useMemo } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { PackageX } from 'lucide-react';

import { useGetV1PhoneLines } from '@/api';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { parseTotalCount } from '@/lib/query-utils';

import { createStockLinesColumns } from './columns';

const routeApi = getRouteApi('/__app/stock/');

const STOCK_LINES_SKELETON_COLUMNS = [
  { header: 'Número', cell: 'text' as const },
  { header: 'Classificação', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

export function StockLinesList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();

  const pageIndex = page - 1;

  const listQuery = useGetV1PhoneLines({
    page_index: pageIndex,
    page_size: pageSize,
    status: 'IN_STOCK'
  });

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
      createStockLinesColumns({
        listSearch: { page, pageSize }
      }),
    [page, pageSize]
  );

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={STOCK_LINES_SKELETON_COLUMNS}
      />
    );
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
        title="Estoque de linhas"
        description="Linhas com situação em estoque (sem vínculo ativo em fatura no período ou conforme regras da operadora)"
      />

      <DataTable
        columns={columns}
        data={items}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <PackageX />
              </EmptyMedia>
              <EmptyTitle>Nenhuma linha em estoque</EmptyTitle>
              <EmptyDescription>
                Quando houver linhas nesse estado, elas aparecerão nesta lista.
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
    </div>
  );
}
