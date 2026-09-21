import { useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { ConfirmDialog } from '@/components/confirm-dialog';
import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import client from '@/lib/client';
import { formatPhoneNumber } from '@/lib/format';

type Finding = {
  phone_line_id?: string | null;
  normalized_number: string;
  internal_status?: string | null;
  external_status?: string | null;
  customer_name?: string | null;
  provider_name?: string | null;
  finding_type: string;
  is_inconclusive: boolean;
  evidence?: string;
};

function RouteComponent() {
  const [source, setSource] = useState('axon');
  const [jobId, setJobId] = useState('');
  const [partial, setPartial] = useState(false);
  const [tab, setTab] = useState<'missing' | 'cancelled'>('missing');
  const [items, setItems] = useState<Finding[]>([]);
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [applying, setApplying] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const p = new URLSearchParams({ source });
      if (jobId.trim()) p.set('import_job_id', jobId.trim());
      if (tab === 'missing') {
        p.set('is_partial_source', String(partial));
        const { data } = await client<Finding[]>({
          url: `/v1/reconciliation/missing-in-operator?${p}`,
          method: 'GET'
        });
        setItems(Array.isArray(data) ? data : []);
      } else {
        const { data } = await client<Finding[]>({
          url: `/v1/reconciliation/cancelled-externally-active?${p}`,
          method: 'GET'
        });
        setItems(Array.isArray(data) ? data : []);
      }
      setSelected({});
    } catch (e: unknown) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  };

  const selectedIds = items
    .filter((it) => it.phone_line_id && selected[it.phone_line_id])
    .map((it) => it.phone_line_id!);

  const applyBatch = async () => {
    setApplying(true);
    try {
      const { data } = await client<{ applied: number; skipped: number }>({
        url: '/v1/reconciliation/cancelled-externally-active/apply',
        method: 'POST',
        data: {
          phone_line_ids: selectedIds,
          source,
          import_job_id: jobId.trim() || undefined,
          confirm: true
        }
      });
      toast.success(
        `Tratamento concluído. Aplicados: ${data.applied}. Ignorados: ${data.skipped}.`
      );
      setConfirmOpen(false);
      await load();
    } catch (e: unknown) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    } finally {
      setApplying(false);
    }
  };

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Relatórios' },
        { label: 'Conciliação de linhas' }
      ]}
    >
      <ListPageHeader
        title="Conciliação de linhas"
        description="Linhas no sistema ausentes na base externa e canceladas na operadora mas ativas internamente. Fontes parciais geram divergências inconclusivas."
      />

      <div className="mb-4 flex flex-wrap items-end gap-3">
        <div className="space-y-1">
          <Label>Fonte</Label>
          <Input value={source} onChange={(e) => setSource(e.target.value)} className="w-40" />
        </div>
        <div className="space-y-1">
          <Label>Job de importação (opcional)</Label>
          <Input value={jobId} onChange={(e) => setJobId(e.target.value)} className="w-72" />
        </div>
        <div className="flex gap-2">
          <Button
            type="button"
            variant={tab === 'missing' ? 'default' : 'outline'}
            onClick={() => setTab('missing')}
          >
            Ausentes na operadora
          </Button>
          <Button
            type="button"
            variant={tab === 'cancelled' ? 'default' : 'outline'}
            onClick={() => setTab('cancelled')}
          >
            Canceladas externamente
          </Button>
        </div>
        {tab === 'missing' ? (
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={partial}
              onChange={(e) => setPartial(e.target.checked)}
            />
            Fonte parcial
          </label>
        ) : null}
        <Button type="button" onClick={() => void load()} disabled={loading}>
          {loading ? 'Carregando…' : 'Consultar'}
        </Button>
        {tab === 'cancelled' ? (
          <Button
            type="button"
            variant="destructive"
            disabled={selectedIds.length === 0}
            onClick={() => setConfirmOpen(true)}
          >
            Cancelar selecionadas ({selectedIds.length})
          </Button>
        ) : null}
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            {tab === 'cancelled' ? <TableHead className="w-10" /> : null}
            <TableHead>Número</TableHead>
            <TableHead>Cliente</TableHead>
            <TableHead>Operadora</TableHead>
            <TableHead>Status interno</TableHead>
            <TableHead>Status externo</TableHead>
            <TableHead>Evidência</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((it, idx) => {
            const id = it.phone_line_id ?? `row-${idx}`;
            return (
              <TableRow key={id}>
                {tab === 'cancelled' ? (
                  <TableCell>
                    {it.phone_line_id && !it.is_inconclusive ? (
                      <input
                        type="checkbox"
                        checked={Boolean(selected[it.phone_line_id])}
                        onChange={(e) =>
                          setSelected((prev) => ({
                            ...prev,
                            [it.phone_line_id!]: e.target.checked
                          }))
                        }
                      />
                    ) : null}
                  </TableCell>
                ) : null}
                <TableCell className="font-mono">
                  {formatPhoneNumber(it.normalized_number) ?? it.normalized_number}
                  {it.is_inconclusive ? (
                    <span className="text-warning ml-2 text-xs">inconclusiva</span>
                  ) : null}
                </TableCell>
                <TableCell>{it.customer_name ?? '—'}</TableCell>
                <TableCell>{it.provider_name ?? '—'}</TableCell>
                <TableCell>{it.internal_status ?? '—'}</TableCell>
                <TableCell>{it.external_status ?? '—'}</TableCell>
                <TableCell className="max-w-xs text-xs">{it.evidence ?? '—'}</TableCell>
                <TableCell>
                  {it.phone_line_id ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      nativeButton={false}
                      render={
                        <Link
                          to="/phone-lines/$phoneLineId"
                          params={{ phoneLineId: it.phone_line_id }}
                          search={{ page: 1, pageSize: 10 }}
                        />
                      }
                    >
                      Histórico
                    </Button>
                  ) : null}
                </TableCell>
              </TableRow>
            );
          })}
          {!items.length ? (
            <TableRow>
              <TableCell colSpan={8} className="text-muted-foreground">
                Nenhum resultado. Importe uma base externa (Axon) e consulte novamente.
              </TableCell>
            </TableRow>
          ) : null}
        </TableBody>
      </Table>

      <ConfirmDialog
        open={confirmOpen}
        title="Cancelar linhas no sistema"
        description={`Confirma o cancelamento de ${selectedIds.length} linha(s) com base na conciliação externa? Não há cancelamento automático sem esta confirmação. A ação será registrada no histórico.`}
        confirmLabel="Cancelar linhas"
        destructive
        requirePhrase="CANCELAR"
        loading={applying}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={applyBatch}
      />
    </PageWrapper>
  );
}

export const Route = createFileRoute('/__app/reports/reconciliation/')({
  component: RouteComponent
});
