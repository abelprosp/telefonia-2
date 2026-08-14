import { useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { toast } from 'sonner';

import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
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
  useCreateExceedanceTerm,
  useExceedanceTerms,
  useUpdateExceedanceTerm
} from '@/lib/fidelity-api';
import { formatMoney } from '@/lib/financial-api';

export const Route = createFileRoute('/__app/exceedance-terms/')({
  component: ExceedanceTermsPage
});

function ExceedanceTermsPage() {
  const query = useExceedanceTerms();
  const createMutation = useCreateExceedanceTerm();
  const updateMutation = useUpdateExceedanceTerm();
  const [term, setTerm] = useState('Roaming');
  const [chargeType, setChargeType] = useState('mirrored');
  const [tabulated, setTabulated] = useState('');

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Termos de excedente' }]}>
      <ListPageHeader
        title="Termos de excedente"
        description="Quando esses termos aparecerem na fatura importada, o sistema aplica cobrança automática nas linhas com a flag “Cobrar excedentes”."
      />
      <div className="mb-6 grid grid-cols-1 gap-3 rounded-lg border p-4 sm:grid-cols-4">
        <div>
          <Label>Termo</Label>
          <Input value={term} onChange={(e) => setTerm(e.target.value)} />
        </div>
        <div>
          <Label>Regra</Label>
          <Select value={chargeType} onValueChange={(v) => setChargeType(v ?? 'mirrored')}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="mirrored">Espelhado</SelectItem>
              <SelectItem value="tabulated">Tabelado</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>Valor tabelado</Label>
          <Input
            inputMode="decimal"
            value={tabulated}
            disabled={chargeType !== 'tabulated'}
            onChange={(e) => setTabulated(e.target.value)}
          />
        </div>
        <div className="flex items-end">
          <Button
            type="button"
            disabled={createMutation.isPending || !term.trim()}
            onClick={() => {
              const amount = tabulated.trim() ? Number(tabulated.replace(',', '.')) : undefined;
              createMutation.mutate(
                {
                  term: term.trim(),
                  charge_type: chargeType,
                  ...(chargeType === 'tabulated' && amount != null ? { tabulated_amount: amount } : {})
                },
                {
                  onSuccess: () => toast.success('Termo cadastrado.'),
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              );
            }}
          >
            Adicionar
          </Button>
        </div>
      </div>
      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Termo</TableHead>
              <TableHead>Regra</TableHead>
              <TableHead>Valor tabelado</TableHead>
              <TableHead>Ativo</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(query.data ?? []).map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.term}</TableCell>
                <TableCell>{item.charge_type === 'tabulated' ? 'Tabelado' : 'Espelhado'}</TableCell>
                <TableCell>
                  {item.tabulated_amount != null ? formatMoney(item.tabulated_amount) : '—'}
                </TableCell>
                <TableCell>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    disabled={updateMutation.isPending}
                    onClick={() =>
                      updateMutation.mutate(
                        { id: item.id, active: !item.active },
                        {
                          onSuccess: () => toast.success('Termo atualizado.'),
                          onError: (e) =>
                            toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                        }
                      )
                    }
                  >
                    {item.active ? 'Desativar' : 'Ativar'}
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </PageWrapper>
  );
}
