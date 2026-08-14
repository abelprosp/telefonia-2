import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { getRouteApi } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
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
import { getImportRequestStatus } from '@/lib/import-request-api';
import { parseTotalCount } from '@/lib/query-utils';

import { createInvoicesColumns } from './columns';
import {
  InvoiceImportProgressBanner,
  type ImportProgressStage,
  type ImportProgressState
} from './invoice-import-progress-banner';
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
const POLL_INTERVAL_MS = 1500;
const STORAGE_KEY = 'luxus_active_import_request';
const MAX_PERSISTED_AGE_MS = 2 * 60 * 60 * 1000; // 2 hours

type PersistedImportState = {
  importId: string;
  fileName: string;
  stage: ImportProgressStage;
  progress: number;
  timestamp: number;
  errorMessage?: string;
};

const IDLE_PROGRESS: ImportProgressState = {
  stage: 'idle',
  progress: 0,
  fileName: ''
};

function loadPersistedImport(): { state: ImportProgressState; importId: string | null } {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { state: IDLE_PROGRESS, importId: null };
    const data = JSON.parse(raw) as PersistedImportState;
    if (Date.now() - data.timestamp > MAX_PERSISTED_AGE_MS) {
      localStorage.removeItem(STORAGE_KEY);
      return { state: IDLE_PROGRESS, importId: null };
    }
    return {
      state: {
        stage: data.stage,
        progress: data.progress,
        fileName: data.fileName,
        errorMessage: data.errorMessage
      },
      importId: data.importId
    };
  } catch {
    return { state: IDLE_PROGRESS, importId: null };
  }
}

function savePersistedImport(importId: string, state: ImportProgressState) {
  try {
    if (state.stage === 'idle') {
      localStorage.removeItem(STORAGE_KEY);
      return;
    }
    const data: PersistedImportState = {
      importId,
      fileName: state.fileName,
      stage: state.stage,
      progress: state.progress,
      timestamp: Date.now(),
      errorMessage: state.errorMessage
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  } catch {
    // Ignore storage errors
  }
}

function clearPersistedImport() {
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Ignore
  }
}

export function InvoicesList() {
  const { page, pageSize, processingMonthId } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const [importOpen, setImportOpen] = useState(false);

  // Initialize from localStorage so refresh (F5) doesn't lose progress
  const persisted = useMemo(() => loadPersistedImport(), []);
  const [importProgress, setImportProgress] = useState<ImportProgressState>(persisted.state);
  const [pollingImportId, setPollingImportId] = useState<string | null>(persisted.importId);

  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const queryClient = useQueryClient();

  // ─── Invalidate the invoices list (predicate-based to bypass object equality) ───
  const invalidateInvoicesList = useCallback(() => {
    void queryClient.invalidateQueries({
      predicate: (query) => {
        const key = query.queryKey;
        return (
          Array.isArray(key) &&
          key.length > 0 &&
          typeof key[0] === 'object' &&
          key[0] !== null &&
          (key[0] as Record<string, unknown>)['url'] === '/v1/provider-invoices'
        );
      }
    });
  }, [queryClient]);

  // ─── Polling: check import request status every POLL_INTERVAL_MS ───
  const stopPolling = useCallback(() => {
    if (pollingIntervalRef.current !== null) {
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
    }
  }, []);

  const checkStatusOnce = useCallback(
    async (importId: string, fileName: string) => {
      try {
        const result = await getImportRequestStatus(importId);
        if (result.status === 'completed') {
          stopPolling();
          setPollingImportId(null);
          const nextState: ImportProgressState = { stage: 'done', progress: 100, fileName };
          setImportProgress(nextState);
          savePersistedImport(importId, nextState);
          invalidateInvoicesList();
          return true;
        } else if (result.status === 'failed') {
          stopPolling();
          setPollingImportId(null);
          const nextState: ImportProgressState = {
            stage: 'error',
            progress: 0,
            fileName,
            errorMessage: result.error ?? 'Falha no processamento do arquivo.'
          };
          setImportProgress(nextState);
          savePersistedImport(importId, nextState);
          return true;
        } else if (result.status === 'pdf_unparsed') {
          stopPolling();
          setPollingImportId(null);
          const nextState: ImportProgressState = {
            stage: 'error',
            progress: 0,
            fileName,
            errorMessage:
              'PDF recebido, mas o parse ainda não está disponível. Use o TXT da operadora para faturar.'
          };
          setImportProgress(nextState);
          savePersistedImport(importId, nextState);
          return true;
        }
        return false;
      } catch (err: unknown) {
        // If the status endpoint returns 404 (e.g. backend container without the new route),
        // stop polling immediately, mark as completed and refresh the invoices list.
        const is404 =
          (typeof err === 'object' &&
            err !== null &&
            'status' in err &&
            (err as { status: number }).status === 404) ||
          (typeof err === 'object' &&
            err !== null &&
            'response' in err &&
            (err as { response?: { status?: number } }).response?.status === 404);

        if (is404) {
          stopPolling();
          setPollingImportId(null);
          clearPersistedImport();
          invalidateInvoicesList();
          const nextState: ImportProgressState = {
            stage: 'done',
            progress: 100,
            fileName
          };
          setImportProgress(nextState);
          return true;
        }
        return false;
      }
    },
    [stopPolling, invalidateInvoicesList]
  );

  const startPolling = useCallback(
    (importId: string, fileName: string) => {
      stopPolling();
      // Check immediately on start
      void checkStatusOnce(importId, fileName).then((finished) => {
        if (finished) return;
        pollingIntervalRef.current = setInterval(async () => {
          await checkStatusOnce(importId, fileName);
        }, POLL_INTERVAL_MS);
      });
    },
    [stopPolling, checkStatusOnce]
  );

  // Resume or start polling whenever pollingImportId changes
  useEffect(() => {
    if (!pollingImportId) return;
    const fileName = importProgress.fileName;
    startPolling(pollingImportId, fileName);
    return () => stopPolling();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pollingImportId]);

  // Cleanup on unmount
  useEffect(() => () => stopPolling(), [stopPolling]);

  // ─── Called by InvoiceImportSheet on every progress change ───
  const handleProgressChange = useCallback(
    (state: ImportProgressState & { importRequestId?: string }) => {
      setImportProgress(state);
      if (state.stage === 'done') {
        stopPolling();
        setPollingImportId(null);
        clearPersistedImport();
        invalidateInvoicesList();
      } else if (state.stage === 'error') {
        stopPolling();
        setPollingImportId(null);
        clearPersistedImport();
      } else if (state.importRequestId) {
        setPollingImportId(state.importRequestId);
        const processingState: ImportProgressState = {
          stage: 'processing',
          progress: 0,
          fileName: state.fileName
        };
        setImportProgress(processingState);
        savePersistedImport(state.importRequestId, processingState);
      }
    },
    [stopPolling, invalidateInvoicesList]
  );

  const handleDismiss = useCallback(() => {
    stopPolling();
    setPollingImportId(null);
    setImportProgress(IDLE_PROGRESS);
    clearPersistedImport();
  }, [stopPolling]);

  const handleRefresh = useCallback(() => {
    invalidateInvoicesList();
  }, [invalidateInvoicesList]);

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
      search: (prev) => ({ ...prev, page: 1, pageSize: next })
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

  const isImportActive =
    importProgress.stage !== 'idle' &&
    importProgress.stage !== 'done' &&
    importProgress.stage !== 'error';

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
          <Button
            type="button"
            onClick={() => setImportOpen(true)}
            disabled={isImportActive}
          >
            <Upload />
            Importar fatura
          </Button>
        }
      />

      {/* ─── Inline import progress banner (resilient to F5/page refresh) ─── */}
      {importProgress.stage !== 'idle' && (
        <InvoiceImportProgressBanner
          state={importProgress}
          onDismiss={handleDismiss}
          onRefresh={handleRefresh}
        />
      )}

      <div className="flex max-w-md flex-col gap-2 sm:flex-row sm:items-end sm:gap-4">
        <Field className="min-w-[220px] flex-1">
          <FieldLabel htmlFor="invoices-filter-pm">
            Mês de processamento
          </FieldLabel>
          <Select
            value={processingMonthId ?? '__all__'}
            onValueChange={(value) => {
              if (value == null) return;
              setProcessingMonthFilter(value);
            }}
          >
            <SelectTrigger
              id="invoices-filter-pm"
              className="border-input bg-background w-full rounded-xl border"
            >
              <SelectValue placeholder="Todos">
                {!processingMonthId || processingMonthId === '__all__'
                  ? 'Todos'
                  : processingMonthLabelById.get(processingMonthId) ?? 'Mês selecionado'}
              </SelectValue>
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
        onProgressChange={handleProgressChange}
      />
    </div>
  );
}
