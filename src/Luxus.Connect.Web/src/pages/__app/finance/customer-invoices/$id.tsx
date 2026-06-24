import { useEffect, useState } from 'react';

import { createFileRoute, Link } from '@tanstack/react-router';
import { Loader2, QrCode, Send, FileDown, Ban, CalendarClock } from 'lucide-react';
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
  downloadCustomerBillingInvoice,
  downloadSicrediBoletoPDF,
  formatBillingStatus,
  formatSicrediBoletoStatus,
  useAlterSicrediBoletoDueDate,
  useBillingSendLog,
  useCancelSicrediBoleto,
  useCustomerBillingDocument,
  useIssueSicrediBoleto,
  useSendCustomerBillingDocument,
  useSyncSicrediPayment,
  useUpdateCustomerBillingDocument
} from '@/lib/billing-api';
import { formatMoney } from '@/lib/financial-api';
import { wrapHtmlDocument } from '@/lib/invoice-layout/render';

export const Route = createFileRoute('/__app/finance/customer-invoices/$id')({
  component: CustomerInvoiceDetailPage
});

function CustomerInvoiceDetailPage() {
  const { id } = Route.useParams();
  const docQuery = useCustomerBillingDocument(id);
  const sendLogQuery = useBillingSendLog(id);
  const updateMutation = useUpdateCustomerBillingDocument();
  const sendMutation = useSendCustomerBillingDocument();
  const boletoMutation = useIssueSicrediBoleto();
  const syncPaymentMutation = useSyncSicrediPayment();
  const cancelBoletoMutation = useCancelSicrediBoleto();
  const alterDueDateMutation = useAlterSicrediBoletoDueDate();

  const [recipient, setRecipient] = useState('');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [status, setStatus] = useState('draft');
  const [newDueDate, setNewDueDate] = useState('');
  const [pdfLoading, setPdfLoading] = useState(false);
  const [invoiceDownloadLoading, setInvoiceDownloadLoading] = useState(false);

  useEffect(() => {
    const d = docQuery.data;
    if (!d) return;
    setRecipient(d.recipient_email);
    setSubject(d.email_subject);
    setBody(d.email_body_html ?? '');
    setStatus(d.status);
    setNewDueDate(d.due_date.slice(0, 10));
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

  const handleIssueBoleto = async () => {
    try {
      const result = await boletoMutation.mutateAsync(id);
      toast.success(result.message);
      await docQuery.refetch();
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleSyncPayment = async () => {
    try {
      const result = await syncPaymentMutation.mutateAsync(id);
      const item = result.items[0];
      if (result.paid > 0) {
        toast.success('Pagamento confirmado no Sicredi.');
      } else {
        toast.message(item?.message ?? 'Pagamento ainda não identificado.');
      }
      await docQuery.refetch();
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleDownloadInvoice = async () => {
    setInvoiceDownloadLoading(true);
    try {
      await downloadCustomerBillingInvoice(id, `fatura-${docQuery.data?.invoice_number ?? id}.html`);
      toast.success('Fatura baixada. Abra no navegador e use Imprimir → Salvar como PDF se precisar.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    } finally {
      setInvoiceDownloadLoading(false);
    }
  };

  const handleDownloadPDF = async () => {
    setPdfLoading(true);
    try {
      await downloadSicrediBoletoPDF(id, `boleto-${docQuery.data?.invoice_number ?? id}.pdf`);
      toast.success('PDF do boleto baixado.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    } finally {
      setPdfLoading(false);
    }
  };

  const handleCancelBoleto = async () => {
    if (!window.confirm('Baixar/cancelar este boleto no Sicredi?')) return;
    try {
      const result = await cancelBoletoMutation.mutateAsync(id);
      toast.success(result.message);
      await docQuery.refetch();
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleAlterDueDate = async () => {
    if (!newDueDate) return;
    try {
      const result = await alterDueDateMutation.mutateAsync({ id, due_date: newDueDate });
      toast.success(result.message);
      await docQuery.refetch();
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

  const boletoActive =
    Boolean(doc.sicredi_nosso_numero) &&
    !doc.sicredi_paid_at &&
    doc.sicredi_boleto_status !== 'paid' &&
    doc.sicredi_boleto_status !== 'cancelled';

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
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{formatBillingStatus(doc.status)}</Badge>
            <Button
              size="sm"
              variant="outline"
              disabled={invoiceDownloadLoading || !doc.email_body_html}
              onClick={() => void handleDownloadInvoice()}
            >
              <FileDown className="mr-2 size-4" />
              {invoiceDownloadLoading ? 'Baixando…' : 'Baixar fatura'}
            </Button>
          </div>
          {(doc.sicredi_paid_at || doc.sicredi_boleto_status === 'paid') && (
            <Badge className="bg-green-600 text-white hover:bg-green-600">
              Pago em {doc.sicredi_paid_at ? new Date(doc.sicredi_paid_at).toLocaleDateString('pt-BR') : '—'}
            </Badge>
          )}
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Conteúdo do e-mail</h2>
            <div className="space-y-2">
              <Label>Destinatário (opcional — necessário para envio por e-mail)</Label>
              <Input
                type="email"
                placeholder="cliente@empresa.com.br"
                value={recipient}
                onChange={(e) => setRecipient(e.target.value)}
              />
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
              <Button
                onClick={() => void handleSend()}
                disabled={
                  sendMutation.isPending ||
                  status === 'cancelled' ||
                  !recipient.trim()
                }
                title={!recipient.trim() ? 'Informe o e-mail do destinatário para enviar' : undefined}
              >
                <Send className="mr-2 size-4" />
                Enviar e-mail
              </Button>
            </div>
          </div>

          <div className="space-y-4">
            <div className="dashboard-card p-5">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 className="font-semibold">Boleto Sicredi / PIX</h2>
                <div className="flex flex-wrap gap-2">
                  {doc.sicredi_linha_digitavel && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={pdfLoading}
                      onClick={() => void handleDownloadPDF()}
                    >
                      <FileDown className="mr-2 size-4" />
                      {pdfLoading ? 'Baixando…' : 'PDF do boleto'}
                    </Button>
                  )}
                  {boletoActive && (
                    <>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={syncPaymentMutation.isPending}
                        onClick={() => void handleSyncPayment()}
                      >
                        {syncPaymentMutation.isPending ? 'Consultando…' : 'Atualizar pagamento'}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={cancelBoletoMutation.isPending}
                        onClick={() => void handleCancelBoleto()}
                      >
                        <Ban className="mr-2 size-4" />
                        {cancelBoletoMutation.isPending ? 'Baixando…' : 'Baixar boleto'}
                      </Button>
                    </>
                  )}
                  {!doc.sicredi_nosso_numero && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={boletoMutation.isPending}
                      onClick={() => void handleIssueBoleto()}
                    >
                      <QrCode className="mr-2 size-4" />
                      {boletoMutation.isPending ? 'Gerando…' : 'Gerar boleto'}
                    </Button>
                  )}
                </div>
              </div>
              {doc.sicredi_boleto_status && (
                <p className="text-muted-foreground mb-3 text-sm">
                  Status: {formatSicrediBoletoStatus(doc.sicredi_boleto_status, doc.sicredi_paid_at)}
                </p>
              )}
              {doc.sicredi_boleto_status === 'failed' && doc.sicredi_boleto_error && (
                <p className="text-destructive mb-3 text-sm">{doc.sicredi_boleto_error}</p>
              )}
              {doc.sicredi_nosso_numero ? (
                <dl className="space-y-2 text-sm">
                  <div>
                    <dt className="text-muted-foreground">Nosso número</dt>
                    <dd className="font-mono">{doc.sicredi_nosso_numero}</dd>
                  </div>
                  {doc.sicredi_linha_digitavel && (
                    <div>
                      <dt className="text-muted-foreground">Linha digitável</dt>
                      <dd className="font-mono break-all text-xs">{doc.sicredi_linha_digitavel}</dd>
                    </div>
                  )}
                  {doc.sicredi_codigo_barras && (
                    <div>
                      <dt className="text-muted-foreground">Código de barras</dt>
                      <dd className="font-mono break-all text-xs">{doc.sicredi_codigo_barras}</dd>
                    </div>
                  )}
                  {doc.sicredi_pix_qr_code && (
                    <div>
                      <dt className="text-muted-foreground">PIX copia e cola</dt>
                      <dd className="font-mono break-all text-xs">{doc.sicredi_pix_qr_code}</dd>
                    </div>
                  )}
                  {boletoActive && (
                    <div className="border-t pt-3">
                      <Label className="text-muted-foreground mb-2 flex items-center gap-1 text-xs">
                        <CalendarClock className="size-3.5" />
                        Alterar vencimento no Sicredi
                      </Label>
                      <div className="flex flex-wrap gap-2">
                        <Input
                          type="date"
                          className="w-auto"
                          value={newDueDate}
                          onChange={(e) => setNewDueDate(e.target.value)}
                        />
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={alterDueDateMutation.isPending || !newDueDate}
                          onClick={() => void handleAlterDueDate()}
                        >
                          {alterDueDateMutation.isPending ? 'Salvando…' : 'Aplicar'}
                        </Button>
                      </div>
                    </div>
                  )}
                </dl>
              ) : (
                <p className="text-muted-foreground text-sm">
                  O boleto híbrido (código de barras + QR PIX) é gerado automaticamente ao criar a fatura,
                  quando a integração Sicredi está configurada.
                </p>
              )}
            </div>
            <div className="dashboard-card p-5">
              <h2 className="mb-3 font-semibold">Pré-visualização</h2>
              <iframe
                className="h-[min(70vh,900px)] w-full rounded-lg border bg-white"
                title="Pré-visualização da fatura"
                srcDoc={wrapHtmlDocument(body)}
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
