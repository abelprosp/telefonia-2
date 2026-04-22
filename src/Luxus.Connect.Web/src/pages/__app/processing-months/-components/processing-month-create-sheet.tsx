import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Controller, useForm, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import {
  getV1ProcessingMonthsQueryKey,
  type ListProvidersResponse,
  useGetV1Providers,
  usePostV1ProcessingMonths
} from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectGroup,
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

const PROVIDER_LIST_PAGE_SIZE = 500;

const formSchema = z.object({
  providerId: z.string().min(1, 'Selecione a operadora'),
  year: z.number().int().min(2000, 'Ano inválido').max(2100, 'Ano inválido'),
  month: z.number().int().min(1, 'Mês inválido').max(12, 'Mês inválido'),
  display_name: z
    .string()
    .min(1, 'Informe o nome de exibição')
    .max(200, 'Nome muito longo')
});

type FormValues = z.infer<typeof formSchema>;

type ProcessingMonthCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function defaultDisplayName(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

const defaultYear = new Date().getFullYear();
const defaultMonth = new Date().getMonth() + 1;

const defaultValues: FormValues = {
  providerId: '',
  year: defaultYear,
  month: defaultMonth,
  display_name: defaultDisplayName(defaultYear, defaultMonth)
};

export function ProcessingMonthCreateSheet({
  open,
  onOpenChange
}: ProcessingMonthCreateSheetProps) {
  const queryClient = useQueryClient();

  const providersQuery = useGetV1Providers(
    {
      page_index: 0,
      page_size: PROVIDER_LIST_PAGE_SIZE
    },
    {
      query: { enabled: open }
    }
  );

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  const watchedYear = useWatch({ control: form.control, name: 'year' });
  const watchedMonth = useWatch({ control: form.control, name: 'month' });

  const createMutation = usePostV1ProcessingMonths({
    mutation: {
      onSuccess: () => {
        toast.success('Mês de processamento criado.');
        void queryClient.invalidateQueries({
          queryKey: getV1ProcessingMonthsQueryKey()
        });
        onOpenChange(false);
        form.reset({
          ...defaultValues,
          year: new Date().getFullYear(),
          month: new Date().getMonth() + 1,
          display_name: defaultDisplayName(
            new Date().getFullYear(),
            new Date().getMonth() + 1
          )
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const onSubmit = form.handleSubmit((values) =>
    createMutation.mutate({
      data: {
        provider_id: values.providerId,
        year: values.year,
        month: values.month,
        display_name: values.display_name.trim()
      }
    })
  );

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          form.reset({
            ...defaultValues,
            display_name: defaultDisplayName(
              watchedYear ?? defaultValues.year,
              watchedMonth ?? defaultValues.month
            )
          });
        }
        onOpenChange(next);
      }}
    >
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>Novo mês de processamento</SheetTitle>
          <SheetDescription>
            Registre a competência (ano/mês) na operadora. O fechamento é feito
            depois, nesta tela ou no detalhe do mês.
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
            <Controller
              control={form.control}
              name="providerId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Operadora</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={field.onChange}
                    disabled={providersQuery.isPending}
                  >
                    <SelectTrigger className="border-input bg-background w-full max-w-none rounded-xl border">
                      <SelectValue placeholder="Selecione">
                        {(providersQuery.data?.items ?? []).find(
                          (p: ListProvidersResponse) => p.id === field.value
                        )?.name ?? 'Selecione'}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {(providersQuery.data?.items ?? []).map(
                          (p: ListProvidersResponse) => (
                            <SelectItem key={p.id} value={p.id}>
                              {p.name}
                            </SelectItem>
                          )
                        )}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={form.control}
                name="year"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="pm-create-year">Ano</FieldLabel>
                    <Input
                      id="pm-create-year"
                      type="number"
                      className="border-input bg-background rounded-xl border"
                      name={field.name}
                      ref={field.ref}
                      onBlur={field.onBlur}
                      value={Number.isFinite(field.value) ? field.value : ''}
                      onChange={(e) => {
                        const raw = e.target.value;
                        const y = raw === '' ? NaN : Number(raw);
                        field.onChange(y);
                        const mo = form.getValues('month');
                        if (
                          Number.isFinite(y) &&
                          Number.isFinite(mo) &&
                          mo >= 1 &&
                          mo <= 12
                        ) {
                          form.setValue(
                            'display_name',
                            defaultDisplayName(y, mo),
                            { shouldValidate: true }
                          );
                        }
                      }}
                    />
                    {fieldState.invalid ? (
                      <FieldError errors={[fieldState.error]} />
                    ) : null}
                  </Field>
                )}
              />
              <Controller
                control={form.control}
                name="month"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="pm-create-month">Mês</FieldLabel>
                    <Input
                      id="pm-create-month"
                      type="number"
                      min={1}
                      max={12}
                      className="border-input bg-background rounded-xl border"
                      name={field.name}
                      ref={field.ref}
                      onBlur={field.onBlur}
                      value={Number.isFinite(field.value) ? field.value : ''}
                      onChange={(e) => {
                        const raw = e.target.value;
                        const mo = raw === '' ? NaN : Number(raw);
                        field.onChange(mo);
                        const y = form.getValues('year');
                        if (
                          Number.isFinite(y) &&
                          Number.isFinite(mo) &&
                          mo >= 1 &&
                          mo <= 12
                        ) {
                          form.setValue(
                            'display_name',
                            defaultDisplayName(y, mo),
                            { shouldValidate: true }
                          );
                        }
                      }}
                    />
                    {fieldState.invalid ? (
                      <FieldError errors={[fieldState.error]} />
                    ) : null}
                  </Field>
                )}
              />
            </div>

            <Controller
              control={form.control}
              name="display_name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="pm-create-display">
                    Nome de exibição
                  </FieldLabel>
                  <Input
                    id="pm-create-display"
                    autoComplete="off"
                    placeholder="Ex.: 04/2026"
                    className="border-input bg-background rounded-xl border"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />
          </div>

          <SheetFooter className="gap-4 border-t pt-6 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Salvando…' : 'Cadastrar'}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
