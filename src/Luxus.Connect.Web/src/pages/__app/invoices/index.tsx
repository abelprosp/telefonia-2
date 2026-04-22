import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

import { InvoicesList } from './-components/invoices-list';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Faturamento' },
        { label: 'Faturas' }
      ]}
    >
      <InvoicesList />
    </PageWrapper>
  );
};

export const Route = createFileRoute('/__app/invoices/')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    const rawPm = search.processingMonthId;
    const processingMonthId: string | undefined =
      typeof rawPm === 'string' && rawPm.trim() !== ''
        ? rawPm.trim()
        : undefined;
    return { page, pageSize, processingMonthId };
  },
  component: RouteComponent
});
