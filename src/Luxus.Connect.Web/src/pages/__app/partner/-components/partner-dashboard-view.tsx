import { Link } from '@tanstack/react-router';
import { ClipboardList, Phone, Users, Wallet } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
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
  formatPhoneLineStatus,
  formatPhoneNumber
} from '@/lib/format';
import {
  usePartnerCustomers,
  usePartnerDashboardStats,
  usePartnerLineRequests
} from '@/lib/partner-api';
import { parseTotalCount } from '@/lib/query-utils';

import { DashboardMetricCard } from '../../-components/dashboard/dashboard-metric-card';

const formatCount = (value: number) =>
  new Intl.NumberFormat('pt-BR').format(value);

const formatMoney = (value: number) =>
  new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  }).format(value);

export function PartnerDashboardView() {
  const statsQuery = usePartnerDashboardStats();
  const customersQuery = usePartnerCustomers({ page_index: 0, page_size: 5 });
  const requestsQuery = usePartnerLineRequests({ page_index: 0, page_size: 5 });

  const pending =
    statsQuery.isPending || customersQuery.isPending || requestsQuery.isPending;

  if (pending) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-10 w-72 rounded-xl" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-2xl" />
          ))}
        </div>
        <Skeleton className="h-80 rounded-2xl" />
      </div>
    );
  }

  const error =
    statsQuery.error ?? customersQuery.error ?? requestsQuery.error;

  if (error) {
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-2xl border px-4 py-3 text-sm">
        {isApiHttpError(error) ? error.message : getErrorMessage(error)}
      </div>
    );
  }

  const stats = statsQuery.data;
  if (!stats) {
    return null;
  }

  const metrics = [
    {
      title: 'Meus clientes',
      value: formatCount(parseTotalCount(stats.customers_count)),
      icon: Users,
      to: '/partner/customers',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Linhas vendidas',
      value: formatCount(parseTotalCount(stats.phone_lines_count)),
      icon: Phone,
      to: '/partner/phone-lines',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Solicitações pendentes',
      value: formatCount(parseTotalCount(stats.pending_operation_requests_count)),
      icon: ClipboardList,
      to: '/partner/requests',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Receita estimada',
      value: formatMoney(stats.total_cost_with_consumption ?? 0),
      icon: Wallet,
      to: '/partner/financial',
      search: { page: 1, pageSize: 10 }
    }
  ];

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          Área do parceiro
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Resumo financeiro e operacional das suas vendas
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <DashboardMetricCard key={metric.title} {...metric} />
        ))}
      </div>

      <div className="dashboard-card overflow-hidden">
        <div className="flex items-center justify-between border-b px-5 py-4">
          <div>
            <h3 className="text-lg font-semibold">Financeiro</h3>
            <p className="text-muted-foreground text-sm">
              Custos das linhas vinculadas aos seus clientes
            </p>
          </div>
        </div>
        <div className="grid gap-4 px-5 py-4 sm:grid-cols-2">
          <div className="rounded-xl border bg-muted/30 p-4">
            <p className="text-muted-foreground text-sm">Custo base</p>
            <p className="mt-1 text-2xl font-semibold">
              {formatMoney(stats.total_base_cost ?? 0)}
            </p>
          </div>
          <div className="rounded-xl border bg-muted/30 p-4">
            <p className="text-muted-foreground text-sm">Com consumo</p>
            <p className="mt-1 text-2xl font-semibold">
              {formatMoney(stats.total_cost_with_consumption ?? 0)}
            </p>
          </div>
        </div>
      </div>

      <RecentCustomers rows={customersQuery.data?.items ?? []} />
      <RecentRequests rows={requestsQuery.data?.items ?? []} />
    </div>
  );
}

function RecentCustomers({
  rows
}: {
  rows: { id: string; name: string; cpf_cnpj: string; active: boolean }[];
}) {
  return (
    <div className="dashboard-card overflow-hidden">
      <div className="flex items-center justify-between border-b px-5 py-4">
        <div>
          <h3 className="text-lg font-semibold">Clientes recentes</h3>
          <p className="text-muted-foreground text-sm">Cadastros sob sua carteira</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          render={
            <Link to="/partner/customers" search={{ page: 1, pageSize: 10 }} />
          }
        >
          Ver todos
        </Button>
      </div>
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>Cliente</TableHead>
            <TableHead>Documento</TableHead>
            <TableHead>Situação</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3} className="text-muted-foreground text-center">
                Nenhum cliente cadastrado ainda.
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell className="font-medium">{row.name}</TableCell>
                <TableCell>{formatCpfCnpj(row.cpf_cnpj)}</TableCell>
                <TableCell>{row.active ? 'Ativo' : 'Inativo'}</TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

function RecentRequests({
  rows
}: {
  rows: {
    id: string;
    phone_line_number: string;
    operation_type: string;
    status: string;
  }[];
}) {
  return (
    <div className="dashboard-card overflow-hidden">
      <div className="flex items-center justify-between border-b px-5 py-4">
        <div>
          <h3 className="text-lg font-semibold">Solicitações recentes</h3>
          <p className="text-muted-foreground text-sm">
            Ativação e desativação de linhas
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          render={
            <Link to="/partner/requests" search={{ page: 1, pageSize: 10 }} />
          }
        >
          Acompanhar
        </Button>
      </div>
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead>Linha</TableHead>
            <TableHead>Operação</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3} className="text-muted-foreground text-center">
                Nenhuma solicitação enviada.
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell className="font-medium">
                  {formatPhoneNumber(row.phone_line_number) ?? row.phone_line_number}
                </TableCell>
                <TableCell>
                  {row.operation_type === 'activation' ? 'Ativação' : 'Desativação'}
                </TableCell>
                <TableCell>{formatOperationStatus(row.status)}</TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

function formatOperationStatus(status: string) {
  switch (status) {
    case 'pending':
      return 'Pendente';
    case 'approved':
      return 'Aprovada';
    case 'rejected':
      return 'Rejeitada';
    case 'cancelled':
      return 'Cancelada';
    default:
      return formatPhoneLineStatus(status) ?? status;
  }
}
