import { useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { defaultTicketStorageBucket, downloadTicketAttachment, putTicketAttachmentFile, requestTicketAttachmentUploadUrl } from '@/lib/ticket-attachment-upload';
import { useAddTicketMessage, useSupportTicket, useUpdateSupportTicket } from '@/lib/ops-api';

export const Route = createFileRoute('/__app/tickets/$ticketId')({
  component: TicketDetailPage
});

function TicketDetailPage() {
  const { ticketId } = Route.useParams();
  const query = useSupportTicket(ticketId);
  const updateMutation = useUpdateSupportTicket(ticketId);
  const msgMutation = useAddTicketMessage(ticketId);
  const [body, setBody] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [formError, setFormError] = useState('');
  const ticket = query.data;
  const bucketConfigured = Boolean(defaultTicketStorageBucket());

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Tickets', to: '/tickets' },
        { label: ticket ? `#${ticket.number}` : 'Detalhe' }
      ]}
    >
      <a href="#ticket-reply" className="skip-link">
        Pular para a resposta
      </a>
      {!ticket ? (
        <p role="status">Carregando ticket…</p>
      ) : (
        <div className="flex max-w-3xl flex-col gap-6">
          <div>
            <h2 className="text-lg font-semibold">{ticket.title}</h2>
            <p className="text-sm">
              Status {ticket.status.replaceAll('_', ' ')} · prioridade {ticket.priority}
              {ticket.sla_due_at ? ` · SLA até ${new Date(ticket.sla_due_at).toLocaleString('pt-BR')}` : ''}
              {ticket.customer_id ? ` · cliente ${ticket.customer_id}` : ''}
              {ticket.phone_line_id ? ` · linha ${ticket.phone_line_id}` : ''}
              {ticket.invoice_id ? ` · fatura ${ticket.invoice_id}` : ''}
            </p>
          </div>
          <div className="flex flex-wrap gap-2" role="group" aria-label="Alterar status do ticket">
            {['em_triagem', 'em_atendimento', 'aguardando_cliente', 'aguardando_operadora', 'resolvido', 'encerrado', 'reaberto'].map(
              (st) => (
                <Button
                  key={st}
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    updateMutation.mutate(
                      { status: st },
                      {
                        onSuccess: () => toast.success('Status atualizado.'),
                        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  {st.replaceAll('_', ' ')}
                </Button>
              )
            )}
          </div>
          <section aria-labelledby="ticket-messages-title">
            <h3 id="ticket-messages-title" className="mb-2 font-semibold">
              Mensagens
            </h3>
            <ul className="space-y-2">
              {(ticket.messages ?? []).map((m) => (
                <li key={m.id} className="rounded-md border p-3 text-sm">
                  <p className="text-muted-foreground text-xs">
                    {m.author_name ?? 'Sistema'} · {new Date(m.created_at).toLocaleString('pt-BR')} · {m.visibility}
                  </p>
                  <p>{m.body}</p>
                  {m.attachment_name ? (
                    <p className="mt-2 text-xs">
                      Anexo:{' '}
                      <Button
                        type="button"
                        variant="link"
                        className="h-auto p-0 text-xs underline"
                        onClick={() =>
                          void downloadTicketAttachment({ ticketId, messageId: m.id }).catch((err) =>
                            toast.error(err instanceof Error ? err.message : 'Falha ao baixar anexo.')
                          )
                        }
                      >
                        {m.attachment_name}
                      </Button>
                    </p>
                  ) : null}
                </li>
              ))}
            </ul>
            <form
              className="mt-3 flex flex-col gap-3"
              onSubmit={async (e) => {
                e.preventDefault();
                setFormError('');
                try {
                  let attachment:
                    | {
                        attachment_key?: string;
                        attachment_name?: string;
                        attachment_bucket?: string;
                        attachment_content_type?: string;
                        attachment_size_bytes?: number;
                      }
                    | undefined;
                  if (file) {
                    if (!bucketConfigured) {
                      setFormError('Defina VITE_STORAGE_BUCKET_NAME para enviar anexos.');
                      return;
                    }
                    const upload = await requestTicketAttachmentUploadUrl({ ticketId, file });
                    await putTicketAttachmentFile(file, upload);
                    attachment = {
                      attachment_key: upload.object_key,
                      attachment_name: file.name,
                      attachment_bucket: upload.bucket_name,
                      attachment_content_type: file.type || 'application/octet-stream',
                      attachment_size_bytes: file.size
                    };
                  }
                  await msgMutation.mutateAsync({
                    body,
                    ...attachment
                  });
                  setBody('');
                  setFile(null);
                  toast.success('Mensagem enviada.');
                } catch (err) {
                  const message = isApiHttpError(err) ? err.message : getErrorMessage(err);
                  setFormError(message);
                  toast.error(message);
                }
              }}
            >
              <div>
                <Label htmlFor="ticket-reply">Resposta</Label>
                <Input
                  id="ticket-reply"
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  aria-invalid={Boolean(formError)}
                  aria-describedby={formError ? 'ticket-reply-error' : undefined}
                />
              </div>
              <div>
                <Label htmlFor="ticket-attachment-file">Anexo (PDF, imagem ou documento)</Label>
                <Input
                  id="ticket-attachment-file"
                  type="file"
                  onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                />
                {file ? <p className="text-muted-foreground mt-1 text-xs">{file.name}</p> : null}
              </div>
              {formError ? (
                <p id="ticket-reply-error" className="text-destructive text-sm" role="alert">
                  {formError}
                </p>
              ) : null}
              <Button type="submit" disabled={msgMutation.isPending}>
                Enviar
              </Button>
            </form>
          </section>
          <section aria-labelledby="ticket-history-title">
            <h3 id="ticket-history-title" className="mb-2 font-semibold">
              Histórico
            </h3>
            <ul className="text-muted-foreground space-y-1 text-sm">
              {(ticket.history ?? []).map((h) => (
                <li key={h.id}>
                  {h.event_type}: {h.from_value ?? '—'} → {h.to_value ?? '—'}
                </li>
              ))}
            </ul>
          </section>
          <Button nativeButton={false} variant="outline" render={<Link to="/tickets" />}>
            Voltar
          </Button>
        </div>
      )}
    </PageWrapper>
  );
}
