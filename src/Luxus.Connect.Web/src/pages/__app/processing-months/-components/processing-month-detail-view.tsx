import { useState } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';
import { ExternalLink, Lock, Download, Play, RotateCcw, Calculator } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import type { GetProcessingMonthResponse } from '@/api';
import {
  getV1ProcessingMonthsQueryKey,
  processingMonthsControllerGetByIdQueryKey,
  usePostV1ProcessingMonthsIdClose,
  usePostV1ProcessingMonthsIdCloseContingency
} from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatProcessingMonthStatus } from '@/lib/format';
import { downloadFinancialExport } from '@/lib/fidelity-api';
import {
  useCloseProcessingMonthWithHash,
  useProcessingMonthLineReadiness,
  useProcessingMonthRuns,
  useReopenProcessingMonth,
  useRunProcessingMonthPipeline,
  useSimulateBillingImpact
} from '@/lib/ops-api';
import { formatMoney } from '@/lib/financial-api';

export type ProcessingMonthListSearch = {
  page: number;
  pageSize: number;
};

type ProcessingMonthDetailViewProps = {
  month: GetProcessingMonthResponse;
  listSearch: ProcessingMonthListSearch;
  providerName?: string;
};

function isProcessingMonthOpen(status: string | number | undefined): boolean {
  if (status === undefined || status === null) {
    return false;
  }
  if (typeof status === 'number') {
    return status === 0;
  }
  const s = String(status).trim().toLowerCase();
  return s === 'open' || s === '0';
}

function formatClosedAt(raw: string | null | undefined): string {
  if (!raw) {
    return '—';
  }
  try {
    return format(parseISO(raw), 'dd/MM/yyyy HH:mm');
  } catch {
    return raw;
  }
}

const contingencySchema = z.object({
  justification: z
    .string()
    .min(5, 'Informe uma justificativa (mínimo 5 caracteres)')
    .max(2000, 'Texto muito longo')
});

type ContingencyForm = z.infer<typeof contingencySchema>;

export function ProcessingMonthDetailView({
  month,
  listSearch,
  providerName
}: ProcessingMonthDetailViewProps) {
  const queryClient = useQueryClient();
  const [closeSheetOpen, setCloseSheetOpen] = useState(false);
  const [contingencySheetOpen, setContingencySheetOpen] = useState(false);

  const open = isProcessingMonthOpen(month.status);

  const contingencyForm = useForm<ContingencyForm>({
    resolver: zodResolver(contingencySchema),
    defaultValues: { justification: '' }
  });

  const closeMutation = usePostV1ProcessingMonthsIdClose({
    mutation: {
      onSuccess: async () => {
        toast.success('Pedido de fechamento enviado para aprovação em dois níveis.');
        setCloseSheetOpen(false);
        await queryClient.invalidateQueries({
          queryKey: getV1ProcessingMonthsQueryKey()
        });
        await queryClient.invalidateQueries({
          queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const contingencyMutation = usePostV1ProcessingMonthsIdCloseContingency({
    mutation: {
      onSuccess: async () => {
        toast.success('Mês fechado em contingência.');
        setContingencySheetOpen(false);
        contingencyForm.reset({ justification: '' });
        await queryClient.invalidateQueries({
          queryKey: getV1ProcessingMonthsQueryKey()
        });
        await queryClient.invalidateQueries({
          queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const invoicesSearch = {
    page: 1,
    pageSize: listSearch.pageSize,
    processingMonthId: month.id
  };

  const pipelineQuery = useProcessingMonthRuns(month.id);
  const runPipeline = useRunProcessingMonthPipeline(month.id);
  const simulateImpact = useSimulateBillingImpact(month.id);
  const closeWithHash = useCloseProcessingMonthWithHash(month.id);
  const reopenMonth = useReopenProcessingMonth(month.id);
  const readinessQuery = useProcessingMonthLineReadiness(month.id, true);
  const runs = pipelineQuery.data ?? [];
  const latest = runs[0];
  const previous = runs[1];

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h3 className="text-foreground text-lg font-semibold">
            {month.display_name}
          </h3>
          <p className="text-muted-foreground mt-1 text-sm">
            {providerName ?? month.provider_id}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            nativeButton={false}
            variant="outline"
            size="sm"
            render={
              <Link
                to="/invoices"
                search={invoicesSearch}
                className="inline-flex items-center gap-2"
              >
                <ExternalLink className="size-4" />
                Faturas deste mês
              </Link>
            }
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              void downloadFinancialExport(month.id, 'csv').then(
                () => toast.success('CSV baixado.'),
                (e: unknown) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              );
            }}
          >
            <Download className="size-4" />
            Exportar CSV
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              void downloadFinancialExport(month.id, 'json').then(
                () => toast.success('JSON baixado.'),
                (e: unknown) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              );
            }}
          >
            <Download className="size-4" />
            Exportar JSON
          </Button>
          {open ? (
            <>
              <Button
                type="button"
                variant="default"
                size="sm"
                onClick={() => setCloseSheetOpen(true)}
              >
                <Lock className="size-4" />
                Fechar mês
              </Button>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => setContingencySheetOpen(true)}
              >
                Fecho em contingência
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={simulateImpact.isPending}
                onClick={() =>
                  simulateImpact.mutate(undefined, {
                    onSuccess: () => toast.success('Simulação calculada com dados reais do mês anterior.'),
                    onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
              >
                <Calculator className="size-4" />
                {simulateImpact.isPending ? 'Simulando…' : 'Simular impacto'}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={closeWithHash.isPending}
                onClick={() =>
                  closeWithHash.mutate(undefined, {
                    onSuccess: async (res) => {
                      toast.success(
                        res.status === 'pending_approval'
                          ? 'Fechamento com hash enviado para aprovação em dois níveis.'
                          : `Hash SHA-256: ${res.consolidation_hash || '—'}`
                      );
                      await queryClient.invalidateQueries({
                        queryKey: getV1ProcessingMonthsQueryKey()
                      });
                      await queryClient.invalidateQueries({
                        queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
                      });
                    },
                    onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
              >
                <Lock className="size-4" />
                Fechar com hash
              </Button>
            </>
          ) : (
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={reopenMonth.isPending}
              onClick={() =>
                reopenMonth.mutate(undefined, {
                  onSuccess: async () => {
                    toast.success('Reabertura enviada para aprovação auditada.');
                    await queryClient.invalidateQueries({
                      queryKey: getV1ProcessingMonthsQueryKey()
                    });
                    await queryClient.invalidateQueries({
                      queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
                    });
                  },
                  onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                })
              }
            >
              <RotateCcw className="size-4" />
              Reabrir competência
            </Button>
          )}
        </div>
      </div>

      <Separator />

      <dl className="grid max-w-2xl grid-cols-1 gap-6 sm:grid-cols-2">
        <div>
          <dt className="text-muted-foreground text-sm">Situação</dt>
          <dd className="mt-1 font-medium">
            {formatProcessingMonthStatus(month.status)}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Competência (API)</dt>
          <dd className="mt-1 font-medium tabular-nums">
            {month.month.toString().padStart(2, '0')}/{month.year}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Fechamento</dt>
          <dd className="mt-1">{formatClosedAt(month.closed_at)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Fechado por</dt>
          <dd className="mt-1">{month.closed_by ?? '—'}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Contingência</dt>
          <dd className="mt-1">
            {month.closed_in_contingency ? 'Sim' : 'Não'}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Hash de consolidação</dt>
          <dd className="mt-1 font-mono text-xs break-all">
            {(month as GetProcessingMonthResponse & { consolidation_hash?: string | null })
              .consolidation_hash ?? '—'}
          </dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-sm">
            Justificativa de contingência
          </dt>
          <dd className="mt-1 whitespace-pre-wrap">
            {month.contingency_justification ?? '—'}
          </dd>
        </div>
      </dl>

      {simulateImpact.data ? (
        <section className="rounded-xl border p-4">
          <h3 className="font-semibold">Impacto simulado</h3>
          <p className="text-muted-foreground mt-1 text-sm">
            Receita, custo e margem com dados reais (mês anterior e faturas do mês, sem heurística).
          </p>
          <dl className="mt-3 grid gap-3 sm:grid-cols-3">
            <div>
              <dt className="text-muted-foreground text-xs">Receita projetada</dt>
              <dd className="font-medium">{formatMoney(simulateImpact.data.projected_revenue)}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Custo</dt>
              <dd className="font-medium">{formatMoney(simulateImpact.data.projected_cost)}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Margem</dt>
              <dd className="font-medium">
                {formatMoney(simulateImpact.data.projected_margin)} (
                {simulateImpact.data.margin_percentage.toFixed(1)}%)
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Receita mês anterior</dt>
              <dd className="font-medium">{formatMoney(simulateImpact.data.previous_revenue)}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Delta vs anterior</dt>
              <dd className="font-medium">
                {formatMoney(simulateImpact.data.revenue_delta)} (
                {simulateImpact.data.revenue_delta_percentage.toFixed(1)}%)
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Linhas ativas</dt>
              <dd className="font-medium">{simulateImpact.data.total_active_lines}</dd>
            </div>
          </dl>
        </section>
      ) : null}

      <section className="rounded-xl border p-4">
        <h3 className="font-semibold">Prontidão das linhas</h3>
        {readinessQuery.isPending ? (
          <p className="text-muted-foreground mt-2 text-sm">Carregando…</p>
        ) : readinessQuery.data ? (
          <p className="mt-2 text-sm">
            {readinessQuery.data.ready_lines} prontas · {readinessQuery.data.blocked_lines} bloqueadas
            · {readinessQuery.data.total_lines} no total
          </p>
        ) : (
          <p className="text-muted-foreground mt-2 text-sm">Sem dados de prontidão.</p>
        )}
        {(readinessQuery.data?.items ?? []).filter((i) => !i.is_ready).slice(0, 8).length > 0 ? (
          <ul className="mt-2 space-y-1 text-sm">
            {(readinessQuery.data?.items ?? [])
              .filter((i) => !i.is_ready)
              .slice(0, 8)
              .map((item) => (
                <li key={item.phone_line_id}>
                  {item.phone_number} — {(item.blocking_rules ?? []).join(', ') || 'bloqueada'}
                </li>
              ))}
          </ul>
        ) : null}
      </section>

      <section aria-labelledby="pipeline-title" className="flex flex-col gap-3">
        <a href="#pipeline-title" className="skip-link">
          Pular para o pipeline
        </a>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h3 id="pipeline-title" className="text-foreground font-semibold">
              Pipeline mensal
            </h3>
            <p className="text-muted-foreground text-sm">
              Importar → validar → simular → identificar linhas → estoque vs clientes → vigências →
              composição → pró-rata → dependentes → contas de origem → pendências → prévia →
              auditoria → liberação → consolidar.
            </p>
          </div>
          {open ? (
            <Button
              type="button"
              size="sm"
              disabled={runPipeline.isPending}
              onClick={() =>
                runPipeline.mutate(undefined, {
                  onSuccess: () => toast.success('Pipeline executado.'),
                  onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                })
              }
            >
              <Play className="size-4" aria-hidden />
              {runPipeline.isPending ? 'Executando…' : 'Reprocessar'}
            </Button>
          ) : null}
        </div>
        {latest ? (
          <p className="text-sm">
            Versão {latest.version} · {latest.status}
            {previous ? ` · comparação com versão ${previous.version} (${previous.status})` : ''}
          </p>
        ) : (
          <p className="text-muted-foreground text-sm">Nenhuma execução registrada.</p>
        )}
        <ol className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {(latest?.steps ?? []).map((step) => {
            const prevStep = previous?.steps?.find((s) => s.key === step.key);
            const changed = Boolean(prevStep && prevStep.status !== step.status);
            let skip: string | null = null;
            if (step.summary_json) {
              try {
                skip = (JSON.parse(step.summary_json) as { skip_reason?: string }).skip_reason ?? null;
              } catch {
                skip = null;
              }
            }
            const statusLabel =
              step.status === 'done'
                ? skip
                  ? `Concluído com lacuna (${skip})`
                  : 'Concluído'
                : step.status === 'failed'
                  ? 'Falhou'
                  : step.status === 'running'
                    ? 'Em execução'
                    : step.status;
            return (
              <li
                key={step.key}
                className="rounded-lg border p-3 text-sm"
                aria-label={`${step.label}: ${statusLabel}`}
              >
                <p className="font-medium">{step.label}</p>
                <p>
                  {statusLabel}
                  {typeof step.duration_ms === 'number' ? ` · ${step.duration_ms} ms` : ''}
                </p>
                {changed ? (
                  <p className="text-xs">
                    Mudou em relação à versão anterior ({prevStep?.status})
                  </p>
                ) : null}
                {step.error ? <p className="text-destructive text-xs">{step.error}</p> : null}
              </li>
            );
          })}
        </ol>
      </section>

      <Sheet open={closeSheetOpen} onOpenChange={setCloseSheetOpen}>
        <SheetContent className="flex w-full flex-col sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Fechar mês</SheetTitle>
            <SheetDescription>
              Após o fechamento, operações mutáveis ligadas a este mês ficam
              bloqueadas conforme as regras do sistema. Esta ação não pode ser
              desfeita aqui.
            </SheetDescription>
          </SheetHeader>
          <SheetFooter className="mt-auto gap-2 border-t pt-4 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              type="button"
              disabled={closeMutation.isPending}
              onClick={() => closeMutation.mutate({ id: month.id })}
            >
              {closeMutation.isPending ? 'Fechando…' : 'Confirmar fechamento'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Sheet
        open={contingencySheetOpen}
        onOpenChange={(next) => {
          if (!next) {
            contingencyForm.reset({ justification: '' });
          }
          setContingencySheetOpen(next);
        }}
      >
        <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
          <SheetHeader>
            <SheetTitle>Fechamento em contingência</SheetTitle>
            <SheetDescription>
              Registre a justificativa do fecho administrativo em contingência.
            </SheetDescription>
          </SheetHeader>

          <form
            className="flex min-h-0 flex-1 flex-col gap-4"
            onSubmit={contingencyForm.handleSubmit((values) =>
              contingencyMutation.mutate({
                id: month.id,
                data: { justification: values.justification.trim() }
              })
            )}
          >
            <Controller
              control={contingencyForm.control}
              name="justification"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="pm-contingency-justification">
                    Justificativa
                  </FieldLabel>
                  <Textarea
                    id="pm-contingency-justification"
                    className="border-input bg-background min-h-32 rounded-xl border"
                    placeholder="Descreva o motivo do fecho em contingência…"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <SheetFooter className="mt-auto gap-2 border-t pt-4 sm:justify-end">
              <SheetClose render={<Button type="button" variant="outline" />}>
                Cancelar
              </SheetClose>
              <Button type="submit" disabled={contingencyMutation.isPending}>
                {contingencyMutation.isPending
                  ? 'Enviando…'
                  : 'Confirmar fecho em contingência'}
              </Button>
            </SheetFooter>
          </form>
        </SheetContent>
      </Sheet>
    </div>
  );
}
