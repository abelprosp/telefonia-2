import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

import { StockLinesList } from './-components/stock-lines-list';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Cadastros' },
        { label: 'Estoque de linhas' }
      ]}
    >
      <StockLinesList />
    </PageWrapper>
  );
};

export const Route = createFileRoute('/__app/stock/')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    return { page, pageSize };
  },
  component: RouteComponent
});
