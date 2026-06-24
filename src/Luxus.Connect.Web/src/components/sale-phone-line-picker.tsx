import { useMemo, useState } from 'react';

import { useGetV1PhoneLines, type ListPhoneLineResponse } from '@/api';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { formatPhoneNumber } from '@/lib/format';

const ALL_PLANS = '__all__';

type SalePhoneLinePickerProps = {
  value?: string;
  excludeLineIds?: string[];
  onChange: (line: ListPhoneLineResponse | null) => void;
};

function lineLabel(line: ListPhoneLineResponse) {
  const number = formatPhoneNumber(line.number) ?? line.number;
  return `${number} · ${line.provider_plan_name}`;
}

export function SalePhoneLinePicker({ value, excludeLineIds = [], onChange }: SalePhoneLinePickerProps) {
  const [planFilter, setPlanFilter] = useState(ALL_PLANS);

  const stockQuery = useGetV1PhoneLines({
    page_index: 0,
    page_size: 100,
    status: 'in_stock'
  });

  const availableLines = useMemo(() => {
    const excluded = new Set(excludeLineIds);
    return (stockQuery.data?.items ?? []).filter((line) => !excluded.has(line.id));
  }, [excludeLineIds, stockQuery.data?.items]);

  const planOptions = useMemo(() => {
    const map = new Map<string, string>();
    for (const line of availableLines) {
      if (!map.has(line.provider_plan_id)) {
        map.set(line.provider_plan_id, line.provider_plan_name);
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name, 'pt-BR'));
  }, [availableLines]);

  const filteredLines = useMemo(() => {
    if (planFilter === ALL_PLANS) return availableLines;
    return availableLines.filter((line) => line.provider_plan_id === planFilter);
  }, [availableLines, planFilter]);

  const handlePlanChange = (next: string) => {
    setPlanFilter(next);
    if (value && next !== ALL_PLANS) {
      const stillVisible = availableLines.some(
        (line) => line.id === value && line.provider_plan_id === next
      );
      if (!stillVisible) {
        onChange(null);
      }
    }
  };

  if (stockQuery.isPending) {
    return <p className="text-muted-foreground text-sm">Carregando linhas em estoque…</p>;
  }

  if (availableLines.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">
        Nenhuma linha em estoque disponível. Cadastre linhas em Cadastros → Estoque de linhas.
      </p>
    );
  }

  return (
    <div className="space-y-3 rounded-lg border border-dashed p-3">
      <p className="text-muted-foreground text-xs">
        Selecione o plano e a linha disponível no estoque.
      </p>

      {planOptions.length > 1 && (
        <div className="space-y-2">
          <Label>Plano</Label>
          <Select value={planFilter} onValueChange={(v) => handlePlanChange(v ?? ALL_PLANS)}>
            <SelectTrigger>
              <SelectValue placeholder="Todos os planos" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_PLANS}>Todos os planos</SelectItem>
              {planOptions.map((plan) => (
                <SelectItem key={plan.id} value={plan.id}>
                  {plan.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="space-y-2">
        <Label>Linha em estoque</Label>
        {filteredLines.length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhuma linha neste plano.</p>
        ) : (
          <Select
            value={value ?? ''}
            onValueChange={(lineId) => {
              const line = filteredLines.find((l) => l.id === lineId) ?? null;
              onChange(line);
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="Selecione a linha" />
            </SelectTrigger>
            <SelectContent>
              {filteredLines.map((line) => (
                <SelectItem key={line.id} value={line.id}>
                  {lineLabel(line)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      {planOptions.length === 1 && (
        <p className="text-muted-foreground text-xs">Plano: {planOptions[0]?.name}</p>
      )}
    </div>
  );
}

export function buildPhoneLineSaleDescription(line: ListPhoneLineResponse) {
  const number = formatPhoneNumber(line.number) ?? line.number;
  return `Linha ${number} — ${line.provider_plan_name}`;
}
