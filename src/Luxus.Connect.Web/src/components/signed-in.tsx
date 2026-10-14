import { useEffect, type ReactNode } from 'react';

import { useAuth } from 'react-oidc-context';

import { AuthConfigError } from './auth-config-error';
import { PageLoader } from './page-loader';

interface SignedInProps {
  children: ReactNode;
}

export const SignedIn = ({ children }: SignedInProps) => {
  const { isAuthenticated, signinRedirect, isLoading, error, activeNavigator } =
    useAuth();

  useEffect(() => {
    if (isLoading || isAuthenticated || error || activeNavigator) {
      return;
    }

    void signinRedirect().catch(() => {
      // error state is set by react-oidc-context
    });
  }, [isAuthenticated, signinRedirect, isLoading, error, activeNavigator]);

  if (error) {
    return (
      <AuthConfigError
        message={error.message}
        hint={
          import.meta.env.DEV
            ? 'Verifique se o Keycloak está a correr em http://localhost:8081 e se o realm «luxus» existe.'
            : 'Verifique se o provedor de identidade está disponível e tente novamente.'
        }
      />
    );
  }

  if (isLoading || activeNavigator === 'signinRedirect') {
    return <PageLoader label="A redirecionar para o login..." />;
  }

  return isAuthenticated ? children : <PageLoader label="Carregando..." />;
};
