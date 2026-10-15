import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useLayoutEffect,
  useState
} from 'react';

import { QueryClientProvider } from '@tanstack/react-query';


import { Toaster } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import {
  registerAuthRolesGetter,
  registerUnauthorizedHandler
} from '@/lib/client';
import { queryClient } from '@/lib/query-client';
import { getItem, setItem } from '@/lib/storage';
import { useAuth } from '@/providers/auth';
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
  const { removeUser, user } = useAuth();

  // Disponibiliza os papéis às guardas de navegação; a API valida a sessão.
  useLayoutEffect(() => {
    return registerAuthRolesGetter(() => user?.roles);
  }, [user?.roles]);

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
      void removeUser();
    });
    return unregister;
  }, [removeUser]);

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
