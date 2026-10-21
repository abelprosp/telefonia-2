import { useState, type FormEvent } from 'react';

import { Fingerprint } from 'lucide-react';

import { PasskeySetup } from '@/components/local-login';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { authRequest, passkeyErrorMessage, useAuth } from '@/providers/auth';

export function AccountSecurity() {
  const auth = useAuth();
  const [current, setCurrent] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [passkeyBusy, setPasskeyBusy] = useState(false);
  const [passkeyError, setPasskeyError] = useState('');
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const addPasskey = async () => {
    setPasskeyBusy(true);
    setPasskeyError('');
    try {
      const result = await auth.registerPasskey();
      setRecoveryCodes(result.recovery_codes);
    } catch (e) {
      setPasskeyError(passkeyErrorMessage(e));
    } finally {
      setPasskeyBusy(false);
    }
  };
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    if (password !== confirm) {
      setError('As senhas não coincidem.');
      return;
    }
    setBusy(true);
    try {
      await authRequest('password', {
        current_password: current,
        new_password: password
      });
      await auth.removeUser();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="space-y-6 p-6" id="account-security">
      <h2 className="text-lg font-semibold">Segurança da conta</h2>
      <form onSubmit={(e) => void submit(e)} className="space-y-3">
        <p className="text-muted-foreground text-sm">
          Use pelo menos 12 caracteres. A troca de senha encerra todas as
          sessões.
        </p>
        <Label htmlFor="current-password">Senha atual</Label>
        <Input
          id="current-password"
          type="password"
          autoComplete="current-password"
          required
          maxLength={256}
          value={current}
          onChange={(e) => setCurrent(e.target.value)}
        />
        <Label htmlFor="new-password">Nova senha</Label>
        <Input
          id="new-password"
          type="password"
          autoComplete="new-password"
          required
          minLength={12}
          maxLength={256}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Label htmlFor="confirm-password">Confirmar nova senha</Label>
        <Input
          id="confirm-password"
          type="password"
          autoComplete="new-password"
          required
          minLength={12}
          maxLength={256}
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
        />
        <Button disabled={busy}>Alterar senha</Button>
        {error && (
          <p role="alert" className="text-destructive text-sm">
            {error}
          </p>
        )}
      </form>
      <h3 className="font-semibold">Passkeys</h3>
      {auth.mfaEnrolled ? (
        <div className="space-y-3">
          <p className="text-sm">
            Passkey ativa ({auth.passkeyCount}{' '}
            {auth.passkeyCount === 1 ? 'cadastrada' : 'cadastradas'}).
          </p>
          <p className="text-sm">
            Você entra com digital, Face ID ou Windows Hello.
          </p>
          <Button
            variant="outline"
            disabled={passkeyBusy}
            onClick={() => void addPasskey()}
          >
            <Fingerprint className="size-4" aria-hidden />
            {passkeyBusy ? 'Adicionando…' : 'Adicionar outra passkey'}
          </Button>
          {recoveryCodes.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm">
                Guarde estes códigos de recuperação; eles só aparecem agora.
              </p>
              <textarea
                aria-label="Códigos de recuperação"
                readOnly
                value={recoveryCodes.join('\n')}
                className="h-40 w-full rounded border p-3 font-mono text-xs"
              />
            </div>
          )}
          {passkeyError && (
            <p role="alert" className="text-destructive text-sm">
              {passkeyError}
            </p>
          )}
        </div>
      ) : (
        <PasskeySetup onComplete={auth.refresh} />
      )}
    </div>
  );
}
