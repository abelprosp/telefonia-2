import { useEffect, useState } from 'react';
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
  useMirrorBillingProcessing,
  usePayoffBillingCompositionItem,
  useUpdateBillingProcessing
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
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [serviceType, setServiceType] = useState('subscription');
  const [proportional, setProportional] = useState(true);
  const [installmentCount, setInstallmentCount] = useState('12');

  const processingId =
    activeProcessingId || query.data?.processings[0]?.id || '';
  const createItem = useCreateBillingCompositionItem(phoneLineId, processingId);
  const deleteItem = useDeleteBillingCompositionItem(phoneLineId, processingId);
  const payoffItem = usePayoffBillingCompositionItem(phoneLineId, processingId);
  const mirror = useMirrorBillingProcessing(phoneLineId, processingId);
  const updateProcessing = useUpdateBillingProcessing(phoneLineId, processingId);
  const [label, setLabel] = useState('');
  const [orgUnit, setOrgUnit] = useState('');
  const [department, setDepartment] = useState('');
  const [costCenter, setCostCenter] = useState('');

  const processings = query.data?.processings ?? [];
  const selected =
    processings.find((p) => p.id === processingId) ?? processings[0];
  const hasEndUser = processings.some((p) => p.perspective === 'customer_end_user');

  useEffect(() => {
    if (!selected) return;
    setLabel(selected.label ?? '');
    setOrgUnit(selected.organizational_unit ?? '');
    setDepartment(selected.department ?? '');
    setCostCenter(selected.cost_center_label ?? '');
  }, [selected?.id, selected?.label, selected?.organizational_unit, selected?.department, selected?.cost_center_label]);

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
            <p className="text-sm font-medium">
              {perspectiveLabel(selected.perspective)}
              {selected.label ? ` · ${selected.label}` : ''}
            </p>
            {selected.perspective === 'customer_end_user' && (
              <div className="flex flex-wrap gap-2">
                <Button
                  type="button"
                  size="sm"
                  variant={selected.mirror_from_primary ? 'default' : 'outline'}
                  disabled={updateProcessing.isPending}
                  onClick={() =>
                    updateProcessing.mutate(
                      { mirror_from_primary: !selected.mirror_from_primary },
                      {
                        onSuccess: () =>
                          toast.success(
                            selected.mirror_from_primary
                              ? 'Espelhamento desativado.'
                              : 'Espelhamento ativado.'
                          ),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  {selected.mirror_from_primary ? 'Espelhamento ligado' : 'Espelhar valores'}
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  disabled={mirror.isPending}
                  onClick={() =>
                    mirror.mutate(undefined, {
                      onSuccess: () => toast.success('Composição copiada do processamento 1.'),
                      onError: (e) =>
                        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                    })
                  }
                >
                  Copiar do processamento 1
                </Button>
              </div>
            )}
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-4">
            <Field>
              <FieldLabel>
                {selected.perspective === 'customer_end_user'
                  ? 'Rótulo (usuário final)'
                  : 'Rótulo'}
              </FieldLabel>
              <Input value={label} onChange={(e) => setLabel(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>UA</FieldLabel>
              <Input value={orgUnit} onChange={(e) => setOrgUnit(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Setor</FieldLabel>
              <Input value={department} onChange={(e) => setDepartment(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Centro de custo</FieldLabel>
              <Input value={costCenter} onChange={(e) => setCostCenter(e.target.value)} />
            </Field>
          </div>
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={updateProcessing.isPending}
            onClick={() =>
              updateProcessing.mutate(
                {
                  label,
                  organizational_unit: orgUnit,
                  department,
                  cost_center_label: costCenter
                },
                {
                  onSuccess: () => toast.success('Classificadores salvos.'),
                  onError: (e) =>
                    toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                }
              )
            }
          >
            Salvar rótulo e classificadores
          </Button>

          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Descrição</TableHead>
                  <TableHead>Vigência</TableHead>
                  <TableHead className="text-right">Valor</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {selected.items.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-muted-foreground text-sm">
                      Nenhum item na composição.
                    </TableCell>
                  </TableRow>
                ) : (
                  selected.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{itemTypeLabel(item.item_type)}</TableCell>
                      <TableCell>
                        {item.description}
                        {item.proportional === false ? (
                          <span className="text-muted-foreground ml-1 text-xs">(sem pró-rata)</span>
                        ) : null}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {(item.start_date ?? '').toString().slice(0, 10) || '—'}
                        {' → '}
                        {(item.end_date ?? '').toString().slice(0, 10) || 'aberto'}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.item_type === 'discount' ? '−' : ''}
                        {formatMoney(item.amount * (item.quantity || 1))}
                        {item.item_type === 'installment' && item.installment_count
                          ? ` · ${item.installment_current ?? 1}/${item.installment_count}`
                          : ''}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.item_type === 'installment' ? (
                          <Button
                            type="button"
                            size="sm"
                            variant="outline"
                            disabled={payoffItem.isPending}
                            onClick={() =>
                              payoffItem.mutate(item.id, {
                                onSuccess: () => toast.success('Quitação das parcelas restantes.'),
                                onError: (e) =>
                                  toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                              })
                            }
                          >
                            Quitar
                          </Button>
                        ) : null}
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
                  <SelectItem value="extra_charge">Cobrança avulsa</SelectItem>
                  <SelectItem value="installment">Parcelamento aparelho</SelectItem>
                  <SelectItem value="exceedance">Excedente (manual)</SelectItem>
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
            {itemType === 'installment' ? (
              <Field>
                <FieldLabel>Parcelas</FieldLabel>
                <Input
                  inputMode="numeric"
                  value={installmentCount}
                  onChange={(e) => setInstallmentCount(e.target.value)}
                />
              </Field>
            ) : null}
            {itemType === 'service' ? (
              <Field>
                <FieldLabel>Tipo de serviço</FieldLabel>
                <Select value={serviceType} onValueChange={(v) => setServiceType(v ?? 'subscription')}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="subscription">Assinatura</SelectItem>
                    <SelectItem value="data">Dados</SelectItem>
                    <SelectItem value="sms">SMS</SelectItem>
                    <SelectItem value="roaming">Roaming</SelectItem>
                    <SelectItem value="other">Outros</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
            ) : null}
            <Field>
              <FieldLabel>Início</FieldLabel>
              <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel>Fim</FieldLabel>
              <Input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
            </Field>
            <Field>
              <FieldLabel className="flex items-center gap-2">
                <input
                  type="checkbox"
                  className="size-4 rounded border"
                  checked={proportional}
                  onChange={(e) => setProportional(e.target.checked)}
                />
                Aceita proporcionalidade
              </FieldLabel>
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
                  amount: parsed,
                  proportional,
                  ...(itemType === 'service' ? { service_type: serviceType } : {}),
                  ...(itemType === 'installment'
                    ? { installment_count: Number(installmentCount) || 1, installment_current: 1 }
                    : {}),
                  ...(startDate ? { start_date: startDate } : {}),
                  ...(endDate ? { end_date: endDate } : {})
                },
                {
                  onSuccess: () => {
                    setDescription('');
                    setAmount('');
                    setStartDate('');
                    setEndDate('');
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
