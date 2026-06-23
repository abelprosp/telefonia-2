import { createFileRoute, Navigate } from '@tanstack/react-router';

import { useAuthRoles } from '@/lib/auth-roles';

import { DashboardView } from './-components/dashboard/dashboard-view';

const RouteComponent = () => {
  const { isPartnerOnly, canAccessOperations, canAccessFinance } = useAuthRoles();

  if (isPartnerOnly) {
    return <Navigate to="/partner" />;
  }

  if (canAccessFinance && !canAccessOperations) {
    return <Navigate to="/finance" />;
  }

  return (
    <div className="@container/main flex flex-1 flex-col gap-4 p-6 md:gap-6 md:p-6">
      <DashboardView />
    </div>
  );
};

export const Route = createFileRoute('/__app/')({
  component: RouteComponent
});
