import { useState } from 'react';
import { Loader2, Webhook } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  WEBHOOK_EVENT_OPTIONS,
  useCreateWebhook,
  useDeleteWebhook,
  useTestWebhook,
  useWebhooks
} from '@/lib/ops-api';

export function WebhooksSettingsPanel() {
  const query = useWebhooks(true);
  const create = useCreateWebhook();
  const remove = useDeleteWebhook();
  const test = useTestWebhook();
  const [url, setUrl] = useState('');
  const [events, setEvents] = useState<string[]>(['BILLING_CLOSED']);
  const [lastSecret, setLastSecret] = useState<string | null>(null);

  const toggleEvent = (value: string) => {
    setEvents((prev) =>
      prev.includes(value) ? prev.filter((e) => e !== value) : [...prev, value]
    );
  };

  return (
    <Card className="rounded-2xl border shadow-xs">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Webhook className="text-primary size-5" />
          Webhooks
        </CardTitle>
        <CardDescription>
          Assinaturas persistidas. O segredo completo (`whsec_…`) aparece só na criação.
          Eventos: transição de linha, fidelidade a vencer, divergência e fechamento.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-3 rounded-xl border p-4">
          <div className="space-y-2">
            <Label htmlFor="webhook-url">URL HTTPS</Label>
            <Input
              id="webhook-url"
              placeholder="https://exemplo.com/hooks/luxus"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
            />
          </div>
          <div className="flex flex-wrap gap-3">
            {WEBHOOK_EVENT_OPTIONS.map((opt) => (
              <label key={opt.value} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  className="size-4 rounded border"
                  checked={events.includes(opt.value)}
                  onChange={() => toggleEvent(opt.value)}
                />
                {opt.label}
              </label>
            ))}
          </div>
          <Button
            type="button"
            size="sm"
            disabled={create.isPending || !url.trim() || events.length === 0}
            onClick={() =>
              create.mutate(
                { url: url.trim(), events },
                {
                  onSuccess: (created) => {
                    setLastSecret(created.secret);
                    setUrl('');
                    toast.success('Webhook criado. Copie o segredo agora.');
                  },
                  onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              )
            }
          >
            {create.isPending ? 'Salvando…' : 'Cadastrar webhook'}
          </Button>
          {lastSecret ? (
            <p className="rounded-lg bg-muted p-3 font-mono text-xs break-all">
              Segredo (única exibição): {lastSecret}
            </p>
          ) : null}
        </div>

        {query.isPending ? (
          <p className="text-muted-foreground flex items-center gap-2 text-sm">
            <Loader2 className="size-4 animate-spin" /> Carregando webhooks…
          </p>
        ) : (query.data ?? []).length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhuma assinatura cadastrada.</p>
        ) : (
          <ul className="space-y-3">
            {(query.data ?? []).map((wh) => (
              <li key={wh.id} className="flex flex-col gap-2 rounded-xl border p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <p className="truncate font-medium">{wh.url}</p>
                  <p className="text-muted-foreground text-xs">
                    {wh.events.join(', ')} · {wh.secret}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={test.isPending}
                    onClick={() =>
                      test.mutate(wh.id, {
                        onSuccess: (r) =>
                          toast.success(
                            r.success
                              ? `Teste OK (HTTP ${r.status_code})`
                              : `Falha (HTTP ${r.status_code})`
                          ),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      })
                    }
                  >
                    Testar
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    disabled={remove.isPending}
                    onClick={() =>
                      remove.mutate(wh.id, {
                        onSuccess: () => toast.success('Webhook removido.'),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      })
                    }
                  >
                    Excluir
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
