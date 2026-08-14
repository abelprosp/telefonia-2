import { useState } from 'react';

import { toast } from 'sonner';

import { useGetV1CustomersIdProcessingMonthsProcessingmonthidBillingReadiness, useGetV1ProcessingMonths } from '@/api';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Field, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { useManualReleaseCustomer } from '@/lib/phone-line-spec-api';

export function BillingReadinessPanel({ customerId }: { customerId: string }) {
  const monthsQuery = useGetV1ProcessingMonths({ page_index: 0, page_size: 200 });
  const months = monthsQuery.data?.items ?? [];
  const [monthId, setMonthId] = useState('');
  const selected = monthId || months[0]?.id || '';
  const [justification, setJustification] = useState('');
  const readinessQuery = useGetV1CustomersIdProcessingMonthsProcessingmonthidBillingReadiness(
    customerId,
    selected,
    { query: { enabled: Boolean(customerId && selected) } }
  );
  const release = useManualReleaseCustomer(customerId, selected);
  const data = readinessQuery.data;

  return (
    <div className="space-y-3">
      <Field>
        <FieldLabel>Mês de processamento</FieldLabel>
        <Select value={selected} onValueChange={(v) => setMonthId(v ?? '')}>
          <SelectTrigger>
            <SelectValue placeholder="Selecione o mês" />
          </SelectTrigger>
          <SelectContent>
            {months.map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.display_name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </Field>
      {readinessQuery.isLoading ? (
        <p className="text-muted-foreground text-sm">Consultando liberação…</p>
      ) : data ? (
        <div className="space-y-2 text-sm">
          <div className="flex items-center gap-2">
            <span>Status:</span>
            <Badge variant={data.is_released_for_billing ? 'default' : 'secondary'}>
              {data.status_display_name}
            </Badge>
          </div>
          <p className="text-muted-foreground">
            Contas esperadas: {data.accounts_expected_for_automatic_rule} · Com fatura:{' '}
            {data.accounts_with_invoice_in_processing_month}
          </p>
          {!data.is_released_for_billing ? (
            <>
              <Field>
                <FieldLabel>Justificativa da liberação manual (mín. 10 caracteres)</FieldLabel>
                <Input
                  value={justification}
                  onChange={(e) => setJustification(e.target.value)}
                />
              </Field>
              <Button
                type="button"
                size="sm"
                disabled={release.isPending || justification.trim().length < 10}
                onClick={() =>
                  release.mutate(justification.trim(), {
                    onSuccess: () => {
                      toast.success('Cliente liberado para faturamento.');
                      setJustification('');
                      void readinessQuery.refetch();
                    },
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
              >
                Liberar manualmente
              </Button>
            </>
          ) : null}
        </div>
      ) : (
        <p className="text-muted-foreground text-sm">
          Sem contexto de liberação para este mês (cliente sem vínculo com a operadora do mês).
        </p>
      )}
    </div>
  );
}
