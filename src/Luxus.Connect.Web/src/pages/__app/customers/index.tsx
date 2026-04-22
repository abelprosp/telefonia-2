import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

import { CustomersList } from './-components/customers-list';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Cadastros' },
        { label: 'Clientes' }
      ]}
    >
      <CustomersList />
    </PageWrapper>
  );
};

export const Route = createFileRoute('/__app/customers/')({
  validateSearch: (search: Record<string, unknown>) => {
    const page = Math.max(1, Number(search.page) || 1);
    const rawSize = Number(search.pageSize) || 10;
    const pageSize = [10, 25, 50].includes(rawSize) ? rawSize : 10;
    const providerId: string | undefined =
      typeof search.providerId === 'string' && search.providerId.length > 0
        ? search.providerId
        : undefined;
    return { page, pageSize, providerId };
  },
  component: RouteComponent
});
