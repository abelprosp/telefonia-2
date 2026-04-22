import { useMemo, useState } from 'react';

import { useQueryClient } from '@tanstack/react-query';
import { getRouteApi } from '@tanstack/react-router';
import { Plus, Users } from 'lucide-react';
import { toast } from 'sonner';

import {
  getV1CustomersQueryKey,
  type ListCustomerResponse,
  useDeleteV1CustomersId,
  useGetV1Customers,
  useProvidersControllerGetById
} from '@/api';
import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyContent,
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

import { createCustomersColumns } from './columns';
import { CustomerCreateSheet } from './customer-create-sheet';

const routeApi = getRouteApi('/__app/customers/');

const CUSTOMERS_SKELETON_COLUMNS = [
  { header: 'Nome', cell: 'text' as const },
  { header: 'Vendedor', cell: 'text' as const },
  { header: 'Documento', cell: 'text' as const },
  { header: 'Tipo', cell: 'text' as const },
  { header: 'Situação', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-44 text-right',
    cell: 'actionsPair' as const
  }
];

export function CustomersList() {
  const { page, pageSize, providerId } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const queryClient = useQueryClient();

  const pageIndex = page - 1;

  const providerQuery = useProvidersControllerGetById(providerId ?? '', {
    query: { enabled: !!providerId }
  });

  const listQuery = useGetV1Customers({
    page_index: pageIndex,
    page_size: pageSize,
    ...(providerId ? { provider_id: providerId } : {})
  });

  const deleteMutation = useDeleteV1CustomersId({
    mutation: {
      onSuccess: () => {
        toast.success('Cliente excluído.');
        void queryClient.invalidateQueries({
          queryKey: getV1CustomersQueryKey()
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
  const [createOpen, setCreateOpen] = useState(false);

  const setPage = (next: number) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: Math.min(Math.max(1, next), totalPages),
        providerId: prev.providerId
      })
    });
  };

  const setPageSize = (next: number) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        pageSize: next,
        providerId: prev.providerId
      })
    });
  };

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={CUSTOMERS_SKELETON_COLUMNS}
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
        title="Clientes"
        description={
          providerId
            ? providerQuery.data
              ? `Filtrando pela operadora ${providerQuery.data.name}.`
              : 'Filtrando por operadora selecionada.'
            : 'Gerencie os clientes cadastrados no sistema'
        }
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            Cadastrar cliente
          </Button>
        }
      />

      <CustomersTable
        rows={items}
        page={page}
        pageSize={pageSize}
        providerId={providerId}
        onCreateClick={() => setCreateOpen(true)}
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

      <CustomerCreateSheet
        open={createOpen}
        onOpenChange={setCreateOpen}
        preferredProviderId={providerId}
      />
    </div>
  );
}

function CustomersTable({
  rows,
  page,
  pageSize,
  providerId,
  onCreateClick,
  onDelete,
  deletePending
}: {
  rows: ListCustomerResponse[];
  page: number;
  pageSize: number;
  providerId?: string;
  onCreateClick: () => void;
  onDelete: (id: string) => void;
  deletePending: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [pendingId, setPendingId] = useState<string | null>(null);

  const columns = useMemo(
    () =>
      createCustomersColumns({
        onDeleteClick: (id) => {
          setPendingId(id);
          setOpen(true);
        },
        deletePending,
        listSearch: {
          page,
          pageSize,
          providerId: providerId ?? undefined
        }
      }),
    [deletePending, page, pageSize, providerId]
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
                <Users />
              </EmptyMedia>
              <EmptyTitle>
                Você ainda não possui nenhum cliente cadastrado
              </EmptyTitle>
              <EmptyDescription>
                Clique no botão abaixo para cadastrar um novo cliente.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <div className="flex flex-wrap gap-2 *:mx-auto">
                <Button variant="outline" onClick={onCreateClick}>
                  <Plus /> Cadastrar cliente
                </Button>
              </div>
            </EmptyContent>
          </Empty>
        }
        getRowId={(row) => row.id}
      />

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Excluir cliente</SheetTitle>
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
