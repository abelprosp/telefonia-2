import { useEffect, useMemo, useState } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { ChevronLeft, Loader2, Trash2 } from 'lucide-react';
import { useForm, type Resolver } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import {
  customersControllerGetByIdQueryKey,
  getV1CustomersQueryKey,
  useDeleteV1CustomersId,
  useGetV1CustomersIdAttachments,
  useGetV1CustomersIdPhoneLines,
  useGetV1CustomersIdProviderLinks,
  usePatchV1CustomersId,
  type ListCustomerResponse
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
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatCpfCnpj,
  formatCustomerType,
  formatPhoneLineStatus
} from '@/lib/format';
import { invalidateDashboardCaches } from '@/lib/query-utils';
import { cn } from '@/lib/utils';

import { CustomerAttachmentsView } from './customer-attachments-view';

type ListSearch = {
  page: number;
  pageSize: number;
  providerId: string | undefined;
};

type FormValues = {
  name: string;
  legal_name: string;
  state_registration: string;
  responsible_salesperson_user_id: string;
};

function buildSchema(isPj: boolean) {
  const obj = z.object({
    name: z.string().min(1, 'Informe o nome'),
    legal_name: z.string(),
    state_registration: z.string(),
    responsible_salesperson_user_id: z
      .string()
      .max(256, 'Identificador muito longo')
  });
  if (!isPj) {
    return obj;
  }
  return obj.superRefine((data, ctx) => {
    if (!data.legal_name.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Informe a razão social',
        path: ['legal_name']
      });
    }
  });
}

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

type CustomerDetailViewProps = {
  customer: ListCustomerResponse;
  listSearch: ListSearch;
};

export function CustomerDetailView({
  customer,
  listSearch
}: CustomerDetailViewProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [deleteOpen, setDeleteOpen] = useState(false);

  const isPj = customer.type.trim().toUpperCase() === 'PJ';
  const schema = useMemo(() => buildSchema(isPj), [isPj]);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema) as unknown as Resolver<FormValues>,
    defaultValues: {
      name: customer.name,
      legal_name: customer.legal_name ?? '',
      state_registration: customer.state_registration ?? '',
      responsible_salesperson_user_id:
        customer.responsible_salesperson_user_id ?? ''
    }
  });

  useEffect(() => {
    form.reset({
      name: customer.name,
      legal_name: customer.legal_name ?? '',
      state_registration: customer.state_registration ?? '',
      responsible_salesperson_user_id:
        customer.responsible_salesperson_user_id ?? ''
    });
  }, [customer, form]);

  const saveMutation = usePatchV1CustomersId({
    mutation: {
      onSuccess: async () => {
        toast.success('Cliente atualizado.');
        await queryClient.invalidateQueries({
          queryKey: customersControllerGetByIdQueryKey(customer.id)
        });
        await queryClient.invalidateQueries({
          queryKey: getV1CustomersQueryKey()
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const deleteMutation = useDeleteV1CustomersId({
    mutation: {
      onSuccess: async () => {
        setDeleteOpen(false);
        toast.success('Cliente desativado.');
        await queryClient.invalidateQueries({
          queryKey: getV1CustomersQueryKey()
        });
        await invalidateDashboardCaches(queryClient);
        navigate({
          to: '/customers',
          search: listSearch
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const isActive = customer.active;
  const customerProviderLinksQuery = useGetV1CustomersIdProviderLinks(
    customer.id
  );
  const customerLinesQuery = useGetV1CustomersIdPhoneLines(customer.id, {
    page_index: 1,
    page_size: 50
  });

  const customerAttachmentsQuery = useGetV1CustomersIdAttachments(customer.id);

  const backTo = {
    to: '/customers' as const,
    search: listSearch
  };

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button
            nativeButton={false}
            variant="outline"
            size="icon"
            render={<Link {...backTo} />}
          >
            <ChevronLeft className="size-4" />
            <span className="sr-only">Voltar</span>
          </Button>
          <div className="flex flex-col">
            <h3 className="text-foreground text-lg font-semibold">
              Editar cliente
            </h3>
            <p className="text-muted-foreground mt-1 text-sm leading-6">
              {customer.name}
            </p>
          </div>
        </div>
      </div>

      {!isActive ? (
        <div className="border-border bg-muted/40 rounded-lg border px-4 py-3 text-sm">
          Este cliente está <strong>inativo</strong> e não aparece nas listagens
          padrão.
        </div>
      ) : null}

      <form
        id="customer-detail-form"
        className="flex flex-col gap-8"
        onSubmit={form.handleSubmit((v: FormValues) =>
          saveMutation.mutate({
            id: customer.id,
            data: {
              name: v.name.trim(),
              legal_name: isPj ? v.legal_name.trim() : null,
              state_registration: v.state_registration.trim() || null,
              birth_or_opening_date: customer.birth_or_opening_date ?? null,
              responsible_salesperson_user_id:
                v.responsible_salesperson_user_id.trim() || null
            }
          })
        )}
      >
        <DetailSection
          title="Dados do cliente"
          description="Tipo e documento são fixos após o cadastro. Nome fantasia, razão social, inscrição estadual e vendedor responsável podem ser ajustados."
        >
          <FieldGroup className="gap-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel>Tipo</FieldLabel>
                <ReadOnlyInput value={formatCustomerType(customer.type)} />
              </Field>
              <Field>
                <FieldLabel>CPF/CNPJ</FieldLabel>
                <ReadOnlyInput
                  value={formatCpfCnpj(customer.cpf_cnpj) || '—'}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="customer-name">
                  {isPj ? 'Nome fantasia' : 'Nome'}
                </FieldLabel>
                <Input
                  id="customer-name"
                  disabled={!isActive}
                  {...form.register('name')}
                />
                <FieldError errors={[form.formState.errors.name]} />
              </Field>
              {isPj ? (
                <Field>
                  <FieldLabel htmlFor="customer-legal">Razão social</FieldLabel>
                  <Input
                    id="customer-legal"
                    disabled={!isActive}
                    {...form.register('legal_name')}
                  />
                  <FieldError errors={[form.formState.errors.legal_name]} />
                </Field>
              ) : null}
              <Field>
                <FieldLabel htmlFor="customer-ie">
                  Inscrição estadual (IE)
                </FieldLabel>
                <Input
                  id="customer-ie"
                  disabled={!isActive}
                  {...form.register('state_registration')}
                />
                <FieldError
                  errors={[form.formState.errors.state_registration]}
                />
              </Field>
              <Field>
                <FieldLabel>Data de nascimento / abertura</FieldLabel>
                <ReadOnlyInput
                  value={
                    customer.birth_or_opening_date?.formatAsDate(
                      'dd/MM/yyyy'
                    ) ?? '—'
                  }
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="customer-salesperson">
                  Vendedor responsável (ID do usuário)
                </FieldLabel>
                <Input
                  id="customer-salesperson"
                  disabled={!isActive}
                  placeholder="Opcional — ex.: sub do Keycloak"
                  {...form.register('responsible_salesperson_user_id')}
                />
                <FieldError
                  errors={[
                    form.formState.errors.responsible_salesperson_user_id
                  ]}
                />
              </Field>
            </div>
          </FieldGroup>
        </DetailSection>

        <Separator />

        <DetailSection
          title="Operadoras vinculadas"
          description="Histórico de vínculos de operadora do cliente."
        >
          {customerProviderLinksQuery.isPending ? (
            <p className="text-muted-foreground text-sm">
              Carregando vínculos de operadora...
            </p>
          ) : customerProviderLinksQuery.isError ? (
            <p className="text-destructive text-sm">
              {isApiHttpError(customerProviderLinksQuery.error)
                ? customerProviderLinksQuery.error.message
                : getErrorMessage(customerProviderLinksQuery.error)}
            </p>
          ) : (customerProviderLinksQuery.data?.length ?? 0) === 0 ? (
            <p className="text-muted-foreground text-sm">
              Nenhum vínculo de operadora encontrado.
            </p>
          ) : (
            <div className="space-y-3">
              <div className="overflow-x-auto rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Operadora</TableHead>
                      <TableHead>Início</TableHead>
                      <TableHead>Fim</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(customerProviderLinksQuery.data ?? []).map((item) => (
                      <TableRow key={`${item.provider_id}-${item.start_date}`}>
                        <TableCell>{item.provider_name}</TableCell>
                        <TableCell>
                          {item.start_date.toDate()?.format('dd/MM/yyyy') ??
                            '—'}
                        </TableCell>
                        <TableCell>
                          {item.end_date?.toDate()?.format('dd/MM/yyyy') ?? '—'}
                        </TableCell>
                        <TableCell>
                          {item.is_active ? 'Ativo' : 'Encerrado'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}
        </DetailSection>

        <Separator />

        <DetailSection
          title="Linhas vinculadas"
          description="Histórico dos vínculos linha-cliente deste cliente."
        >
          {customerLinesQuery.isPending ? (
            <p className="text-muted-foreground text-sm">
              Carregando vínculos...
            </p>
          ) : customerLinesQuery.isError ? (
            <p className="text-destructive text-sm">
              {isApiHttpError(customerLinesQuery.error)
                ? customerLinesQuery.error.message
                : getErrorMessage(customerLinesQuery.error)}
            </p>
          ) : (customerLinesQuery.data?.items?.length ?? 0) === 0 ? (
            <div className="border-input rounded-md border border-dashed p-6 text-center text-sm">
              <p className="text-muted-foreground">
                Nenhuma linha vinculada a este cliente.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="overflow-x-auto rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Linha</TableHead>
                      <TableHead>Status da linha</TableHead>
                      <TableHead>Classificação</TableHead>
                      <TableHead>Início</TableHead>
                      <TableHead>Fim</TableHead>
                      <TableHead>Vínculo</TableHead>
                      <TableHead />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(customerLinesQuery.data?.items ?? []).map((item) => (
                      <TableRow
                        key={`${item.phone_line_id}-${item.start_date}`}
                      >
                        <TableCell>{item.phone_line_number}</TableCell>
                        <TableCell>
                          {formatPhoneLineStatus(item.phone_line_status)}
                        </TableCell>
                        <TableCell>{item.line_classification}</TableCell>
                        <TableCell>
                          {item.start_date.toDate()?.format('dd/MM/yyyy') ??
                            '—'}
                        </TableCell>
                        <TableCell>
                          {item.end_date?.toDate()?.format('dd/MM/yyyy') ?? '—'}
                        </TableCell>
                        <TableCell>
                          {item.is_active ? 'Ativo' : 'Encerrado'}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            nativeButton={false}
                            variant="outline"
                            size="sm"
                            render={
                              <Link
                                to="/phone-lines/$phoneLineId"
                                params={{ phoneLineId: item.phone_line_id }}
                                search={{ page: 1, pageSize: 10 }}
                              />
                            }
                          >
                            Abrir linha
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <p className="text-muted-foreground text-xs">
                Mostrando {customerLinesQuery.data?.items?.length ?? 0} de{' '}
                {customerLinesQuery.data?.total_count ?? 0} vínculos.
              </p>
            </div>
          )}
        </DetailSection>

        <Separator />

        <DetailSection
          title="Endereços"
          description="O contrato OpenAPI atual não inclui endereços no detalhe do cliente."
        >
          <div className="border-input rounded-md border border-dashed p-6 text-center text-sm">
            <p className="text-muted-foreground">Nenhum endereço encontrado.</p>
          </div>
        </DetailSection>

        <Separator />

        <DetailSection
          title="Documentação e arquivos"
          description="Documentos anexados ao cadastro do cliente."
        >
          {customerAttachmentsQuery.isPending ? (
            <div className="border-input rounded-md border border-dashed p-6 text-center text-sm">
              <p className="text-muted-foreground">Carregando arquivos...</p>
            </div>
          ) : customerAttachmentsQuery.isError ? (
            <div className="border-input rounded-md border border-dashed p-6 text-center">
              <p className="text-destructive text-sm">
                {String(customerAttachmentsQuery.error)}
              </p>
            </div>
          ) : (
            <CustomerAttachmentsView
              customerId={customer.id}
              attachments={customerAttachmentsQuery.data!}
            />
          )}
        </DetailSection>

        {isActive ? (
          <>
            <Separator />

            <DetailSection
              title="Danger Zone"
              description="Ações irreversíveis ou que alteram o status do cliente."
            >
              <div
                className={cn(
                  'border-destructive/50 bg-destructive/5 flex flex-col gap-4 rounded-lg border p-4',
                  deleteMutation.isPending && 'opacity-80'
                )}
              >
                <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                  <div className="min-w-0">
                    <p className="text-foreground font-semibold">
                      Desativar cliente
                    </p>
                    <p className="text-muted-foreground mt-0.5 text-sm leading-6">
                      O cliente deixará de aparecer nas listagens ativas.
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="destructive"
                    size="sm"
                    className="shrink-0"
                    onClick={() => setDeleteOpen(true)}
                    disabled={deleteMutation.isPending}
                  >
                    <Trash2 className="mr-2 size-4" />
                    Desativar
                  </Button>
                </div>
              </div>
            </DetailSection>
          </>
        ) : null}

        {isActive ? (
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              className="whitespace-nowrap"
              onClick={() => {
                form.reset({
                  name: customer.name,
                  legal_name: customer.legal_name ?? '',
                  state_registration: customer.state_registration ?? '',
                  responsible_salesperson_user_id:
                    customer.responsible_salesperson_user_id ?? ''
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
              render={<Link {...backTo} />}
            >
              Voltar
            </Button>
            <Button
              type="submit"
              form="customer-detail-form"
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
        ) : (
          <div className="mt-8 flex justify-end">
            <Button
              nativeButton={false}
              type="button"
              variant="outline"
              className="whitespace-nowrap"
              render={<Link {...backTo} />}
            >
              Voltar à lista
            </Button>
          </div>
        )}
      </form>

      <Sheet open={deleteOpen} onOpenChange={setDeleteOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Desativar cliente</SheetTitle>
            <SheetDescription>
              Tem certeza? O cliente não aparecerá mais nas listagens de ativos.
            </SheetDescription>
          </SheetHeader>
          <SheetFooter className="gap-2 sm:justify-end">
            <SheetClose render={<Button variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => {
                deleteMutation.mutate({ id: customer.id });
              }}
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  Desativando…
                </>
              ) : (
                'Desativar'
              )}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
