import { useQuery } from '@tanstack/react-query';

import { client } from '@/lib/client';

export type MovementReportItem = {
  phone_line_id: string;
  number: string;
  status: string;
  line_classification: string;
  customer_id?: string | null;
  customer_name?: string | null;
  last_invoice_id?: string | null;
  last_invoice_number?: string | null;
  pending_cycles?: number;
  transition_sub_status?: string | null;
  transition_started_at?: string | null;
};

export type MovementReportsResponse = {
  processing_month_id: string;
  processing_month_name: string;
  entries: MovementReportItem[];
  exits: MovementReportItem[];
  activation_pending: MovementReportItem[];
};

export function useMovementReports(processingMonthId: string) {
  return useQuery({
    queryKey: ['reports', 'line-movements', processingMonthId],
    queryFn: async () => {
      const { data } = await client<MovementReportsResponse>({
        url: '/v1/reports/line-movements',
        method: 'GET',
        params: { processing_month_id: processingMonthId }
      });
      return data;
    },
    enabled: Boolean(processingMonthId)
  });
}
