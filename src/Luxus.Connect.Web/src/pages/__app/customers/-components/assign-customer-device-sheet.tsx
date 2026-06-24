import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
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
import { useAssignCustomerDevice } from '@/lib/customer-devices-api';
import { useDeviceStockList } from '@/lib/device-stock-api';

const formSchema = z.object({
  source: z.enum(['stock', 'manual']),
  deviceStockItemId: z.string().optional(),
  brand: z.string().optional(),
  model: z.string().optional(),
  description: z.string().optional(),
  monthly_amount: z.string().min(1, 'Informe o valor mensal')
});

type FormValues = z.infer<typeof formSchema>;

type AssignCustomerDeviceSheetProps = {
  customerId: string;
  customerName?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
};

const defaultValues: FormValues = {
  source: 'stock',
  deviceStockItemId: '',
  brand: '',
  model: '',
  description: '',
  monthly_amount: ''
};

function parseMoney(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return NaN;
  const normalized = trimmed.replace(/\./g, '').replace(',', '.');
  const n = Number(normalized);
  return Number.isFinite(n) && n >= 0 ? n : NaN;
}

export function AssignCustomerDeviceSheet({
  customerId,
  customerName,
  open,
  onOpenChange,
  onSuccess
}: AssignCustomerDeviceSheetProps) {
  const assignMutation = useAssignCustomerDevice(customerId);
  const stockQuery = useDeviceStockList({ page_index: 0, page_size: 200, status: 'in_stock' });

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  const source = form.watch('source');
  const deviceStockItemId = form.watch('deviceStockItemId');
  const stockItems = stockQuery.data?.items ?? [];

  useEffect(() => {
    if (!open) form.reset(defaultValues);
  }, [open, form]);

  useEffect(() => {
    if (source !== 'stock' || !deviceStockItemId) return;
    const device = stockItems.find((d) => d.id === deviceStockItemId);
    if (!device) return;
    if (device.sale_price != null) {
      form.setValue('monthly_amount', device.sale_price.toLocaleString('pt-BR', { minimumFractionDigits: 2 }));
    }
    form.setValue('description', `Aparelho ${device.brand} ${device.model}${device.storage_capacity ? ` ${device.storage_capacity}` : ''}`);
  }, [deviceStockItemId, form, source, stockItems]);

  const onSubmit = form.handleSubmit((values) => {
    const monthlyAmount = parseMoney(values.monthly_amount);
    if (Number.isNaN(monthlyAmount)) {
      toast.error('Informe um valor mensal válido.');
      return;
    }

    if (values.source === 'stock' && !values.deviceStockItemId) {
      toast.error('Selecione um aparelho do estoque.');
      return;
    }
    if (values.source === 'manual' && (!values.brand?.trim() || !values.model?.trim())) {
      toast.error('Informe marca e modelo.');
      return;
    }

    assignMutation.mutate(
      {
        device_stock_item_id: values.source === 'stock' ? values.deviceStockItemId : null,
        brand: values.source === 'manual' ? values.brand?.trim() : null,
        model: values.source === 'manual' ? values.model?.trim() : null,
        description: values.description?.trim() || null,
        monthly_amount: monthlyAmount
      },
      {
        onSuccess: () => {
          toast.success('Aparelho vinculado ao cliente.');
          onOpenChange(false);
          onSuccess?.();
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="overflow-y-auto sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>Vincular aparelho</SheetTitle>
          <SheetDescription>
            Cadastre um aparelho para cobrança na fatura{customerName ? ` de ${customerName}` : ''}.
          </SheetDescription>
        </SheetHeader>

        <form onSubmit={onSubmit} className="flex flex-col gap-4 px-4">
          <Controller
            control={form.control}
            name="source"
            render={({ field }) => (
              <Field>
                <FieldLabel>Origem</FieldLabel>
                <Select value={field.value} onValueChange={(v) => field.onChange(v ?? 'stock')}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="stock">Do estoque</SelectItem>
                    <SelectItem value="manual">Cadastro manual</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
            )}
          />

          {source === 'stock' ? (
            <Controller
              control={form.control}
              name="deviceStockItemId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Aparelho em estoque</FieldLabel>
                  <Select value={field.value} onValueChange={(v) => field.onChange(v ?? '')}>
                    <SelectTrigger>
                      <SelectValue placeholder={stockItems.length === 0 ? 'Nenhum aparelho disponível' : 'Selecione'} />
                    </SelectTrigger>
                    <SelectContent>
                      {stockItems.map((d) => (
                        <SelectItem key={d.id} value={d.id}>
                          {d.brand} {d.model}
                          {d.storage_capacity ? ` · ${d.storage_capacity}` : ''}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </Field>
              )}
            />
          ) : (
            <>
              <Controller
                control={form.control}
                name="brand"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Marca</FieldLabel>
                    <Input {...field} placeholder="Ex.: Samsung" />
                  </Field>
                )}
              />
              <Controller
                control={form.control}
                name="model"
                render={({ field }) => (
                  <Field>
                    <FieldLabel>Modelo</FieldLabel>
                    <Input {...field} placeholder="Ex.: Galaxy A54" />
                  </Field>
                )}
              />
            </>
          )}

          <Controller
            control={form.control}
            name="description"
            render={({ field }) => (
              <Field>
                <FieldLabel>Descrição na fatura</FieldLabel>
                <Input {...field} placeholder="Texto exibido na fatura" />
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="monthly_amount"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Valor mensal na fatura</FieldLabel>
                <Input {...field} placeholder="Ex.: 150,00" inputMode="decimal" />
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <SheetFooter>
            <SheetClose render={<Button type="button" variant="outline" />}>Cancelar</SheetClose>
            <Button type="submit" disabled={assignMutation.isPending}>
              Vincular
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
