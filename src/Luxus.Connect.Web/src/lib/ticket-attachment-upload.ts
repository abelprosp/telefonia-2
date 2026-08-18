import { client } from '@/lib/client';
import { env } from '@/env';

const PRESIGNED_UPLOAD_EXPIRES_SECONDS = 60 * 5;

export function defaultTicketStorageBucket(): string {
  return (env.VITE_STORAGE_BUCKET_NAME ?? '').trim();
}

export async function requestTicketAttachmentUploadUrl(opts: {
  ticketId: string;
  file: File;
  portal?: boolean;
}): Promise<{ url: string; http_method: string; object_key: string; bucket_name: string }> {
  const bucket = defaultTicketStorageBucket();
  if (!bucket) {
    throw new Error('Defina VITE_STORAGE_BUCKET_NAME para enviar anexos.');
  }
  const base = opts.portal ? '/v1/portal/tickets' : '/v1/tickets';
  const { data } = await client<{
    url: string;
    http_method: string;
    object_key: string;
    bucket_name: string;
  }>({
    url: `${base}/${opts.ticketId}/attachments/upload-url`,
    method: 'POST',
    data: {
      bucket_name: bucket,
      file_name: opts.file.name,
      content_type: opts.file.type || 'application/octet-stream',
      expires_in_seconds: PRESIGNED_UPLOAD_EXPIRES_SECONDS,
      size_bytes: opts.file.size
    }
  });
  return data;
}

export async function putTicketAttachmentFile(
  file: File,
  upload: { url: string; http_method: string }
): Promise<void> {
  const contentType = file.type || 'application/octet-stream';
  const res = await fetch(upload.url, {
    method: upload.http_method,
    body: file,
    headers: { 'Content-Type': contentType }
  });
  if (!res.ok) {
    throw new Error(`Falha ao enviar o anexo (${res.status}).`);
  }
}

export async function downloadTicketAttachment(opts: {
  ticketId: string;
  messageId: string;
  portal?: boolean;
}): Promise<void> {
  const base = opts.portal ? '/v1/portal/tickets' : '/v1/tickets';
  const { data } = await client<{ url: string; file_name?: string }>({
    url: `${base}/${opts.ticketId}/messages/${opts.messageId}/attachment`,
    method: 'GET'
  });
  window.open(data.url, '_blank', 'noopener,noreferrer');
}
