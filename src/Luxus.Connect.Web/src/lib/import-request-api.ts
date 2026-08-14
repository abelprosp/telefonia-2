/**
 * GET /v1/provider-invoices/import-requests/{id}
 * Returns the status of a provider invoice import request.
 */
import { client } from '@/lib/client';

export type ImportRequestStatus =
  | 'pending'
  | 'processing'
  | 'completed'
  | 'failed'
  | 'pdf_unparsed';

export type ImportRequestStatusResponse = {
  id: string;
  processing_month_id: string;
  status: ImportRequestStatus;
  error: string | null;
  completed_at: string | null;
};

export async function getImportRequestStatus(id: string): Promise<ImportRequestStatusResponse> {
  const { data } = await client<ImportRequestStatusResponse>({
    url: `/v1/provider-invoices/import-requests/${id}`,
    method: 'GET'
  });
  return data;
}
