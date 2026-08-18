import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import client from '@/lib/client';

export interface InAppNotification {
  id: string;
  title: string;
  description: string;
  category: 'billing' | 'operations' | 'security' | 'system';
  severity: 'info' | 'warning' | 'success' | 'critical';
  action_url?: string;
  action_label?: string;
  is_read: boolean;
  created_at: string;
}

export interface NotificationListResponse {
  items: InAppNotification[];
  unread_count: number;
  total_count: number;
}

export const useGetNotifications = () => {
  return useQuery<NotificationListResponse>({
    queryKey: ['inapp-notifications'],
    queryFn: async () => {
      try {
        const { data } = await client<NotificationListResponse>({
          url: '/v1/inapp-notifications',
          method: 'GET'
        });
        return data;
      } catch {
        return { items: [], unread_count: 0, total_count: 0 };
      }
    },
    refetchInterval: 30000 // Atualiza a cada 30 segundos
  });
};

export const useMarkNotificationRead = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await client<{ success: boolean }>({
        url: `/v1/inapp-notifications/${id}/read`,
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['inapp-notifications'] });
    }
  });
};

export const useMarkAllNotificationsRead = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client<{ success: boolean }>({
        url: '/v1/inapp-notifications/read-all',
        method: 'POST'
      });
      return data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['inapp-notifications'] });
    }
  });
};
