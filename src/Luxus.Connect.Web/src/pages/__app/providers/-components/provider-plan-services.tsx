import type { GetProviderPlanServiceResponse } from '@/api';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';

function formatBrl(value: number | string | null | undefined) {
  if (value === null || value === undefined) {
    return '—';
  }
  const n = typeof value === 'string' ? Number(value) : value;
  if (!Number.isFinite(n)) {
    return '—';
  }
  return n.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

type Props = {
  services: GetProviderPlanServiceResponse[];
};

export function ProviderPlanServices({ services }: Props) {
  if (services.length === 0) {
    return (
      <p className="text-muted-foreground p-4 text-sm">
        Nenhum serviço cadastrado neste plano.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto p-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Recorrente</TableHead>
            <TableHead>Preço</TableHead>
            <TableHead>Ativo</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {services.map((row) => (
            <TableRow key={row.id}>
              <TableCell className="font-medium">{row.name}</TableCell>
              <TableCell>{row.recurring ? 'Sim' : 'Não'}</TableCell>
              <TableCell>{formatBrl(row.price)}</TableCell>
              <TableCell>{row.active ? 'Sim' : 'Não'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
