import { useEffect, useMemo, useState } from 'react';

import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import {
  getV1CustomersIdPhoneLinesQueryKey,
  getV1PhoneLinesIdCustomerLinksQueryKey,
  getV1PhoneLinesQueryKey,
  useGetV1Customers,
  useGetV1PhoneLines,
  usePostV1PhoneLinesIdCustomerLinks
} from '@/api';
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
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatPhoneNumber } from '@/lib/format';
import { parseMoneyInput } from '@/lib/phone-line-api';

type LinkCustomerLineSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
} & (
  | {
      mode: 'line-to-customer';
      phoneLineId: string;
      phoneLineNumber?: string;
    }
  | {
      mode: 'customer-to-line';
      customerId: string;
      customerName?: string;
    }
);

export function LinkCustomerLineSheet(props: LinkCustomerLineSheetProps) {
  const { open, onOpenChange, onSuccess, mode } = props;
  const queryClient = useQueryClient();

  const [selectedCustomerId, setSelectedCustomerId] = useState('');
  const [selectedLineId, setSelectedLineId] = useState('');
  const [effectiveDate, setEffectiveDate] = useState('');
  const [monthlyAmount, setMonthlyAmount] = useState('');

  const customersQuery = useGetV1Customers({
    page_index: 0,
    page_size: 100
  });

  const stockLinesQuery = useGetV1PhoneLines({
    page_index: 0,
    page_size: 100,
    status: 'in_stock'
  });

  const customersOptions = useMemo(
    () => (customersQuery.data?.items ?? []).filter((c) => c.active),
    [customersQuery.data?.items]
  );

  const lineOptions = useMemo(
    () => stockLinesQuery.data?.items ?? [],
    [stockLinesQuery.data?.items]
  );

  useEffect(() => {
    if (!open) {
      setSelectedCustomerId('');
      setSelectedLineId('');
      setEffectiveDate('');
      setMonthlyAmount('');
      return;
    }
    if (mode === 'customer-to-line') {
      setSelectedCustomerId(props.customerId);
    }
  }, [open, mode, mode === 'customer-to-line' ? props.customerId : null]);

  const assignMutation = usePostV1PhoneLinesIdCustomerLinks({
    mutation: {
      onSuccess: async (_data, variables) => {
        toast.success(
          mode === 'customer-to-line'
            ? 'Linha vinculada ao cliente.'
            : 'Cliente vinculado à linha.'
        );
        onOpenChange(false);
        onSuccess?.();
        await queryClient.invalidateQueries({
          queryKey: getV1PhoneLinesIdCustomerLinksQueryKey(variables.id)
        });
        await queryClient.invalidateQueries({ queryKey: getV1PhoneLinesQueryKey() });
        if (mode === 'customer-to-line') {
          await queryClient.invalidateQueries({
            queryKey: getV1CustomersIdPhoneLinesQueryKey(props.customerId, {
              page_index: 0,
              page_size: 50
            })
          });
        }
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const phoneLineId =
    mode === 'line-to-customer' ? props.phoneLineId : selectedLineId;
  const customerId =
    mode === 'customer-to-line' ? props.customerId : selectedCustomerId;

  const canSubmit =
    Boolean(phoneLineId) &&
    Boolean(customerId) &&
    !assignMutation.isPending &&
    (mode !== 'line-to-customer' || customersOptions.length > 0) &&
    (mode !== 'customer-to-line' || lineOptions.length > 0);

  const title =
    mode === 'customer-to-line' ? 'Vincular linha ao cliente' : 'Vincular cliente à linha';

  const description =
    mode === 'customer-to-line'
      ? `Selecione uma linha em estoque para vincular${props.customerName ? ` a ${props.customerName}` : ''}.`
      : `Associe um cliente à linha${props.phoneLineNumber ? ` ${formatPhoneNumber(props.phoneLineNumber) ?? props.phoneLineNumber}` : ''}.`;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>{title}</SheetTitle>
          <SheetDescription>{description}</SheetDescription>
        </SheetHeader>

        <div className="space-y-4 py-4">
          {mode === 'line-to-customer' ? (
            <Field>
              <FieldLabel>Cliente</FieldLabel>
              {customersQuery.isPending ? (
                <p className="text-muted-foreground text-sm">Carregando clientes...</p>
              ) : customersOptions.length === 0 ? (
                <p className="text-muted-foreground text-sm">
                  Nenhum cliente ativo encontrado. Cadastre ou reative um cliente primeiro.
                </p>
              ) : (
                <Select
                  value={selectedCustomerId}
                  onValueChange={(value) => setSelectedCustomerId(value ?? '')}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Selecione um cliente" />
                  </SelectTrigger>
                  <SelectContent>
                    {customersOptions.map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </Field>
          ) : (
            <Field>
              <FieldLabel>Linha em estoque</FieldLabel>
              {stockLinesQuery.isPending ? (
                <p className="text-muted-foreground text-sm">Carregando linhas...</p>
              ) : lineOptions.length === 0 ? (
                <p className="text-muted-foreground text-sm">
                  Nenhuma linha em estoque disponível. Cadastre ou libere uma linha no estoque.
                </p>
              ) : (
                <Select
                  value={selectedLineId}
                  onValueChange={(value) => setSelectedLineId(value ?? '')}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Selecione uma linha" />
                  </SelectTrigger>
                  <SelectContent>
                    {lineOptions.map((line) => (
                      <SelectItem key={line.id} value={line.id}>
                        {formatPhoneNumber(line.number) ?? line.number}
                        {line.provider_plan_name ? ` · ${line.provider_plan_name}` : ''}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </Field>
          )}

          <Field>
            <FieldLabel>Valor mensal cobrado do cliente (R$)</FieldLabel>
            <Input
              inputMode="decimal"
              placeholder="0,00"
              value={monthlyAmount}
              onChange={(e) => setMonthlyAmount(e.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              Valor de refaturamento mensal desta linha para o cliente.
            </p>
          </Field>

          <Field>
            <FieldLabel>Data de início (opcional)</FieldLabel>
            <Input
              type="date"
              value={effectiveDate}
              onChange={(e) => setEffectiveDate(e.target.value)}
            />
          </Field>
        </div>

        <SheetFooter className="gap-2">
          <SheetClose render={<Button variant="outline" />}>Cancelar</SheetClose>
          <Button
            disabled={!canSubmit}
            onClick={() =>
              assignMutation.mutate({
                id: phoneLineId,
                data: {
                  customer_id: customerId,
                  start_date: effectiveDate || null,
                  monthly_amount: parseMoneyInput(monthlyAmount)
                } as { customer_id: string; start_date: string | null; monthly_amount?: number | null }
              })
            }
          >
            Confirmar vínculo
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
