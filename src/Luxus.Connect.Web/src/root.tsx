import '@/index.css';
import '@/polyfills';

import { RouterProvider, createRouter } from '@tanstack/react-router';
import { createRoot } from 'react-dom/client';
import { AuthProvider } from 'react-oidc-context';

// Import the generated route tree
import { env } from '@/env';
import { AppProvider } from '@/providers/app';
// import { AuthProvider } from '@/providers/auth';
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

const oidcConfig = {
  authority: `${env.VITE_AUTH_URL}/realms/luxus`,
  client_id: env.VITE_CLIENT_ID,
  redirect_uri: window.location.origin,
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
  automaticSilentRenew: true // Automatically refresh the access token
};

createRoot(document.getElementById('root')!).render(
  <AuthProvider {...oidcConfig}>
    <AppProvider defaultTheme="dark" storageKey="luxus-connect">
      <RouterProvider router={router} />
    </AppProvider>
  </AuthProvider>
);
