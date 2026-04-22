import { useEffect, type ReactNode } from 'react';

import { useAuth } from 'react-oidc-context';

import { PageLoader } from './page-loader';

interface SignedInProps {
  children: ReactNode;
}

export const SignedIn = ({ children }: SignedInProps) => {
  const { isAuthenticated, signinRedirect } = useAuth();

  useEffect(() => {
    if (!isAuthenticated) {
      signinRedirect();
      return;
    }
  }, [isAuthenticated, signinRedirect]);

  return isAuthenticated ? children : <PageLoader label="Carregando..." />;
};
