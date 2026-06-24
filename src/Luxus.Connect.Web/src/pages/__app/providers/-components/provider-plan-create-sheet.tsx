import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
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
import { useCreateProviderPlan } from '@/lib/providers-api';

const formSchema = z.object({
  code: z
    .string()
    .min(1, 'Informe o código')
    .max(64, 'Código deve ter no máximo 64 caracteres'),
  name: z
    .string()
    .min(1, 'Informe o nome')
    .max(256, 'Nome deve ter no máximo 256 caracteres'),
  monthly_price: z.string().optional()
});

type FormValues = z.infer<typeof formSchema>;

type ProviderPlanCreateSheetProps = {
  providerId: string;
  providerName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (planId: string) => void;
};

const defaultValues: FormValues = {
  code: '',
  name: '',
  monthly_price: ''
};

function parseMonthlyPrice(value?: string) {
  const trimmed = value?.trim();
  if (!trimmed) return undefined;
  const normalized = trimmed.replace(/\./g, '').replace(',', '.');
  const n = Number(normalized);
  if (!Number.isFinite(n) || n < 0) {
    return NaN;
  }
  return n;
}

export function ProviderPlanCreateSheet({
  providerId,
  providerName,
  open,
  onOpenChange,
  onSuccess
}: ProviderPlanCreateSheetProps) {
  const createMutation = useCreateProviderPlan(providerId);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  useEffect(() => {
    if (!open) {
      form.reset(defaultValues);
    }
  }, [open, form]);

  const onSubmit = form.handleSubmit((values) => {
    const monthlyPrice = parseMonthlyPrice(values.monthly_price);
    if (Number.isNaN(monthlyPrice)) {
      toast.error('Informe um valor mensal válido ou deixe em branco.');
      return;
    }

    createMutation.mutate(
      {
        code: values.code.trim(),
        name: values.name.trim(),
        monthly_price: monthlyPrice ?? null
      },
      {
        onSuccess: (plan) => {
          toast.success('Plano cadastrado.');
          onOpenChange(false);
          onSuccess?.(plan.id);
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>Novo plano</SheetTitle>
          <SheetDescription>
            Cadastre um plano em {providerName} para vincular às linhas do estoque e às vendas.
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
            <Controller
              control={form.control}
              name="code"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Código</FieldLabel>
                  <Input {...field} placeholder="Ex.: SMART-20GB" autoComplete="off" />
                  <FieldError>{fieldState.error?.message}</FieldError>
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Nome</FieldLabel>
                  <Input {...field} placeholder="Ex.: Plano Smart 20GB" />
                  <FieldError>{fieldState.error?.message}</FieldError>
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="monthly_price"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Valor mensal de venda (opcional)</FieldLabel>
                  <Input {...field} placeholder="Ex.: 89,90" inputMode="decimal" />
                  <FieldError>{fieldState.error?.message}</FieldError>
                </Field>
              )}
            />
          </div>

          <SheetFooter className="gap-4 border-t pt-6 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>Cancelar</SheetClose>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Salvando…' : 'Cadastrar plano'}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
