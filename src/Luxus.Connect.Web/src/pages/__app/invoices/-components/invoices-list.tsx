import { useMemo, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { FileText, Upload } from 'lucide-react';

import { useGetV1ProcessingMonths, useGetV1ProviderInvoices } from '@/api';
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
import { Field, FieldLabel } from '@/components/ui/field';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { parseTotalCount } from '@/lib/query-utils';

import { createInvoicesColumns } from './columns';
import { InvoiceImportSheet } from './invoice-import-sheet';

const routeApi = getRouteApi('/__app/invoices/');

const INVOICES_SKELETON_COLUMNS = [
  { header: 'Conta', cell: 'text' as const },
  { header: 'Emissão', cell: 'text' as const },
  { header: 'Vencimento', cell: 'text' as const },
  { header: 'Mês proc.', cell: 'text' as const },
  { header: 'Valor', cell: 'text' as const },
  { header: 'Situação', cell: 'text' as const },
  {
    header: 'Ações',
    headClassName: 'w-24 text-right',
    cell: 'actionsLink' as const
  }
];

const PROCESSING_MONTHS_PAGE_SIZE = 500;

export function InvoicesList() {
  const { page, pageSize, processingMonthId } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [importOpen, setImportOpen] = useState(false);

  const pageIndex = page - 1;

  const processingMonthsQuery = useGetV1ProcessingMonths({
    page_index: 0,
    page_size: PROCESSING_MONTHS_PAGE_SIZE
  });

  const processingMonthLabelById = useMemo(() => {
    const map = new Map<string, string>();
    for (const m of processingMonthsQuery.data?.items ?? []) {
      map.set(m.id, m.display_name);
    }
    return map;
  }, [processingMonthsQuery.data?.items]);

  const listQuery = useGetV1ProviderInvoices({
    page_index: pageIndex,
    page_size: pageSize,
    processing_month_id: processingMonthId
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

  const setProcessingMonthFilter = (value: string) => {
    navigate({
      search: (prev) => ({
        ...prev,
        page: 1,
        processingMonthId: value === '__all__' ? undefined : value
      })
    });
  };

  const columns = useMemo(
    () =>
      createInvoicesColumns({
        page,
        pageSize,
        processingMonthId,
        processingMonthLabelById
      }),
    [page, pageSize, processingMonthId, processingMonthLabelById]
  );

  if (listQuery.isPending || processingMonthsQuery.isPending) {
    return (
      <ListPageSkeleton
        pageSize={pageSize}
        columns={INVOICES_SKELETON_COLUMNS}
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
        title="Faturas importadas"
        description="Faturas de origem por operadora (endpoint /v1/providers/{id}/invoices)"
        action={
          <Button type="button" onClick={() => setImportOpen(true)}>
            <Upload />
            Importar fatura
          </Button>
        }
      />

      <div className="flex max-w-md flex-col gap-2 sm:flex-row sm:items-end sm:gap-4">
        <Field className="min-w-[220px] flex-1">
          <FieldLabel htmlFor="invoices-filter-pm">
            Mês de processamento
          </FieldLabel>
          <Select
            value={processingMonthId ?? '__all__'}
            onValueChange={(value) => {
              if (value == null) {
                return;
              }
              setProcessingMonthFilter(value);
            }}
          >
            <SelectTrigger
              id="invoices-filter-pm"
              className="border-input bg-background w-full rounded-xl border"
            >
              <SelectValue placeholder="Todos" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="__all__">Todos</SelectItem>
                {(processingMonthsQuery.data?.items ?? []).map((m) => (
                  <SelectItem key={m.id} value={m.id}>
                    {m.display_name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
      </div>

      <DataTable
        columns={columns}
        data={items}
        emptyMessage={
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <FileText />
              </EmptyMedia>
              <EmptyTitle>Nenhuma fatura encontrada</EmptyTitle>
              <EmptyDescription>
                Quando houver faturas para esta operadora, elas aparecerão aqui.
                Use &quot;Importar fatura&quot; para registrar um novo arquivo.
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

      <InvoiceImportSheet
        open={importOpen}
        onOpenChange={setImportOpen}
        preferredProcessingMonthId={processingMonthId}
      />
    </div>
  );
}
