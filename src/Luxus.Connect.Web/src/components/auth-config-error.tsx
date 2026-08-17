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

  const isStateError = message.toLowerCase().includes('state') || message.toLowerCase().includes('matching');

  const handleRetry = async () => {
    try {
      // Limpa os parâmetros de código e estado da URL para não reenviar state corrompido/antigo
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
      <div className="rounded-full bg-destructive/10 p-3 text-destructive">
        <RefreshCw className="size-6 animate-pulse" />
      </div>
      <h1 className="text-xl font-bold">{title}</h1>
      <p className="text-muted-foreground max-w-md text-sm">
        {isStateError
          ? 'A sessão de login anterior expirou ou o redirecionamento foi interrompido.'
          : message}
      </p>
      {hint && (
        <p className="text-muted-foreground max-w-lg text-xs whitespace-pre-wrap bg-muted/50 p-3 rounded-lg border font-mono">
          {hint}
        </p>
      )}
      <Button onClick={handleRetry} className="gap-2 mt-2">
        <LogIn className="size-4" />
        Fazer Login Novamente
      </Button>
    </div>
  );
};
