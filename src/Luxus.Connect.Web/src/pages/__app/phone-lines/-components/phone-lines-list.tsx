import { useMemo } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { Phone } from 'lucide-react';

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

import { createPhoneLinesColumns } from './columns';

const routeApi = getRouteApi('/__app/phone-lines/');

const PHONE_LINES_SKELETON_COLUMNS = [
  { header: 'Número', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  { header: 'Classificação', cell: 'text' as const },
  { header: 'Conta operadora', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

export function PhoneLinesList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();

  const pageIndex = page - 1;

  const { data, isPending, isError, error } = useGetV1PhoneLines({
    page_index: pageIndex,
    page_size: pageSize
  });

  const total = data?.total_count ?? 0;
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
      createPhoneLinesColumns({
        listSearch: { page, pageSize }
      }),
    [page, pageSize]
  );

  if (isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={PHONE_LINES_SKELETON_COLUMNS}
      />
    );
  }

  if (isError) {
    const err = error;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-lg border px-4 py-3 text-sm">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const items = data?.items ?? [];

  return (
    <div className="flex flex-col gap-6">
      <ListPageHeader
        title="Linhas telefônicas"
        description="Consulte e abra o detalhe das linhas cadastradas no sistema"
      />

      <DataTable
        columns={columns}
        data={items}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Phone />
              </EmptyMedia>
              <EmptyTitle>Nenhuma linha encontrada</EmptyTitle>
              <EmptyDescription>
                Quando houver linhas vinculadas às contas e clientes, elas
                aparecerão nesta lista.
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
