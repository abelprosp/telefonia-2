import { createFileRoute } from '@tanstack/react-router';

import { PageWrapper } from '@/components/page-wrapper';

import { DashboardView } from './-components/dashboard/dashboard-view';

const RouteComponent = () => {
  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }]}>
      <DashboardView />
    </PageWrapper>
  );
};

export const Route = createFileRoute('/__app/')({
  component: RouteComponent
});
