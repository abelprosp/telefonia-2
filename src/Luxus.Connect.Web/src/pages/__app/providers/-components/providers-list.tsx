import { useMemo, useState } from 'react';

import { useQueryClient } from '@tanstack/react-query';
import { getRouteApi } from '@tanstack/react-router';
import { Building2, Plus } from 'lucide-react';
import { toast } from 'sonner';

import {
  getV1ProvidersQueryKey,
  type ListProvidersResponse,
  useDeleteV1ProvidersId,
  useGetV1Providers
} from '@/api';
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
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { invalidateDashboardCaches, parseTotalCount } from '@/lib/query-utils';

import { createProvidersColumns } from './columns';
import { ProviderCreateSheet } from './provider-create-sheet';

const routeApi = getRouteApi('/__app/providers/');

const PROVIDERS_SKELETON_COLUMNS = [
  { header: 'Nome', cell: 'text' as const },
  { header: 'Slug', cell: 'text' as const },
  { header: 'Ativa', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-44 text-right',
    cell: 'actionsPair' as const
  }
];

export function ProvidersList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);

  const pageIndex = page - 1;

  const listQuery = useGetV1Providers({
    page_index: pageIndex,
    page_size: pageSize
  });

  const deleteMutation = useDeleteV1ProvidersId({
    mutation: {
      onSuccess: () => {
        toast.success('Operadora excluída.');
        void queryClient.invalidateQueries({
          queryKey: getV1ProvidersQueryKey()
        });
        void invalidateDashboardCaches(queryClient);
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
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

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={PROVIDERS_SKELETON_COLUMNS}
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
        title="Operadoras"
        description="Gerencie as operadoras cadastradas no sistema"
        action={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <Plus />
            Cadastrar operadora
          </Button>
        }
      />

      <ProvidersTable
        rows={items}
        page={page}
        pageSize={pageSize}
        onDelete={(id) => deleteMutation.mutate({ id })}
        deletePending={deleteMutation.isPending}
      />

      <DataTablePagination
        page={page}
        totalPages={totalPages}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <ProviderCreateSheet open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
}

function ProvidersTable({
  rows,
  page,
  pageSize,
  onDelete,
  deletePending
}: {
  rows: ListProvidersResponse[];
  page: number;
  pageSize: number;
  onDelete: (id: string) => void;
  deletePending: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [pendingId, setPendingId] = useState<string | null>(null);

  const columns = useMemo(
    () =>
      createProvidersColumns({
        onDeleteClick: (id) => {
          setPendingId(id);
          setOpen(true);
        },
        deletePending,
        listSearch: { page, pageSize }
      }),
    [deletePending, page, pageSize]
  );

  return (
    <>
      <DataTable
        columns={columns}
        data={rows}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Building2 />
              </EmptyMedia>
              <EmptyTitle>
                Você ainda não possui nenhuma operadora cadastrada
              </EmptyTitle>
              <EmptyDescription>
                Use "Cadastrar operadora" para cadastrar uma nova operadora.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        }
        getRowId={(row) => row.id}
      />

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Excluir operadora</SheetTitle>
            <SheetDescription>
              Esta ação não pode ser desfeita. Confirma a exclusão?
            </SheetDescription>
          </SheetHeader>
          <SheetFooter className="gap-2 sm:justify-end">
            <SheetClose render={<Button variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              variant="destructive"
              disabled={deletePending || !pendingId}
              onClick={() => {
                if (pendingId) {
                  onDelete(pendingId);
                  setOpen(false);
                  setPendingId(null);
                }
              }}
            >
              Excluir
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </>
  );
}
