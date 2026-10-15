import { useState, type FormEvent } from 'react';

import { MfaSetup } from '@/components/local-login';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { authRequest, useAuth } from '@/providers/auth';

export function AccountSecurity() {
  const auth = useAuth();
  const [current, setCurrent] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
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
      <h3 className="font-semibold">Autenticação em dois fatores</h3>
      {auth.mfaEnrolled ? (
        <p className="text-sm">
          Autenticador ativo. Para recuperar o acesso, use um dos códigos
          guardados ou contate o administrador da VPS.
        </p>
      ) : (
        <MfaSetup onComplete={auth.refresh} />
      )}
    </div>
  );
}
