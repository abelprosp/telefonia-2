import { useEffect, useMemo, useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { CheckCircle2, Download, FileStack, XCircle } from 'lucide-react';
import { toast } from 'sonner';

import { useGetV1ProcessingMonths } from '@/api';
import { DataTable } from '@/components/data-table';
import { ListPageHeader } from '@/components/list-page';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  type BulkBillingPreviewItem,
  type BulkGenerateResult,
  useBulkBillingPreview,
  useBulkGenerateBillingDocuments,
  useManualBillingPreview,
  useManualGenerateBillingDocuments
} from '@/lib/billing-api';
import { formatMoney, todayISO } from '@/lib/financial-api';
import { downloadFinancialExport } from '@/lib/fidelity-api';

const EMPTY_PREVIEW_ITEMS: BulkBillingPreviewItem[] = [];

export const Route = createFileRoute('/__app/finance/customer-invoices/bulk-generate')({
  component: BulkGenerateInvoicesPage
});

type GenerateMode = 'manual' | 'refaturamento';

function previewRowId(row: BulkBillingPreviewItem) {
  return row.billing_group_id || row.customer_id;
}

function groupLabel(row: BulkBillingPreviewItem) {
  if (row.group_label) return row.group_label;
  if (row.phone_line_number) return `Linha ${row.phone_line_number}`;
  return 'Cliente';
}
function skipReasonLabel(reason?: string) {
  const map: Record<string, string> = {
    no_billing_email: 'Sem e-mail de cobrança',
    no_monthly_amount: 'Sem valor mensal',
    no_lines_on_invoice: 'Linha não está na fatura importada',
    no_provider_invoice: 'Nenhuma fatura da operadora neste mês',
    no_active_lines: 'Sem linhas ou aparelhos ativos',
    already_billed: 'Já faturado neste mês',
    not_billing_ready: 'Pendente de liberação para faturamento'
  };
  return reason ? (map[reason] ?? reason) : '—';
}

function BulkGenerateInvoicesPage() {
  const [mode, setMode] = useState<GenerateMode>('manual');
  const [processingMonthId, setProcessingMonthId] = useState('');
  const [issueDate, setIssueDate] = useState(todayISO());
  const [dueDate, setDueDate] = useState(todayISO());
  const [description, setDescription] = useState('');
  const [result, setResult] = useState<BulkGenerateResult | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const monthsQuery = useGetV1ProcessingMonths({ page_index: 0, page_size: 500 });
  const refaturamentoPreviewQuery = useBulkBillingPreview(
    mode === 'refaturamento' ? processingMonthId : ''
  );
  const manualPreviewQuery = useManualBillingPreview(undefined, mode === 'manual');
  const bulkGenerateMutation = useBulkGenerateBillingDocuments();
  const manualGenerateMutation = useManualGenerateBillingDocuments();

  const preview =
    mode === 'refaturamento' ? refaturamentoPreviewQuery.data : manualPreviewQuery.data;
  const previewItems = preview?.items ?? EMPTY_PREVIEW_ITEMS;
  const previewLoading =
    mode === 'refaturamento' ? refaturamentoPreviewQuery.isPending : manualPreviewQuery.isPending;

  useEffect(() => {
    setSelectedIds(new Set());
    setResult(null);
  }, [mode, processingMonthId]);

  useEffect(() => {
    if (!previewItems.length) {
      setSelectedIds(new Set());
      return;
    }
    setSelectedIds(
      new Set(previewItems.filter((i) => i.eligible).map((i) => previewRowId(i)))
    );
  }, [previewItems]);

  const toggleCustomer = (customerId: string, checked: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (checked) next.add(customerId);
      else next.delete(customerId);
      return next;
    });
  };

  const toggleAllEligible = (checked: boolean) => {
    if (!preview) return;
    if (checked) {
      setSelectedIds(new Set(previewItems.filter((i) => i.eligible).map((i) => previewRowId(i))));
    } else {
      setSelectedIds(new Set());
    }
  };

  const selectedEligibleCount = useMemo(() => {
    if (!preview) return 0;
    return previewItems.filter((i) => i.eligible && selectedIds.has(previewRowId(i))).length;
  }, [previewItems, selectedIds]);

  const columns = useMemo<ColumnDef<BulkBillingPreviewItem>[]>(
    () => [
      {
        id: 'select',
        header: () => (
          <input
            type="checkbox"
            aria-label="Selecionar todos elegíveis"
            checked={
              previewItems.some((i) => i.eligible) &&
              previewItems.filter((i) => i.eligible).every((i) => selectedIds.has(previewRowId(i)))
            }
            onChange={(e) => toggleAllEligible(e.target.checked)}
          />
        ),
        cell: ({ row }) => (
          <input
            type="checkbox"
            aria-label={`Selecionar ${row.original.customer_name}`}
            disabled={!row.original.eligible}
            checked={selectedIds.has(previewRowId(row.original))}
            onChange={(e) => toggleCustomer(previewRowId(row.original), e.target.checked)}
          />
        )
      },
      { accessorKey: 'customer_name', header: 'Cliente' },
      {
        id: 'group',
        header: 'Boleto',
        cell: ({ row }) => groupLabel(row.original)
      },
      { accessorKey: 'billing_email', header: 'E-mail cobrança' },
      {
        accessorKey: 'line_count',
        header: 'Linhas',
        cell: ({ row }) => row.original.line_count
      },
      {
        accessorKey: 'device_count',
        header: 'Aparelhos',
        cell: ({ row }) => row.original.device_count ?? 0
      },
      ...(mode === 'refaturamento'
        ? [
            {
              accessorKey: 'provider_cost',
              header: 'Custo operadora',
              cell: ({ row }: { row: { original: BulkBillingPreviewItem } }) =>
                formatMoney(row.original.provider_cost ?? 0)
            } satisfies ColumnDef<BulkBillingPreviewItem>
          ]
        : []),
      {
        accessorKey: 'monthly_amount',
        header: 'Valor a faturar',
        cell: ({ row }) => formatMoney(row.original.monthly_amount)
      },
      {
        id: 'readiness',
        header: 'Liberação',
        cell: ({ row }) =>
          row.original.is_released_for_billing ? (
            <Badge variant="default">
              {row.original.billing_readiness_label || 'Liberado'}
            </Badge>
          ) : (
            <Badge variant="secondary">
              {row.original.billing_readiness_label || 'Pendente'}
            </Badge>
          )
      },
      {
        id: 'status',
        header: 'Situação',
        cell: ({ row }) =>
          row.original.eligible ? (
            <Badge variant="default">Pronto</Badge>
          ) : (
            <Badge variant="secondary">{skipReasonLabel(row.original.skip_reason)}</Badge>
          )
      }
    ],
    [mode, previewItems, selectedIds]
  );

  const handleGenerate = () => {
    const billingGroupIds = [...selectedIds];
    if (billingGroupIds.length === 0) {
      toast.error('Selecione ao menos um boleto.');
      return;
    }
    if (mode === 'refaturamento' && !processingMonthId) {
      toast.error('Selecione o mês de processamento.');
      return;
    }

    const body = {
      issue_date: issueDate,
      due_date: dueDate,
      billing_group_ids: billingGroupIds,
      ...(description.trim() ? { description: description.trim() } : {})
    };

    const onSuccess = (data: BulkGenerateResult) => {
      setResult(data);
      toast.success(
        `${data.created} fatura(s) criada(s) com boleto, ${data.skipped} ignorada(s), ${data.failed} falha(s).`
      );
    };
    const onError = (e: unknown) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));

    if (mode === 'refaturamento') {
      bulkGenerateMutation.mutate(
        { ...body, processing_month_id: processingMonthId },
        { onSuccess, onError }
      );
    } else {
      manualGenerateMutation.mutate(body, { onSuccess, onError });
    }
  };

  const showTable = mode === 'manual' || Boolean(processingMonthId);
  const isGenerating = bulkGenerateMutation.isPending || manualGenerateMutation.isPending;

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Gerar faturas"
      description="Crie faturas com boleto Sicredi: um boleto por linha Normal e um por grupo Titular+dependentes."
        action={
          <div className="flex flex-wrap gap-2">
            {processingMonthId ? (
              <>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    void downloadFinancialExport(processingMonthId, 'csv').then(
                      () => toast.success('CSV baixado.'),
                      (e: unknown) =>
                        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                    );
                  }}
                >
                  <Download className="mr-2 size-4" />
                  Exportar CSV
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    void downloadFinancialExport(processingMonthId, 'json').then(
                      () => toast.success('JSON baixado.'),
                      (e: unknown) =>
                        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                    );
                  }}
                >
                  <Download className="mr-2 size-4" />
                  Exportar JSON
                </Button>
              </>
            ) : null}
            <Button variant="outline" render={<Link to="/finance/customer-invoices" search={{ page: 1, pageSize: 10 }} />}>
              Voltar à lista
            </Button>
          </div>
        }
      />

      <div className="bg-card grid gap-4 rounded-lg border p-4 md:grid-cols-2 lg:grid-cols-5">
        <div>
          <Label>Modo</Label>
          <Select
            value={mode}
            onValueChange={(v) => {
              setMode((v as GenerateMode) ?? 'manual');
              setResult(null);
            }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="manual">Avulso (sem fatura operadora)</SelectItem>
              <SelectItem value="refaturamento">Refaturamento (fatura importada)</SelectItem>
            </SelectContent>
          </Select>
        </div>
        {mode === 'refaturamento' && (
          <div>
            <Label>Mês de processamento</Label>
            <Select
              value={processingMonthId}
              onValueChange={(v) => {
                setProcessingMonthId(v ?? '');
                setResult(null);
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Selecione" />
              </SelectTrigger>
              <SelectContent>
                {(monthsQuery.data?.items ?? []).map((pm) => (
                  <SelectItem key={pm.id} value={pm.id}>
                    {pm.display_name} — {pm.status}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
        <div>
          <Label>Data de emissão</Label>
          <Input type="date" value={issueDate} onChange={(e) => setIssueDate(e.target.value)} />
        </div>
        <div>
          <Label>Data de vencimento</Label>
          <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
        </div>
        <div>
          <Label>Descrição (opcional)</Label>
          <Input
            placeholder={
              mode === 'refaturamento' && preview?.processing_month_name
                ? `Mensalidade telefonia — ${preview.processing_month_name}`
                : 'Mensalidade telefonia'
            }
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
      </div>

      {showTable && previewLoading && (
        <p className="text-muted-foreground text-sm">Carregando clientes…</p>
      )}

      {showTable && preview && (
        <div className="flex flex-col gap-3">
          {mode === 'refaturamento' && (preview.provider_invoices_count ?? 0) === 0 && (
            <p className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
              Nenhuma fatura da operadora importada para este mês. Use o modo{' '}
              <strong>Avulso</strong> ou importe a fatura em <strong>Faturamento → Faturas</strong>.
            </p>
          )}
          {mode === 'manual' && (
            <p className="text-muted-foreground text-sm">
              Clientes com linhas ou aparelhos ativos. Selecione quem receberá fatura com boleto.
            </p>
          )}
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-sm">
              <span className="font-medium">{previewItems.length}</span> boleto(s) ·{' '}
              <span className="font-medium text-green-600">{selectedEligibleCount}</span>{' '}
              selecionado(s) para gerar
            </p>
            <Button
              onClick={handleGenerate}
              disabled={selectedEligibleCount === 0 || isGenerating}
            >
              <FileStack className="mr-2 size-4" />
              {isGenerating
                ? 'Gerando…'
                : `Gerar ${selectedEligibleCount} fatura(s) + boleto`}
            </Button>
          </div>
          <DataTable columns={columns} data={previewItems} getRowId={(r) => previewRowId(r)} />
        </div>
      )}

      {result && (
        <div className="bg-card rounded-lg border p-4">
          <h3 className="mb-3 font-semibold">Resultado da geração</h3>
          <div className="mb-4 flex flex-wrap gap-4 text-sm">
            <span className="flex items-center gap-1 text-green-600">
              <CheckCircle2 className="size-4" /> {result.created} criada(s)
            </span>
            <span className="text-muted-foreground">{result.skipped} ignorada(s)</span>
            {result.failed > 0 && (
              <span className="flex items-center gap-1 text-destructive">
                <XCircle className="size-4" /> {result.failed} falha(s)
              </span>
            )}
          </div>
          <ul className="max-h-64 space-y-1 overflow-y-auto text-sm">
            {result.items.map((item) => (
              <li key={item.customer_id} className="flex flex-wrap items-center gap-2">
                <span className="font-medium">{item.customer_name}</span>
                <Badge variant={item.status === 'created' ? 'default' : 'secondary'}>
                  {item.status === 'created'
                    ? 'Criada'
                    : item.status === 'skipped'
                      ? 'Ignorada'
                      : 'Falha'}
                </Badge>
                {item.message && <span className="text-muted-foreground">{item.message}</span>}
                {item.document_id && (
                  <Link
                    to="/finance/customer-invoices/$id"
                    params={{ id: item.document_id }}
                    className="text-primary hover:underline"
                  >
                    Ver fatura
                  </Link>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
