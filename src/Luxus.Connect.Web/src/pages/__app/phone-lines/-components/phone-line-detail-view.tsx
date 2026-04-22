import { useMemo, useState } from 'react';

import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { ChevronLeft } from 'lucide-react';
import { toast } from 'sonner';

import {
  getV1PhoneLinesIdCustomerLinksQueryKey,
  useDeleteV1PhoneLinesIdCustomerLinksActive,
  useGetV1Customers,
  useGetV1PhoneLinesIdCustomerLinks,
  usePostV1PhoneLinesIdCustomerLinks,
  usePostV1PhoneLinesIdCustomerLinksTransfer,
  type GetPhoneLineResponse,
  type GetPhoneLineServiceResponse
} from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
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
  formatLineClassification,
  formatPhoneLineStatus,
  formatPhoneNumber,
  formatTransitionSubStatus
} from '@/lib/format';
import { cn } from '@/lib/utils';

type PhoneLinesListSearch = {
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

function ReadOnlyField({
  value,
  className
}: {
  value: string;
  className?: string;
}) {
  return (
    <Input
      readOnly
      value={value}
      className={cn(
        'bg-muted/50 pointer-events-none border-transparent shadow-none',
        className
      )}
    />
  );
}

export type PhoneLineRelatedLabels = {
  titular: string;
};

type PhoneLineListOrigin = '/phone-lines' | '/stock';

type PhoneLineDetailViewProps = {
  line: GetPhoneLineResponse;
  listSearch: PhoneLinesListSearch;
  backListTo?: PhoneLineListOrigin;
};

export function PhoneLineDetailView({
  line,
  listSearch,
  backListTo = '/phone-lines'
}: PhoneLineDetailViewProps) {
  const queryClient = useQueryClient();
  const [assignOpen, setAssignOpen] = useState(false);
  const [transferOpen, setTransferOpen] = useState(false);
  const [unassignOpen, setUnassignOpen] = useState(false);
  const [selectedCustomerId, setSelectedCustomerId] = useState('');
  const [effectiveDate, setEffectiveDate] = useState('');

  const customerLinksQuery = useGetV1PhoneLinesIdCustomerLinks(line.id);
  const customersQuery = useGetV1Customers({
    page_index: 1,
    page_size: 500
  });

  const activeLink = useMemo(
    () => customerLinksQuery.data?.find((l) => l.is_active) ?? null,
    [customerLinksQuery.data]
  );

  const customersOptions = useMemo(
    () => (customersQuery.data?.items ?? []).filter((c) => c.active),
    [customersQuery.data?.items]
  );

  const resetLinkActionForm = () => {
    setSelectedCustomerId('');
    setEffectiveDate('');
  };

  const invalidateLinks = async () => {
    await queryClient.invalidateQueries({
      queryKey: getV1PhoneLinesIdCustomerLinksQueryKey(line.id)
    });
  };

  const assignMutation = usePostV1PhoneLinesIdCustomerLinks({
    mutation: {
      onSuccess: async () => {
        toast.success('Cliente vinculado à linha.');
        setAssignOpen(false);
        resetLinkActionForm();
        await invalidateLinks();
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const transferMutation = usePostV1PhoneLinesIdCustomerLinksTransfer({
    mutation: {
      onSuccess: async () => {
        toast.success('Linha transferida com sucesso.');
        setTransferOpen(false);
        resetLinkActionForm();
        await invalidateLinks();
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const unassignMutation = useDeleteV1PhoneLinesIdCustomerLinksActive({
    mutation: {
      onSuccess: async () => {
        toast.success('Vínculo ativo removido.');
        setUnassignOpen(false);
        resetLinkActionForm();
        await invalidateLinks();
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const backLink = {
    to: backListTo,
    search: listSearch
  };

  const lineDetailRoute =
    backListTo === '/stock'
      ? ('/stock/$phoneLineId' as const)
      : ('/phone-lines/$phoneLineId' as const);

  const displayNumber = formatPhoneNumber(line.number) ?? '—';

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-4">
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
              Linha telefônica
            </h3>
            <p className="text-muted-foreground mt-1 text-sm leading-6">
              {displayNumber}
            </p>
          </div>
        </div>
      </div>

      <DetailSection
        title="Identificação"
        description="Número, plano e vínculos da linha telefônica."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Número</FieldLabel>
              <ReadOnlyField value={displayNumber} />
            </Field>
            <Field>
              <FieldLabel>Centro de custo</FieldLabel>
              <ReadOnlyField value={line.cost_center_name ?? '—'} />
            </Field>
            <Field>
              <FieldLabel>Plano</FieldLabel>
              <ReadOnlyField
                value={line.provider_plan_name}
                className="text-xs"
              />
            </Field>
            <Field>
              <FieldLabel>Conta operadora</FieldLabel>
              <ReadOnlyField
                value={line.provider_account_number}
                className="text-xs"
              />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Vínculo com cliente"
        description="Gestão do cliente ativo da linha e histórico de vínculos."
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Cliente ativo</FieldLabel>
              <ReadOnlyField
                value={activeLink?.customer_name ?? 'Sem vínculo ativo'}
              />
            </Field>
            <Field>
              <FieldLabel>Documento</FieldLabel>
              <ReadOnlyField
                value={
                  formatCpfCnpj(activeLink?.customer_document ?? '') ?? '—'
                }
              />
            </Field>
          </div>

          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setAssignOpen(true)}
            >
              Vincular cliente
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setTransferOpen(true)}
              disabled={!activeLink}
            >
              Transferir vínculo
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={() => setUnassignOpen(true)}
              disabled={!activeLink}
            >
              Desvincular ativo
            </Button>
          </div>

          {(customerLinksQuery.data ?? []).length === 0 ? (
            <p className="text-muted-foreground text-sm">
              Sem histórico de vínculos para esta linha.
            </p>
          ) : (
            <div className="overflow-x-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Cliente</TableHead>
                    <TableHead>Documento</TableHead>
                    <TableHead>Início</TableHead>
                    <TableHead>Fim</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(customerLinksQuery.data ?? []).map((link) => (
                    <TableRow key={`${link.customer_id}-${link.start_date}`}>
                      <TableCell>{link.customer_name}</TableCell>
                      <TableCell>
                        {formatCpfCnpj(link.customer_document ?? '') ?? '—'}
                      </TableCell>
                      <TableCell>
                        {link.start_date.toDate()?.format('dd/MM/yyyy') ?? '—'}
                      </TableCell>
                      <TableCell>
                        {link.end_date?.toDate()?.format('dd/MM/yyyy') ?? '—'}
                      </TableCell>
                      <TableCell>
                        {link.is_active ? 'Ativo' : 'Encerrado'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Serviços na linha"
        description="Composição da linha telefônica."
      >
        {(line.services ?? []).length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhum serviço.</p>
        ) : (
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Código</TableHead>
                  <TableHead>Preço</TableHead>
                  <TableHead>Ativo</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(line.services ?? []).map((s: GetPhoneLineServiceResponse) => (
                  <TableRow key={s.id}>
                    <TableCell className="text-sm">{s.name}</TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {s.code}
                    </TableCell>
                    <TableCell className="text-sm tabular-nums">
                      {s.price?.toCurrency()}
                    </TableCell>
                    <TableCell className="text-sm">
                      {s.active ? 'Sim' : 'Não'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </DetailSection>

      <Separator />

      <DetailSection
        title="Ciclo e faturamento"
        description="Referência à última fatura quando informada."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Última fatura</FieldLabel>
              <ReadOnlyField value={line.last_invoice_number ?? '—'} />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Hierarquia e status"
        description="Classificação de cobrança, linha titular e estado da linha."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Classificação</FieldLabel>
              <ReadOnlyField
                value={formatLineClassification(line.line_classification)}
              />
            </Field>
            <Field>
              <FieldLabel>Linha titular</FieldLabel>
              {line.titular_line_id ? (
                <div className="flex flex-wrap items-end gap-2">
                  <ReadOnlyField
                    value={
                      line.titular_line_number
                        ? (formatPhoneNumber(line.titular_line_number) ?? '—')
                        : '—'
                    }
                    className="min-w-0 flex-1"
                  />
                  <Button
                    nativeButton={false}
                    variant="outline"
                    size="sm"
                    className="shrink-0"
                    render={
                      <Link
                        to={lineDetailRoute}
                        params={{ phoneLineId: line.titular_line_id }}
                        search={listSearch}
                      />
                    }
                  >
                    Abrir titular
                  </Button>
                </div>
              ) : (
                <ReadOnlyField value="—" />
              )}
            </Field>
            <Field>
              <FieldLabel>Status</FieldLabel>
              <ReadOnlyField value={formatPhoneLineStatus(line.status)} />
            </Field>
            <Field>
              <FieldLabel>Substatus transição</FieldLabel>
              <ReadOnlyField
                value={
                  line.transition_sub_status === null ||
                  line.transition_sub_status === undefined
                    ? '—'
                    : formatTransitionSubStatus(line.transition_sub_status)
                }
              />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Transição e datas"
        description="Controle de transição e marcos de ativação ou cancelamento."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Início da transição</FieldLabel>
              <ReadOnlyField
                value={line.transition_started_at?.formatAsDate() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Ativação</FieldLabel>
              <ReadOnlyField
                value={line.activation_date?.formatAsDate() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Cancelamento</FieldLabel>
              <ReadOnlyField
                value={line.cancellation_date?.formatAsDate() ?? '—'}
              />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <div className="flex flex-wrap items-center justify-end gap-2">
        <Button
          nativeButton={false}
          type="button"
          variant="outline"
          className="whitespace-nowrap"
          render={<Link {...backLink} />}
        >
          Voltar
        </Button>
      </div>

      <Sheet open={assignOpen} onOpenChange={setAssignOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Vincular cliente</SheetTitle>
            <SheetDescription>
              Associa um cliente à linha nesta data.
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 py-4">
            <Field>
              <FieldLabel>Cliente</FieldLabel>
              <Select
                value={selectedCustomerId}
                onValueChange={(value) => setSelectedCustomerId(value ?? '')}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Selecione um cliente" />
                </SelectTrigger>
                <SelectContent>
                  {customersOptions.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Data de início (opcional)</FieldLabel>
              <Input
                type="date"
                value={effectiveDate}
                onChange={(e) => setEffectiveDate(e.target.value)}
              />
            </Field>
          </div>
          <SheetFooter className="gap-2">
            <SheetClose render={<Button variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              disabled={assignMutation.isPending || !selectedCustomerId}
              onClick={() =>
                assignMutation.mutate({
                  id: line.id,
                  data: {
                    customer_id: selectedCustomerId,
                    start_date: effectiveDate || null
                  }
                })
              }
            >
              Confirmar
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Sheet open={transferOpen} onOpenChange={setTransferOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Transferir vínculo</SheetTitle>
            <SheetDescription>
              Fecha o vínculo atual e inicia o novo na data informada.
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 py-4">
            <Field>
              <FieldLabel>Novo cliente</FieldLabel>
              <Select
                value={selectedCustomerId}
                onValueChange={(value) => setSelectedCustomerId(value ?? '')}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Selecione um cliente" />
                </SelectTrigger>
                <SelectContent>
                  {customersOptions
                    .filter((c) => c.id !== activeLink?.customer_id)
                    .map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel>Data da transferência (opcional)</FieldLabel>
              <Input
                type="date"
                value={effectiveDate}
                onChange={(e) => setEffectiveDate(e.target.value)}
              />
            </Field>
          </div>
          <SheetFooter className="gap-2">
            <SheetClose render={<Button variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              disabled={transferMutation.isPending || !selectedCustomerId}
              onClick={() =>
                transferMutation.mutate({
                  id: line.id,
                  data: {
                    customer_id: selectedCustomerId,
                    transfer_date: effectiveDate || null
                  }
                })
              }
            >
              Transferir
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Sheet open={unassignOpen} onOpenChange={setUnassignOpen}>
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Desvincular cliente ativo</SheetTitle>
            <SheetDescription>
              Encerra o vínculo ativo da linha na data informada.
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-4 py-4">
            <Field>
              <FieldLabel>Data de encerramento (opcional)</FieldLabel>
              <Input
                type="date"
                value={effectiveDate}
                onChange={(e) => setEffectiveDate(e.target.value)}
              />
            </Field>
          </div>
          <SheetFooter className="gap-2">
            <SheetClose render={<Button variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              variant="destructive"
              disabled={unassignMutation.isPending}
              onClick={() =>
                unassignMutation.mutate({
                  id: line.id,
                  data: {
                    end_date: effectiveDate || null
                  }
                })
              }
            >
              Desvincular
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
