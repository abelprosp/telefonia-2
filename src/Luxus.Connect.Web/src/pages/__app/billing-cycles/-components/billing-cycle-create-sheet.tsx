import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import {
  getV1BillingCyclesQueryKey,
  type ListProvidersResponse,
  useGetV1Providers,
  usePostV1BillingCycles
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
import { invalidateDashboardCaches } from '@/lib/query-utils';

const PROVIDER_LIST_PAGE_SIZE = 500;

const formSchema = z
  .object({
    providerId: z.string().min(1, 'Selecione a operadora'),
    code: z.string().min(1, 'Informe o código').max(64, 'Código muito longo'),
    name: z.string().min(1, 'Informe o nome').max(200, 'Nome muito longo'),
    start_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Use a data completa'),
    end_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Use a data completa')
  })
  .refine(
    (d) => {
      const a = new Date(`${d.start_date}T12:00:00`);
      const b = new Date(`${d.end_date}T12:00:00`);
      return !Number.isNaN(a.getTime()) && !Number.isNaN(b.getTime()) && b >= a;
    },
    {
      message: 'A data de fim deve ser igual ou posterior à de início',
      path: ['end_date']
    }
  );

type FormValues = z.infer<typeof formSchema>;

type BillingCycleCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

const defaultValues: FormValues = {
  providerId: '',
  code: '',
  name: '',
  start_date: '',
  end_date: ''
};

export function BillingCycleCreateSheet({
  open,
  onOpenChange
}: BillingCycleCreateSheetProps) {
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

  const createMutation = usePostV1BillingCycles({
    mutation: {
      onSuccess: () => {
        toast.success('Ciclo de faturamento cadastrado.');
        void queryClient.invalidateQueries({
          queryKey: getV1BillingCyclesQueryKey()
        });
        void invalidateDashboardCaches(queryClient);
        onOpenChange(false);
        form.reset(defaultValues);
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
        code: values.code.trim(),
        name: values.name.trim(),
        start_date: values.start_date,
        end_date: values.end_date
      }
    })
  );

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          form.reset(defaultValues);
        }
        onOpenChange(next);
      }}
    >
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>Novo ciclo de faturamento</SheetTitle>
          <SheetDescription>
            Informe código, nome e intervalo de datas do ciclo na operadora.
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

            <Controller
              control={form.control}
              name="code"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="billing-cycle-create-code">
                    Código
                  </FieldLabel>
                  <Input
                    id="billing-cycle-create-code"
                    autoComplete="off"
                    className="border-input bg-background rounded-xl border text-sm"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="billing-cycle-create-name">
                    Nome
                  </FieldLabel>
                  <Input
                    id="billing-cycle-create-name"
                    autoComplete="off"
                    placeholder="Ex.: Ciclo março/2026"
                    className="border-input bg-background rounded-xl border"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="start_date"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="billing-cycle-create-start">
                    Data de início
                  </FieldLabel>
                  <Input
                    id="billing-cycle-create-start"
                    type="date"
                    className="border-input bg-background rounded-xl border"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="end_date"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="billing-cycle-create-end">
                    Data de fim
                  </FieldLabel>
                  <Input
                    id="billing-cycle-create-end"
                    type="date"
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
