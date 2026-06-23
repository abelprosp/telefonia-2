import { useEffect, useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { Loader2, Send } from 'lucide-react';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
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
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatBillingStatus,
  useBillingSendLog,
  useCustomerBillingDocument,
  useSendCustomerBillingDocument,
  useUpdateCustomerBillingDocument
} from '@/lib/billing-api';
import { formatMoney } from '@/lib/financial-api';

export const Route = createFileRoute('/__app/finance/customer-invoices/$id')({
  component: CustomerInvoiceDetailPage
});

function CustomerInvoiceDetailPage() {
  const { id } = Route.useParams();
  const docQuery = useCustomerBillingDocument(id);
  const sendLogQuery = useBillingSendLog(id);
  const updateMutation = useUpdateCustomerBillingDocument();
  const sendMutation = useSendCustomerBillingDocument();

  const [recipient, setRecipient] = useState('');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [status, setStatus] = useState('draft');

  useEffect(() => {
    const d = docQuery.data;
    if (!d) return;
    setRecipient(d.recipient_email);
    setSubject(d.email_subject);
    setBody(d.email_body_html ?? '');
    setStatus(d.status);
  }, [docQuery.data]);

  const handleSave = async () => {
    try {
      await updateMutation.mutateAsync({
        id,
        recipient_email: recipient,
        email_subject: subject,
        email_body_html: body,
        status
      });
      toast.success('Fatura atualizada.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleSend = async () => {
    try {
      await updateMutation.mutateAsync({
        id,
        recipient_email: recipient,
        email_subject: subject,
        email_body_html: body,
        status: status === 'cancelled' ? 'cancelled' : 'ready'
      });
      const result = await sendMutation.mutateAsync(id);
      toast.success(result.message);
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  if (docQuery.isPending) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Loader2 className="text-muted-foreground size-8 animate-spin" />
      </div>
    );
  }

  const doc = docQuery.data;
  if (!doc) {
    return <p className="p-6 text-sm">Fatura não encontrada.</p>;
  }

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Financeiro', to: '/finance' },
        { label: 'Faturas para envio', to: '/finance/customer-invoices' },
        { label: doc.invoice_number }
      ]}
    >
      <div className="flex flex-col gap-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">{doc.invoice_number}</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              {doc.customer_name} · {formatMoney(doc.amount)} · venc.{' '}
              {new Date(doc.due_date).toLocaleDateString('pt-BR')}
            </p>
          </div>
          <Badge variant="outline">{formatBillingStatus(doc.status)}</Badge>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Conteúdo do e-mail</h2>
            <div className="space-y-2">
              <Label>Destinatário</Label>
              <Input value={recipient} onChange={(e) => setRecipient(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Assunto</Label>
              <Input value={subject} onChange={(e) => setSubject(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Corpo (HTML)</Label>
              <Textarea
                className="min-h-[280px] font-mono text-xs"
                value={body}
                onChange={(e) => setBody(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Status</Label>
              <Select value={status} onValueChange={(v) => setStatus(v ?? 'draft')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">Rascunho</SelectItem>
                  <SelectItem value="ready">Pronta para envio</SelectItem>
                  <SelectItem value="cancelled">Cancelada</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" onClick={() => void handleSave()} disabled={updateMutation.isPending}>
                Salvar rascunho
              </Button>
              <Button onClick={() => void handleSend()} disabled={sendMutation.isPending || status === 'cancelled'}>
                <Send className="mr-2 size-4" />
                Enviar e-mail
              </Button>
            </div>
          </div>

          <div className="space-y-4">
            <div className="dashboard-card p-5">
              <h2 className="mb-3 font-semibold">Pré-visualização</h2>
              <div
                className="prose prose-sm dark:prose-invert max-w-none rounded-lg border p-4"
                dangerouslySetInnerHTML={{ __html: body }}
              />
            </div>
            <div className="dashboard-card p-5">
              <h2 className="mb-3 font-semibold">Histórico de envios</h2>
              {(sendLogQuery.data ?? []).length === 0 ? (
                <p className="text-muted-foreground text-sm">Nenhum envio registrado.</p>
              ) : (
                <ul className="space-y-2 text-sm">
                  {(sendLogQuery.data ?? []).map((log) => (
                    <li key={log.id} className="border-b pb-2 last:border-0">
                      <span className={log.success ? 'text-green-600' : 'text-destructive'}>
                        {log.success ? 'Enviado' : 'Falhou'}
                      </span>
                      {' · '}
                      {new Date(log.sent_at).toLocaleString('pt-BR')} · {log.recipient_email}
                      {log.error_message ? (
                        <p className="text-destructive text-xs">{log.error_message}</p>
                      ) : null}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        </div>

        <Link
          to="/finance/customer-invoices"
          search={{ page: 1, pageSize: 10 }}
          className="text-primary text-sm font-medium hover:underline"
        >
          Voltar à lista
        </Link>
      </div>
    </PageWrapper>
  );
}
