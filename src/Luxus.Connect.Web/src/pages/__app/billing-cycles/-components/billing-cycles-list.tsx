import { useMemo, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { Calendar, Plus } from 'lucide-react';

import { useGetV1BillingCycles } from '@/api';
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

import { BillingCycleCreateSheet } from './billing-cycle-create-sheet';
import { createBillingCyclesColumns } from './columns';

const routeApi = getRouteApi('/__app/billing-cycles/');

const BILLING_CYCLES_SKELETON_COLUMNS = [
  { header: 'Código', cell: 'text' as const },
  { header: 'Nome', cell: 'text' as const },
  { header: 'Início', cell: 'text' as const },
  { header: 'Fim', cell: 'text' as const },
  { header: 'Situação', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

export function BillingCyclesList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);

  const pageIndex = page - 1;

  const listQuery = useGetV1BillingCycles({
    page_index: pageIndex,
    page_size: pageSize
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
      createBillingCyclesColumns({
        listSearch: { page, pageSize }
      }),
    [page, pageSize]
  );

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={BILLING_CYCLES_SKELETON_COLUMNS}
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
        title="Ciclos de faturamento"
        description="Cadastre e gerencie os ciclos de faturamento das contas na operadora"
        action={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <Plus />
            Novo ciclo
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
                <Calendar />
              </EmptyMedia>
              <EmptyTitle>Nenhum ciclo cadastrado</EmptyTitle>
              <EmptyDescription>
                Use &quot;Novo ciclo&quot; para cadastrar o primeiro ciclo de
                faturamento.
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

      <BillingCycleCreateSheet open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
}
