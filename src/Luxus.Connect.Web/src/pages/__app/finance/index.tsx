import { createFileRoute } from '@tanstack/react-router';

import { FinanceDashboardView } from './-components/finance-dashboard-view';

export const Route = createFileRoute('/__app/finance/')({
  component: () => (
    <div className="@container/main flex flex-1 flex-col gap-4 p-6 md:gap-6 md:p-6">
      <FinanceDashboardView />
    </div>
  )
});
