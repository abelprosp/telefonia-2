import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode
} from 'react';

import { env } from '@/env';
import { authNetworkMessage, isExpiredSession } from '@/lib/auth-response';
import { queryClient } from '@/lib/query-client';

export type SessionUser = {
  profile: {
    sub: string;
    name: string;
    email: string;
    preferred_username: string;
    picture?: string;
  };
  roles: string[];
  organization_id: string;
};
type Session = {
  user: SessionUser;
  mfa_required: boolean;
  mfa_enrolled: boolean;
};

async function loadSession(): Promise<Session | undefined> {
  const response = await fetch(
    `${env.VITE_API_URL.replace(/\/+$/, '')}/v1/auth/session`,
    {
      credentials: 'include',
      cache: 'no-store',
      signal: AbortSignal.timeout(15000)
    }
  ).catch((error: unknown) => {
    throw new Error(authNetworkMessage(error));
  });
  if (response.status === 401) return undefined;
  if (!response.ok)
    throw new Error('Não foi possível verificar a sessão. Tente novamente.');
  return response.json() as Promise<Session>;
}

export async function authRequest<T = Record<string, never>>(
  path: string,
  data?: unknown
): Promise<T> {
  const response = await fetch(
    `${env.VITE_API_URL.replace(/\/+$/, '')}/v1/auth/${path}`,
    {
      method: data === undefined ? 'GET' : 'POST',
      credentials: 'include',
      signal: AbortSignal.timeout(15000),
      cache: 'no-store',
      headers: { 'Content-Type': 'application/json' },
      body: data === undefined ? undefined : JSON.stringify(data)
    }
  ).catch((error: unknown) => {
    throw new Error(authNetworkMessage(error));
  });
  const result = await response.json().catch(() => undefined);
  if (isExpiredSession(response.status, result?.messages?.[0]?.code)) {
    window.dispatchEvent(new Event('telefonia:session-expired'));
    throw new Error('Sua sessão expirou. Entre novamente para continuar.');
  }
  if (!response.ok)
    throw new Error(
      result?.messages?.[0]?.message ||
        'Não foi possível acessar sua conta. Tente novamente.'
    );
  return result as T;
}
type AuthState = {
  user?: SessionUser;
  isAuthenticated: boolean;
  isLoading: boolean;
  error?: Error;
  mfaRequired: boolean;
  mfaEnrolled: boolean;
  sessionExpired: boolean;
  login: (username: string, password: string) => Promise<void>;
  refresh: () => Promise<void>;
  removeUser: () => Promise<void>;
  expireSession: () => void;
  logout: () => Promise<void>;
};
const Context = createContext<AuthState | undefined>(undefined);
export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session>();
  const [isLoading, setLoading] = useState(true);
  const [error, setError] = useState<Error>();
  const [sessionExpired, setSessionExpired] = useState(false);
  const refresh = useCallback(async () => {
    setSession(await loadSession());
    setError(undefined);
  }, []);
  useEffect(() => {
    let active = true;
    void loadSession()
      .then((next) => {
        if (active) setSession(next);
      })
      .catch((e: Error) => {
        if (active) setError(e);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);
  const login = async (username: string, password: string) => {
    queryClient.clear();
    setSession(await authRequest<Session>('login', { username, password }));
    setError(undefined);
    setSessionExpired(false);
  };
  const removeUser = useCallback(async () => {
    queryClient.clear();
    setSession(undefined);
    setError(undefined);
  }, []);
  const expireSession = useCallback(() => {
    void removeUser();
    setSessionExpired(true);
  }, [removeUser]);
  useEffect(() => {
    window.addEventListener('telefonia:session-expired', expireSession);
    return () =>
      window.removeEventListener('telefonia:session-expired', expireSession);
  }, [expireSession]);
  const logout = async () => {
    await authRequest('logout', {});
    await removeUser();
    window.location.assign('/');
  };
  return (
    <Context.Provider
      value={{
        user: session?.user,
        isAuthenticated: !!session?.user,
        isLoading,
        error,
        mfaRequired: session?.mfa_required ?? false,
        mfaEnrolled: session?.mfa_enrolled ?? false,
        sessionExpired,
        login,
        refresh,
        removeUser,
        expireSession,
        logout
      }}
    >
      {children}
    </Context.Provider>
  );
}
export function useAuth() {
  const auth = useContext(Context);
  if (!auth) throw new Error('AuthProvider ausente');
  return auth;
}
