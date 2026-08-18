import { createFileRoute, Outlet } from '@tanstack/react-router';
import { useAuth } from 'react-oidc-context';

import { AppTopBar } from '@/components/app-top-bar';
import { AuthConfigError } from '@/components/auth-config-error';
import { LayoutSidebar } from '@/components/layout-sidebar';
import { MfaGuidanceBanner } from '@/components/mfa-guidance-banner';
import { PageLoader } from '@/components/page-loader';
import { SignedIn } from '@/components/signed-in';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { getAuthConfigHint } from '@/lib/auth-config-hint';

const RouteComponent = () => {
  const { isLoading, error } = useAuth();

  return (
    <SidebarProvider>
      <a href="#conteudo-principal" className="skip-link">
        Pular para o conteúdo
      </a>
      <LayoutSidebar variant="inset" />
      <SidebarInset className="bg-muted/30">
        {error ? (
          <AuthConfigError
            message={error.message}
            hint={getAuthConfigHint()}
          />
        ) : isLoading ? (
          <PageLoader label="Carregando..." />
        ) : (
          <SignedIn>
            <AppTopBar />
            <MfaGuidanceBanner />
            <div id="conteudo-principal" className="flex flex-1 flex-col" tabIndex={-1}>
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
