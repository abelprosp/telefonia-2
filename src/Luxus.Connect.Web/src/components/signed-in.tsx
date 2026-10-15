import type { ReactNode } from 'react';

import { useAuth } from '@/providers/auth';

import { LocalLogin } from './local-login';
import { PageLoader } from './page-loader';

export function SignedIn({ children }: { children: ReactNode }) {
  const auth = useAuth();
  if (auth.isLoading) return <PageLoader label="Verificando sessão..." />;
  return auth.isAuthenticated && !auth.mfaRequired ? children : <LocalLogin />;
}
