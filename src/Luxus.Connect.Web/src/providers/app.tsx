import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
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
import { type Theme } from '@/types/app';

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

  // Regista o token antes dos filhos dispararem queries (useEffect corre depois dos filhos).
  registerAuthTokenGetter(() => user?.access_token ?? null);

  const [theme, setTheme] = useState<Theme>(
    () => getItem(`${storageKey}.theme`, defaultTheme)!
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
      signinRedirect({ redirect_uri: window.location.href });
    });
    return unregister;
  }, [signinRedirect]);

  useEffect(() => {
    return registerAuthTokenGetter(() => user?.access_token ?? null);
  }, [user?.access_token]);

  const value = {
    theme,
    setTheme: (theme_: Theme) => {
      setTheme(theme_);
    }
  };

  return (
    <AppProviderContext.Provider {...props} value={value}>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          {children}
          <Toaster />
        </TooltipProvider>
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
