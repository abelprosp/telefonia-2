import { Link } from '@tanstack/react-router';
import { ChevronLeft } from 'lucide-react';

import type { GetProviderInvoiceResponse } from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import {
  formatInvoiceItemType,
  formatInvoiceStatus,
  formatPhoneNumber
} from '@/lib/format';
import { useAuthRoles } from '@/lib/auth-roles';
import { cn } from '@/lib/utils';

import { InvoiceFinancialActions } from './invoice-financial-actions';

type InvoicesListSearch = {
  page: number;
  pageSize: number;
  processingMonthId: string | undefined;
};

export type InvoiceRelatedLabels = {
  provider: string;
  customer: string;
  billingCycle: string;
  processingMonth: string;
  contractingCompany: string;
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

function numOrDash(v: null | number | string | undefined) {
  if (v === null || v === undefined) {
    return '—';
  }
  const n = typeof v === 'string' ? Number(v) : v;
  return Number.isFinite(n) ? String(n) : String(v);
}

type InvoiceDetailViewProps = {
  invoice: GetProviderInvoiceResponse;
  listSearch: InvoicesListSearch;
};

export function InvoiceDetailView({
  invoice,
  listSearch
}: InvoiceDetailViewProps) {
  const { canAccessFinance } = useAuthRoles();
  const backLink = {
    to: '/invoices' as const,
    search: {
      page: listSearch.page,
      pageSize: listSearch.pageSize,
      processingMonthId: listSearch.processingMonthId
    }
  };

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            nativeButton={false}
            size="icon"
            render={<Link {...backLink} />}
          >
            <ChevronLeft className="size-4" />
            <span className="sr-only">Voltar</span>
          </Button>
          <div className="flex flex-col">
            <h3 className="text-foreground text-lg font-semibold">Fatura</h3>
            <p className="text-muted-foreground mt-1 text-sm leading-6">
              {invoice.number}
            </p>
          </div>
        </div>
      </div>

      <DetailSection
        title="Identificação e valores"
        description="Dados da fatura de origem importada da operadora e totais consolidados."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Conta </FieldLabel>
              <ReadOnlyField value={invoice.provider_account_number} />
            </Field>
            <Field>
              <FieldLabel>Situação</FieldLabel>
              <ReadOnlyField value={formatInvoiceStatus(invoice.status!)} />
            </Field>
            <Field>
              <FieldLabel>Emissão</FieldLabel>
              <ReadOnlyField
                value={
                  invoice.issue_date?.toDate()?.format('dd/MM/yyyy') ?? '-'
                }
              />
            </Field>
            <Field>
              <FieldLabel>Vencimento</FieldLabel>
              <ReadOnlyField
                value={
                  invoice.due_date?.toDate()?.format('dd/MM/yyyy') ?? '-'
                }
              />
            </Field>
            <Field>
              <FieldLabel>Valor total</FieldLabel>
              <ReadOnlyField value={invoice.total_amount.toCurrency()!} />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      {canAccessFinance ? (
        <>
          <Separator />
          <DetailSection
            title="Financeiro"
            description="Vincule esta fatura da operadora a uma conta a pagar para o refaturamento."
          >
            <InvoiceFinancialActions
              invoiceId={invoice.id}
              totalAmount={invoice.total_amount}
              accountPayableId={invoice.account_payable_id}
              accountPayableStatus={invoice.account_payable_status}
            />
          </DetailSection>
        </>
      ) : null}

      <Separator />

      <DetailSection
        title="Vínculos"
        description="Operadora, cliente, empresa contratante, ciclo e mês de processamento."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Operadora</FieldLabel>
              <div className="flex flex-wrap items-end gap-2">
                <ReadOnlyField
                  value={invoice.provider_name}
                  className="min-w-0 flex-1"
                />
                <Button
                  nativeButton={false}
                  variant="outline"
                  size="sm"
                  className="shrink-0"
                  render={
                    <Link
                      to="/providers/$providerId"
                      params={{ providerId: invoice.provider_id }}
                      search={{ page: 1, pageSize: 10 }}
                    />
                  }
                >
                  Abrir
                </Button>
              </div>
            </Field>
            <Field>
              <FieldLabel>Empresa contratante</FieldLabel>
              <ReadOnlyField value={invoice.contracting_company_name} />
            </Field>
            <Field>
              <FieldLabel>Ciclo de faturamento</FieldLabel>
              <div className="flex flex-wrap items-end gap-2">
                <ReadOnlyField
                  value={invoice.billing_cycle_name}
                  className="min-w-0 flex-1"
                />
                <Button
                  nativeButton={false}
                  variant="outline"
                  size="sm"
                  className="shrink-0"
                  render={
                    <Link
                      to="/billing-cycles/$cycleId"
                      params={{ cycleId: invoice.billing_cycle_id }}
                      search={{ page: 1, pageSize: 10 }}
                    />
                  }
                >
                  Abrir
                </Button>
              </div>
            </Field>
            <Field className="sm:col-span-2">
              <FieldLabel>Mês de processamento</FieldLabel>
              <ReadOnlyField value={invoice.processing_month_name ?? '—'} />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Composição e referências"
        description="Subtotais informados na fatura e linhas vinculadas."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Serviços</FieldLabel>
              <ReadOnlyField
                value={invoice.subtotal_services?.toCurrency() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Uso / consumo</FieldLabel>
              <ReadOnlyField
                value={invoice.subtotal_usage?.toCurrency() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Impostos</FieldLabel>
              <ReadOnlyField
                value={invoice.subtotal_taxes?.toCurrency() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Descontos</FieldLabel>
              <ReadOnlyField
                value={invoice.subtotal_discounts?.toCurrency() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Parcelas / aparelhos</FieldLabel>
              <ReadOnlyField
                value={invoice.subtotal_installments?.toCurrency() ?? '—'}
              />
            </Field>
            <Field>
              <FieldLabel>Centro de custo (ID)</FieldLabel>
              <ReadOnlyField value={invoice.cost_center_name ?? '—'} />
            </Field>
          </div>
          {invoice.phone_lines.length && (
            <Field>
              <FieldLabel>
                Linhas telefônicas ({invoice.phone_lines.length})
              </FieldLabel>
              <ul className="grid grid-cols-2 gap-2 md:grid-cols-4">
                {invoice.phone_lines.map((line) => (
                  <li key={line.id}>
                    <Button
                      nativeButton={false}
                      variant="outline"
                      size="sm"
                      className="text-xs"
                      render={
                        <Link
                          to="/phone-lines/$phoneLineId"
                          params={{ phoneLineId: line.id }}
                          search={{ page: 1, pageSize: 10 }}
                        />
                      }
                    >
                      {formatPhoneNumber(line.number)}
                    </Button>
                  </li>
                ))}
              </ul>
            </Field>
          )}
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Itens da fatura"
        description="Linhas de itens importadas com a fatura."
      >
        {invoice.provider_invoice_items.length === 0 ? (
          <p className="text-muted-foreground text-sm">
            Nenhum item retornado.
          </p>
        ) : (
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Descrição</TableHead>
                  <TableHead className="text-right">Qtd.</TableHead>
                  <TableHead className="text-right">Total</TableHead>
                  <TableHead className="text-right">Tipo</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {invoice.provider_invoice_items.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell className="max-w-[min(100vw,28rem)]">
                      <span className="text-sm leading-snug">
                        {row.description}
                      </span>
                    </TableCell>
                    <TableCell className="text-right text-sm tabular-nums">
                      {numOrDash(row.quantity)}
                    </TableCell>
                    <TableCell className="text-right text-sm tabular-nums">
                      {row.total_price.toCurrency()}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-right text-sm">
                      {formatInvoiceItemType(row.item_type)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
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
    </div>
  );
}
