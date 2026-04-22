import { postV1PreSignedUrlsUpload } from '@/api';

const PRESIGNED_UPLOAD_EXPIRES_SECONDS = 60 * 5;

/**
 * Chave de objeto no bucket para anexos de cliente (prefixo por cliente + UUID).
 */
export function buildCustomerAttachmentStorageObjectKey(
  customerId: string,
  file: File
): string {
  const safe = file.name.replace(/[^a-zA-Z0-9._-]/g, '_');
  return `customers/${customerId}/${crypto.randomUUID()}_${safe}`;
}

/**
 * Obtém URL pré-assinada na API, envia o arquivo com o método indicado e o mesmo
 * Content-Type usado na assinatura.
 */
export async function uploadCustomerAttachmentFileViaPresignedUrl(
  file: File,
  bucket: string,
  objectKey: string
): Promise<void> {
  const contentType = file.type || 'application/octet-stream';
  const presigned = await postV1PreSignedUrlsUpload({
    bucket_name: bucket,
    object_key: objectKey,
    content_type: contentType,
    expires_in_seconds: PRESIGNED_UPLOAD_EXPIRES_SECONDS
  });

  const res = await fetch(presigned.url, {
    method: presigned.http_method,
    body: file,
    headers: {
      'Content-Type': contentType
    }
  });

  if (!res.ok) {
    throw new Error(
      `Falha ao enviar o arquivo para o armazenamento (${res.status}).`
    );
  }
}
