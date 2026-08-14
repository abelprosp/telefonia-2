import { useEffect, useState } from 'react';

import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Field, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  useFidelityRenewalDecision,
  useFidelityRenewalTriggers,
  useLineFidelity,
  useUpsertLineFidelity
} from '@/lib/fidelity-api';
import { todayISO } from '@/lib/financial-api';

type FidelityPanelProps = {
  phoneLineId: string;
};

function eventTypeLabel(type: string) {
  switch (type) {
    case 'automatic':
      return 'Automática';
    case 'contractual_change':
      return 'Por alteração';
    case 'declined':
      return 'Não renovar';
    case 'created':
      return 'Cadastro';
    default:
      return type;
  }
}

export function FidelityPanel({ phoneLineId }: FidelityPanelProps) {
  const query = useLineFidelity(phoneLineId);
  const upsert = useUpsertLineFidelity(phoneLineId);
  const decision = useFidelityRenewalDecision(phoneLineId);
  const triggers = useFidelityRenewalTriggers();

  const [startDate, setStartDate] = useState(todayISO());
  const [months, setMonths] = useState('12');
  const [autoRenew, setAutoRenew] = useState(false);
  const [renewalMonths, setRenewalMonths] = useState('12');

  useEffect(() => {
    if (!query.data) return;
    setStartDate(String(query.data.start_date).slice(0, 10));
    setMonths(String(query.data.initial_months));
    setAutoRenew(query.data.auto_renew);
    setRenewalMonths(String(query.data.renewal_period_months ?? 12));
  }, [query.data]);

  const missing = query.isError;

  return (
    <div className="space-y-4">
      {query.isLoading ? (
        <p className="text-muted-foreground text-sm">Carregando fidelidade…</p>
      ) : null}

      {missing ? (
        <p className="text-muted-foreground text-sm">
          Nenhuma fidelidade cadastrada. Informe o início e o prazo para criar.
        </p>
      ) : null}

      {query.data ? (
        <p className="text-sm">
          Status:{' '}
          <strong>
            {query.data.status === 'expired' ? 'Contrato expirado' : 'Com contrato ativo'}
          </strong>
          {' · '}
          Término previsto:{' '}
          {String(query.data.predicted_end_date).slice(0, 10)}
        </p>
      ) : null}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <Field>
          <FieldLabel>Início da fidelidade</FieldLabel>
          <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
        </Field>
        <Field>
          <FieldLabel>Prazo inicial (meses)</FieldLabel>
          <Input
            inputMode="numeric"
            value={months}
            onChange={(e) => setMonths(e.target.value)}
          />
        </Field>
        <Field>
          <FieldLabel className="flex items-center gap-2">
            <input
              type="checkbox"
              className="size-4 rounded border"
              checked={autoRenew}
              onChange={(e) => setAutoRenew(e.target.checked)}
            />
            Renovação automática
          </FieldLabel>
        </Field>
        <Field>
          <FieldLabel>Período de renovação (meses)</FieldLabel>
          <Input
            inputMode="numeric"
            value={renewalMonths}
            disabled={!autoRenew}
            onChange={(e) => setRenewalMonths(e.target.value)}
          />
        </Field>
      </div>

      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          size="sm"
          disabled={upsert.isPending}
          onClick={() => {
            const n = Number(months);
            if (!Number.isFinite(n) || n <= 0) {
              toast.error('Informe o prazo em meses.');
              return;
            }
            upsert.mutate(
              {
                start_date: startDate,
                initial_months: n,
                auto_renew: autoRenew,
                ...(autoRenew ? { renewal_period_months: Number(renewalMonths) || 12 } : {})
              },
              {
                onSuccess: () => toast.success('Fidelidade salva.'),
                onError: (e) =>
                  toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              }
            );
          }}
        >
          Salvar fidelidade
        </Button>
        {query.data ? (
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={decision.isPending}
            onClick={() => {
              const ok = window.confirm(
                'Esta alteração pode gerar renovação contratual. Deseja renovar a fidelidade pelo prazo inicial?'
              );
              decision.mutate(
                {
                  renew: ok,
                  trigger: 'plan',
                  notes: ok
                    ? 'Renovação por alteração contratual.'
                    : 'Operador optou por não renovar.'
                },
                {
                  onSuccess: () =>
                    toast.success(ok ? 'Fidelidade renovada.' : 'Decisão registrada.'),
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              );
            }}
          >
            Renovar por alteração
          </Button>
        ) : null}
      </div>

      {query.data && query.data.history.length > 0 ? (
        <div>
          <p className="mb-2 text-sm font-medium">Histórico</p>
          <ul className="text-muted-foreground space-y-1 text-sm">
            {query.data.history.map((ev) => (
              <li key={ev.id}>
                {String(ev.occurred_at).slice(0, 10)} · {eventTypeLabel(ev.event_type)}
                {ev.notes ? ` — ${ev.notes}` : ''}
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {(triggers.data ?? []).length > 0 ? (
        <p className="text-muted-foreground text-xs">
          Gatilhos de pergunta:{' '}
          {triggers.data
            ?.filter((t) => t.prompt_enabled)
            .map((t) => t.label)
            .join(', ') || 'nenhum'}
          .
        </p>
      ) : null}
    </div>
  );
}
