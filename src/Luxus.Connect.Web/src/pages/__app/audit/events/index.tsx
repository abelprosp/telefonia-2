import { useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';

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
import client from '@/lib/client';

type AuditEvent = {
  id: string;
  entity_type: string;
  entity_id: string;
  action: string;
  actor_user_id?: string | null;
  created_at: string;
};

function RouteComponent() {
  const [entityType, setEntityType] = useState('phone_line');
  const [entityId, setEntityId] = useState('');
  const [items, setItems] = useState<AuditEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const p = new URLSearchParams({
        entity_type: entityType,
        page_index: '0',
        page_size: '50'
      });
      if (entityId.trim()) p.set('entity_id', entityId.trim());
      const { data } = await client<{ items: AuditEvent[] }>({
        url: `/v1/audit/events?${p}`,
        method: 'GET'
      });
      setItems(data.items ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Falha ao carregar histórico');
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageWrapper
      breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Auditoria' }, { label: 'Histórico' }]}
    >
      <div className="flex flex-col gap-4">
        <div>
          <h1 className="text-xl font-semibold">Histórico de movimentações</h1>
          <p className="text-muted-foreground text-sm">
            Eventos de domínio (criação, vínculo, importações). Acesso restrito a master, admin e
            financial.
          </p>
        </div>
        <div className="flex flex-wrap items-end gap-3">
          <div className="space-y-1">
            <Label>Tipo de entidade</Label>
            <Input value={entityType} onChange={(e) => setEntityType(e.target.value)} className="w-48" />
          </div>
          <div className="space-y-1">
            <Label>ID da entidade</Label>
            <Input value={entityId} onChange={(e) => setEntityId(e.target.value)} className="w-72" />
          </div>
          <Button type="button" onClick={() => void load()} disabled={loading}>
            {loading ? 'Carregando…' : 'Consultar'}
          </Button>
        </div>
        {error ? <p className="text-destructive text-sm">{error}</p> : null}
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Quando</TableHead>
              <TableHead>Ação</TableHead>
              <TableHead>Entidade</TableHead>
              <TableHead>Responsável</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((ev) => (
              <TableRow key={ev.id}>
                <TableCell className="whitespace-nowrap">
                  {new Date(ev.created_at).toLocaleString('pt-BR')}
                </TableCell>
                <TableCell>{ev.action}</TableCell>
                <TableCell className="font-mono text-xs">
                  {ev.entity_type}:{ev.entity_id}
                </TableCell>
                <TableCell className="font-mono text-xs">{ev.actor_user_id ?? '—'}</TableCell>
              </TableRow>
            ))}
            {!items.length ? (
              <TableRow>
                <TableCell colSpan={4} className="text-muted-foreground">
                  Nenhum evento encontrado para os filtros.
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </div>
    </PageWrapper>
  );
}

export const Route = createFileRoute('/__app/audit/events/')({
  component: RouteComponent
});
