import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { format } from 'date-fns';
import { Controller, useForm, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { withMask } from 'use-mask-input';
import { z } from 'zod';

import {
  getV1CustomersQueryKey,
  type ListProvidersResponse,
  useGetV1Providers,
  usePostV1Customers
} from '@/api';
import { Button } from '@/components/ui/button';
import { DatePicker } from '@/components/ui/date-picker';
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

const PROVIDERS_PAGE_SIZE = 500;

const formSchema = z
  .object({
    providerId: z.string().min(1, 'Selecione a operadora inicial'),
    type: z.enum(['PF', 'PJ']),
    name: z.string().min(1, 'Informe o nome').max(256, 'Nome muito longo'),
    legal_name: z.string().optional(),
    document: z
      .string()
      .min(11, 'Documento inválido')
      .max(20, 'Documento inválido'),
    state_registration: z.string().optional(),
    birth_or_opening_date: z.string().optional(),
    responsible_salesperson_user_id: z
      .string()
      .max(256, 'Identificador muito longo')
      .optional()
  })
  .superRefine((data, ctx) => {
    if (data.type === 'PJ' && !data.legal_name?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Razão social é obrigatória para PJ',
        path: ['legal_name']
      });
    }
  });

type FormValues = z.infer<typeof formSchema>;

type CustomerCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  preferredProviderId?: string;
};

const defaultValues: FormValues = {
  providerId: '',
  type: 'PJ',
  name: '',
  legal_name: '',
  document: '',
  state_registration: '',
  birth_or_opening_date: '',
  responsible_salesperson_user_id: ''
};

export function CustomerCreateSheet({
  open,
  onOpenChange,
  preferredProviderId
}: CustomerCreateSheetProps) {
  const queryClient = useQueryClient();

  const providersQuery = useGetV1Providers(
    { page_index: 0, page_size: PROVIDERS_PAGE_SIZE },
    { query: { enabled: open } }
  );

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });
  const customerType =
    useWatch({ control: form.control, name: 'type' }) ?? 'PJ';

  useEffect(() => {
    if (open && preferredProviderId && preferredProviderId.length > 0) {
      form.setValue('providerId', preferredProviderId);
    }
  }, [form, open, preferredProviderId]);

  const createMutation = usePostV1Customers({
    mutation: {
      onSuccess: async () => {
        toast.success('Cliente cadastrado.');
        await queryClient.invalidateQueries({
          queryKey: getV1CustomersQueryKey()
        });
        await invalidateDashboardCaches(queryClient);
        onOpenChange(false);
        form.reset(defaultValues);
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        provider_id: values.providerId,
        type: values.type,
        name: values.name.trim(),
        legal_name: values.legal_name?.trim() || null,
        document: values.document.trim(),
        state_registration: values.state_registration?.trim() || null,
        birth_or_opening_date: values.birth_or_opening_date?.trim() || null,
        responsible_salesperson_user_id:
          values.responsible_salesperson_user_id?.trim() || null,
        addresses: []
      }
    });
  });

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
          <SheetTitle>Novo cliente</SheetTitle>
          <SheetDescription>
            Cadastre o cliente com operadora inicial, tipo e documento.
            Opcional: identificador do vendedor responsável (ex.: sub do
            Keycloak).
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
            <Controller
              control={form.control}
              name="providerId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Operadora inicial</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={(value) => field.onChange(value ?? '')}
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
              name="type"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Tipo</FieldLabel>
                  <Select
                    value={field.value}
                    onValueChange={(value) =>
                      field.onChange((value ?? 'PJ') as 'PF' | 'PJ')
                    }
                  >
                    <SelectTrigger className="border-input bg-background rounded-xl border">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="PF">PF</SelectItem>
                      <SelectItem value="PJ">PJ</SelectItem>
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
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Nome</FieldLabel>
                  <Input
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
              name="legal_name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Razão social</FieldLabel>
                  <Input
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
              name="document"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>
                    {customerType === 'PF' ? 'CPF' : 'CNPJ'}
                  </FieldLabel>
                  <Input
                    {...field}
                    className="border-input bg-background rounded-xl border"
                    placeholder={customerType === 'PF' ? 'CPF' : 'CNPJ'}
                    ref={withMask(
                      customerType === 'PF'
                        ? '999.999.999-99'
                        : '99.999.999/9999-99',
                      {
                        placeholder: '',
                        showMaskOnHover: false
                      }
                    )}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="state_registration"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Inscrição estadual</FieldLabel>
                  <Input
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
              name="birth_or_opening_date"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Data de nascimento/abertura</FieldLabel>
                  <DatePicker
                    className="border-input rounded-xl border"
                    name={field.name}
                    ref={field.ref}
                    value={field.value}
                    onBlur={field.onBlur}
                    onChange={(date) =>
                      field.onChange(date ? format(date, 'yyyy-MM-dd') : '')
                    }
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="responsible_salesperson_user_id"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Vendedor responsável (ID do usuário)</FieldLabel>
                  <Input
                    className="border-input bg-background rounded-xl border"
                    placeholder="Opcional"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />
          </div>

          <SheetFooter className="mt-auto gap-2 border-t px-6 py-4 sm:justify-end">
            <SheetClose render={<Button variant="outline" type="button" />}>
              Cancelar
            </SheetClose>
            <Button type="submit" disabled={createMutation.isPending}>
              Cadastrar
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
