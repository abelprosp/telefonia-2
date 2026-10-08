import { createContext, useContext, useEffect, type ReactNode } from 'react';

import {
  useOrganizationBrandingQuery,
  type WhitelabelSettings
} from '@/api/settings-api';

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
  const { data, isLoading } = useOrganizationBrandingQuery();

  const settings = data?.whitelabel ?? defaultWhitelabel;

  useEffect(() => {
    const color = settings.primary_color;
    if (!/^#[0-9a-f]{6}$/i.test(color)) return;
    const rgb = [1, 3, 5].map(
      (offset) => parseInt(color.slice(offset, offset + 2), 16) / 255
    );
    const linear = rgb.map((v) =>
      v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4
    );
    const luminance =
      0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2];
    const foreground = luminance > 0.179 ? '#000000' : '#ffffff';
    const root = document.documentElement;
    const tokens = {
      '--primary': color,
      '--primary-foreground': foreground,
      '--sidebar-primary': color,
      '--sidebar-primary-foreground': foreground
    };
    for (const [key, value] of Object.entries(tokens))
      root.style.setProperty(key, value);
    return () => {
      for (const key of Object.keys(tokens)) root.style.removeProperty(key);
    };
  }, [settings.primary_color]);

  useEffect(() => {
    // Atualiza o título do documento dinamicamente
    if (settings.app_name) {
      document.title = `${settings.app_name} — Gestão de Telefonia`;
    }

    // Atualiza o favicon se fornecido
    if (settings.favicon_url) {
      let link: HTMLLinkElement | null =
        document.querySelector("link[rel~='icon']");
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
