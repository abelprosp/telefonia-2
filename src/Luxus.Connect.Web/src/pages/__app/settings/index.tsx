import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

const RouteComponent = () => {
  return (
    <PageWrapper
      breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Configurações' }]}
    />
  );
};

export const Route = createFileRoute('/__app/settings/')({
  component: RouteComponent
});
