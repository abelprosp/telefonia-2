import { useState } from 'react';

import axios, { type AxiosProgressEvent } from 'axios';
import {
  CircleAlertIcon,
  CloudUploadIcon,
  DownloadIcon,
  FileArchiveIcon,
  FileSpreadsheetIcon,
  FileTextIcon,
  HeadphonesIcon,
  ImageIcon,
  RefreshCwIcon,
  Trash2Icon,
  UploadIcon,
  VideoIcon
} from 'lucide-react';
import { toast } from 'sonner';

import {
  getV1CustomersIdAttachmentsQueryKey,
  useDeleteV1CustomersIdAttachmentsAttachmentid,
  usePostV1CustomersIdAttachments,
  usePostV1PreSignedUrlsDownload,
  usePostV1PreSignedUrlsUpload,
  type CustomerAttachmentResponse
} from '@/api';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { env } from '@/env';
import { formatBytes, useFileUpload } from '@/hooks/use-file-upload';
import { queryClient } from '@/lib/query-client';
import { cn } from '@/lib/utils';

const DEFAULT_STORAGE_BUCKET_NAME = (env.VITE_STORAGE_BUCKET_NAME ?? '').trim();

function buildCustomerAttachmentStorageObjectKey(
  customerId: string,
  fileName: string
): string {
  const safe = fileName.replace(/[^a-zA-Z0-9._-]/g, '_');
  return `customers/${customerId}/${crypto.randomUUID()}_${safe}`;
}

// Represents a file currently being uploaded (not yet persisted in the API)
interface PendingUploadItem {
  id: string;
  file: File;
  progress: number;
  status: 'uploading' | 'error';
  error?: string;
}

interface CustomerAttachmentsViewProps {
  customerId: string;
  maxFiles?: number;
  maxSize?: number;
  accept?: string;
  multiple?: boolean;
  className?: string;
  attachments: CustomerAttachmentResponse[];
}

export function CustomerAttachmentsView({
  customerId,
  maxFiles = 10,
  maxSize = 256 * 1024 * 1024, // 256MB
  accept = '*',
  multiple = true,
  className,
  attachments = []
}: CustomerAttachmentsViewProps) {
  // pendingUploads: files in-flight. Persisted attachments come from the API.
  const [pendingUploads, setPendingUploads] = useState<PendingUploadItem[]>([]);

  const postAttachmentMutation = usePostV1CustomersIdAttachments({
    mutation: {
      onError: (e) => {
        toast.error(String(e));
      }
    }
  });

  const deleteAttachmentMutation =
    useDeleteV1CustomersIdAttachmentsAttachmentid({
      mutation: {
        onSuccess: async () => {
          await queryClient.invalidateQueries({
            queryKey: getV1CustomersIdAttachmentsQueryKey(customerId)
          });
        },
        onError: (e) => {
          toast.error(String(e));
        }
      }
    });

  const presignedUploadMutation = usePostV1PreSignedUrlsUpload();
  const presignedDownloadMutation = usePostV1PreSignedUrlsDownload();

  const [
    { isDragging, errors },
    {
      removeFile,
      handleDragEnter,
      handleDragLeave,
      handleDragOver,
      handleDrop,
      openFileDialog,
      getInputProps
    }
  ] = useFileUpload({
    maxFiles,
    maxSize,
    accept,
    multiple,
    onFilesAdded: async (newFiles) => {
      newFiles.forEach(async (fileWithPreview) => {
        const file = fileWithPreview.file as File;
        const tempId = fileWithPreview.id;

        if (
          attachments.some(
            (attachment) =>
              attachment.original_file_name === file.name &&
              attachment.size_bytes === file.size
          )
        ) {
          return;
        }

        // Immediately add to pending so the user sees the progress row
        setPendingUploads((prev) => [
          ...prev,
          { id: tempId, file, progress: 0, status: 'uploading' }
        ]);

        const objectKey = buildCustomerAttachmentStorageObjectKey(
          customerId,
          file.name
        );

        try {
          const presigned = await presignedUploadMutation.mutateAsync({
            data: {
              bucket_name: DEFAULT_STORAGE_BUCKET_NAME,
              object_key: objectKey,
              expires_in_seconds: 60
            }
          });

          await axios.request({
            url: presigned.url,
            method: presigned.http_method,
            data: file,
            headers: {
              'Content-Type': file.type || 'application/octet-stream'
            },
            onUploadProgress: (progressEvent: AxiosProgressEvent) => {
              if (progressEvent.total) {
                const pct = Math.min(
                  Math.round(
                    (progressEvent.loaded * 100) / progressEvent.total
                  ),
                  99 // stay at 99% until the API call succeeds
                );
                setPendingUploads((prev) =>
                  prev.map((item) =>
                    item.id === tempId ? { ...item, progress: pct } : item
                  )
                );
              }
            }
          });

          await postAttachmentMutation.mutateAsync({
            id: customerId,
            data: {
              title: file.name.trim() || null,
              original_file_name: file.name,
              storage_bucket: DEFAULT_STORAGE_BUCKET_NAME,
              storage_object_key: objectKey,
              content_type: file.type || null,
              size_bytes: file.size
            }
          });

          // Remove from pending — the invalidation below will show it in the API list
          setPendingUploads((prev) =>
            prev.filter((item) => item.id !== tempId)
          );
          removeFile(tempId);
        } catch (err) {
          setPendingUploads((prev) =>
            prev.map((item) =>
              item.id === tempId
                ? {
                    ...item,
                    status: 'error' as const,
                    error: String(err)
                  }
                : item
            )
          );
          toast.error(String(err));
        } finally {
          await queryClient.invalidateQueries({
            queryKey: getV1CustomersIdAttachmentsQueryKey(customerId)
          });
        }
      });
    }
  });

  const handleDownloadAttachment = async (row: CustomerAttachmentResponse) => {
    try {
      const presigned = await presignedDownloadMutation.mutateAsync({
        data: {
          bucket_name: row.storage_bucket,
          object_key: row.storage_object_key,
          expires_in_seconds: 60
        }
      });
      window.open(presigned.url, '_blank', 'noopener,noreferrer');
    } catch (err) {
      toast.error(String(err));
    }
  };

  const handleDeleteAllAttachments = () => {
    attachments?.forEach((attachment) => {
      handleDeleteAttachment(attachment.id!);
    });
  };

  const handleDeleteAttachment = (attachmentId: string) => {
    deleteAttachmentMutation.mutate({
      id: customerId,
      attachmentId: attachmentId
    });
  };

  const handleDismissFailedUpload = (tempId: string) => {
    setPendingUploads((prev) => prev.filter((item) => item.id !== tempId));
  };

  const getFileIcon = (contentType: string) => {
    if (contentType.startsWith('image/'))
      return <ImageIcon className="size-4" />;
    if (contentType.startsWith('video/'))
      return <VideoIcon className="size-4" />;
    if (contentType.startsWith('audio/'))
      return <HeadphonesIcon className="size-4" />;
    if (contentType.includes('pdf')) return <FileTextIcon className="size-4" />;
    if (contentType.includes('word') || contentType.includes('doc'))
      return <FileTextIcon className="size-4" />;
    if (contentType.includes('excel') || contentType.includes('sheet'))
      return <FileSpreadsheetIcon className="size-4" />;
    if (contentType.includes('zip') || contentType.includes('rar'))
      return <FileArchiveIcon className="size-4" />;
    return <FileTextIcon className="size-4" />;
  };

  const getFileTypeLabel = (contentType: string) => {
    if (contentType.startsWith('image/')) return 'Image';
    if (contentType.startsWith('video/')) return 'Video';
    if (contentType.startsWith('audio/')) return 'Audio';
    if (contentType.includes('pdf')) return 'PDF';
    if (contentType.includes('word') || contentType.includes('doc'))
      return 'Word';
    if (contentType.includes('excel') || contentType.includes('sheet'))
      return 'Excel';
    if (contentType.includes('zip') || contentType.includes('rar'))
      return 'Archive';
    if (contentType.includes('json')) return 'JSON';
    if (contentType.includes('text')) return 'Text';
    return 'File';
  };

  const totalCount = (attachments?.length ?? 0) + pendingUploads.length;

  return (
    <div className={cn('w-full space-y-4', className)}>
      {/* Upload Area */}
      <div
        className={cn(
          'relative rounded-lg border border-dashed p-6 text-center transition-colors',
          isDragging
            ? 'border-primary bg-primary/5'
            : 'border-muted-foreground/25'
        )}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <input {...getInputProps()} className="sr-only" />
        <div className="flex flex-col items-center gap-4">
          <div
            className={cn(
              'bg-muted flex h-12 w-12 items-center justify-center rounded-full transition-colors',
              isDragging
                ? 'border-primary bg-primary/10'
                : 'border-muted-foreground/25'
            )}
          >
            <UploadIcon className="text-muted-foreground h-5 w-5" />
          </div>
          <div className="space-y-2">
            <p className="text-sm font-medium">
              Solte arquivos aqui ou{' '}
              <button
                type="button"
                onClick={openFileDialog}
                className="text-primary cursor-pointer underline-offset-4 hover:underline"
              >
                selecione arquivos
              </button>
            </p>
            <p className="text-muted-foreground text-xs">
              Tamanho máximo: {formatBytes(maxSize)} • Quantidade máxima:{' '}
              {maxFiles}
            </p>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Arquivos ({totalCount})</h3>
        <div className="flex gap-2">
          <Button onClick={openFileDialog} variant="outline" size="sm">
            <CloudUploadIcon className="h-4 w-4" />
            Adicionar arquivos
          </Button>
          {totalCount > 0 && (
            <Button
              onClick={handleDeleteAllAttachments}
              variant="outline"
              size="sm"
            >
              <Trash2Icon className="h-4 w-4" />
              Remover todos
            </Button>
          )}
        </div>
      </div>

      {/* Files Table */}
      {totalCount > 0 && (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow className="text-xs">
                <TableHead className="h-9 ps-4">Nome</TableHead>
                <TableHead className="h-9">Tipo</TableHead>
                <TableHead className="h-9">Tamanho</TableHead>
                <TableHead className="h-9 w-[100px] ps-4">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {/* Persisted attachments — source of truth from the API */}
              {attachments?.map((attachment) => (
                <TableRow key={attachment.id}>
                  <TableCell className="py-2 ps-1.5">
                    <div className="flex items-center gap-1">
                      <div className="text-muted-foreground/80 flex size-8 shrink-0 items-center justify-center">
                        {getFileIcon(attachment.content_type ?? '')}
                      </div>
                      <p className="truncate text-sm font-medium">
                        {attachment.original_file_name}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell className="py-2">
                    <Badge variant="secondary" className="text-xs">
                      {getFileTypeLabel(attachment.content_type ?? '')}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground py-2 text-sm">
                    {formatBytes(attachment.size_bytes ?? 0)}
                  </TableCell>
                  <TableCell className="py-2">
                    <div className="flex items-center gap-1">
                      <Button
                        size="icon"
                        variant="ghost"
                        className="size-8"
                        onClick={() => handleDownloadAttachment(attachment)}
                        disabled={presignedDownloadMutation.isPending}
                      >
                        <DownloadIcon className="size-3.5" />
                      </Button>
                      <Button
                        onClick={() => handleDeleteAttachment(attachment.id!)}
                        variant="ghost"
                        size="icon"
                        className="size-8"
                        disabled={deleteAttachmentMutation.isPending}
                      >
                        <Trash2Icon className="size-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}

              {/* In-progress uploads */}
              {pendingUploads.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="py-2 ps-1.5">
                    <div className="flex items-center gap-1">
                      <div className="text-muted-foreground/80 relative flex size-8 shrink-0 items-center justify-center">
                        {item.status === 'uploading' ? (
                          <div className="relative">
                            <svg
                              className="size-8 -rotate-90"
                              viewBox="0 0 32 32"
                            >
                              <circle
                                cx="16"
                                cy="16"
                                r="14"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                className="text-muted-foreground/20"
                              />
                              <circle
                                cx="16"
                                cy="16"
                                r="14"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeDasharray={`${2 * Math.PI * 14}`}
                                strokeDashoffset={`${2 * Math.PI * 14 * (1 - item.progress / 100)}`}
                                className="text-primary transition-all duration-300"
                                strokeLinecap="round"
                              />
                            </svg>
                            <div className="absolute inset-0 flex items-center justify-center">
                              {getFileIcon(item.file.type)}
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-center justify-center">
                            {getFileIcon(item.file.type)}
                          </div>
                        )}
                      </div>
                      <p className="flex items-center gap-1 truncate text-sm font-medium">
                        {item.file.name}
                        {item.status === 'error' && (
                          <Badge variant="destructive-light" size="sm">
                            Erro
                          </Badge>
                        )}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell className="py-2">
                    <Badge variant="secondary" className="text-xs">
                      {getFileTypeLabel(item.file.type)}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground py-2 text-sm">
                    {formatBytes(item.file.size)}
                  </TableCell>
                  <TableCell className="py-2">
                    {item.status === 'error' && (
                      <Button
                        onClick={() => handleDismissFailedUpload(item.id)}
                        variant="ghost"
                        size="icon"
                        className="text-destructive/80 hover:text-destructive size-8"
                        title="Descartar"
                      >
                        <RefreshCwIcon className="size-3.5" />
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Validation errors */}
      {errors.length > 0 && (
        <Alert variant="destructive" className="mt-5">
          <CircleAlertIcon />
          <AlertTitle>Erro no upload</AlertTitle>
          <AlertDescription>
            {errors.map((error, index) => (
              <p key={index} className="last:mb-0">
                {error}
              </p>
            ))}
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
