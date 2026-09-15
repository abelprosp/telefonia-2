import { postV1PreSignedUrlsDownload, postV1PreSignedUrlsUpload } from '@/api';

const PRESIGNED_UPLOAD_EXPIRES_SECONDS = 60 * 5; // 5 minutos

/**
 * Chave relativa no bucket (provider + nome).
 * A API prefixa automaticamente com o OrganizationId do token.
 */
export function buildInvoiceStorageObjectKey(
  file: File,
  providerId: string
): string {
  const safe = file.name.replace(/[^a-zA-Z0-9._-]/g, '_');
  const unique = crypto.randomUUID().slice(0, 8);
  return `invoices/${providerId}/${unique}_${safe}`;
}

/**
 * Obtém URL pré-assinada na API, envia o arquivo com `PUT` e usa o mesmo
 * `Content-Type` da assinatura. Retorna a object_key efetiva (com prefixo da org).
 */
export async function uploadFileFromPresignedUrl(
  file: File,
  bucket: string,
  objectKey: string,
  onProgress?: (percent: number) => void
): Promise<string> {
  const contentType = file.type || 'application/octet-stream';
  const presigned = await postV1PreSignedUrlsUpload({
    bucket_name: bucket,
    object_key: objectKey,
    content_type: contentType,
    expires_in_seconds: PRESIGNED_UPLOAD_EXPIRES_SECONDS
  });

  const effectiveKey = (presigned as { object_key?: string }).object_key || objectKey;

  await new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    if (onProgress) {
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          onProgress(Math.round((event.loaded / event.total) * 100));
        }
      });
    }

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress?.(100);
        resolve();
      } else {
        reject(
          new Error(`Falha ao enviar o arquivo para o armazenamento (${xhr.status}).`)
        );
      }
    });

    xhr.addEventListener('error', () => {
      reject(new Error('Falha de rede ao enviar o arquivo.'));
    });

    xhr.addEventListener('abort', () => {
      reject(new Error('Upload cancelado.'));
    });

    xhr.open(presigned.http_method, presigned.url);
    xhr.setRequestHeader('Content-Type', contentType);
    xhr.send(file);
  });

  return effectiveKey;
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
