import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';

export type FinancialAgentSettings = {
  enabled: boolean;
  evolution_api_url: string;
  evolution_api_key_set: boolean;
  evolution_instance: string;
  n8n_webhook_url: string;
  updated_at?: string;
};

export type FinancialAgentSettingsInput = {
  enabled?: boolean;
  evolution_api_url?: string;
  evolution_api_key?: string;
  evolution_instance?: string;
  n8n_webhook_url?: string;
};

export type FinancialAgentWhatsAppStatus = {
  configured: boolean;
  state: string;
  connected: boolean;
  qr_code?: string;
  pairing_code?: string;
  instance_name?: string;
  profile_name?: string;
  owner_jid?: string;
  message?: string;
};

export type FinancialAgentEvent = {
  id: string;
  event_type: string;
  event_label: string;
  whatsapp_number?: string;
  customer_name?: string;
  invoice_number?: string;
  success: boolean;
  summary: string;
  created_at: string;
};

export type FinancialAgentPanelStats = {
  total_today: number;
  success_today: number;
  failed_today: number;
  lookups_today: number;
  receipts_today: number;
  payments_today: number;
};

export type FinancialAgentPanel = {
  tools_ready: boolean;
  organization_set: boolean;
  sicredi_enabled: boolean;
  sicredi_connected: boolean;
  settings: FinancialAgentSettings;
  whatsapp: FinancialAgentWhatsAppStatus;
  stats: FinancialAgentPanelStats;
  message: string;
};

export const financialAgentKeys = {
  panel: ['financial-agent', 'panel'] as const,
  events: ['financial-agent', 'events'] as const,
  whatsapp: ['financial-agent', 'whatsapp'] as const
};

export function useFinancialAgentPanel(opts?: { refetchInterval?: number | false }) {
  return useQuery({
    queryKey: financialAgentKeys.panel,
    queryFn: async () => {
      const { data } = await client<FinancialAgentPanel>({
        url: '/v1/financial-agent',
        method: 'GET'
      });
      return data;
    },
    refetchInterval: opts?.refetchInterval
  });
}

export function useFinancialAgentEvents(opts?: { refetchInterval?: number | false }) {
  return useQuery({
    queryKey: financialAgentKeys.events,
    queryFn: async () => {
      const { data } = await client<{ items: FinancialAgentEvent[] }>({
        url: '/v1/financial-agent/events',
        method: 'GET'
      });
      return data.items ?? [];
    },
    refetchInterval: opts?.refetchInterval
  });
}

export function useUpdateFinancialAgentSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: FinancialAgentSettingsInput) => {
      const { data } = await client<FinancialAgentPanel, unknown, FinancialAgentSettingsInput>({
        url: '/v1/financial-agent/settings',
        method: 'PUT',
        data: input
      });
      return data;
    },
    onSuccess: (panel) => {
      queryClient.setQueryData(financialAgentKeys.panel, panel);
      void queryClient.invalidateQueries({ queryKey: financialAgentKeys.events });
    }
  });
}

export function useConnectFinancialAgentWhatsApp() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<FinancialAgentWhatsAppStatus>({
        url: '/v1/financial-agent/whatsapp/connect',
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: financialAgentKeys.panel });
      void queryClient.invalidateQueries({ queryKey: financialAgentKeys.events });
    }
  });
}

export function useDisconnectFinancialAgentWhatsApp() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<FinancialAgentWhatsAppStatus>({
        url: '/v1/financial-agent/whatsapp/disconnect',
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: financialAgentKeys.panel });
      void queryClient.invalidateQueries({ queryKey: financialAgentKeys.events });
    }
  });
}
