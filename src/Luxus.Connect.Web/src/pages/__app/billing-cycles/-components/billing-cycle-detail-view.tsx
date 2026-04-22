import { useEffect } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { ChevronLeft, Loader2 } from 'lucide-react';
import { Controller, useForm, type Resolver } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import {
  billingCyclesControllerGetByIdQueryKey,
  getV1BillingCyclesQueryKey,
  usePatchV1BillingCyclesId,
  type ListBillingCycleResponse
} from '@/api';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel
} from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { invalidateDashboardCaches } from '@/lib/query-utils';

type BillingCyclesListSearch = {
  page: number;
  pageSize: number;
};

function DetailSection({
  title,
  description,
  children
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
      <div>
        <h2 className="text-foreground font-semibold">{title}</h2>
        <p className="text-muted-foreground mt-1 text-sm leading-6">
          {description}
        </p>
      </div>
      <div className="sm:max-w-3xl md:col-span-2">{children}</div>
    </div>
  );
}

function ReadOnlyInput({ value }: { value: string }) {
  return (
    <Input
      readOnly
      value={value}
      className="bg-muted/50 pointer-events-none border-transparent shadow-none"
    />
  );
}

function formatDateTimePtBr(iso: string | null | undefined) {
  if (!iso) {
    return '—';
  }
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString('pt-BR');
}

const formSchema = z
  .object({
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

type BillingCycleDetailViewProps = {
  cycle: ListBillingCycleResponse;
  listSearch: BillingCyclesListSearch;
};

export function BillingCycleDetailView({
  cycle,
  listSearch
}: BillingCycleDetailViewProps) {
  const queryClient = useQueryClient();

  const backLink = {
    to: '/billing-cycles' as const,
    search: listSearch
  };

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema) as unknown as Resolver<FormValues>,
    defaultValues: {
      code: cycle.code,
      name: cycle.name,
      start_date: cycle.start_date.slice(0, 10),
      end_date: cycle.end_date.slice(0, 10)
    }
  });

  useEffect(() => {
    form.reset({
      code: cycle.code,
      name: cycle.name,
      start_date: cycle.start_date.slice(0, 10),
      end_date: cycle.end_date.slice(0, 10)
    });
  }, [cycle, form]);

  const saveMutation = usePatchV1BillingCyclesId({
    mutation: {
      onSuccess: async () => {
        toast.success('Ciclo atualizado.');
        await queryClient.invalidateQueries({
          queryKey: billingCyclesControllerGetByIdQueryKey(cycle.id)
        });
        await queryClient.invalidateQueries({
          queryKey: getV1BillingCyclesQueryKey()
        });
        await invalidateDashboardCaches(queryClient);
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-start gap-4">
          <Button
            nativeButton={false}
            variant="outline"
            size="icon"
            render={<Link {...backLink} />}
          >
            <ChevronLeft className="size-4" />
            <span className="sr-only">Voltar</span>
          </Button>
          <div className="flex flex-col">
            <h3 className="text-foreground text-lg font-semibold">
              Ciclo de faturamento
            </h3>
            <p className="text-muted-foreground mt-1 text-sm leading-6">
              {cycle.name}
            </p>
          </div>
        </div>
      </div>

      <form
        id="billing-cycle-detail-form"
        className="flex flex-col gap-8"
        onSubmit={form.handleSubmit((v: FormValues) =>
          saveMutation.mutate({
            id: cycle.id,
            data: {
              provider_id: cycle.provider_id,
              code: v.code.trim(),
              name: v.name.trim(),
              start_date: v.start_date,
              end_date: v.end_date
            }
          })
        )}
      >
        <DetailSection
          title="Dados do ciclo"
          description="Código, nome e período do ciclo na operadora. Alterações são enviadas à API (processamento assíncrono)."
        >
          <FieldGroup className="gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Controller
                control={form.control}
                name="code"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="billing-cycle-code">Código</FieldLabel>
                    <Input
                      id="billing-cycle-code"
                      autoComplete="off"
                      className="text-sm"
                      {...field}
                    />
                    <FieldError errors={[fieldState.error]} />
                  </Field>
                )}
              />
              <Controller
                control={form.control}
                name="name"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="billing-cycle-name">Nome</FieldLabel>
                    <Input
                      id="billing-cycle-name"
                      autoComplete="off"
                      {...field}
                    />
                    <FieldError errors={[fieldState.error]} />
                  </Field>
                )}
              />
              <Controller
                control={form.control}
                name="start_date"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="billing-cycle-start">
                      Data de início
                    </FieldLabel>
                    <Input id="billing-cycle-start" type="date" {...field} />
                    <FieldError errors={[fieldState.error]} />
                  </Field>
                )}
              />
              <Controller
                control={form.control}
                name="end_date"
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="billing-cycle-end">
                      Data de fim
                    </FieldLabel>
                    <Input id="billing-cycle-end" type="date" {...field} />
                    <FieldError errors={[fieldState.error]} />
                  </Field>
                )}
              />
            </div>
          </FieldGroup>
        </DetailSection>

        <Separator />

        <DetailSection
          title="Situação e fechamento"
          description="Estado atual do ciclo e registro de fechamento, quando aplicável."
        >
          <FieldGroup className="gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel>Situação</FieldLabel>
                <ReadOnlyInput value={cycle.status ?? '—'} />
              </Field>
              <Field>
                <FieldLabel>Fechado em</FieldLabel>
                <ReadOnlyInput value={formatDateTimePtBr(cycle.closed_at)} />
              </Field>
              <Field className="sm:col-span-2">
                <FieldLabel>Fechado por</FieldLabel>
                <ReadOnlyInput value={cycle.closed_by ?? '—'} />
              </Field>
            </div>
          </FieldGroup>
        </DetailSection>

        <Separator />

        <DetailSection
          title="Danger Zone"
          description="Operações que encerram ou removem o ciclo ainda não estão expostas nesta versão da API."
        >
          <div className="border-border bg-muted/30 flex flex-col gap-2 rounded-lg border p-4">
            <p className="text-muted-foreground text-sm leading-6">
              Não há ações destrutivas disponíveis nesta tela. Alterações ao
              período e identificação do ciclo usam o formulário acima e a API
              responde de forma assíncrona.
            </p>
          </div>
        </DetailSection>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            className="whitespace-nowrap"
            onClick={() => {
              form.reset({
                code: cycle.code,
                name: cycle.name,
                start_date: cycle.start_date.slice(0, 10),
                end_date: cycle.end_date.slice(0, 10)
              });
            }}
          >
            Cancelar
          </Button>
          <Button
            nativeButton={false}
            type="button"
            variant="outline"
            className="whitespace-nowrap"
            render={<Link {...backLink} />}
          >
            Voltar
          </Button>
          <Button
            type="submit"
            className="whitespace-nowrap"
            disabled={saveMutation.isPending}
          >
            {saveMutation.isPending ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                Salvando…
              </>
            ) : (
              'Salvar'
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}
