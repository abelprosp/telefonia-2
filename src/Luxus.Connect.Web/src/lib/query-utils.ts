import type { QueryClient } from '@tanstack/react-query';

import {
  getV1CustomersQueryKey,
  getV1PhoneLinesQueryKey,
  getV1StatsDashboardQueryKey
} from '@/api';

export function parseTotalCount(v: number | string | undefined) {
  if (v === undefined) {
    return 0;
  }
  const n = typeof v === 'string' ? Number(v) : v;
  return Number.isFinite(n) ? n : 0;
}

/** Alinha-se ao antigo `queryKeys.dashboard.summary`: métricas + amostras recentes. */
export function invalidateDashboardCaches(queryClient: QueryClient) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: getV1StatsDashboardQueryKey() }),
    queryClient.invalidateQueries({ queryKey: getV1CustomersQueryKey() }),
    queryClient.invalidateQueries({ queryKey: getV1PhoneLinesQueryKey() })
  ]);
}
