interface AuthConfigErrorProps {
  title?: string;
  message: string;
  hint?: string;
}

export const AuthConfigError = ({
  title = 'Falha na autenticação',
  message,
  hint
}: AuthConfigErrorProps) => (
  <div className="flex h-full min-h-[50vh] w-full flex-col items-center justify-center gap-3 p-8 text-center">
    <h1 className="text-lg font-semibold">{title}</h1>
    <p className="text-muted-foreground max-w-md text-sm">{message}</p>
    {hint && (
      <p className="text-muted-foreground max-w-lg text-xs whitespace-pre-wrap">{hint}</p>
    )}
  </div>
);
