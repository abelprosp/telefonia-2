import { zodResolver } from '@hookform/resolvers/zod';
import { format } from 'date-fns';
import { Controller, useForm, useWatch } from 'react-hook-form';
import { toast } from 'sonner';
import { withMask } from 'use-mask-input';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { DatePicker } from '@/components/ui/date-picker';
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
import { useCreatePartnerCustomer } from '@/lib/partner-api';

const formSchema = z
  .object({
    type: z.enum(['PF', 'PJ']),
    name: z.string().min(1, 'Informe o nome').max(256, 'Nome muito longo'),
    legal_name: z.string().optional(),
    document: z
      .string()
      .min(11, 'Documento inválido')
      .max(20, 'Documento inválido'),
    state_registration: z.string().optional(),
    birth_or_opening_date: z.string().optional()
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

type PartnerCustomerCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

const defaultValues: FormValues = {
  type: 'PJ',
  name: '',
  legal_name: '',
  document: '',
  state_registration: '',
  birth_or_opening_date: ''
};

export function PartnerCustomerCreateSheet({
  open,
  onOpenChange
}: PartnerCustomerCreateSheetProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  const customerType = useWatch({ control: form.control, name: 'type' }) ?? 'PJ';

  const createMutation = useCreatePartnerCustomer();

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate(
      {
        type: values.type,
        name: values.name.trim(),
        legal_name: values.legal_name?.trim() || null,
        document: values.document.replace(/\D/g, ''),
        state_registration: values.state_registration?.trim() || null,
        birth_or_opening_date: values.birth_or_opening_date?.trim() || null,
        addresses: []
      },
      {
        onSuccess: () => {
          toast.success('Cliente cadastrado na sua carteira.');
          onOpenChange(false);
          form.reset(defaultValues);
        },
        onError: (e) => {
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
        }
      }
    );
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
            O cliente será vinculado automaticamente à sua carteira de vendas. A
            operadora pode ser associada depois, ao vincular uma linha.
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
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
                    ref={withMask(
                      customerType === 'PF'
                        ? '999.999.999-99'
                        : '99.999.999/9999-99',
                      { placeholder: '', showMaskOnHover: false }
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
