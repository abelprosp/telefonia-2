import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useLayoutEffect,
  useState
} from 'react';

import { QueryClientProvider } from '@tanstack/react-query';
import { useAuth } from 'react-oidc-context';

import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import {
  registerAuthTokenGetter,
  registerUnauthorizedHandler
} from '@/lib/client';
import { queryClient } from '@/lib/query-client';
import { getItem, setItem } from '@/lib/storage';
import { WhitelabelProvider } from '@/providers/whitelabel-provider';
import { ThemeSchema, type Theme } from '@/types/app';

const readStoredTheme = (key: string, fallback: Theme): Theme => {
  const parsed = ThemeSchema.safeParse(getItem<string>(key));
  return parsed.success ? parsed.data : fallback;
};

type AppProviderProps = {
  children: ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
};

type AppProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: AppProviderState = {
  theme: 'system',
  setTheme: () => null
};

const AppProviderContext = createContext<AppProviderState>(initialState);

export const AppProvider = ({
  children,
  defaultTheme = 'system',
  storageKey = 'luxus-connect',
  ...props
}: AppProviderProps) => {
  const { signinRedirect, user } = useAuth();

  // Regista o token antes de os filhos dispararem queries. Um layout effect corre
  // antes dos useEffect dos filhos, evitando a primeira chamada sem Authorization.
  useLayoutEffect(() => {
    return registerAuthTokenGetter(() => user?.access_token ?? null);
  }, [user?.access_token]);

  const [theme, setTheme] = useState<Theme>(() =>
    readStoredTheme(`${storageKey}.theme`, defaultTheme)
  );

  useEffect(() => {
    setItem(`${storageKey}.theme`, theme);

    const root = window.document.documentElement;

    root.classList.remove('light', 'dark');

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
        .matches
        ? 'dark'
        : 'light';

      root.classList.add(systemTheme);
      return;
    }

    root.classList.add(theme);
  }, [theme, storageKey]);

  useEffect(() => {
    const unregister = registerUnauthorizedHandler(() => {
      const now = Date.now();
      const lastRedirect = Number(
        sessionStorage.getItem('luxus_last_auth_redirect') || '0'
      );
      if (now - lastRedirect < 10000) {
        // Prevent redirect/refresh storm if 401 persists
        return;
      }
      sessionStorage.setItem('luxus_last_auth_redirect', String(now));
      void signinRedirect();
    });
    return unregister;
  }, [signinRedirect]);

  const value = {
    theme,
    setTheme: (theme_: Theme) => {
      setTheme(theme_);
    }
  };

  return (
    <AppProviderContext.Provider {...props} value={value}>
      <QueryClientProvider client={queryClient}>
        <WhitelabelProvider>
          <TooltipProvider>
            {children}
            <Toaster />
          </TooltipProvider>
        </WhitelabelProvider>
      </QueryClientProvider>
    </AppProviderContext.Provider>
  );
};



export const useApp = () => {
  const context = useContext(AppProviderContext);

  if (context === undefined)
    throw new Error('useApp must be used within a AppProvider');

  return context;
};
