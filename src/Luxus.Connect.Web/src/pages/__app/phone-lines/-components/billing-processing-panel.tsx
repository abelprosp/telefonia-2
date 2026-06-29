import { useState } from 'react';
import { toast } from 'sonner';

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
  itemTypeLabel,
  perspectiveLabel,
  useCreateBillingCompositionItem,
  useDeleteBillingCompositionItem,
  useEnableEndUserProcessing,
  useLineBillingProcessings,
  useMirrorBillingProcessing
} from '@/lib/billing-processing-api';
import { formatMoney } from '@/lib/financial-api';

type BillingProcessingPanelProps = {
  phoneLineId: string;
  hasActiveCustomer: boolean;
};

export function BillingProcessingPanel({
  phoneLineId,
  hasActiveCustomer
}: BillingProcessingPanelProps) {
  const query = useLineBillingProcessings(phoneLineId, hasActiveCustomer);
  const enableEndUser = useEnableEndUserProcessing(phoneLineId);
  const [activeProcessingId, setActiveProcessingId] = useState<string>('');
  const [itemType, setItemType] = useState('service');
  const [description, setDescription] = useState('');
  const [amount, setAmount] = useState('');

  const processingId =
    activeProcessingId || query.data?.processings[0]?.id || '';
  const createItem = useCreateBillingCompositionItem(phoneLineId, processingId);
  const deleteItem = useDeleteBillingCompositionItem(phoneLineId, processingId);
  const mirror = useMirrorBillingProcessing(phoneLineId, processingId);

  if (!hasActiveCustomer) {
    return (
      <p className="text-muted-foreground text-sm">
        Vincule um cliente para configurar os processamentos financeiros.
      </p>
    );
  }

  if (query.isLoading) {
    return <p className="text-muted-foreground text-sm">Carregando processamentos…</p>;
  }

  const processings = query.data?.processings ?? [];
  const selected =
    processings.find((p) => p.id === processingId) ?? processings[0];
  const hasEndUser = processings.some((p) => p.perspective === 'customer_end_user');

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {processings.map((p) => (
          <Button
            key={p.id}
            type="button"
            size="sm"
            variant={selected?.id === p.id ? 'default' : 'outline'}
            onClick={() => setActiveProcessingId(p.id)}
          >
            {perspectiveLabel(p.perspective)}
            {' · '}
            {formatMoney(p.total_amount)}
          </Button>
        ))}
        {!hasEndUser && (
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={enableEndUser.isPending}
            onClick={() =>
              enableEndUser.mutate(undefined, {
                onSuccess: () => toast.success('2º processamento ativado.'),
                onError: (e) =>
                  toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
              })
            }
          >
            Ativar revenda (2º processamento)
          </Button>
        )}
      </div>

      {selected && (
        <>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-sm font-medium">{perspectiveLabel(selected.perspective)}</p>
            {selected.perspective === 'customer_end_user' && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                disabled={mirror.isPending}
                onClick={() =>
                  mirror.mutate(undefined, {
                    onSuccess: () => toast.success('Composição espelhada do processamento 1.'),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  })
                }
              >
                Espelhar do processamento 1
              </Button>
            )}
          </div>

          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Descrição</TableHead>
                  <TableHead className="text-right">Valor</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {selected.items.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-muted-foreground text-sm">
                      Nenhum item na composição.
                    </TableCell>
                  </TableRow>
                ) : (
                  selected.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{itemTypeLabel(item.item_type)}</TableCell>
                      <TableCell>{item.description}</TableCell>
                      <TableCell className="text-right">
                        {item.item_type === 'discount' ? '−' : ''}
                        {formatMoney(item.amount * (item.quantity || 1))}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          disabled={deleteItem.isPending}
                          onClick={() =>
                            deleteItem.mutate(item.id, {
                              onError: (e) =>
                                toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                            })
                          }
                        >
                          Remover
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-4">
            <Field>
              <FieldLabel>Tipo</FieldLabel>
              <Select value={itemType} onValueChange={(v) => setItemType(v ?? 'service')}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="service">Serviço</SelectItem>
                  <SelectItem value="discount">Desconto</SelectItem>
                  <SelectItem value="extra_charge">Cobrança extra</SelectItem>
                  <SelectItem value="installment">Parcelamento</SelectItem>
                </SelectContent>
              </Select>
            </Field>
            <Field className="sm:col-span-2">
              <FieldLabel>Descrição</FieldLabel>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Valor (R$)</FieldLabel>
              <Input
                inputMode="decimal"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </Field>
          </div>
          <Button
            type="button"
            size="sm"
            disabled={createItem.isPending || !description.trim()}
            onClick={() => {
              const parsed = Number(amount.replace(',', '.'));
              if (!Number.isFinite(parsed) || parsed < 0) {
                toast.error('Informe um valor válido.');
                return;
              }
              createItem.mutate(
                {
                  item_type: itemType,
                  description: description.trim(),
                  amount: parsed
                },
                {
                  onSuccess: () => {
                    setDescription('');
                    setAmount('');
                    toast.success('Item adicionado.');
                  },
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              );
            }}
          >
            Adicionar item
          </Button>
        </>
      )}
    </div>
  );
}
