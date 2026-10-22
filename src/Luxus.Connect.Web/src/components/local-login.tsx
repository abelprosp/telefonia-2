import { useState, type FormEvent } from 'react';

import { browserSupportsWebAuthn } from '@simplewebauthn/browser';
import {
  Copy,
  Download,
  Eye,
  EyeOff,
  Fingerprint,
  KeyRound,
  ShieldCheck
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { passkeyErrorMessage, useAuth } from '@/providers/auth';

export function PasskeySetup({
  onComplete
}: {
  onComplete: () => Promise<void>;
}) {
  const auth = useAuth();
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
      setError(passkeyErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };
  if (!browserSupportsWebAuthn())
    return (
      <p role="alert" className="text-destructive text-sm">
        Este navegador ou dispositivo não suporta passkeys. Use um navegador
        moderno ou outro dispositivo para entrar com digital, Face ID ou Windows
        Hello.
      </p>
    );
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
        Use a digital, o Face ID, o Windows Hello ou uma chave de segurança.
        Nada de código para digitar.
      </p>
      <Button
        className="w-full"
        autoFocus
        disabled={busy}
        onClick={() =>
          void run(async () => {
            const result = await auth.registerPasskey();
            if (result.recovery_codes.length > 0)
              setCodes(result.recovery_codes);
            else await onComplete();
          })
        }
      >
        <Fingerprint className="size-4" aria-hidden />
        {busy ? 'Cadastrando passkey…' : 'Cadastrar passkey'}
      </Button>
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
  const [recoveryCode, setRecoveryCode] = useState('');
  const [useRecovery, setUseRecovery] = useState(false);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    setError('');
    try {
      await action();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Não foi possível continuar.');
    } finally {
      setBusy(false);
    }
  };
  const submitPassword = (e: FormEvent) => {
    e.preventDefault();
    void run(async () => {
      await auth.login(username, password);
      setPassword('');
    });
  };
  const submitRecovery = (e: FormEvent) => {
    e.preventDefault();
    void run(async () => {
      await auth.verifyRecoveryCode(recoveryCode);
      setRecoveryCode('');
    });
  };
  return (
    <main className="bg-muted/30 flex min-h-screen w-full items-center justify-center p-4">
      <section className="bg-background w-full max-w-md space-y-6 rounded-2xl border p-6 shadow-sm">
        <div>
          <ShieldCheck className="text-primary mb-3 size-8" aria-hidden />
          <h1 className="text-2xl font-semibold">
            {auth.isAuthenticated
              ? auth.mfaRequired
                ? auth.mfaEnrolled
                  ? 'Confirme seu acesso'
                  : 'Cadastre sua passkey'
                : 'Sessão ativa'
              : 'Entrar no Telefonia'}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm">
            {auth.isAuthenticated
              ? auth.mfaRequired
                ? auth.mfaEnrolled
                  ? 'Use sua passkey para continuar.'
                  : 'Proteja seu acesso com uma passkey. Use a digital, o Face ID ou o Windows Hello.'
                : 'Sua sessão já está autenticada.'
              : 'Use seu usuário e senha. Tudo é feito aqui no Telefonia.'}
          </p>
        </div>
        {auth.sessionExpired && !auth.isAuthenticated && (
          <p role="status" className="bg-muted rounded-lg p-3 text-sm">
            Sua sessão expirou. Entre novamente para continuar.
          </p>
        )}
        {auth.isAuthenticated && auth.mfaRequired ? (
          auth.mfaEnrolled ? (
            <div className="space-y-4">
              <Button
                className="w-full"
                autoFocus
                disabled={busy}
                onClick={() =>
                  void run(async () => {
                    await auth.loginWithPasskey();
                  })
                }
              >
                <Fingerprint className="size-4" aria-hidden />
                {busy ? 'Verificando…' : 'Entrar com passkey'}
              </Button>
              {useRecovery ? (
                <form onSubmit={submitRecovery} className="space-y-3">
                  <Label htmlFor="recovery-code">Código de recuperação</Label>
                  <Input
                    id="recovery-code"
                    autoComplete="one-time-code"
                    required
                    autoFocus
                    maxLength={64}
                    aria-describedby="recovery-code-help"
                    value={recoveryCode}
                    onChange={(e) => setRecoveryCode(e.target.value)}
                  />
                  <p
                    id="recovery-code-help"
                    className="text-muted-foreground text-xs"
                  >
                    Cada código de recuperação funciona uma única vez.
                  </p>
                  <Button type="submit" className="w-full" disabled={busy}>
                    <KeyRound className="size-4" aria-hidden />
                    {busy ? 'Verificando…' : 'Usar código de recuperação'}
                  </Button>
                  <Button
                    type="button"
                    variant="link"
                    className="h-auto px-0 text-sm"
                    disabled={busy}
                    onClick={() => {
                      setUseRecovery(false);
                      setRecoveryCode('');
                      setError('');
                    }}
                  >
                    Voltar para a passkey
                  </Button>
                </form>
              ) : (
                <Button
                  type="button"
                  variant="link"
                  className="h-auto px-0 text-sm"
                  onClick={() => {
                    setUseRecovery(true);
                    setError('');
                  }}
                >
                  Não consigo usar a passkey
                </Button>
              )}
            </div>
          ) : (
            <PasskeySetup onComplete={auth.refresh} />
          )
        ) : auth.isAuthenticated ? (
          <p className="text-sm">Sua sessão já está autenticada.</p>
        ) : (
          <form onSubmit={submitPassword} className="space-y-4">
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
                aria-label={showPassword ? 'Ocultar senha' : 'Mostrar senha'}
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
            <Button type="submit" className="w-full" disabled={busy}>
              {busy ? 'Verificando…' : 'Continuar'}
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
