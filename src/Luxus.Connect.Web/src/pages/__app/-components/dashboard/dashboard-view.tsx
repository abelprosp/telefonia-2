import { Link } from '@tanstack/react-router';
import {
  Building2,
  Calendar,
  FilePenLine,
  FileText,
  Import,
  Phone,
  Plus,
  Receipt,
  Users
} from 'lucide-react';

import {
  useGetV1Customers,
  useGetV1PhoneLines,
  useGetV1StatsDashboard
} from '@/api';
import { Button } from '@/components/ui/button';
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '@/components/ui/tooltip';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatCpfCnpj,
  formatPhoneLineStatus,
  formatPhoneNumber
} from '@/lib/format';
import { parseTotalCount } from '@/lib/query-utils';

import { DashboardMetricCard } from './dashboard-metric-card';

const formatCount = (value: number) =>
  new Intl.NumberFormat('pt-BR').format(value);

export const DashboardView = () => {
  const statsQuery = useGetV1StatsDashboard();
  const customersQuery = useGetV1Customers({
    page_index: 0,
    page_size: 10
  });
  const phoneLinesQuery = useGetV1PhoneLines({
    page_index: 0,
    page_size: 10
  });

  const summaryPending =
    statsQuery.isPending ||
    customersQuery.isPending ||
    phoneLinesQuery.isPending;

  if (summaryPending) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-10 w-72 rounded-xl" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-2xl" />
          ))}
        </div>
        <Skeleton className="h-80 rounded-2xl" />
        <Skeleton className="h-80 rounded-2xl" />
      </div>
    );
  }

  const summaryError =
    statsQuery.error ?? customersQuery.error ?? phoneLinesQuery.error;

  if (summaryError) {
    const err = summaryError;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-2xl border px-4 py-3 text-sm">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const stats = statsQuery.data;
  const customers = customersQuery.data;
  const phoneLines = phoneLinesQuery.data;

  if (!stats || !customers || !phoneLines) {
    return null;
  }

  const metrics = [
    {
      title: 'Clientes',
      value: formatCount(parseTotalCount(stats.customers_count)),
      icon: Users,
      to: '/customers'
    },
    {
      title: 'Operadoras',
      value: formatCount(parseTotalCount(stats.providers_count)),
      icon: Building2,
      to: '/providers',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Linhas telefônicas',
      value: formatCount(parseTotalCount(stats.phone_lines_count)),
      icon: Phone,
      to: '/phone-lines',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Faturas emitidas',
      value: formatCount(parseTotalCount(stats.provider_invoices_count)),
      icon: FileText,
      to: '/invoices',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Ciclos de faturamento',
      value: formatCount(parseTotalCount(stats.billing_cycles_count)),
      icon: Calendar,
      to: '/billing-cycles',
      search: { page: 1, pageSize: 10 }
    }
  ];

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          Visão geral
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Indicadores consolidados da operação de telefonia
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        {metrics.map((metric) => (
          <DashboardMetricCard key={metric.title} {...metric} />
        ))}
      </div>

      <RecentCustomersPanel rows={customers.items ?? []} />
      <RecentPhoneLinesPanel rows={phoneLines.items ?? []} />
    </div>
  );
};

const RecentCustomersPanel = ({
  rows
}: {
  rows: { id: string; name: string; cpf_cnpj: string; active: boolean }[];
}) => {
  return (
    <div className="dashboard-card overflow-hidden">
      <div className="border-b px-5 py-4">
        <h3 className="text-lg font-semibold">Clientes recentes</h3>
        <p className="text-muted-foreground text-sm">
          Últimos cadastros na plataforma
        </p>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>Cliente</TableHead>
            <TableHead>Documento</TableHead>
            <TableHead>Situação</TableHead>
            <TableHead className="w-24 text-right" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Receipt />
                    </EmptyMedia>
                    <EmptyTitle>Nenhum cliente encontrado</EmptyTitle>
                    <EmptyDescription>
                      Comece cadastrando seu primeiro cliente ou importe uma
                      fatura.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent>
                    <div className="flex flex-wrap gap-2 *:mx-auto">
                      <Button>
                        <Plus /> Cadastrar cliente
                      </Button>
                      <Button variant="outline">
                        <Import /> Importar fatura
                      </Button>
                    </div>
                  </EmptyContent>
                </Empty>
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell className="font-medium">{row.name}</TableCell>
                <TableCell className="text-muted-foreground">
                  {formatCpfCnpj(row.cpf_cnpj)}
                </TableCell>
                <TableCell>
                  <span
                    className={
                      row.active
                        ? 'inline-flex rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700'
                        : 'text-muted-foreground inline-flex rounded-full bg-muted px-2.5 py-1 text-xs font-medium'
                    }
                  >
                    {row.active ? 'Ativo' : 'Inativo'}
                  </span>
                </TableCell>
                <TableCell>
                  <div className="flex justify-end">
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            nativeButton={false}
                            variant="ghost"
                            size="sm"
                            className="text-primary hover:text-primary"
                            render={
                              <Link
                                to="/customers/$customerId"
                                params={{ customerId: row.id }}
                                search={{
                                  page: 1,
                                  pageSize: 10,
                                  providerId: undefined
                                }}
                              >
                                <FilePenLine />
                              </Link>
                            }
                          />
                        }
                      />
                      <TooltipContent>Abrir</TooltipContent>
                    </Tooltip>
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};

const RecentPhoneLinesPanel = ({
  rows
}: {
  rows: { id: string; number: string | null; status?: string | null }[];
}) => {
  return (
    <div className="dashboard-card overflow-hidden">
      <div className="border-b px-5 py-4">
        <h3 className="text-lg font-semibold">Linhas recentes</h3>
        <p className="text-muted-foreground text-sm">
          Últimas linhas telefônicas cadastradas
        </p>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>Número</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-24 text-right" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Phone />
                    </EmptyMedia>
                    <EmptyTitle>Nenhuma linha encontrada</EmptyTitle>
                    <EmptyDescription>
                      Cadastre uma linha telefônica ou importe uma fatura para
                      começar.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              </TableCell>
            </TableRow>
          ) : (
            rows.map((line) => (
              <TableRow key={line.id}>
                <TableCell className="font-medium">
                  {formatPhoneNumber(line.number!) ?? '—'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatPhoneLineStatus(line.status) ?? '—'}
                </TableCell>
                <TableCell>
                  <div className="flex justify-end">
                    <Tooltip>
                      <TooltipTrigger
                        render={
                          <Button
                            nativeButton={false}
                            variant="ghost"
                            size="sm"
                            className="text-primary hover:text-primary"
                            render={
                              <Link
                                to="/phone-lines/$phoneLineId"
                                params={{ phoneLineId: line.id }}
                                search={{ page: 1, pageSize: 10 }}
                              >
                                <FilePenLine />
                              </Link>
                            }
                          />
                        }
                      />
                      <TooltipContent>Abrir</TooltipContent>
                    </Tooltip>
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};
