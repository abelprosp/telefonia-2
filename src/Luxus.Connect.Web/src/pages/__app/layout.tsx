import { createFileRoute, Outlet } from '@tanstack/react-router';
import { useAuth } from 'react-oidc-context';

import { AppTopBar } from '@/components/app-top-bar';
import { AuthConfigError } from '@/components/auth-config-error';
import { LayoutSidebar } from '@/components/layout-sidebar';
import { PageLoader } from '@/components/page-loader';
import { SignedIn } from '@/components/signed-in';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';

const RouteComponent = () => {
  const { isLoading, error } = useAuth();

  return (
    <SidebarProvider>
      <LayoutSidebar variant="inset" />
      <SidebarInset className="bg-muted/30">
        {error ? (
          <AuthConfigError
            message={error.message}
            hint="Confirme VITE_AUTH_URL=http://localhost:8081 no ficheiro .env do frontend."
          />
        ) : isLoading ? (
          <PageLoader label="Carregando..." />
        ) : (
          <SignedIn>
            <AppTopBar />
            <div className="flex flex-1 flex-col">
              <Outlet />
            </div>
          </SignedIn>
        )}
      </SidebarInset>
    </SidebarProvider>
  );
};

export const Route = createFileRoute('/__app')({
  component: RouteComponent
});
