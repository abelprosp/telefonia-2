import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { withMask } from 'use-mask-input';
import { z } from 'zod';

import { useGetV1Providers, useProvidersControllerGetById } from '@/api';
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
import { useCreateStockPhoneLine } from '@/lib/stock-api';

const formSchema = z.object({
  providerId: z.string().min(1, 'Selecione a operadora'),
  providerAccountNumber: z.string().min(1, 'Informe o número da conta'),
  providerPlanId: z.string().min(1, 'Selecione o plano'),
  number: z.string().min(10, 'Informe o número da linha')
});

type FormValues = z.infer<typeof formSchema>;

type StockLineCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
};

export function StockLineCreateSheet({ open, onOpenChange, onSuccess }: StockLineCreateSheetProps) {
  const createMutation = useCreateStockPhoneLine();
  const providersQuery = useGetV1Providers({ page_index: 0, page_size: 500 });

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      providerId: '',
      providerAccountNumber: '',
      providerPlanId: '',
      number: ''
    }
  });

  const providerId = form.watch('providerId');
  const providerDetailQuery = useProvidersControllerGetById(providerId, {
    query: { enabled: Boolean(providerId) }
  });

  useEffect(() => {
    form.setValue('providerPlanId', '');
  }, [providerId, form]);

  useEffect(() => {
    if (!open) {
      form.reset();
    }
  }, [open, form]);

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate(
      {
        number: values.number.replace(/\D/g, ''),
        provider_id: values.providerId,
        provider_account_number: values.providerAccountNumber.trim(),
        provider_plan_id: values.providerPlanId
      },
      {
        onSuccess: () => {
          toast.success('Linha cadastrada no estoque.');
          onOpenChange(false);
          onSuccess?.();
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  });

  const plans = providerDetailQuery.data?.plans ?? [];

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="overflow-y-auto sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>Cadastrar linha no estoque</SheetTitle>
          <SheetDescription>
            Informe operadora, conta, plano e número. A conta precisa existir no sistema (geralmente
            criada ao importar a primeira fatura da operadora).
          </SheetDescription>
        </SheetHeader>

        <form onSubmit={onSubmit} className="flex flex-col gap-4 px-4">
          <Controller
            control={form.control}
            name="providerId"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Operadora</FieldLabel>
                <Select value={field.value} onValueChange={(v) => field.onChange(v ?? '')}>
                  <SelectTrigger>
                    <SelectValue placeholder="Selecione" />
                  </SelectTrigger>
                  <SelectContent>
                    {(providersQuery.data?.items ?? []).map((p) => (
                      <SelectItem key={p.id} value={p.id}>
                        {p.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="providerAccountNumber"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Número da conta operadora</FieldLabel>
                <Input {...field} placeholder="Ex.: 123456789" />
                <FieldError>{fieldState.error?.message}</FieldError>
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="providerPlanId"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Plano</FieldLabel>
                <Select
                  value={field.value}
                  onValueChange={(v) => field.onChange(v ?? '')}
                  disabled={!providerId || plans.length === 0}
                >
                  <SelectTrigger>
                    <SelectValue
                      placeholder={
                        !providerId
                          ? 'Selecione a operadora primeiro'
                          : plans.length === 0
                            ? 'Nenhum plano cadastrado'
                            : 'Selecione o plano'
                      }
                    />
                  </SelectTrigger>
                  <SelectContent>
                    {plans.map((plan) => (
                      <SelectItem key={plan.id} value={plan.id}>
                        {plan.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldError>{fieldState.error?.message}</FieldError>
                {providerId && plans.length === 0 && !providerDetailQuery.isPending && (
                  <p className="text-muted-foreground text-xs">
                    Nenhum plano nesta operadora.{' '}
                    <Link
                      to="/providers/$providerId"
                      params={{ providerId }}
                      search={{ page: 1, pageSize: 10 }}
                      className="text-primary font-medium hover:underline"
                    >
                      Cadastre um plano
                    </Link>{' '}
                    antes de incluir a linha no estoque.
                  </p>
                )}
              </Field>
            )}
          />

          <Controller
            control={form.control}
            name="number"
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel>Número da linha</FieldLabel>
                <Input
                  {...field}
                  ref={withMask('(99) 99999-9999', {
                    showMaskOnHover: false,
                    showMaskOnFocus: false
                  })}
                  placeholder="(11) 99999-9999"
                />
                <FieldError>{fieldState.error?.message}</FieldError>
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
