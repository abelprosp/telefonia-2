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
  Users,
  type LucideIcon
} from 'lucide-react';

import {
  useGetV1Customers,
  useGetV1PhoneLines,
  useGetV1StatsDashboard
} from '@/api';
import {
  StatsHero,
  StatsMetricCard,
  StatsMetricGrid,
  StatsPanel
} from '@/components/stats-layout';
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
        <StatsHero
          align="left"
          title="Visão geral"
          subtitle="Indicadores consolidados da API Luxus.Connect."
        />
        <StatsMetricGrid className="mt-8 grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-5">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-44 rounded-xl" />
          ))}
        </StatsMetricGrid>
        <Skeleton className="h-48 rounded-xl" />
        <Skeleton className="h-48 rounded-xl" />
      </div>
    );
  }

  const summaryError =
    statsQuery.error ?? customersQuery.error ?? phoneLinesQuery.error;

  if (summaryError) {
    const err = summaryError;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-lg border px-4 py-3 text-sm">
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

  const metrics: {
    variant: 'blue' | 'green' | 'red' | 'amber' | 'purple';
    icon: LucideIcon;
    value: number;
    title: string;
    description?: string;
    to: string;
    search?: Record<string, unknown>;
    featured?: boolean;
  }[] = [
    {
      variant: 'blue',
      icon: Users,
      value: parseTotalCount(stats.customers_count),
      title: 'Clientes',
      to: '/customers'
    },
    {
      variant: 'green',
      icon: Building2,
      value: parseTotalCount(stats.providers_count),
      title: 'Operadoras',
      to: '/providers',
      search: { page: 1, pageSize: 10 }
    },
    {
      variant: 'red',
      icon: Phone,
      value: parseTotalCount(stats.phone_lines_count),
      title: 'Linhas telefônicas',
      to: '/phone-lines',
      search: { page: 1, pageSize: 10 },
      featured: true
    },
    {
      variant: 'amber',
      icon: FileText,
      value: parseTotalCount(stats.provider_invoices_count),
      title: 'Faturas emitidas pelas operadoras',
      to: '/invoices',
      search: { page: 1, pageSize: 10 }
    },
    {
      variant: 'purple',
      icon: Calendar,
      value: parseTotalCount(stats.billing_cycles_count),
      title: 'Ciclos de faturamento',
      to: '/billing-cycles',
      search: { page: 1, pageSize: 10 }
    }
  ];

  return (
    <div className="flex flex-col gap-6">
      <StatsHero align="left" title="Dashboard" />

      <StatsPanel title="Visão geral" description="Indicadores consolidados">
        <StatsMetricGrid className="grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3">
          {metrics.map((m) => (
            <StatsMetricCard
              key={m.title}
              variant={m.variant}
              icon={m.icon}
              value={m.value}
              title={m.title}
              description={m.description}
              to={m.to}
              search={m.search}
              featured={m.featured}
            />
          ))}
        </StatsMetricGrid>
      </StatsPanel>

      <StatsPanel
        title="Clientes recentes"
        description="Últimos 10 clientes cadastrados"
      >
        <RecentCustomersTable rows={customers.items ?? []} />
      </StatsPanel>

      <StatsPanel
        title="Linhas recentes"
        description="Últimas 10 linhas telefônicas cadastradas"
      >
        <RecentPhoneLinesTable rows={phoneLines.items ?? []} />
      </StatsPanel>
    </div>
  );
};

const RecentCustomersTable = ({
  rows
}: {
  rows: { id: string; name: string; cpf_cnpj: string; active: boolean }[];
}) => {
  return (
    <div className="w-full overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nome</TableHead>
            <TableHead>Documento</TableHead>
            <TableHead className="w-28">Situação</TableHead>
            <TableHead className="w-24 text-right"> </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="rounded-b-lg">
              <TableCell className="font-medium" colSpan={4}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Receipt />
                    </EmptyMedia>
                    <EmptyTitle>Nenhum cliente encontrado</EmptyTitle>
                    <EmptyDescription>
                      Você ainda não cadastrou nenhum cliente. Comece
                      cadastrando seu primeiro cliente ou importe uma fatura.
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
              <TableRow className="odd:bg-muted/50" key={row.id}>
                <TableCell className="font-medium">{row.name}</TableCell>
                <TableCell className="text-muted-foreground">
                  {formatCpfCnpj(row.cpf_cnpj)}
                </TableCell>
                <TableCell>
                  <span
                    className={
                      row.active
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-muted-foreground'
                    }
                  >
                    {row.active ? 'Ativo' : 'Inativo'}
                  </span>
                </TableCell>
                <TableCell>
                  <div className="flex justify-end gap-2">
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
                                className="text-primary hover:underline"
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

const RecentPhoneLinesTable = ({
  rows
}: {
  rows: { id: string; number: string | null; status?: string | null }[];
}) => {
  return (
    <div className="w-full overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Número</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-24 text-right"> </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="rounded-b-lg">
              <TableCell className="font-medium" colSpan={4}>
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Receipt />
                    </EmptyMedia>
                    <EmptyTitle>Nenhuma linha encontrada</EmptyTitle>
                    <EmptyDescription>
                      Você ainda não cadastrou nenhuma linha telefônica. Comece
                      cadastrando sua primeira linha telefônica ou importe uma
                      fatura.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent>
                    <div className="flex flex-wrap gap-2 *:mx-auto">
                      <Button>
                        <Plus /> Cadastrar linha
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
            rows.map((line) => (
              <TableRow key={line.id} className="odd:bg-muted/50">
                <TableCell className="text-sm font-medium">
                  {formatPhoneNumber(line.number!) ?? '—'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  <span
                    className={
                      line.status === 'ACTIVE'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-muted-foreground'
                    }
                  >
                    {formatPhoneLineStatus(line.status) ?? '—'}
                  </span>
                </TableCell>
                <TableCell>
                  <div className="flex justify-end gap-2">
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
                                className="text-primary hover:underline"
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
