import { createFileRoute } from '@tanstack/react-router';

import { PartnerDashboardView } from './-components/partner-dashboard-view';

const RouteComponent = () => (
  <div className="@container/main flex flex-1 flex-col gap-4 p-6 md:gap-6 md:p-6">
    <PartnerDashboardView />
  </div>
);

export const Route = createFileRoute('/__app/partner/')({
  component: RouteComponent
});
