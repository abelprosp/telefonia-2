import '@/index.css';
import '@/polyfills';

import { RouterProvider, createRouter } from '@tanstack/react-router';
import { WebStorageStateStore } from 'oidc-client-ts';
import { createRoot } from 'react-dom/client';
import { AuthProvider, type AuthProviderProps } from 'react-oidc-context';

// Import the generated route tree
import { env } from '@/env';
import { MemoryStorage } from '@/lib/oidc-memory-storage';
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

// Tokens stay in memory (XSS blast radius). OIDC redirect state still needs
// sessionStorage so the auth callback survives the Keycloak round-trip.
const oidcUserStore = new WebStorageStateStore({ store: new MemoryStorage() });
const oidcStateStore = new WebStorageStateStore({ store: window.sessionStorage });

const oidcConfig: AuthProviderProps = {
  authority: `${env.VITE_AUTH_URL.replace(/\/+$/, '')}/realms/luxus`,
  client_id: env.VITE_CLIENT_ID,
  scope: 'openid tenant-organization',
  redirect_uri: window.location.origin,
  userStore: oidcUserStore,
  stateStore: oidcStateStore,
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
  automaticSilentRenew: true
};


createRoot(document.getElementById('root')!).render(
  <AuthProvider {...oidcConfig}>
    <AppProvider defaultTheme="light" storageKey="luxus-connect">
      <RouterProvider router={router} />
    </AppProvider>
  </AuthProvider>
);
