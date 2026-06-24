import { createFileRoute, Link } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';

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
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-2">
          <Button variant="secondary" size="sm" render={<Link to="/stock" />}>
            Estoque de linhas
          </Button>
          <Button variant="outline" size="sm" render={<Link to="/stock/devices" />}>
            Estoque de aparelhos
          </Button>
        </div>
        <StockLinesList />
      </div>
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
