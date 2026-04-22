import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

import { ProcessingMonthsList } from './-components/processing-months-list';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Faturamento' },
        { label: 'Meses de processamento' }
      ]}
    >
      <ProcessingMonthsList />
    </PageWrapper>
  );
};

export const Route = createFileRoute('/__app/processing-months/')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    return { page, pageSize };
  },
  component: RouteComponent
});
