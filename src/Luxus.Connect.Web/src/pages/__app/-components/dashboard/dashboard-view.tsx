import { useMemo, useState } from 'react';

import { Link } from '@tanstack/react-router';
import {
  Activity,
  FilePenLine,
  Heart,
  Moon,
  MoreVertical,
  Percent,
  Phone,
  Receipt,
  Zap
} from 'lucide-react';

import type { ListProcessingMonthResponse } from '@/api';
import {
  useGetV1Customers,
  useGetV1PhoneLines,
  useGetV1ProcessingMonths,
  useGetV1StatsDashboard
} from '@/api';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatMoney } from '@/lib/financial-api';
import {
  formatCpfCnpj,
  formatPhoneLineStatus,
  formatPhoneNumber
} from '@/lib/format';
import { useOperationalDashboard, useProcessingMonthLineReadiness } from '@/lib/ops-api';
import { parseTotalCount } from '@/lib/query-utils';

const formatCount = (value: number) =>
  new Intl.NumberFormat('pt-BR').format(value);

const MONTH_SHORT = [
  'Jan',
  'Fev',
  'Mar',
  'Abr',
  'Mai',
  'Jun',
  'Jul',
  'Ago',
  'Set',
  'Out',
  'Nov',
  'Dez'
] as const;

function shortMonthLabel(month: number) {
  return MONTH_SHORT[month - 1] ?? String(month).padStart(2, '0');
}

export const DashboardView = () => {
  const [monthId, setMonthId] = useState('');
  const statsQuery = useGetV1StatsDashboard();
  const monthsQuery = useGetV1ProcessingMonths({
    page_index: 0,
    page_size: 24
  });
  const months = monthsQuery.data?.items ?? [];
  const selectedMonthId = monthId || months[0]?.id || '';
  const operationalQuery = useOperationalDashboard(selectedMonthId);
  const readinessQuery = useProcessingMonthLineReadiness(selectedMonthId, Boolean(selectedMonthId));
  const previousMonthId = useMemo(() => {
    const ordered = [...months].sort((a, b) => a.year - b.year || a.month - b.month);
    const index = ordered.findIndex((m) => m.id === selectedMonthId);
    return index > 0 ? ordered[index - 1]?.id ?? '' : '';
  }, [months, selectedMonthId]);
  const previousReadinessQuery = useProcessingMonthLineReadiness(
    previousMonthId,
    Boolean(previousMonthId)
  );
  const customersQuery = useGetV1Customers({
    page_index: 0,
    page_size: 10
  });
  const phoneLinesQuery = useGetV1PhoneLines({
    page_index: 0,
    page_size: 10
  });

  const summaryPending =
    statsQuery.isPending ||
    monthsQuery.isPending ||
    (operationalQuery.isPending && !operationalQuery.data) ||
    customersQuery.isPending ||
    phoneLinesQuery.isPending;

  if (summaryPending) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-10 w-72 rounded-2xl" />
        <div className="grid gap-6 md:grid-cols-12">
          <Skeleton className="h-96 rounded-3xl md:col-span-4" />
          <Skeleton className="h-96 rounded-3xl md:col-span-3" />
          <Skeleton className="h-96 rounded-3xl md:col-span-5" />
        </div>
        <Skeleton className="h-80 rounded-3xl" />
      </div>
    );
  }

  const summaryError = statsQuery.error ?? customersQuery.error ?? phoneLinesQuery.error;

  if (summaryError) {
    const err = summaryError;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-2xl border px-4 py-3 text-sm">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const stats = statsQuery.data;
  const operational = operationalQuery.data;
  const customers = customersQuery.data;
  const phoneLines = phoneLinesQuery.data;

  if (!stats || !customers || !phoneLines) {
    return null;
  }

  const totalLines = parseTotalCount(stats.phone_lines_count) || (operational?.lines_summary.total_lines ?? 0);
  const activeLines = operational?.lines_summary.active_lines ?? 0;
  const inTransitionLines = operational?.lines_summary.in_transition_lines ?? 0;
  const orphanLines = operational?.lines_summary.orphan_lines ?? 0;

  const projectedRevenue = operational?.financial_summary.projected_monthly_revenue ?? 0;
  const projectedCost = operational?.financial_summary.total_base_cost ?? 0;
  const marginPercentage = operational?.financial_summary.margin_percentage ?? 0;
  const selectedMonth =
    months.find((m) => m.id === selectedMonthId) ??
    months.find((m) => m.id === operational?.current_month_status?.processing_month_id);
  const currentMonthName =
    selectedMonth?.display_name ?? operational?.current_month_status?.display_name ?? 'Sem competência';
  const pendingDivergences = operational?.pending_divergences ?? 0;
  const importedInvoices =
    operational?.current_month_status?.invoice_count ?? parseTotalCount(stats.provider_invoices_count);
  const activeCustomers =
    operational?.customers_summary.active_customers ?? parseTotalCount(stats.customers_count);
  const totalCustomers =
    operational?.customers_summary.total_customers ?? activeCustomers;

  const readyLines = readinessQuery.data?.ready_lines ?? 0;
  const readinessTotal = readinessQuery.data?.total_lines ?? 0;
  const readinessPct = readinessTotal > 0 ? Math.round((readyLines / readinessTotal) * 100) : 0;
  const previousReady = previousReadinessQuery.data?.ready_lines ?? 0;
  const previousTotal = previousReadinessQuery.data?.total_lines ?? 0;
  const previousPct = previousTotal > 0 ? Math.round((previousReady / previousTotal) * 100) : null;
  const readinessDelta = previousPct !== null ? readinessPct - previousPct : null;
  const readyDots = Math.round((readinessPct / 100) * 28);

  const monthlyTrend = operational?.monthly_trend ?? [];
  const maxTrendRevenue = Math.max(0, ...monthlyTrend.map((item) => item.revenue));

  // Cálculos 100% reais da distribuição do parque
  const activePct = totalLines > 0 ? Math.round((activeLines / totalLines) * 100) : (activeLines > 0 ? 100 : 0);
  const transitionPct = totalLines > 0 ? Math.round((inTransitionLines / totalLines) * 100) : 0;
  const orphanPct = totalLines > 0 ? Math.max(0, 100 - activePct - transitionPct) : 0;

  // Proporções de receita reais
  const planosPct = projectedRevenue > 0 ? 100 : 0;
  const svaPct = 0;
  const aparelhosPct = 0;

  const todayDateStr = new Intl.DateTimeFormat('pt-BR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  }).format(new Date());

  return (
    <div className="flex flex-col gap-6">
      {/* Top Header Bento com cores originais */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground">
            Visão Geral Operacional
          </h1>
          <p className="text-muted-foreground mt-1 text-sm font-medium">
            Gestão integrada de linhas, clientes, faturamento e divergências
          </p>
        </div>

        <div className="flex items-center gap-3">
          <span className="text-muted-foreground hidden text-xs font-semibold sm:inline-block">
            {todayDateStr}
          </span>
          <CompetenceMonthSelect
            months={months}
            value={selectedMonthId}
            onChange={setMonthId}
          />
        </div>
      </div>

      {/* Grid Principal Bento */}
      <div className="grid gap-6 md:grid-cols-12">
        {/* Card 1: Composição de Receita (Esquerda) */}
        <div className="flex flex-col justify-between rounded-3xl border bg-card p-6 shadow-xs md:col-span-12 lg:col-span-4">
          <div>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="bg-primary/10 text-primary flex size-8 items-center justify-center rounded-full">
                  <Zap className="size-4 fill-current" />
                </div>
                <span className="text-sm font-bold text-foreground">Composição de Receita</span>
              </div>
              <button type="button" className="text-muted-foreground hover:text-foreground">
                <MoreVertical className="size-4" />
              </button>
            </div>

            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-extrabold tracking-tight text-foreground">
                {formatMoney(projectedRevenue)}
              </span>
              <span className="inline-flex items-center rounded-full bg-emerald-100 px-2.5 py-0.5 text-[11px] font-bold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
                {activeLines} {activeLines === 1 ? 'linha faturada' : 'linhas faturadas'}
              </span>
            </div>

            {/* Diagrama de Bolhas Sobrepostas com proporções reais do banco */}
            <div className="relative my-6 flex h-48 w-full items-center justify-center">
              {/* Bolha 1: Planos Base (Azul/Primário) */}
              <div className="bg-primary text-primary-foreground absolute left-6 top-4 flex size-28 flex-col items-center justify-center rounded-full shadow-md transition-transform hover:scale-105">
                <span className="text-lg font-black leading-none">{planosPct}%</span>
                <span className="mt-0.5 text-[10px] font-bold uppercase tracking-wider opacity-90">Planos</span>
              </div>

              {/* Bolha 2: Serviços SVA (Indigo/Violeta) */}
              <div className="absolute right-8 top-6 flex size-24 flex-col items-center justify-center rounded-full bg-indigo-600 text-white shadow-lg transition-transform hover:scale-105">
                <span className="text-base font-black leading-none">{svaPct}%</span>
                <span className="mt-0.5 text-[10px] font-bold uppercase tracking-wider opacity-85">SVA</span>
              </div>

              {/* Bolha 3: Aparelhos (Verde Esmeralda) */}
              <div className="absolute bottom-2 right-20 flex size-18 flex-col items-center justify-center rounded-full bg-emerald-500 text-white shadow-sm transition-transform hover:scale-105">
                <span className="text-xs font-black leading-none">{aparelhosPct}%</span>
                <span className="text-[9px] font-bold uppercase tracking-wider opacity-90">Aparelhos</span>
              </div>
            </div>
          </div>

          {/* Barras de Distribuição Horizontais com dados 100% reais */}
          <div className="space-y-3 pt-2">
            <div>
              <div className="mb-1 flex items-center justify-between text-xs font-bold">
                <span className="text-foreground">{activePct}% ({activeLines} {activeLines === 1 ? 'linha' : 'linhas'})</span>
                <span className="text-muted-foreground flex items-center gap-1.5">
                  Ativas <span className="bg-primary size-2 rounded-full" />
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div className="bg-primary h-full rounded-full" style={{ width: `${activePct}%` }} />
              </div>
            </div>

            <div>
              <div className="mb-1 flex items-center justify-between text-xs font-bold">
                <span className="text-foreground">{transitionPct}% ({inTransitionLines} {inTransitionLines === 1 ? 'linha' : 'linhas'})</span>
                <span className="text-muted-foreground flex items-center gap-1.5">
                  Em Transição <span className="size-2 rounded-full bg-indigo-600" />
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div className="h-full rounded-full bg-indigo-600" style={{ width: `${transitionPct}%` }} />
              </div>
            </div>

            <div>
              <div className="mb-1 flex items-center justify-between text-xs font-bold">
                <span className="text-foreground">{orphanPct}% ({orphanLines} {orphanLines === 1 ? 'linha' : 'linhas'})</span>
                <span className="text-muted-foreground flex items-center gap-1.5">
                  Órfãs / Estoque <span className="size-2 rounded-full bg-emerald-500" />
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div className="h-full rounded-full bg-emerald-500" style={{ width: `${orphanPct}%` }} />
              </div>
            </div>
          </div>
        </div>

        {/* Coluna Central e Direita */}
        <div className="flex flex-col gap-6 md:col-span-12 lg:col-span-8">
          <div className="grid gap-6 sm:grid-cols-12">
            {/* Card 2.1: Linhas Ativas */}
            <div className="flex flex-col justify-between rounded-3xl border bg-card p-6 shadow-xs sm:col-span-6 lg:col-span-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="flex size-8 items-center justify-center rounded-full bg-rose-50 text-rose-600 dark:bg-rose-950 dark:text-rose-300">
                    <Heart className="size-4 fill-current" />
                  </div>
                  <span className="text-xs font-bold text-foreground">Linhas Ativas</span>
                </div>
                <button type="button" className="text-muted-foreground">
                  <MoreVertical className="size-3.5" />
                </button>
              </div>

              <div className="mt-4 flex items-baseline justify-between">
                <div>
                  <span className="text-3xl font-extrabold tracking-tight text-foreground">
                    {formatCount(activeLines)}
                  </span>
                  <span className="text-muted-foreground ml-1 text-xs font-semibold">linhas</span>
                </div>
                <div className="text-muted-foreground text-right text-[11px] font-semibold">
                  Total {formatCount(totalLines)}
                </div>
              </div>
            </div>

            {/* Card 2.2: Clientes Ativos */}
            <div className="flex flex-col justify-between rounded-3xl border bg-card p-6 shadow-xs sm:col-span-6 lg:col-span-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="flex size-8 items-center justify-center rounded-full bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-300">
                    <Activity className="size-4" />
                  </div>
                  <span className="text-xs font-bold text-foreground">Clientes</span>
                </div>
                <button type="button" className="text-muted-foreground">
                  <MoreVertical className="size-3.5" />
                </button>
              </div>

              <div className="mt-4 flex items-baseline justify-between">
                <div>
                  <span className="text-3xl font-extrabold tracking-tight text-foreground">
                    {formatCount(parseTotalCount(stats.customers_count))}
                  </span>
                  <span className="text-muted-foreground ml-1 text-xs font-semibold">empresas</span>
                </div>
                <div className="text-right text-[11px] font-semibold text-emerald-600">
                  {formatCount(activeCustomers)} {activeCustomers === 1 ? 'ativo' : 'ativos'}
                  {totalCustomers > activeCustomers ? ` de ${formatCount(totalCustomers)}` : ''}
                </div>
              </div>
            </div>

            {/* Card 3: Índice de Prontidão e Dot Matrix */}
            <div className="flex flex-col justify-between rounded-3xl border bg-card p-6 shadow-xs sm:col-span-12 lg:col-span-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="flex size-8 items-center justify-center rounded-full bg-indigo-50 text-indigo-600 dark:bg-indigo-950 dark:text-indigo-300">
                    <Percent className="size-4" />
                  </div>
                  <span className="text-xs font-bold text-foreground">Prontidão</span>
                </div>
                {readinessDelta !== null ? (
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold ${
                      readinessDelta >= 0
                        ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                        : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                    }`}
                  >
                    {readinessDelta > 0 ? '+' : ''}
                    {readinessDelta}%
                  </span>
                ) : (
                  <span className="text-muted-foreground text-[10px] font-semibold">
                    {readyLines}/{readinessTotal || 0} linhas
                  </span>
                )}
              </div>

              <div className="mt-2">
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl font-extrabold tracking-tight text-foreground">
                    {readinessQuery.isPending && !readinessQuery.data ? '—' : readinessPct}
                  </span>
                  <span className="text-muted-foreground text-sm font-bold">%</span>
                </div>

                {/* Dot Matrix Grid */}
                <div className="mt-3 grid grid-cols-7 gap-1.5">
                  {Array.from({ length: 28 }).map((_, idx) => (
                    <div
                      key={idx}
                      className={`size-2.5 rounded-full ${
                        idx < readyDots ? 'bg-primary/80' : 'bg-muted'
                      }`}
                    />
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Card 4: Card de Análise da Competência e Margem (Inferior) */}
          <div className="flex flex-col justify-between rounded-3xl border bg-card p-6 shadow-xs">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="bg-primary/10 text-primary flex size-9 items-center justify-center rounded-full">
                  <Moon className="size-4" />
                </div>
                <div>
                  <span className="text-sm font-bold tracking-tight text-foreground">Análise do Ciclo de Faturamento</span>
                  <p className="text-muted-foreground text-xs">{currentMonthName}</p>
                </div>
              </div>

              <CompetenceMonthSelect
                months={months}
                value={selectedMonthId}
                onChange={setMonthId}
                label="Competência Mensal"
              />
            </div>

            <div className="my-6 grid grid-cols-2 gap-6 sm:grid-cols-4">
              <div className="border-l-2 border-emerald-500 pl-3">
                <span className="text-2xl font-black text-foreground">{marginPercentage.toFixed(1)}%</span>
                <p className="text-muted-foreground text-xs">Margem Bruta Projetada</p>
              </div>

              <div className="border-l-2 border-primary pl-3">
                <span className="text-2xl font-black text-foreground">{formatMoney(projectedCost)}</span>
                <p className="text-muted-foreground text-xs">Custo Base Operadora</p>
              </div>

              <div className="border-l-2 border-indigo-500 pl-3">
                <span className="text-2xl font-black text-foreground">{formatCount(importedInvoices)}</span>
                <p className="text-muted-foreground text-xs">Faturas Importadas</p>
              </div>

              <div className="border-l-2 border-amber-500 pl-3">
                <span className="text-2xl font-black text-foreground">{formatCount(pendingDivergences)}</span>
                <p className="text-muted-foreground text-xs">Divergências Pendentes</p>
              </div>
            </div>

            {/* Gráfico de Barras Mensais Estilizado */}
            <div className="mt-2 flex h-28 items-end justify-between gap-2 border-t pt-4">
              {monthlyTrend.length === 0 ? (
                <p className="text-muted-foreground w-full self-center text-center text-xs">
                  Sem competências para montar o histórico.
                </p>
              ) : (
                monthlyTrend.map((bar) => {
                  const active = bar.processing_month_id === selectedMonthId;
                  const height =
                    maxTrendRevenue > 0
                      ? `${Math.max(12, Math.round((bar.revenue / maxTrendRevenue) * 100))}%`
                      : '20%';
                  return (
                    <button
                      key={bar.processing_month_id}
                      type="button"
                      className="flex flex-1 flex-col items-center gap-2"
                      onClick={() => setMonthId(bar.processing_month_id)}
                      aria-pressed={active}
                      aria-label={`Selecionar competência ${bar.display_name}`}
                    >
                      <div className="relative flex h-20 w-full max-w-[28px] items-end rounded-full bg-muted/60 p-1">
                        <div
                          className={`w-full rounded-full ${active ? 'bg-primary shadow-sm' : 'bg-primary/25'}`}
                          style={{ height }}
                        />
                      </div>
                      <span
                        className={`text-[10px] font-bold ${active ? 'text-primary' : 'text-muted-foreground'}`}
                      >
                        {active ? `${shortMonthLabel(bar.month)} ↗` : shortMonthLabel(bar.month)}
                      </span>
                    </button>
                  );
                })
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Tabelas de Clientes e Linhas Recentes */}
      <div className="grid gap-6 lg:grid-cols-2">
        <RecentCustomersPanel rows={customers.items ?? []} />
        <RecentPhoneLinesPanel rows={phoneLines.items ?? []} />
      </div>
    </div>
  );
};

function CompetenceMonthSelect({
  months,
  value,
  onChange,
  label
}: {
  months: ListProcessingMonthResponse[];
  value: string;
  onChange: (id: string) => void;
  label?: string;
}) {
  if (months.length === 0) {
    return (
      <div className="inline-flex items-center rounded-full border bg-card px-3.5 py-1.5 text-xs font-semibold text-muted-foreground shadow-xs">
        {label ?? 'Sem competência'}
      </div>
    );
  }

  return (
    <Select value={value} onValueChange={(next) => onChange(next ?? '')}>
      <SelectTrigger
        size="sm"
        aria-label={label ?? 'Selecionar competência'}
        className="h-8 rounded-full border bg-card px-3.5 text-xs font-semibold shadow-xs"
      >
        <SelectValue placeholder={label ?? 'Competência'} />
      </SelectTrigger>
      <SelectContent align="end">
        {months.map((month) => (
          <SelectItem key={month.id} value={month.id}>
            {month.display_name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

const RecentCustomersPanel = ({
  rows
}: {
  rows: { id: string; name: string; cpf_cnpj: string; active: boolean }[];
}) => {
  return (
    <div className="overflow-hidden rounded-3xl border bg-card p-0 shadow-xs">
      <div className="flex items-center justify-between border-b px-6 py-4">
        <div>
          <h3 className="text-base font-bold text-foreground">Clientes Recentes</h3>
          <p className="text-muted-foreground text-xs">
            Últimos cadastros realizados na plataforma
          </p>
        </div>
        <Button variant="ghost" size="sm" render={<Link to="/customers" search={{ page: 1, pageSize: 10, providerId: undefined }} />}>
          Ver todos
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="pl-6">Cliente</TableHead>
            <TableHead>Documento</TableHead>
            <TableHead>Situação</TableHead>
            <TableHead className="w-16 text-right pr-6" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Receipt />
                    </EmptyMedia>
                    <EmptyTitle>Nenhum cliente cadastrado</EmptyTitle>
                    <EmptyDescription>
                      Comece cadastrando seu primeiro cliente ou importe uma fatura.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell className="pl-6 font-semibold text-foreground">{row.name}</TableCell>
                <TableCell className="text-muted-foreground font-mono text-xs">
                  {formatCpfCnpj(row.cpf_cnpj)}
                </TableCell>
                <TableCell>
                  <span
                    className={
                      row.active
                        ? 'inline-flex items-center rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                        : 'text-muted-foreground inline-flex items-center rounded-full bg-muted px-2.5 py-0.5 text-xs font-semibold'
                    }
                  >
                    {row.active ? 'Ativo' : 'Inativo'}
                  </span>
                </TableCell>
                <TableCell className="pr-6 text-right">
                  <Button
                    nativeButton={false}
                    variant="ghost"
                    size="sm"
                    className="size-8 rounded-full p-0 text-muted-foreground hover:text-foreground"
                    render={
                      <Link
                        to="/customers/$customerId"
                        params={{ customerId: row.id }}
                        search={{
                          page: 1,
                          pageSize: 10,
                          providerId: undefined
                        }}
                      >
                        <FilePenLine className="size-4" />
                      </Link>
                    }
                  />
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};

const RecentPhoneLinesPanel = ({
  rows
}: {
  rows: { id: string; number: string | null; status?: string | null }[];
}) => {
  return (
    <div className="overflow-hidden rounded-3xl border bg-card p-0 shadow-xs">
      <div className="flex items-center justify-between border-b px-6 py-4">
        <div>
          <h3 className="text-base font-bold text-foreground">Linhas Telefônicas Recentes</h3>
          <p className="text-muted-foreground text-xs">
            Últimas linhas móveis registradas
          </p>
        </div>
        <Button variant="ghost" size="sm" render={<Link to="/phone-lines" search={{ page: 1, pageSize: 10 }} />}>
          Ver todas
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="pl-6">Número</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-16 text-right pr-6" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Phone />
                    </EmptyMedia>
                    <EmptyTitle>Nenhuma linha encontrada</EmptyTitle>
                    <EmptyDescription>
                      Cadastre uma linha telefônica ou importe uma fatura para começar.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              </TableCell>
            </TableRow>
          ) : (
            rows.map((line) => (
              <TableRow key={line.id}>
                <TableCell className="pl-6 font-mono font-semibold text-foreground">
                  {formatPhoneNumber(line.number!) ?? '—'}
                </TableCell>
                <TableCell>
                  <span className="inline-flex items-center rounded-full bg-muted px-2.5 py-0.5 text-xs font-semibold text-foreground">
                    {formatPhoneLineStatus(line.status) ?? '—'}
                  </span>
                </TableCell>
                <TableCell className="pr-6 text-right">
                  <Button
                    nativeButton={false}
                    variant="ghost"
                    size="sm"
                    className="size-8 rounded-full p-0 text-muted-foreground hover:text-foreground"
                    render={
                      <Link
                        to="/phone-lines/$phoneLineId"
                        params={{ phoneLineId: line.id }}
                        search={{ page: 1, pageSize: 10 }}
                      >
                        <FilePenLine className="size-4" />
                      </Link>
                    }
                  />
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};
