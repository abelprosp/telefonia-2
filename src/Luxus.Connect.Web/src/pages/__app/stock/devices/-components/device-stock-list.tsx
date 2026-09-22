import { useCallback, useMemo, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { PackageX, Plus } from 'lucide-react';
import { toast } from 'sonner';

import { ConfirmDialog } from '@/components/confirm-dialog';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { DynamicSearchInput } from '@/components/dynamic-search-input';
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
import { useDeviceStockList, useUpdateDeviceStockItem } from '@/lib/device-stock-api';
import { parseTotalCount } from '@/lib/query-utils';

import { createDeviceStockColumns } from './columns';
import { DeviceStockCreateSheet } from './device-stock-create-sheet';

const routeApi = getRouteApi('/__app/stock/devices/');

const SKELETON_COLUMNS = [
  { header: 'SKU', cell: 'text' as const },
  { header: 'Aparelho', cell: 'text' as const },
  { header: 'IMEI', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const }
];

export function DeviceStockList() {
  const { page, pageSize, q } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);
  const [pendingSoldId, setPendingSoldId] = useState<string | null>(null);
  const markSoldMutation = useUpdateDeviceStockItem();

  const listQuery = useDeviceStockList({
    page_index: page - 1,
    page_size: pageSize,
    status: 'in_stock',
    ...(q ? { q } : {})
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

  const setSearch = (next: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        ...(next ? { q: next } : { q: undefined })
      })
    });
  };

  const handleMarkSold = useCallback(
    (id: string) => {
      markSoldMutation.mutate(
        { id, status: 'sold' },
        {
          onSuccess: () => {
            toast.success('Aparelho marcado como vendido.');
            setPendingSoldId(null);
            void listQuery.refetch();
          },
          onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
        }
      );
    },
    [listQuery, markSoldMutation]
  );

  const columns = useMemo(
    () =>
      createDeviceStockColumns({
        onMarkSold: (item) => setPendingSoldId(item.id)
      }),
    []
  );

  if (listQuery.isPending) {
    return <ListPageSkeleton pageSize={pageSize} columns={SKELETON_COLUMNS} />;
  }

  if (listQuery.isError) {
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-lg border px-4 py-3 text-sm">
        {isApiHttpError(listQuery.error)
          ? listQuery.error.message
          : getErrorMessage(listQuery.error)}
      </div>
    );
  }

  const items = listQuery.data?.items ?? [];

  return (
    <div className="flex flex-col gap-6">
      <ListPageHeader
        title="Estoque de aparelhos"
        description="Aparelhos disponíveis para venda. Cadastre manualmente ou marque como vendido após a comercialização."
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            Cadastrar aparelho
          </Button>
        }
      />

      <DeviceStockCreateSheet
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSuccess={() => void listQuery.refetch()}
      />

      <DynamicSearchInput
        value={q ?? ''}
        onChange={setSearch}
        placeholder="Pesquisar por SKU, marca, modelo ou IMEI…"
        aria-label="Pesquisar aparelhos"
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
              <EmptyTitle>Nenhum aparelho em estoque</EmptyTitle>
              <EmptyDescription>
                Cadastre aparelhos para disponibilizá-los nas vendas.
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

      <ConfirmDialog
        open={Boolean(pendingSoldId)}
        title="Marcar como vendido"
        description="Tem certeza que deseja marcar este aparelho como vendido? Ele sairá do estoque disponível."
        confirmLabel="Confirmar venda"
        destructive
        loading={markSoldMutation.isPending}
        onCancel={() => setPendingSoldId(null)}
        onConfirm={() => {
          if (pendingSoldId) handleMarkSold(pendingSoldId);
        }}
      />
    </div>
  );
}
