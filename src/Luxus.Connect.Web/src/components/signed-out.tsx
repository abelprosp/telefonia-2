import { useEffect, type ReactNode } from 'react';

import { useNavigate } from '@tanstack/react-router';

import { useAuth } from '@/providers/auth';

import { PageLoader } from './page-loader';

interface SignedOutProps {
  redirectTo: string;
  children: ReactNode;
}

export const SignedOut = ({
  redirectTo = '/',
  children
}: SignedOutProps): ReactNode => {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: redirectTo });
    }
  }, [isAuthenticated, redirectTo, navigate]);

  if (isLoading) {
    return <PageLoader label="Carregando..." />;
  }

  return isAuthenticated ? null : children;
};
