import { useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { useCreateSupportTicket, useSupportTickets } from '@/lib/ops-api';

export const Route = createFileRoute('/__app/tickets/')({
  component: TicketsPage
});

const STATUS_LABEL: Record<string, string> = {
  aberto: 'Aberto',
  em_triagem: 'Em triagem',
  em_atendimento: 'Em atendimento',
  aguardando_cliente: 'Aguardando cliente',
  aguardando_operadora: 'Aguardando operadora',
  resolvido: 'Resolvido',
  encerrado: 'Encerrado',
  reaberto: 'Reaberto'
};

function TicketsPage() {
  const query = useSupportTickets();
  const createMutation = useCreateSupportTicket();
  const [title, setTitle] = useState('');
  const [category, setCategory] = useState('geral');
  const [priority, setPriority] = useState('media');
  const [message, setMessage] = useState('');

  const items = query.data?.items ?? [];

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Tickets' }]}>
      <a href="#ticket-title" className="skip-link">
        Pular para abrir ticket
      </a>
      <ListPageHeader
        title="Tickets de suporte"
        description="Atendimento interno com categorias, prioridade, SLA, histórico e vínculos."
      />
      <form
        className="mb-6 grid grid-cols-1 gap-3 rounded-lg border p-4 sm:grid-cols-2 lg:grid-cols-5"
        onSubmit={(e) => {
          e.preventDefault();
          createMutation.mutate(
            { title, category, priority, message },
            {
              onSuccess: () => {
                toast.success('Ticket criado.');
                setTitle('');
                setMessage('');
              },
              onError: (err) => toast.error(isApiHttpError(err) ? err.message : getErrorMessage(err))
            }
          );
        }}
      >
        <div className="lg:col-span-2">
          <Label htmlFor="ticket-title">Título</Label>
          <Input id="ticket-title" value={title} onChange={(e) => setTitle(e.target.value)} required />
        </div>
        <div>
          <Label htmlFor="ticket-category">Categoria</Label>
          <Select value={category} onValueChange={(v) => setCategory(v ?? 'geral')}>
            <SelectTrigger id="ticket-category">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="geral">Geral</SelectItem>
              <SelectItem value="fatura">Fatura</SelectItem>
              <SelectItem value="linha">Linha</SelectItem>
              <SelectItem value="cobranca">Cobrança</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label htmlFor="ticket-priority">Prioridade</Label>
          <Select value={priority} onValueChange={(v) => setPriority(v ?? 'media')}>
            <SelectTrigger id="ticket-priority">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="baixa">Baixa</SelectItem>
              <SelectItem value="media">Média</SelectItem>
              <SelectItem value="alta">Alta</SelectItem>
              <SelectItem value="critica">Crítica</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-end">
          <Button type="submit" disabled={createMutation.isPending || !title.trim()}>
            Abrir ticket
          </Button>
        </div>
        <div className="sm:col-span-2 lg:col-span-5">
          <Label htmlFor="ticket-message">Mensagem inicial</Label>
          <Input id="ticket-message" value={message} onChange={(e) => setMessage(e.target.value)} />
        </div>
      </form>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>#</TableHead>
            <TableHead>Título</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Prioridade</TableHead>
            <TableHead>Atualizado</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((t) => (
            <TableRow key={t.id}>
              <TableCell>{t.number}</TableCell>
              <TableCell>
                <Link to="/tickets/$ticketId" params={{ ticketId: t.id }} className="underline">
                  {t.title}
                </Link>
              </TableCell>
              <TableCell>{STATUS_LABEL[t.status] ?? t.status}</TableCell>
              <TableCell>{t.priority}</TableCell>
              <TableCell>{new Date(t.updated_at).toLocaleString('pt-BR')}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </PageWrapper>
  );
}
