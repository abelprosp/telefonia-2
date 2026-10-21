import { LogIn, RefreshCw } from 'lucide-react';

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
  const handleRetry = () => window.location.reload();

  return (
    <div className="flex h-full min-h-[60vh] w-full flex-col items-center justify-center gap-4 p-8 text-center">
      <a href="#login-retry" className="skip-link">
        Pular para o login
      </a>
      <div className="bg-destructive/10 text-destructive rounded-full p-3">
        <RefreshCw className="size-6" aria-hidden />
      </div>
      <h1 className="text-xl font-bold">{title}</h1>
      <p className="max-w-md text-sm">{message}</p>
      {hint && (
        <p className="bg-muted/50 max-w-lg rounded-lg border p-3 text-xs whitespace-pre-wrap">
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
