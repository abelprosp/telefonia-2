import { useState, type FormEvent } from 'react';

import { Copy, Download, Eye, EyeOff, ShieldCheck } from 'lucide-react';

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
  const [saved, setSaved] = useState(false);
  const [copied, setCopied] = useState(false);
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
        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            onClick={() =>
              void run(async () => {
                await navigator.clipboard.writeText(codes.join('\n'));
                setCopied(true);
              })
            }
          >
            <Copy className="size-4" aria-hidden />{' '}
            {copied ? 'Copiados' : 'Copiar códigos'}
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              const url = URL.createObjectURL(
                new Blob([codes.join('\n')], { type: 'text/plain' })
              );
              const link = document.createElement('a');
              link.href = url;
              link.download = 'telefonia-codigos-recuperacao.txt';
              link.click();
              setTimeout(() => URL.revokeObjectURL(url), 1000);
            }}
          >
            <Download className="size-4" aria-hidden /> Baixar códigos
          </Button>
        </div>
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            checked={saved}
            onChange={(e) => setSaved(e.target.checked)}
            className="mt-1"
          />
          Guardei os códigos em um local seguro, separado do celular.
        </label>
        <Button disabled={busy || !saved} onClick={() => void run(onComplete)}>
          Guardei os códigos. Continuar
        </Button>
        {error && <p role="alert">{error}</p>}
      </div>
    );
  return (
    <div className="space-y-4">
      <p className="text-sm">
        Configure uma vez. Nos próximos acessos, informe o código de seis
        dígitos do Google Authenticator, Microsoft Authenticator ou outro
        autenticador. O código funciona mesmo sem internet no celular.
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
          <p className="text-sm">
            1. No autenticador, adicione uma conta e escaneie este QR code.
          </p>
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
          <p className="text-muted-foreground text-xs">
            Se estiver no mesmo celular, copie a chave acima e adicione uma
            conta manualmente. Nunca compartilhe esta chave ou o QR code.
          </p>
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
            <Label htmlFor="mfa-enroll-code">
              2. Digite o código que aparece no aplicativo
            </Label>
            <Input
              id="mfa-enroll-code"
              inputMode="numeric"
              autoComplete="one-time-code"
              pattern="[0-9]{6}"
              maxLength={6}
              required
              autoFocus
              value={code}
              onChange={(e) =>
                setCode(e.target.value.replace(/\D/g, '').slice(0, 6))
              }
            />
            <Button disabled={busy}>Confirmar autenticador</Button>
          </form>
          <Button
            variant="ghost"
            disabled={busy}
            onClick={() => {
              setSetup(undefined);
              setCode('');
              setError('');
            }}
          >
            Reiniciar configuração
          </Button>
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
  const [showPassword, setShowPassword] = useState(false);
  const [useRecovery, setUseRecovery] = useState(false);
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      if (auth.isAuthenticated) {
        await authRequest('mfa/verify', { code: code.trim() });
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
          <ShieldCheck className="text-primary mb-3 size-8" aria-hidden />
          <h1 className="text-2xl font-semibold">
            {auth.isAuthenticated
              ? auth.mfaEnrolled
                ? 'Confirme seu acesso'
                : 'Configure a segurança da conta'
              : 'Entrar no Telefonia'}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm">
            {auth.isAuthenticated
              ? auth.mfaEnrolled
                ? 'Abra seu aplicativo autenticador para continuar.'
                : 'Proteja seu acesso com um autenticador no celular.'
              : 'Use seu usuário e senha. Tudo é feito aqui no Telefonia.'}
          </p>
        </div>
        {auth.sessionExpired && !auth.isAuthenticated && (
          <p role="status" className="bg-muted rounded-lg p-3 text-sm">
            Sua sessão expirou. Entre novamente para continuar.
          </p>
        )}
        {auth.isAuthenticated && !auth.mfaEnrolled ? (
          <MfaSetup onComplete={auth.refresh} />
        ) : (
          <form onSubmit={(e) => void submit(e)} className="space-y-4">
            {auth.isAuthenticated ? (
              <>
                <Label htmlFor="login-otp">
                  {useRecovery
                    ? 'Código de recuperação'
                    : 'Código do autenticador'}
                </Label>
                <Input
                  id="login-otp"
                  autoComplete="one-time-code"
                  required
                  key={useRecovery ? 'recovery' : 'totp'}
                  autoFocus
                  inputMode={useRecovery ? 'text' : 'numeric'}
                  pattern={useRecovery ? undefined : '[0-9]{6}'}
                  maxLength={useRecovery ? 64 : 6}
                  placeholder={
                    useRecovery ? 'Cole um código guardado' : '000000'
                  }
                  aria-describedby="mfa-code-help"
                  value={code}
                  onChange={(e) =>
                    setCode(
                      useRecovery
                        ? e.target.value
                        : e.target.value.replace(/\D/g, '').slice(0, 6)
                    )
                  }
                />
                <p id="mfa-code-help" className="text-muted-foreground text-xs">
                  {useRecovery
                    ? 'Cada código de recuperação funciona uma única vez.'
                    : 'O código muda a cada 30 segundos. Se não funcionar, ative a data e a hora automáticas no celular.'}
                </p>
                <Button
                  type="button"
                  variant="link"
                  className="h-auto px-0 text-sm"
                  onClick={() => {
                    setUseRecovery(!useRecovery);
                    setCode('');
                    setError('');
                  }}
                >
                  {useRecovery
                    ? 'Usar o aplicativo autenticador'
                    : 'Perdi o celular. Usar código de recuperação'}
                </Button>
              </>
            ) : (
              <>
                <Label htmlFor="login-user">Usuário</Label>
                <Input
                  id="login-user"
                  autoComplete="username"
                  autoCapitalize="none"
                  spellCheck={false}
                  autoFocus
                  required
                  maxLength={128}
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
                <Label htmlFor="login-password">Senha</Label>
                <div className="relative">
                  <Input
                    id="login-password"
                    type={showPassword ? 'text' : 'password'}
                    className="pr-12"
                    autoComplete="current-password"
                    required
                    maxLength={256}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="absolute top-0 right-0"
                    aria-label={
                      showPassword ? 'Ocultar senha' : 'Mostrar senha'
                    }
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <EyeOff className="size-4" aria-hidden />
                    ) : (
                      <Eye className="size-4" aria-hidden />
                    )}
                  </Button>
                </div>
              </>
            )}
            <Button className="w-full" disabled={busy}>
              {busy
                ? 'Verificando…'
                : auth.isAuthenticated
                  ? 'Confirmar e entrar'
                  : 'Continuar'}
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
              void auth.logout().catch((e: Error) => setError(e.message))
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
