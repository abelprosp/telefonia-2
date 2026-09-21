import { LogIn, RefreshCw } from 'lucide-react';
import { useAuth } from 'react-oidc-context';

import { Button } from '@/components/ui/button';

interface AuthConfigErrorProps {
  title?: string;
  message: string;
  hint?: string;
}

export const AuthConfigError = ({
  title = 'Falha na autenticação',
  message,
  hint
}: AuthConfigErrorProps) => {
  const { signinRedirect, removeUser } = useAuth();

  const isStateError =
    message.toLowerCase().includes('no matching state') ||
    message.toLowerCase().includes('state does not match') ||
    message.toLowerCase().includes('stale state');
  const isScopeError =
    message.toLowerCase().includes('invalid_scope') ||
    message.toLowerCase().includes('invalid scopes');

  const handleRetry = async () => {
    try {
      // Drop OIDC error/code/state query params so retry does not re-parse a failed callback.
      window.history.replaceState({}, document.title, window.location.pathname);
      sessionStorage.clear();
      localStorage.removeItem('luxus_last_auth_redirect');
      if (removeUser) {
        await removeUser();
      }
      await signinRedirect();
    } catch {
      window.location.href = window.location.origin;
    }
  };

  return (
    <div className="flex h-full min-h-[60vh] w-full flex-col items-center justify-center gap-4 p-8 text-center">
      <a href="#login-retry" className="skip-link">
        Pular para o login
      </a>
      <div className="rounded-full bg-destructive/10 p-3 text-destructive">
        <RefreshCw className="size-6" aria-hidden />
      </div>
      <h1 className="text-xl font-bold">{title}</h1>
      <p className="max-w-md text-sm">
        {isStateError
          ? 'A sessão de login anterior expirou ou o redirecionamento foi interrompido.'
          : isScopeError
            ? 'O Keycloak rejeitou os scopes do cliente (realm desatualizado). Veja o detalhe abaixo.'
            : message}
      </p>
      {hint && (
        <p className="max-w-lg rounded-lg border bg-muted/50 p-3 text-xs whitespace-pre-wrap">
          {hint}
        </p>
      )}
      <Button id="login-retry" onClick={handleRetry} className="mt-2 gap-2">
        <LogIn className="size-4" aria-hidden />
        Fazer Login Novamente
      </Button>
    </div>
  );
};
