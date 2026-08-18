import { useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { toast } from 'sonner';

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
import { client } from '@/lib/client';
import { formatMoney } from '@/lib/financial-api';
import { formatCpfCnpj, formatPhoneNumber } from '@/lib/format';
import {
  downloadPortalInvoice,
  useAddTicketMessage,
  useCreatePortalTicket,
  usePortalContracts,
  usePortalInvoices,
  usePortalLines,
  usePortalMe,
  usePortalTicket,
  usePortalTickets,
  useUpdatePortalMe
} from '@/lib/ops-api';
import {
  defaultTicketStorageBucket,
  downloadTicketAttachment,
  putTicketAttachmentFile,
  requestTicketAttachmentUploadUrl
} from '@/lib/ticket-attachment-upload';

export const Route = createFileRoute('/__app/portal/')({
  component: PortalPage
});

function PortalPage() {
  const me = usePortalMe();
  const lines = usePortalLines();
  const invoices = usePortalInvoices();
  const contracts = usePortalContracts();
  const tickets = usePortalTickets();
  const updateMe = useUpdatePortalMe();
  const createTicket = useCreatePortalTicket();
  const [email, setEmail] = useState('');
  const [title, setTitle] = useState('');
  const [message, setMessage] = useState('');
  const [ticketFile, setTicketFile] = useState<File | null>(null);
  const [openTicketId, setOpenTicketId] = useState('');
  const [reply, setReply] = useState('');
  const [replyFile, setReplyFile] = useState<File | null>(null);
  const [ticketError, setTicketError] = useState('');
  const openTicket = usePortalTicket(openTicketId);
  const replyMutation = useAddTicketMessage(openTicketId, true);

  const currentEmail = me.data?.customer.billing_email ?? '';

  return (
    <PageWrapper breadcrumbs={[{ label: 'Portal do cliente' }]}>
      <a href="#portal-tickets" className="skip-link">
        Pular para tickets
      </a>
      <ListPageHeader
        title="Portal do cliente"
        description="Consulta exclusiva do seu CPF/CNPJ: linhas, cobranças, faturas, contratos e tickets. Sem custos de operadora nem dados de outros clientes."
      />
      {me.data ? (
        <p className="mb-6 text-sm">
          {me.data.customer.name} · {formatCpfCnpj(me.data.customer.cpf_cnpj)} · {me.data.active_lines_count} linhas
          ativas · mensalidade {formatMoney(me.data.total_monthly_amount)}
        </p>
      ) : (
        <p className="mb-6 text-sm">
          Vínculo 1:1 por CPF/CNPJ. Se o documento não estiver no token, o acesso é recusado.
        </p>
      )}

      <section className="mb-8" aria-labelledby="portal-profile">
        <h2 id="portal-profile" className="mb-2 font-semibold">
          Dados permitidos
        </h2>
        <form
          className="flex max-w-lg flex-wrap items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            updateMe.mutate(
              { billing_email: email || currentEmail },
              {
                onSuccess: () => toast.success('E-mail de cobrança atualizado.'),
                onError: (err) => toast.error(isApiHttpError(err) ? err.message : getErrorMessage(err))
              }
            );
          }}
        >
          <div className="min-w-64 flex-1">
            <Label htmlFor="portal-billing-email">E-mail de cobrança</Label>
            <Input
              id="portal-billing-email"
              type="email"
              autoComplete="email"
              defaultValue={currentEmail}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <Button type="submit" disabled={updateMe.isPending}>
            Salvar
          </Button>
        </form>
      </section>

      <section className="mb-8" aria-labelledby="portal-lines">
        <h2 id="portal-lines" className="mb-2 font-semibold">
          Linhas
        </h2>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Número</TableHead>
              <TableHead>Plano</TableHead>
              <TableHead>Mensalidade</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(lines.data ?? []).map((l) => (
              <TableRow key={l.id}>
                <TableCell>{formatPhoneNumber(l.number)}</TableCell>
                <TableCell>{l.plan_name}</TableCell>
                <TableCell>{formatMoney(l.monthly_amount)}</TableCell>
                <TableCell>{l.status}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </section>

      <section className="mb-8" aria-labelledby="portal-invoices">
        <h2 id="portal-invoices" className="mb-2 font-semibold">
          Faturas
        </h2>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Número</TableHead>
              <TableHead>Vencimento</TableHead>
              <TableHead>Valor</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Download</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(invoices.data ?? []).map((inv) => (
              <TableRow key={inv.id}>
                <TableCell>{inv.invoice_number}</TableCell>
                <TableCell>{new Date(inv.due_date).toLocaleDateString('pt-BR')}</TableCell>
                <TableCell>{formatMoney(inv.total_amount)}</TableCell>
                <TableCell>{inv.status}</TableCell>
                <TableCell>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      void downloadPortalInvoice(inv.id, `fatura-${inv.invoice_number}.html`).then(
                        () => toast.success('Fatura baixada.'),
                        (e: unknown) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      );
                    }}
                  >
                    Baixar
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </section>

      <section className="mb-8" aria-labelledby="portal-contracts">
        <h2 id="portal-contracts" className="mb-2 font-semibold">
          Contratos
        </h2>
        <ul className="text-sm">
          {(contracts.data ?? []).map((c) => (
            <li key={c.id}>
              {c.trigger ?? 'contrato'} · {c.status} · {new Date(c.created_at).toLocaleDateString('pt-BR')}
            </li>
          ))}
        </ul>
      </section>

      <section id="portal-tickets" aria-labelledby="portal-tickets-title">
        <h2 id="portal-tickets-title" className="mb-2 font-semibold">
          Tickets
        </h2>
        <form
          className="mb-4 grid max-w-2xl gap-2 sm:grid-cols-2"
          onSubmit={(e) => {
            e.preventDefault();
            setTicketError('');
            createTicket.mutate(
              { title, message, category: 'geral', priority: 'media' },
              {
                onSuccess: async (created) => {
                  try {
                    if (ticketFile) {
                      const upload = await requestTicketAttachmentUploadUrl({
                        ticketId: created.id,
                        file: ticketFile,
                        portal: true
                      });
                      await putTicketAttachmentFile(ticketFile, upload);
                      await client({
                        url: `/v1/portal/tickets/${created.id}/messages`,
                        method: 'POST',
                        data: {
                          body: '',
                          attachment_key: upload.object_key,
                          attachment_name: ticketFile.name,
                          attachment_bucket: upload.bucket_name
                        }
                      });
                    }
                    toast.success('Ticket aberto.');
                    setTitle('');
                    setMessage('');
                    setTicketFile(null);
                    setOpenTicketId(created.id);
                  } catch (err) {
                    toast.error(isApiHttpError(err) ? err.message : getErrorMessage(err));
                  }
                },
                onError: (err) => {
                  const msg = isApiHttpError(err) ? err.message : getErrorMessage(err);
                  setTicketError(msg);
                  toast.error(msg);
                }
              }
            );
          }}
        >
          <div className="sm:col-span-2">
            <Label htmlFor="portal-ticket-title">Título</Label>
            <Input
              id="portal-ticket-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              aria-invalid={Boolean(ticketError)}
              aria-describedby={ticketError ? 'portal-ticket-error' : undefined}
            />
          </div>
          <div className="sm:col-span-2">
            <Label htmlFor="portal-ticket-message">Mensagem</Label>
            <Input id="portal-ticket-message" value={message} onChange={(e) => setMessage(e.target.value)} />
          </div>
          <div className="sm:col-span-2">
            <Label htmlFor="portal-ticket-file">Anexo inicial (opcional)</Label>
            <Input
              id="portal-ticket-file"
              type="file"
              onChange={(e) => setTicketFile(e.target.files?.[0] ?? null)}
            />
            <p className="text-muted-foreground mt-1 text-xs">
              O arquivo é enviado após abrir o ticket, com URL autorizada.
            </p>
          </div>
          {ticketError ? (
            <p id="portal-ticket-error" className="text-destructive sm:col-span-2 text-sm" role="alert">
              {ticketError}
            </p>
          ) : null}
          <Button type="submit" disabled={createTicket.isPending || !title.trim()}>
            Abrir ticket
          </Button>
        </form>
        <ul className="text-sm">
          {(tickets.data?.items ?? []).map((t) => (
            <li key={t.id} className="mb-2">
              <button
                type="button"
                className="underline underline-offset-4"
                onClick={() => setOpenTicketId(t.id)}
              >
                #{t.number} {t.title}
              </button>{' '}
              · {t.status}
            </li>
          ))}
        </ul>
        {openTicket.data ? (
          <div className="mt-4 max-w-2xl rounded-lg border p-4">
            <h3 className="font-semibold">#{openTicket.data.number} {openTicket.data.title}</h3>
            <ul className="mt-2 space-y-2">
              {(openTicket.data.messages ?? []).map((m) => (
                <li key={m.id} className="text-sm">
                  <p>{m.body}</p>
                  {m.attachment_name ? (
                    <Button
                      type="button"
                      variant="link"
                      className="h-auto p-0 text-xs underline"
                      onClick={() =>
                        void downloadTicketAttachment({
                          ticketId: openTicket.data.id,
                          messageId: m.id,
                          portal: true
                        }).catch((err) => toast.error(err instanceof Error ? err.message : 'Falha ao baixar anexo.'))
                      }
                    >
                      Baixar {m.attachment_name}
                    </Button>
                  ) : null}
                </li>
              ))}
            </ul>
            <form
              className="mt-3 flex flex-col gap-2"
              onSubmit={async (e) => {
                e.preventDefault();
                try {
                  let extra: Record<string, string | number> = {};
                  if (replyFile) {
                    if (!defaultTicketStorageBucket()) {
                      toast.error('Defina VITE_STORAGE_BUCKET_NAME para enviar anexos.');
                      return;
                    }
                    const upload = await requestTicketAttachmentUploadUrl({
                      ticketId: openTicket.data.id,
                      file: replyFile,
                      portal: true
                    });
                    await putTicketAttachmentFile(replyFile, upload);
                    extra = {
                      attachment_key: upload.object_key,
                      attachment_name: replyFile.name,
                      attachment_bucket: upload.bucket_name
                    };
                  }
                  await replyMutation.mutateAsync({ body: reply, ...extra });
                  setReply('');
                  setReplyFile(null);
                  toast.success('Mensagem enviada.');
                } catch (err) {
                  toast.error(isApiHttpError(err) ? err.message : getErrorMessage(err));
                }
              }}
            >
              <Label htmlFor="portal-ticket-reply">Resposta</Label>
              <Input id="portal-ticket-reply" value={reply} onChange={(e) => setReply(e.target.value)} />
              <Label htmlFor="portal-ticket-reply-file">Anexo</Label>
              <Input
                id="portal-ticket-reply-file"
                type="file"
                onChange={(e) => setReplyFile(e.target.files?.[0] ?? null)}
              />
              <Button type="submit" disabled={replyMutation.isPending || (!reply.trim() && !replyFile)}>
                Enviar
              </Button>
            </form>
          </div>
        ) : null}
      </section>
    </PageWrapper>
  );
}
