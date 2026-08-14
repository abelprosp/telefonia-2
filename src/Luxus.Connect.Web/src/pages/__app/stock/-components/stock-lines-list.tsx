import { useMemo, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { Filter, PackageX, Plus } from 'lucide-react';

import { useGetV1PhoneLines, type ListPhoneLineResponse } from '@/api';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { LinkCustomerLineSheet } from '@/components/link-customer-line-sheet';
import { parseTotalCount } from '@/lib/query-utils';

import { createStockLinesColumns } from './columns';
import { StockLineCreateSheet } from './stock-line-create-sheet';

const routeApi = getRouteApi('/__app/stock/');

const STOCK_LINES_SKELETON_COLUMNS = [
  { header: 'Número', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  { header: 'Operadora', cell: 'text' as const },
  { header: 'CNPJ Luxus', cell: 'text' as const },
  { header: 'Conta', cell: 'text' as const },
  { header: 'Custo Base', cell: 'text' as const },
  { header: 'Com Consumo', cell: 'text' as const },
  { header: 'Última Fatura', cell: 'text' as const },
  { header: 'Classificação', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

export function StockLinesList() {
  const { page, pageSize } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [createOpen, setCreateOpen] = useState(false);
  const [linkLine, setLinkLine] = useState<ListPhoneLineResponse | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('in_stock');

  const pageIndex = page - 1;

  const listQuery = useGetV1PhoneLines({
    page_index: pageIndex,
    page_size: pageSize,
    status: statusFilter === 'all' ? undefined : (statusFilter as any)
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
        listSearch: { page, pageSize },
        onLinkCustomer: (line) => setLinkLine(line)
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
        description="Linhas sob controle sem vínculo ativo com cliente. Entram automaticamente em estoque ao importar faturas da operadora (§3.2 e §4.1)."
        action={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus />
            Cadastrar linha
          </Button>
        }
      />

      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <Filter className="text-muted-foreground size-4" />
          <span className="text-sm font-medium">Filtrar status:</span>
          <Select
            value={statusFilter}
            onValueChange={(val) => {
              if (val) {
                setStatusFilter(val);
                setPage(1);
              }
            }}
          >
            <SelectTrigger className="w-44">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="in_stock">Em estoque</SelectItem>
              <SelectItem value="inactive">Inativa</SelectItem>
              <SelectItem value="all">Todas as linhas</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <StockLineCreateSheet
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSuccess={() => void listQuery.refetch()}
      />

      {linkLine ? (
        <LinkCustomerLineSheet
          mode="line-to-customer"
          phoneLineId={linkLine.id}
          phoneLineNumber={linkLine.number}
          open={Boolean(linkLine)}
          onOpenChange={(open) => {
            if (!open) setLinkLine(null);
          }}
          onSuccess={() => void listQuery.refetch()}
        />
      ) : null}

      <DataTable
        columns={columns}
        data={items}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <PackageX />
              </EmptyMedia>
              <EmptyTitle>Nenhuma linha encontrada</EmptyTitle>
              <EmptyDescription>
                {statusFilter === 'in_stock'
                  ? 'Nenhuma linha disponível em estoque no momento.'
                  : 'Nenhuma linha atende aos critérios selecionados.'}
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
