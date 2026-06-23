import { Link } from '@tanstack/react-router';
import { ArrowDownCircle, ArrowUpCircle, FileText, Handshake, Layers, Mail, Wallet } from 'lucide-react';

import { Skeleton } from '@/components/ui/skeleton';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatMoney, useFinancialSummary } from '@/lib/financial-api';

import { DashboardMetricCard } from '../../-components/dashboard/dashboard-metric-card';

export function FinanceDashboardView() {
  const summaryQuery = useFinancialSummary();

  if (summaryQuery.isPending) {
    return (
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-32 rounded-2xl" />
        ))}
      </div>
    );
  }

  if (summaryQuery.error) {
    const err = summaryQuery.error;
    return (
      <div className="border-destructive/40 bg-destructive/10 text-destructive rounded-2xl border px-4 py-3 text-sm">
        {isApiHttpError(err) ? err.message : getErrorMessage(err)}
      </div>
    );
  }

  const s = summaryQuery.data;
  if (!s) return null;

  const metrics = [
    {
      title: 'Contas a pagar',
      value: formatMoney(s.total_payable_open),
      icon: ArrowUpCircle,
      to: '/finance/payables',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Contas a receber',
      value: formatMoney(s.total_receivable_open),
      icon: ArrowDownCircle,
      to: '/finance/receivables',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Comissões pendentes',
      value: formatMoney(s.total_partner_commission_accrued),
      icon: Handshake,
      to: '/finance/partner-sales',
      search: { page: 1, pageSize: 10, status: 'accrued' }
    },
    {
      title: 'Vencidos',
      value: `${s.payable_overdue_count + s.receivable_overdue_count}`,
      icon: Wallet,
      to: '/finance/payables',
      search: { page: 1, pageSize: 10, status: 'overdue' }
    }
  ];

  const billingMetrics = [
    {
      title: 'Faturas operadora',
      value: String(s.provider_invoices_count ?? 0),
      icon: FileText,
      to: '/invoices',
      search: { page: 1, pageSize: 10, processingMonthId: undefined }
    },
    {
      title: 'Valor das faturas',
      value: formatMoney(s.provider_invoices_total_amount ?? 0),
      icon: Layers,
      to: '/invoices',
      search: { page: 1, pageSize: 10, processingMonthId: undefined }
    },
    {
      title: 'Sem conta a pagar',
      value: String(s.provider_invoices_without_payable_count ?? 0),
      icon: ArrowUpCircle,
      to: '/invoices',
      search: { page: 1, pageSize: 10, processingMonthId: undefined }
    },
    {
      title: 'Meses abertos',
      value: String(s.open_processing_months_count ?? 0),
      icon: Layers,
      to: '/processing-months',
      search: { page: 1, pageSize: 10 }
    }
  ];

  const emailMetrics = [
    {
      title: 'Faturas rascunho',
      value: String(s.billing_documents_draft_count ?? 0),
      icon: Mail,
      to: '/finance/customer-invoices',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Prontas p/ envio',
      value: String(s.billing_documents_ready_count ?? 0),
      icon: Mail,
      to: '/finance/customer-invoices',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Faturas enviadas',
      value: String(s.billing_documents_sent_count ?? 0),
      icon: Mail,
      to: '/finance/customer-invoices',
      search: { page: 1, pageSize: 10 }
    },
    {
      title: 'Inadimplentes',
      value: String(s.receivable_overdue_count ?? 0),
      icon: Wallet,
      to: '/finance/collections',
      search: { page: 1, pageSize: 10 }
    }
  ];

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Financeiro</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Controle financeiro integrado ao faturamento e refaturamento da operação
        </p>
      </div>
      <div>
        <h2 className="text-muted-foreground mb-3 text-xs font-semibold tracking-wide uppercase">
          Fluxo de caixa
        </h2>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {metrics.map((m) => (
            <DashboardMetricCard key={m.title} {...m} />
          ))}
        </div>
      </div>
      <div>
        <h2 className="text-muted-foreground mb-3 text-xs font-semibold tracking-wide uppercase">
          Faturamento e refaturamento
        </h2>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {billingMetrics.map((m) => (
            <DashboardMetricCard key={m.title} {...m} />
          ))}
        </div>
      </div>
      <div>
        <h2 className="text-muted-foreground mb-3 text-xs font-semibold tracking-wide uppercase">
          Envio de faturas e cobrança
        </h2>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {emailMetrics.map((m) => (
            <DashboardMetricCard key={m.title} {...m} />
          ))}
        </div>
      </div>
      <div className="dashboard-card p-5">
        <h3 className="text-lg font-semibold">Acesso rápido</h3>
        <div className="mt-4 flex flex-wrap gap-3">
          <Link
            to="/finance/payables"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Contas a pagar
          </Link>
          <Link
            to="/finance/receivables"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Contas a receber
          </Link>
          <Link
            to="/finance/customer-invoices"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Faturas para envio
          </Link>
          <Link
            to="/finance/collections"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Inadimplentes
          </Link>
          <Link
            to="/finance/invoice-email-templates"
            className="text-primary text-sm font-medium hover:underline"
          >
            Templates de e-mail
          </Link>
          <Link
            to="/finance/partner-sales"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Vendas de parceiros
          </Link>
          <Link
            to="/invoices"
            search={{ page: 1, pageSize: 10, processingMonthId: undefined }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Faturas da operadora
          </Link>
          <Link
            to="/processing-months"
            search={{ page: 1, pageSize: 10 }}
            className="text-primary text-sm font-medium hover:underline"
          >
            Meses de processamento
          </Link>
        </div>
      </div>
    </div>
  );
}
