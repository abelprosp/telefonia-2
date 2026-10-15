import '@/index.css';
import '@/polyfills';

import { type ReactNode } from 'react';

import { RouterProvider, createRouter } from '@tanstack/react-router';
import { createRoot } from 'react-dom/client';

import { LocalLogin } from '@/components/local-login';
import { PageLoader } from '@/components/page-loader';
import { AppProvider } from '@/providers/app';
import { AuthProvider, useAuth } from '@/providers/auth';
import { routeTree } from '@/route-tree.gen';

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

// Create a new router instance
const router = createRouter({
  routeTree,
  context: {},
  defaultPreload: 'intent',
  scrollRestoration: true,
  defaultStructuralSharing: true,
  defaultPreloadStaleTime: 0
});

function AuthGate({ children }: { children: ReactNode }) {
  const auth = useAuth();
  if (auth.isLoading) return <PageLoader label="Verificando sessão..." />;
  if (auth.error)
    return (
      <div role="alert" className="p-8">
        <p>{auth.error.message}</p>
        <button onClick={() => window.location.reload()}>
          Tentar novamente
        </button>
      </div>
    );
  if (!auth.isAuthenticated || auth.mfaRequired) return <LocalLogin />;
  return children;
}

createRoot(document.getElementById('root')!).render(
  <AuthProvider>
    <AuthGate>
      <AppProvider defaultTheme="light" storageKey="luxus-connect">
        <RouterProvider router={router} />
      </AppProvider>
    </AuthGate>
  </AuthProvider>
);
