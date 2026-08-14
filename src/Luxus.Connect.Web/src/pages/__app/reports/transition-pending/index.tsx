import { useMemo, useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowDownRight, ArrowUpRight, Clock, Download } from 'lucide-react';
import { toast } from 'sonner';

import { useGetV1ProcessingMonths } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { formatLineClassification, formatPhoneLineStatus, formatPhoneNumber } from '@/lib/format';
import {
  type MovementReportItem,
  useMovementReports
} from '@/lib/reports-api';

const RouteComponent = () => {
  const monthsQuery = useGetV1ProcessingMonths({ page_index: 0, page_size: 200 });
  const months = monthsQuery.data?.items ?? [];
  const [monthId, setMonthId] = useState('');
  const selected = monthId || months[0]?.id || '';
  const reports = useMovementReports(selected);

  const selectedMonthName = months.find((m) => m.id === selected)?.display_name ?? 'Mês';

  const tabs = useMemo(
    () =>
      [
        { key: 'entries', label: 'Entradas', items: reports.data?.entries ?? [] },
        { key: 'exits', label: 'Saídas', items: reports.data?.exits ?? [] },
        {
          key: 'pending',
          label: 'Pendências de ativação',
          items: reports.data?.activation_pending ?? []
        }
      ] as const,
    [reports.data]
  );
  const [tab, setTab] = useState<(typeof tabs)[number]['key']>('entries');
  const current = tabs.find((t) => t.key === tab) ?? tabs[0];

  const exportCsv = () => {
    if (!current.items || current.items.length === 0) {
      toast.error('Nenhum dado para exportar neste relatório.');
      return;
    }
    const headers = ['Linha', 'Status', 'Classificacao', 'Cliente', 'Ultima_Fatura'];
    if (tab === 'pending') {
      headers.push('Ciclos_Pendentes');
    }
    const rows = current.items.map((item) => {
      const line = [
        item.number,
        item.status,
        item.line_classification,
        `"${(item.customer_name ?? '').replace(/"/g, '""')}"`,
        `"${(item.last_invoice_number ?? '').replace(/"/g, '""')}"`
      ];
      if (tab === 'pending') {
        line.push(String(item.pending_cycles ?? 1));
      }
      return line.join(';');
    });

    const csvContent = '\uFEFF' + [headers.join(';'), ...rows].join('\r\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute(
      'download',
      `movimentacao_${current.key}_${selectedMonthName.replace(/[\s/]/g, '_')}.csv`
    );
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    toast.success('Arquivo CSV exportado com sucesso.');
  };

  const entriesCount = reports.data?.entries?.length ?? 0;
  const exitsCount = reports.data?.exits?.length ?? 0;
  const pendingCount = reports.data?.activation_pending?.length ?? 0;

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Relatórios' },
        { label: 'Movimentação de linhas' }
      ]}
    >
      <div className="flex flex-col gap-6">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold">Movimentação de Linhas (§2.3 e §10.3)</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Entradas, saídas e pendências de ativação/portabilidade do mês de processamento.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <div className="w-72">
              <Select value={selected} onValueChange={(v) => setMonthId(v ?? '')}>
                <SelectTrigger>
                  <SelectValue placeholder="Mês de processamento" />
                </SelectTrigger>
                <SelectContent>
                  {months.map((m) => (
                    <SelectItem key={m.id} value={m.id}>
                      {m.display_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button
              type="button"
              variant="outline"
              disabled={reports.isLoading || current.items.length === 0}
              onClick={exportCsv}
            >
              <Download className="mr-1.5 size-4" />
              Exportar CSV
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Card
            className={`cursor-pointer transition-all ${
              tab === 'entries' ? 'ring-2 ring-primary border-transparent' : ''
            }`}
            onClick={() => setTab('entries')}
          >
            <CardContent className="flex items-center justify-between p-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Entradas do Mês
                </p>
                <p className="mt-1 text-2xl font-bold">{entriesCount}</p>
              </div>
              <div className="rounded-full bg-emerald-500/10 p-2.5 text-emerald-600 dark:text-emerald-400">
                <ArrowDownRight className="size-5" />
              </div>
            </CardContent>
          </Card>

          <Card
            className={`cursor-pointer transition-all ${
              tab === 'exits' ? 'ring-2 ring-primary border-transparent' : ''
            }`}
            onClick={() => setTab('exits')}
          >
            <CardContent className="flex items-center justify-between p-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Saídas do Mês
                </p>
                <p className="mt-1 text-2xl font-bold">{exitsCount}</p>
              </div>
              <div className="rounded-full bg-amber-500/10 p-2.5 text-amber-600 dark:text-amber-400">
                <ArrowUpRight className="size-5" />
              </div>
            </CardContent>
          </Card>

          <Card
            className={`cursor-pointer transition-all ${
              tab === 'pending' ? 'ring-2 ring-primary border-transparent' : ''
            }`}
            onClick={() => setTab('pending')}
          >
            <CardContent className="flex items-center justify-between p-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Pendências de Ativação
                </p>
                <p className="mt-1 text-2xl font-bold">{pendingCount}</p>
              </div>
              <div className="rounded-full bg-blue-500/10 p-2.5 text-blue-600 dark:text-blue-400">
                <Clock className="size-5" />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="flex flex-wrap gap-2 border-b border-border pb-3">
          {tabs.map((t) => (
            <button
              key={t.key}
              type="button"
              className={`rounded-md px-3.5 py-1.5 text-sm font-medium transition-colors ${
                tab === t.key
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-muted/60 text-muted-foreground hover:bg-muted'
              }`}
              onClick={() => setTab(t.key)}
            >
              {t.label} ({t.items.length})
            </button>
          ))}
        </div>

        {reports.isLoading ? (
          <p className="text-muted-foreground py-4 text-sm">Carregando relatório…</p>
        ) : (
          <MovementTable items={current.items} showCycles={tab === 'pending'} />
        )}
      </div>
    </PageWrapper>
  );
};

function MovementTable({
  items,
  showCycles
}: {
  items: MovementReportItem[];
  showCycles: boolean;
}) {
  if (items.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-border p-8 text-center text-muted-foreground text-sm">
        Nenhum registro encontrado neste recorte.
      </div>
    );
  }
  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Linha</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Classificação</TableHead>
            <TableHead>Cliente</TableHead>
            <TableHead>Última Fatura</TableHead>
            {showCycles ? <TableHead>Ciclos Pendentes</TableHead> : null}
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => (
            <TableRow key={item.phone_line_id}>
              <TableCell className="font-medium">
                <Link
                  to="/phone-lines/$phoneLineId"
                  params={{ phoneLineId: item.phone_line_id }}
                  search={{ page: 1, pageSize: 20 }}
                  className="text-primary underline-offset-4 hover:underline"
                >
                  {formatPhoneNumber(item.number) ?? item.number}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant="secondary">{formatPhoneLineStatus(item.status)}</Badge>
              </TableCell>
              <TableCell>{formatLineClassification(item.line_classification)}</TableCell>
              <TableCell>{item.customer_name ?? '—'}</TableCell>
              <TableCell className="text-xs text-muted-foreground">
                {item.last_invoice_number ?? '—'}
              </TableCell>
              {showCycles ? (
                <TableCell>
                  <Badge variant={(item.pending_cycles ?? 1) > 1 ? 'warning-light' : 'outline'} size="sm">
                    {item.pending_cycles ?? 1} {item.pending_cycles === 1 ? 'ciclo' : 'ciclos'}
                  </Badge>
                </TableCell>
              ) : null}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

export const Route = createFileRoute('/__app/reports/transition-pending/')({
  component: RouteComponent
});
