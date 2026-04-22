import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Relatórios' },
        { label: 'Linhas em transição' }
      ]}
    />
  );
};

export const Route = createFileRoute('/__app/reports/transition-pending/')({
  component: RouteComponent
});
