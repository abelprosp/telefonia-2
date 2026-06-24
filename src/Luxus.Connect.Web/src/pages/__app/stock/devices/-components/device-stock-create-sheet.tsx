import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
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
import { useCreateDeviceStockItem } from '@/lib/device-stock-api';

const formSchema = z.object({
  sku: z.string().max(64, 'SKU muito longo').optional(),
  brand: z.string().min(1, 'Informe a marca').max(128, 'Marca muito longa'),
  model: z.string().min(1, 'Informe o modelo').max(256, 'Modelo muito longo'),
  imei: z.string().optional(),
  color: z.string().optional(),
  storage_capacity: z.string().optional(),
  unit_cost: z.string().optional(),
  sale_price: z.string().optional(),
  notes: z.string().optional()
});

type FormValues = z.infer<typeof formSchema>;

type DeviceStockCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
};

const defaultValues: FormValues = {
  sku: '',
  brand: '',
  model: '',
  imei: '',
  color: '',
  storage_capacity: '',
  unit_cost: '',
  sale_price: '',
  notes: ''
};

function parseMoney(value?: string) {
  const trimmed = value?.trim();
  if (!trimmed) return undefined;
  const normalized = trimmed.replace(/\./g, '').replace(',', '.');
  const n = Number(normalized);
  if (!Number.isFinite(n) || n < 0) return NaN;
  return n;
}

export function DeviceStockCreateSheet({
  open,
  onOpenChange,
  onSuccess
}: DeviceStockCreateSheetProps) {
  const createMutation = useCreateDeviceStockItem();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  useEffect(() => {
    if (!open) form.reset(defaultValues);
  }, [open, form]);

  const onSubmit = form.handleSubmit((values) => {
    const unitCost = parseMoney(values.unit_cost);
    const salePrice = parseMoney(values.sale_price);
    if (Number.isNaN(unitCost) || Number.isNaN(salePrice)) {
      toast.error('Informe valores monetários válidos ou deixe em branco.');
      return;
    }

    createMutation.mutate(
      {
        sku: values.sku?.trim() || null,
        brand: values.brand.trim(),
        model: values.model.trim(),
        imei: values.imei?.replace(/\D/g, '') || null,
        color: values.color?.trim() || null,
        storage_capacity: values.storage_capacity?.trim() || null,
        unit_cost: unitCost ?? null,
        sale_price: salePrice ?? null,
        notes: values.notes?.trim() || null
      },
      {
        onSuccess: () => {
          toast.success('Aparelho cadastrado no estoque.');
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
          <SheetTitle>Cadastrar aparelho</SheetTitle>
          <SheetDescription>
            Registre aparelhos disponíveis para venda. O SKU é gerado automaticamente se não for
            informado.
          </SheetDescription>
        </SheetHeader>

        <form onSubmit={onSubmit} className="flex flex-col gap-4 px-4">
          <Controller
            control={form.control}
            name="brand"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Marca</FieldLabel>
                <Input {...field} placeholder="Ex.: Samsung" />
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="model"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Modelo</FieldLabel>
                <Input {...field} placeholder="Ex.: Galaxy A54" />
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Controller
              control={form.control}
              name="storage_capacity"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Armazenamento</FieldLabel>
                  <Input {...field} placeholder="Ex.: 128GB" />
                </Field>
              )}
            />
            <Controller
              control={form.control}
              name="color"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Cor</FieldLabel>
                  <Input {...field} placeholder="Ex.: Preto" />
                </Field>
              )}
            />
          </div>

          <Controller
            control={form.control}
            name="imei"
            render={({ field }) => (
              <Field>
                <FieldLabel>IMEI (opcional)</FieldLabel>
                <Input {...field} placeholder="15 dígitos" inputMode="numeric" />
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="sku"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>SKU (opcional)</FieldLabel>
                <Input {...field} placeholder="Gerado automaticamente se vazio" />
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Controller
              control={form.control}
              name="unit_cost"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Custo (opcional)</FieldLabel>
                  <Input {...field} placeholder="Ex.: 1.200,00" inputMode="decimal" />
                </Field>
              )}
            />
            <Controller
              control={form.control}
              name="sale_price"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Preço de venda (opcional)</FieldLabel>
                  <Input {...field} placeholder="Ex.: 1.599,00" inputMode="decimal" />
                </Field>
              )}
            />
          </div>

          <Controller
            control={form.control}
            name="notes"
            render={({ field }) => (
              <Field>
                <FieldLabel>Observações</FieldLabel>
                <Textarea {...field} rows={3} placeholder="Condição, acessórios, etc." />
              </Field>
            )}
          />

          <SheetFooter>
            <SheetClose render={<Button type="button" variant="outline" />}>Cancelar</SheetClose>
            <Button type="submit" disabled={createMutation.isPending}>
              Cadastrar
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
