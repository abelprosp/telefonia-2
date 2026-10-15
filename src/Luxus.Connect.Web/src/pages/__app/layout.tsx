import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';


import { AppTopBar } from '@/components/app-top-bar';
import { AuthConfigError } from '@/components/auth-config-error';
import { LayoutSidebar } from '@/components/layout-sidebar';
import { MfaGuidanceBanner } from '@/components/mfa-guidance-banner';
import { PageLoader } from '@/components/page-loader';
import { SignedIn } from '@/components/signed-in';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { getAuthConfigHint } from '@/lib/auth-config-hint';
import { resolveAuthRoleState } from '@/lib/auth-roles';
import { getCurrentRoles } from '@/lib/client';
import { useAuth } from '@/providers/auth';

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
            hint={getAuthConfigHint(error.message)}
          />
        ) : isLoading ? (
          <PageLoader label="Carregando..." />
        ) : (
          <SignedIn>
            <AppTopBar />
            <MfaGuidanceBanner />
            <div
              id="conteudo-principal"
              className="flex flex-1 flex-col"
              tabIndex={-1}
            >
              <Outlet />
            </div>
          </SignedIn>
        )}
      </SidebarInset>
    </SidebarProvider>
  );
};

export const Route = createFileRoute('/__app')({
  // Guarda de papéis no nível da rota (defesa em profundidade; a API também valida).
  // Espelha a política já aplicada em LayoutSidebar, mas antes de renderizar a página.
  beforeLoad: ({ location }) => {
    const token = getCurrentRoles();
    // Sem token ainda não há papéis para avaliar; o componente SignedIn trata o login.
    if (!token) {
      return;
    }
    const state = resolveAuthRoleState(token);
    const path = location.pathname;

    if (state.isCustomerPortal) {
      if (!path.startsWith('/portal') && path !== '/settings') {
        throw redirect({ to: '/portal' });
      }
      return;
    }

    if (state.isInternalStaff && path.startsWith('/portal')) {
      throw redirect({ to: '/' });
    }

    if (
      state.isViewer &&
      !state.canAccessOperations &&
      !state.canAccessFinance &&
      path !== '/' &&
      path !== '/settings'
    ) {
      throw redirect({ to: '/' });
    }

    if (state.isPartnerOnly) {
      if (
        path.startsWith('/portal') ||
        (!path.startsWith('/partner') && path !== '/settings')
      ) {
        throw redirect({ to: '/partner' });
      }
      return;
    }

    if (
      state.canAccessFinance &&
      !state.canAccessOperations &&
      !path.startsWith('/finance') &&
      path !== '/settings'
    ) {
      throw redirect({ to: '/finance' });
    }

    if (
      state.canAccessOperations &&
      !state.canAccessFinance &&
      (path.startsWith('/finance') || path.startsWith('/users'))
    ) {
      throw redirect({ to: '/' });
    }
  },
  component: RouteComponent
});
