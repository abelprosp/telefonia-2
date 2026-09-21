import { useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';

import { useGetV1ProcessingMonths } from '@/api';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
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
import client from '@/lib/client';
import { formatPhoneNumber } from '@/lib/format';

type ConsumptionItem = {
  phone_line_id: string;
  phone_number: string;
  customer_name?: string | null;
  provider_name: string;
  year: number;
  month: number;
  status: string;
  base_cost?: number | null;
  cost_with_consumption?: number | null;
  invoice_total: number;
};

function RouteComponent() {
  const monthsQuery = useGetV1ProcessingMonths({ page_index: 0, page_size: 200 });
  const months = monthsQuery.data?.items ?? [];
  const [monthId, setMonthId] = useState('');
  const selected = monthId || months[0]?.id || '';
  const [items, setItems] = useState<ConsumptionItem[]>([]);
  const [totalAmount, setTotalAmount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const queryString = useMemo(() => {
    const p = new URLSearchParams({ page_index: '0', page_size: '200' });
    if (selected) p.set('processing_month_id', selected);
    return p.toString();
  }, [selected]);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await client<{ items: ConsumptionItem[]; total_amount: number }>({
        url: `/v1/reports/line-consumption?${queryString}`,
        method: 'GET'
      });
      setItems(data.items ?? []);
      setTotalAmount(data.total_amount ?? 0);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Falha ao carregar relatório');
    } finally {
      setLoading(false);
    }
  };

  const exportFmt = async (format: string) => {
    const p = new URLSearchParams(queryString);
    p.set('format', format);
    const res = await client<Blob>({
      url: `/v1/reports/line-consumption/export?${p}`,
      method: 'GET',
      responseType: 'blob'
    });
    const blob = new Blob([res.data as BlobPart], {
      type: String((res.headers as Record<string, string>)?.['content-type'] ?? 'application/octet-stream')
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `consumo-linhas.${format === 'xlsx' ? 'xlsx' : format}`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Relatórios' },
        { label: 'Consumo de linhas' }
      ]}
    >
      <div className="flex flex-col gap-4">
        <div>
          <h1 className="text-xl font-semibold">Relatório de linhas e consumo</h1>
          <p className="text-muted-foreground text-sm">
            Dados derivados das faturas importadas (custo base e com consumo). Exportação PDF, TXT e
            Excel com os mesmos filtros.
          </p>
        </div>
        <div className="flex flex-wrap items-end gap-3">
          <div className="space-y-1">
            <Label>Mês de processamento</Label>
            <Select value={selected} onValueChange={(v) => setMonthId(v ?? '')}>
              <SelectTrigger className="w-72">
                <SelectValue placeholder="Selecione" />
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
          <Button type="button" onClick={() => void load()} disabled={loading || !selected}>
            {loading ? 'Carregando…' : 'Consultar'}
          </Button>
          <Button type="button" variant="outline" disabled={!items.length} onClick={() => void exportFmt('pdf')}>
            PDF
          </Button>
          <Button type="button" variant="outline" disabled={!items.length} onClick={() => void exportFmt('txt')}>
            TXT
          </Button>
          <Button type="button" variant="outline" disabled={!items.length} onClick={() => void exportFmt('xlsx')}>
            Excel
          </Button>
        </div>
        {error ? <p className="text-destructive text-sm">{error}</p> : null}
        <p className="text-sm">Total (custo linhas): {totalAmount.toFixed(2)}</p>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Linha</TableHead>
              <TableHead>Cliente</TableHead>
              <TableHead>Operadora</TableHead>
              <TableHead>Competência</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Base</TableHead>
              <TableHead>Com consumo</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((it) => (
              <TableRow key={`${it.phone_line_id}-${it.year}-${it.month}`}>
                <TableCell className="font-mono">
                  {formatPhoneNumber(it.phone_number) ?? it.phone_number}
                </TableCell>
                <TableCell>{it.customer_name ?? '—'}</TableCell>
                <TableCell>{it.provider_name}</TableCell>
                <TableCell>
                  {String(it.month).padStart(2, '0')}/{it.year}
                </TableCell>
                <TableCell>{it.status}</TableCell>
                <TableCell>{it.base_cost?.toFixed(2) ?? '—'}</TableCell>
                <TableCell>{it.cost_with_consumption?.toFixed(2) ?? '—'}</TableCell>
              </TableRow>
            ))}
            {!items.length ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground">
                  Nenhum registro. Selecione o mês e consulte.
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </div>
    </PageWrapper>
  );
}

export const Route = createFileRoute('/__app/reports/line-consumption/')({
  component: RouteComponent
});
