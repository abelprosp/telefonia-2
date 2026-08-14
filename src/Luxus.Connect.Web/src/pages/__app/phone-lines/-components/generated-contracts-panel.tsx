import { usePhoneLineGeneratedContracts } from '@/lib/fidelity-api';

type GeneratedContractsPanelProps = {
  phoneLineId: string;
};

function triggerLabel(trigger?: string | null) {
  switch (trigger) {
    case 'sale':
      return 'Venda';
    case 'line_assign':
      return 'Vínculo da linha';
    case 'service_add':
      return 'Novo serviço';
    case 'device':
      return 'Aparelho';
    default:
      return trigger || 'Automático';
  }
}

export function GeneratedContractsPanel({ phoneLineId }: GeneratedContractsPanelProps) {
  const query = usePhoneLineGeneratedContracts(phoneLineId);

  if (query.isLoading) {
    return <p className="text-muted-foreground text-sm">Carregando contratos…</p>;
  }

  const items = query.data ?? [];
  if (items.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">
        Nenhum contrato gerado para esta linha. Contratos são criados automaticamente quando há
        template ativo, na venda, no vínculo da linha ou na inclusão de serviço.
      </p>
    );
  }

  return (
    <ul className="space-y-2 text-sm">
      {items.map((c) => (
        <li key={c.id} className="rounded-md border px-3 py-2">
          <span className="font-medium">{c.contract_template_name || 'Contrato'}</span>
          {' · '}
          {triggerLabel(c.trigger)}
          {' · '}
          {String(c.generated_at || c.created_at).slice(0, 10)}
        </li>
      ))}
    </ul>
  );
}
