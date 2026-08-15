import { createContext, useContext, useEffect, type ReactNode } from 'react';
import { useOrganizationSettingsQuery, type WhitelabelSettings } from '@/api/settings-api';

type WhitelabelContextType = {
  settings: WhitelabelSettings;
  isLoading: boolean;
};

const defaultWhitelabel: WhitelabelSettings = {
  app_name: 'Luxus Connect',
  app_slogan: 'Gestão de Telefonia Inteligente',
  logo_url: '',
  dark_logo_url: '',
  favicon_url: '',
  primary_color: '#0f766e',
  support_email: 'suporte@luxusconnect.com.br',
  support_phone: '(11) 99999-9999',
  footer_text: '© 2026 Luxus Connect. Todos os direitos reservados.'
};

const WhitelabelContext = createContext<WhitelabelContextType>({
  settings: defaultWhitelabel,
  isLoading: false
});

export const WhitelabelProvider = ({ children }: { children: ReactNode }) => {
  const { data, isLoading } = useOrganizationSettingsQuery();

  const settings = data?.whitelabel ?? defaultWhitelabel;

  useEffect(() => {
    // Atualiza o título do documento dinamicamente
    if (settings.app_name) {
      document.title = `${settings.app_name} — Gestão de Telefonia`;
    }

    // Atualiza o favicon se fornecido
    if (settings.favicon_url) {
      let link: HTMLLinkElement | null = document.querySelector("link[rel~='icon']");
      if (!link) {
        link = document.createElement('link');
        link.rel = 'icon';
        document.head.appendChild(link);
      }
      link.href = settings.favicon_url;
    }
  }, [settings.app_name, settings.favicon_url]);

  return (
    <WhitelabelContext.Provider value={{ settings, isLoading }}>
      {children}
    </WhitelabelContext.Provider>
  );
};

export const useWhitelabel = () => {
  return useContext(WhitelabelContext);
};
