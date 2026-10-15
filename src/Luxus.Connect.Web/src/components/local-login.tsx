import { useState, type FormEvent } from 'react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { authRequest, useAuth } from '@/providers/auth';

export function MfaSetup({ onComplete }: { onComplete: () => Promise<void> }) {
  const [setup, setSetup] = useState<{ secret: string; qr: string }>();
  const [code, setCode] = useState('');
  const [codes, setCodes] = useState<string[]>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    setError('');
    try {
      await action();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  if (codes)
    return (
      <div className="space-y-4">
        <h2 className="font-semibold">Guarde seus códigos de recuperação</h2>
        <p className="text-sm">
          Cada código pode ser usado uma vez se você perder o celular. Guarde em
          local seguro; eles só aparecem agora.
        </p>
        <textarea
          aria-label="Códigos de recuperação"
          readOnly
          value={codes.join('\n')}
          className="h-56 w-full rounded border p-3 font-mono text-xs"
        />
        <Button disabled={busy} onClick={() => void run(onComplete)}>
          Guardei os códigos. Continuar
        </Button>
        {error && <p role="alert">{error}</p>}
      </div>
    );
  return (
    <div className="space-y-4">
      <p className="text-sm">
        Use Google Authenticator, Microsoft Authenticator ou outro aplicativo
        compatível.
      </p>
      {!setup ? (
        <Button
          disabled={busy}
          onClick={() =>
            void run(async () => setSetup(await authRequest('mfa/setup', {})))
          }
        >
          Configurar autenticador
        </Button>
      ) : (
        <>
          <img
            src={`data:image/png;base64,${setup.qr}`}
            alt="QR code para configurar o autenticador"
            className="size-56"
          />
          <Label htmlFor="mfa-secret">Chave para adicionar manualmente</Label>
          <Input
            id="mfa-secret"
            readOnly
            value={setup.secret}
            className="font-mono"
          />
          <form
            className="space-y-3"
            onSubmit={(e) => {
              e.preventDefault();
              void run(async () => {
                const result = await authRequest<{ recovery_codes: string[] }>(
                  'mfa/confirm',
                  { code }
                );
                setSetup(undefined);
                setCode('');
                setCodes(result.recovery_codes);
              });
            }}
          >
            <Label htmlFor="mfa-enroll-code">Código de seis dígitos</Label>
            <Input
              id="mfa-enroll-code"
              inputMode="numeric"
              autoComplete="one-time-code"
              pattern="[0-9]{6}"
              maxLength={6}
              required
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
            <Button disabled={busy}>Confirmar autenticador</Button>
          </form>
        </>
      )}
      {error && (
        <p role="alert" className="text-destructive text-sm">
          {error}
        </p>
      )}
    </div>
  );
}

export function LocalLogin() {
  const auth = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      if (auth.isAuthenticated) {
        await authRequest('mfa/verify', { code });
        await auth.refresh();
      } else {
        await auth.login(username, password);
        setPassword('');
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <main className="bg-muted/30 flex min-h-screen w-full items-center justify-center p-4">
      <section className="bg-background w-full max-w-md space-y-6 rounded-2xl border p-6 shadow-sm">
        <div>
          <h1 className="text-2xl font-semibold">
            {auth.isAuthenticated ? 'Proteja sua conta' : 'Entrar no Telefonia'}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm">
            {auth.isAuthenticated
              ? 'Confirme sua identidade para continuar.'
              : 'Use seu usuário e senha.'}
          </p>
        </div>
        {auth.isAuthenticated && !auth.mfaEnrolled ? (
          <MfaSetup onComplete={auth.refresh} />
        ) : (
          <form onSubmit={(e) => void submit(e)} className="space-y-4">
            {auth.isAuthenticated ? (
              <>
                <Label htmlFor="login-otp">
                  Código do autenticador ou de recuperação
                </Label>
                <Input
                  id="login-otp"
                  autoComplete="one-time-code"
                  required
                  maxLength={64}
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                />
              </>
            ) : (
              <>
                <Label htmlFor="login-user">Usuário</Label>
                <Input
                  id="login-user"
                  autoComplete="username"
                  required
                  maxLength={128}
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
                <Label htmlFor="login-password">Senha</Label>
                <Input
                  id="login-password"
                  type="password"
                  autoComplete="current-password"
                  required
                  maxLength={256}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </>
            )}
            <Button className="w-full" disabled={busy}>
              {busy ? 'Verificando…' : 'Entrar'}
            </Button>
          </form>
        )}
        {error && (
          <p role="alert" className="text-destructive text-sm">
            {error}
          </p>
        )}
        {auth.isAuthenticated ? (
          <Button
            variant="outline"
            onClick={() =>
              void auth
                .signoutRedirect()
                .catch((e: Error) => setError(e.message))
            }
          >
            Sair desta conta
          </Button>
        ) : (
          <p className="text-muted-foreground text-sm">
            Esqueceu a senha? Solicite a redefinição ao administrador.
          </p>
        )}
      </section>
    </main>
  );
}
