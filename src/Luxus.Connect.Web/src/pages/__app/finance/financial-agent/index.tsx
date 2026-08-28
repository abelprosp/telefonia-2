import { useEffect, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import {
  Bot,
  CheckCircle2,
  Loader2,
  MessageCircle,
  QrCode,
  Wallet
} from 'lucide-react';
import { toast } from 'sonner';

import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
import {
  type FinancialAgentWhatsAppStatus,
  useConnectFinancialAgentWhatsApp,
  useDisconnectFinancialAgentWhatsApp,
  useFinancialAgentEvents,
  useFinancialAgentPanel,
  useUpdateFinancialAgentSettings
} from '@/lib/financial-agent-api';

export const Route = createFileRoute('/__app/finance/financial-agent/')({
  component: FinancialAgentPage
});

function qrImageSrc(code?: string) {
  if (!code) return '';
  if (code.startsWith('data:') || code.startsWith('http')) return code;
  return `data:image/png;base64,${code}`;
}

function FinancialAgentPage() {
  const [awaitingPairing, setAwaitingPairing] = useState(false);
  const [qrOverride, setQrOverride] = useState<string>('');
  const panelQuery = useFinancialAgentPanel({
    refetchInterval: awaitingPairing ? 4000 : 12000
  });
  const eventsQuery = useFinancialAgentEvents({
    refetchInterval: awaitingPairing ? 4000 : 10000
  });
  const saveMutation = useUpdateFinancialAgentSettings();
  const connectMutation = useConnectFinancialAgentWhatsApp();
  const disconnectMutation = useDisconnectFinancialAgentWhatsApp();

  const panel = panelQuery.data;
  const wa = panel?.whatsapp;

  const [enabled, setEnabled] = useState(true);
  const [evolutionUrl, setEvolutionUrl] = useState('');
  const [evolutionKey, setEvolutionKey] = useState('');
  const [instance, setInstance] = useState('luxus');
  const [n8nUrl, setN8nUrl] = useState('');

  useEffect(() => {
    if (!panel) return;
    setEnabled(panel.settings.enabled);
    setEvolutionUrl(panel.settings.evolution_api_url);
    setInstance(panel.settings.evolution_instance || 'luxus');
    setN8nUrl(panel.settings.n8n_webhook_url);
  }, [
    panel?.settings.updated_at,
    panel?.settings.enabled,
    panel?.settings.evolution_api_url,
    panel?.settings.evolution_instance,
    panel?.settings.n8n_webhook_url
  ]);

  useEffect(() => {
    if (wa?.connected) {
      setAwaitingPairing(false);
      setQrOverride('');
    }
  }, [wa?.connected]);

  const applyWhatsApp = (status: FinancialAgentWhatsAppStatus) => {
    if (status.qr_code) {
      setQrOverride(status.qr_code);
      setAwaitingPairing(!status.connected);
    }
    if (status.connected) {
      setAwaitingPairing(false);
      setQrOverride('');
    }
  };

  const handleSave = async () => {
    try {
      const input: Parameters<typeof saveMutation.mutateAsync>[0] = {
        enabled,
        evolution_api_url: evolutionUrl.trim(),
        evolution_instance: instance.trim() || 'luxus',
        n8n_webhook_url: n8nUrl.trim()
      };
      if (evolutionKey.trim()) {
        input.evolution_api_key = evolutionKey.trim();
      }
      await saveMutation.mutateAsync(input);
      setEvolutionKey('');
      toast.success('Configuração do agente salva.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleConnect = async () => {
    try {
      const status = await connectMutation.mutateAsync();
      applyWhatsApp(status);
      toast.success(status.connected ? 'WhatsApp conectado.' : 'QR Code gerado. Escaneie no celular.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const handleDisconnect = async () => {
    try {
      await disconnectMutation.mutateAsync();
      setQrOverride('');
      setAwaitingPairing(false);
      toast.success('WhatsApp desconectado.');
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  if (panelQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Agente WhatsApp' }]}>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-2xl" />
          ))}
        </div>
      </PageWrapper>
    );
  }

  if (panelQuery.error) {
    const err = panelQuery.error;
    return (
      <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Agente WhatsApp' }]}>
        <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-2xl border px-4 py-3 text-sm">
          {isApiHttpError(err) ? err.message : getErrorMessage(err)}
        </div>
      </PageWrapper>
    );
  }

  if (!panel) return null;

  const qrSrc = qrImageSrc(qrOverride || wa?.qr_code);
  const connected = Boolean(wa?.connected);
  const configured = Boolean(wa?.configured);
  const events = eventsQuery.data ?? [];

  return (
    <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Agente WhatsApp' }]}>
      <ListPageHeader
        title="Agente financeiro no WhatsApp"
        description="Configure a Evolution, conecte o número e acompanhe o atendimento automático de faturas e comprovantes."
      />

      {panel.message ? (
        <p className="text-muted-foreground text-sm">{panel.message}</p>
      ) : null}

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatusCard
          title="Ferramentas da API"
          value={panel.tools_ready ? 'Ativas' : 'Pendentes'}
          ok={panel.tools_ready}
          icon={Bot}
          hint={
            panel.tools_ready
              ? 'Chave e organização do agente definidas no servidor.'
              : 'Defina FINANCIAL_AGENT_API_KEY e FINANCIAL_AGENT_ORG_ID na API.'
          }
        />
        <StatusCard
          title="WhatsApp"
          value={connected ? 'Conectado' : configured ? 'Desconectado' : 'Sem Evolution'}
          ok={connected}
          icon={MessageCircle}
          hint={wa?.profile_name || wa?.instance_name || wa?.message}
        />
        <StatusCard
          title="Sicredi"
          value={
            panel.sicredi_connected ? 'Conectado' : panel.sicredi_enabled ? 'Sem conexão' : 'Desligado'
          }
          ok={panel.sicredi_connected}
          icon={Wallet}
          hint={
            panel.sicredi_connected
              ? 'Consulta de boletos e liquidação disponível.'
              : 'O agente usa o Sicredi para confirmar pagamentos.'
          }
        />
        <StatusCard
          title="Atendimentos hoje"
          value={String(panel.stats.total_today)}
          ok={panel.stats.failed_today === 0}
          icon={CheckCircle2}
          hint={`${panel.stats.success_today} ok · ${panel.stats.failed_today} falha · ${panel.stats.payments_today} baixas`}
        />
      </div>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)]">
        <Card className="rounded-2xl border shadow-xs">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <QrCode className="text-primary size-5" />
              Conectar WhatsApp
            </CardTitle>
            <CardDescription>
              A Evolution gera o QR da instância. Escaneie com o WhatsApp do número que vai atender os
              clientes.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant={connected ? 'success-light' : configured ? 'warning-light' : 'secondary'}>
                {connected ? 'Conectado' : configured ? 'Aguardando pareamento' : 'Não configurado'}
              </Badge>
              {wa?.instance_name ? (
                <Badge variant="outline">Instância {wa.instance_name}</Badge>
              ) : null}
              {wa?.profile_name ? <Badge variant="outline">{wa.profile_name}</Badge> : null}
            </div>
            {wa?.message ? <p className="text-muted-foreground text-sm">{wa.message}</p> : null}
            {qrSrc && !connected ? (
              <div className="bg-background flex flex-col items-center gap-3 rounded-xl border p-4">
                <img src={qrSrc} alt="QR Code do WhatsApp" className="size-56 rounded-lg bg-white p-2" />
                <p className="text-muted-foreground text-center text-xs">
                  WhatsApp → Aparelhos conectados → Conectar um aparelho
                </p>
                {wa?.pairing_code ? (
                  <p className="font-mono text-sm">Código: {wa.pairing_code}</p>
                ) : null}
                {awaitingPairing ? (
                  <p className="text-muted-foreground flex items-center gap-2 text-xs">
                    <Loader2 className="size-3.5 animate-spin" /> Aguardando leitura do QR…
                  </p>
                ) : null}
              </div>
            ) : null}
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                disabled={!configured || !panel.settings.enabled || connectMutation.isPending || connected}
                onClick={() => void handleConnect()}
              >
                {connectMutation.isPending ? 'Gerando QR…' : connected ? 'Já conectado' : 'Conectar WhatsApp'}
              </Button>
              {qrSrc && !connected ? (
                <Button
                  type="button"
                  variant="outline"
                  disabled={!configured || connectMutation.isPending}
                  onClick={() => void handleConnect()}
                >
                  Atualizar QR
                </Button>
              ) : null}
              <Button
                type="button"
                variant="outline"
                disabled={!connected || disconnectMutation.isPending}
                onClick={() => void handleDisconnect()}
              >
                {disconnectMutation.isPending ? 'Desconectando…' : 'Desconectar'}
              </Button>
            </div>
            {!panel.settings.enabled ? (
              <p className="text-muted-foreground text-xs">
                Marque “Agente habilitado” e salve para conectar o WhatsApp.
              </p>
            ) : !configured ? (
              <p className="text-muted-foreground text-xs">
                Salve a URL e a chave da Evolution ao lado para habilitar o botão de conectar.
              </p>
            ) : null}
          </CardContent>
        </Card>

        <Card className="rounded-2xl border shadow-xs">
          <CardHeader>
            <CardTitle className="text-lg">Evolution e n8n</CardTitle>
            <CardDescription>
              A chave da API não é exibida depois de salva. Deixe em branco para manter a atual.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                className="size-4 rounded border"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
              />
              Agente habilitado nesta organização
            </label>
            <div className="space-y-2">
              <Label htmlFor="evolution-url">URL da Evolution API</Label>
              <Input
                id="evolution-url"
                placeholder="https://evolution.exemplo.com"
                value={evolutionUrl}
                onChange={(e) => setEvolutionUrl(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="evolution-key">
                Chave da Evolution {panel.settings.evolution_api_key_set ? '(já cadastrada)' : ''}
              </Label>
              <Input
                id="evolution-key"
                type="password"
                autoComplete="new-password"
                placeholder={panel.settings.evolution_api_key_set ? '••••••••' : 'apikey'}
                value={evolutionKey}
                onChange={(e) => setEvolutionKey(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="evolution-instance">Nome da instância</Label>
              <Input
                id="evolution-instance"
                placeholder="luxus"
                value={instance}
                onChange={(e) => setInstance(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="n8n-webhook">Webhook do n8n (MESSAGES_UPSERT)</Label>
              <Input
                id="n8n-webhook"
                placeholder="https://n8n.exemplo.com/webhook/agente-financeiro-whatsapp"
                value={n8nUrl}
                onChange={(e) => setN8nUrl(e.target.value)}
              />
            </div>
            <Button type="button" disabled={saveMutation.isPending} onClick={() => void handleSave()}>
              {saveMutation.isPending ? 'Salvando…' : 'Salvar configuração'}
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card className="rounded-2xl border shadow-xs">
        <CardHeader>
          <CardTitle className="text-lg">Histórico de atendimento</CardTitle>
          <CardDescription>
            Consultas, envio de faturas, comprovantes e baixas feitos pelo agente.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {eventsQuery.isPending ? (
            <p className="text-muted-foreground flex items-center gap-2 text-sm">
              <Loader2 className="size-4 animate-spin" /> Carregando eventos…
            </p>
          ) : events.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              Nenhum atendimento registrado ainda. Quando o WhatsApp estiver conectado, as conversas
              aparecem aqui.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Quando</TableHead>
                  <TableHead>Ação</TableHead>
                  <TableHead>Cliente</TableHead>
                  <TableHead>Fatura</TableHead>
                  <TableHead>WhatsApp</TableHead>
                  <TableHead>Resumo</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.map((ev) => (
                  <TableRow key={ev.id}>
                    <TableCell className="whitespace-nowrap">
                      {new Date(ev.created_at).toLocaleString('pt-BR')}
                    </TableCell>
                    <TableCell>
                      <Badge variant={ev.success ? 'success-light' : 'destructive-light'}>
                        {ev.event_label || ev.event_type}
                      </Badge>
                    </TableCell>
                    <TableCell>{ev.customer_name || '—'}</TableCell>
                    <TableCell className="font-mono text-xs">{ev.invoice_number || '—'}</TableCell>
                    <TableCell className="font-mono text-xs">{ev.whatsapp_number || '—'}</TableCell>
                    <TableCell className="max-w-md text-muted-foreground">{ev.summary}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </PageWrapper>
  );
}

function StatusCard({
  title,
  value,
  ok,
  icon: Icon,
  hint
}: {
  title: string;
  value: string;
  ok: boolean;
  icon: typeof Bot;
  hint?: string;
}) {
  return (
    <div className="dashboard-card p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-2">
          <p className="text-muted-foreground text-sm font-medium">{title}</p>
          <p className="text-2xl font-semibold tracking-tight">{value}</p>
          {hint ? <p className="text-muted-foreground text-xs">{hint}</p> : null}
        </div>
        <div
          className={
            ok
              ? 'bg-success/10 text-success flex size-10 shrink-0 items-center justify-center rounded-xl'
              : 'bg-primary/10 text-primary flex size-10 shrink-0 items-center justify-center rounded-xl'
          }
        >
          <Icon className="size-5" />
        </div>
      </div>
    </div>
  );
}
