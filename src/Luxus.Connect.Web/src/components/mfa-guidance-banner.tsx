import { ShieldAlert } from 'lucide-react';

import { useUserProfileQuery } from '@/api/settings-api';
import { env } from '@/env';

export function MfaGuidanceBanner() {
  const profile = useUserProfileQuery();
  const data = profile.data;
  if (!data || data.mfa_enrolled) {
    return null;
  }

  const href = `${env.VITE_AUTH_URL.replace(/\/+$/, '')}/realms/luxus/account/#/security/signingin`;
  const privileged = data.privileged_access;

  return (
    <div
      role="status"
      className="border-border bg-muted/60 mx-4 mt-3 flex flex-wrap items-start gap-3 rounded-lg border px-4 py-3 text-sm md:mx-6"
    >
      <ShieldAlert className="text-foreground mt-0.5 size-5 shrink-0" aria-hidden />
      <div className="min-w-0 flex-1">
        <p className="font-medium">
          {privileged
            ? 'Conta privilegiada sem MFA no Keycloak'
            : 'Ative a autenticação em dois fatores'}
        </p>
        <p className="text-muted-foreground mt-1">
          O login usa OTP do Keycloak (não há TOTP próprio neste sistema). Configure o autenticador
          na segurança da conta.
        </p>
      </div>
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="text-primary underline underline-offset-4 focus-visible:outline-none"
      >
        Abrir segurança da conta Keycloak
      </a>
    </div>
  );
}
