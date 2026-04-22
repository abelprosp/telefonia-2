import { postV1PreSignedUrlsDownload, postV1PreSignedUrlsUpload } from '@/api';

const PRESIGNED_UPLOAD_EXPIRES_SECONDS = 60 * 5; // 5 minutos

/**
 * Chave de objeto no bucket (prefixo UUID + nome sanitizado).
 */
export function buildInvoiceStorageObjectKey(
  file: File,
  providerId: string
): string {
  const safe = file.name.replace(/[^a-zA-Z0-9._-]/g, '_');
  return `${providerId}/${safe}`;
}

/**
 * Obtém URL pré-assinada na API, envia o arquivo com `PUT` e usa o mesmo
 * `Content-Type` da assinatura.
 */
export async function uploadFileFromPresignedUrl(
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

export async function createPresignedDownloadUrl(
  bucket: string,
  objectKey: string
): Promise<string> {
  const presigned = await postV1PreSignedUrlsDownload({
    bucket_name: bucket,
    object_key: objectKey,
    expires_in_seconds: PRESIGNED_UPLOAD_EXPIRES_SECONDS
  });

  return presigned.url;
}
