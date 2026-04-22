import { createFileRoute, Outlet } from '@tanstack/react-router';
import { useAuth } from 'react-oidc-context';

import { LayoutSidebar } from '@/components/layout-sidebar';
import { PageLoader } from '@/components/page-loader';
import { SignedIn } from '@/components/signed-in';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';

const RouteComponent = () => {
  const { isLoading } = useAuth();

  return (
    <SidebarProvider>
      <LayoutSidebar variant="inset" />
      <SidebarInset>
        {isLoading ? (
          <PageLoader label="Carregando..." />
        ) : (
          <SignedIn>
            <Outlet />
          </SignedIn>
        )}
      </SidebarInset>
    </SidebarProvider>
  );
};

export const Route = createFileRoute('/__app')({
  component: RouteComponent
});
