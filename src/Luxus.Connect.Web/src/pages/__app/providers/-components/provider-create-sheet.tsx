import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import { getV1ProvidersQueryKey, usePostV1Providers } from '@/api';
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
import { invalidateDashboardCaches } from '@/lib/query-utils';

const formSchema = z.object({
  name: z
    .string()
    .min(1, 'Informe o nome')
    .max(100, 'Nome deve ter no máximo 100 caracteres'),
  slug: z
    .string()
    .min(1, 'Informe o slug')
    .max(50, 'Slug deve ter no máximo 50 caracteres')
    .regex(
      /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
      'Use letras minúsculas, números e hífens (ex.: vivo-sp)'
    )
});

type FormValues = z.infer<typeof formSchema>;

type ProviderCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

const defaultValues: FormValues = {
  name: '',
  slug: ''
};

export function ProviderCreateSheet({
  open,
  onOpenChange
}: ProviderCreateSheetProps) {
  const queryClient = useQueryClient();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues
  });

  const createMutation = usePostV1Providers({
    mutation: {
      onSuccess: () => {
        toast.success('Operadora cadastrada.');
        void queryClient.invalidateQueries({
          queryKey: getV1ProvidersQueryKey()
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
        name: values.name.trim(),
        slug: values.slug.trim().toLowerCase()
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
          <SheetTitle>Nova operadora</SheetTitle>
          <SheetDescription>
            Cadastre uma operadora (VIVO, TIM, etc.) com nome e identificador
            único (slug). O slug é usado em integrações e URLs internas.
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
            <Controller
              control={form.control}
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="provider-create-name">Nome</FieldLabel>
                  <Input
                    id="provider-create-name"
                    autoComplete="organization"
                    placeholder="Ex.: Vivo Empresas"
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
              name="slug"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="provider-create-slug">Slug</FieldLabel>
                  <Input
                    id="provider-create-slug"
                    autoComplete="off"
                    placeholder="Ex.: vivo-empresas"
                    className="border-input bg-background rounded-xl border text-sm"
                    {...field}
                    onChange={(e) => {
                      field.onChange(e.target.value.toLowerCase());
                    }}
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
