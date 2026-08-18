import {
  type ChangeEvent,
  type DragEvent,
  useEffect,
  useRef,
  useState
} from 'react';

import { zodResolver } from '@hookform/resolvers/zod';

import { File, FileText, X } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import type { ListProcessingMonthResponse, ListProvidersResponse } from '@/api';
import {
  useGetV1ProcessingMonths,
  useGetV1Providers,
  usePostV1ProviderInvoices
} from '@/api';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { env } from '@/env';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatMoney } from '@/lib/financial-api';
import {
  buildInvoiceStorageObjectKey,
  uploadFileFromPresignedUrl
} from '@/lib/invoice-import-upload';
import { previewProviderInvoiceImport, type ImportPreview } from '@/lib/ops-api';

import type { ImportProgressState } from './invoice-import-progress-banner';

const MAX_IMPORT_FILE_BYTES = 256 * 1024 * 1024;

const INVOICE_IMPORT_ACCEPT = '.txt,.pdf,application/pdf,text/plain';

const validInvoiceFile = (file: File) => {
  const ext = file.name.split('.').pop()?.toLowerCase() ?? '';
  if (ext !== 'txt' && ext !== 'pdf') {
    return false;
  }
  const mime = file.type;
  if (!mime) {
    return true;
  }
  if (ext === 'pdf') {
    return mime === 'application/pdf';
  }
  return mime === 'text/plain';
};

const formSchema = z.object({
  providerId: z.string().min(1, 'Selecione a operadora'),
  processingMonthId: z.string().min(1, 'Selecione o mês de processamento'),
  originalFileName: z.string().optional(),
  allowSubstitute: z.boolean()
});

type FormValues = z.infer<typeof formSchema>;

const LIST_PAGE_SIZE = 500;
const PROCESSING_MONTHS_PAGE_SIZE = 500;

type InvoiceImportSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Pré-seleciona a operadora quando a lista já tem um filtro ativo. */
  preferredProviderId?: string;
  /** Pré-seleciona o mês quando a lista de faturas está filtrada por mês. */
  preferredProcessingMonthId?: string;
  /** Chamado quando o progresso muda — permite a página pai exibir o banner e iniciar polling. */
  onProgressChange?: (state: ImportProgressState & { importRequestId?: string }) => void;
};

export function InvoiceImportSheet({
  open,
  onOpenChange,
  preferredProviderId = '',
  preferredProcessingMonthId = '',
  onProgressChange
}: InvoiceImportSheetProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<ImportPreview | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const [uploadedKey, setUploadedKey] = useState<string | null>(null);

  const providersQuery = useGetV1Providers(
    {
      page_index: 0,
      page_size: LIST_PAGE_SIZE
    },
    {
      query: { enabled: open }
    }
  );

  const processingMonthsQuery = useGetV1ProcessingMonths(
    {
      page_index: 0,
      page_size: PROCESSING_MONTHS_PAGE_SIZE
    },
    {
      query: { enabled: open }
    }
  );

  const defaultBucket =
    (env.VITE_STORAGE_BUCKET_NAME as string | undefined) ?? '';

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      providerId: '',
      processingMonthId: '',
      originalFileName: '',
      allowSubstitute: false
    }
  });

  useEffect(() => {
    if (open && preferredProviderId.trim().length > 0) {
      form.setValue('providerId', preferredProviderId);
    }
  }, [open, preferredProviderId, form]);

  useEffect(() => {
    if (open && preferredProcessingMonthId.trim().length > 0) {
      form.setValue('processingMonthId', preferredProcessingMonthId);
    }
  }, [open, preferredProcessingMonthId, form]);

  const importMutation = usePostV1ProviderInvoices({
    mutation: {
      onSuccess: (data) => {
        if (data.status === 'completed') {
          onProgressChange?.({
            stage: 'done',
            progress: 100,
            fileName: importFile?.name ?? ''
          });
        } else if (data.status === 'failed') {
          onProgressChange?.({
            stage: 'error',
            progress: 0,
            fileName: importFile?.name ?? '',
            errorMessage: data.error ?? 'Falha no processamento do arquivo.'
          });
        } else if (data.status === 'pdf_unparsed') {
          onProgressChange?.({
            stage: 'error',
            progress: 0,
            fileName: importFile?.name ?? '',
            errorMessage:
              'PDF recebido. O parse ainda não está disponível — use o TXT da operadora para faturar.'
          });
        } else {
          // Still processing in background -> pass importRequestId so the parent can poll
          onProgressChange?.({
            stage: 'registering',
            progress: 95,
            fileName: importFile?.name ?? '',
            importRequestId: data.id
          });
        }
        // Reset local state
        form.reset({ providerId: '', processingMonthId: '', originalFileName: '', allowSubstitute: false });
        setImportFile(null);
        setPreview(null);
        setUploadedKey(null);
        if (fileInputRef.current) fileInputRef.current.value = '';
      },
      onError: (e) => {
        onProgressChange?.({
          stage: 'error',
          progress: 0,
          fileName: importFile?.name ?? '',
          errorMessage: isApiHttpError(e) ? e.message : getErrorMessage(e)
        });
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const file = importFile;
      const storageObjectKey =
        uploadedKey ??
        (file
          ? buildInvoiceStorageObjectKey(file, values.providerId)
          : `manual/${values.providerId}/${crypto.randomUUID()}`);

      // Close the sheet immediately so the user sees the page progress banner
      onOpenChange(false);

      if (file && !uploadedKey) {
        if (!defaultBucket.trim()) {
          onProgressChange?.({
            stage: 'error',
            progress: 0,
            fileName: file.name,
            errorMessage: 'Defina VITE_STORAGE_BUCKET_NAME no ambiente para enviar o arquivo.'
          });
          toast.error('Defina VITE_STORAGE_BUCKET_NAME no ambiente para enviar o arquivo.');
          return;
        }

        // Stage 1: presigning (0 → 10%)
        onProgressChange?.({ stage: 'presigning', progress: 5, fileName: file.name });

        // Stage 2: upload with real XHR progress (10 → 85%)
        onProgressChange?.({ stage: 'uploading', progress: 10, fileName: file.name });
        await uploadFileFromPresignedUrl(file, defaultBucket, storageObjectKey, (pct) => {
          // Map XHR 0–100% to the 10–85% range
          onProgressChange?.({
            stage: 'uploading',
            progress: 10 + Math.round(pct * 0.75),
            fileName: file.name
          });
        });
      }

      // Stage 3: API registration (85 → ~95%)
      onProgressChange?.({
        stage: 'registering',
        progress: file ? 85 : 50,
        fileName: file?.name ?? values.originalFileName ?? ''
      });

      importMutation.mutate({
        data: {
          provider_id: values.providerId,
          processing_month_id: values.processingMonthId,
          storage_bucket: defaultBucket,
          storage_object_key: storageObjectKey,
          original_file_name: values.originalFileName ?? null,
          allow_substitute: values.allowSubstitute
        }
      });
    } catch (e) {
      onProgressChange?.({
        stage: 'error',
        progress: 0,
        fileName: importFile?.name ?? '',
        errorMessage: isApiHttpError(e) ? e.message : getErrorMessage(e)
      });
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  });

  const handlePreview = form.handleSubmit(async (values) => {
    const file = importFile;
    if (!file) {
      toast.error('Anexe o TXT da operadora para pré-visualizar.');
      return;
    }
    if (!defaultBucket.trim()) {
      toast.error('Defina VITE_STORAGE_BUCKET_NAME no ambiente para enviar o arquivo.');
      return;
    }
    setPreviewing(true);
    try {
      const storageObjectKey =
        uploadedKey ?? buildInvoiceStorageObjectKey(file, values.providerId);
      if (!uploadedKey) {
        await uploadFileFromPresignedUrl(file, defaultBucket, storageObjectKey);
        setUploadedKey(storageObjectKey);
      }
      const result = await previewProviderInvoiceImport({
        provider_id: values.providerId,
        processing_month_id: values.processingMonthId,
        storage_bucket: defaultBucket,
        storage_object_key: storageObjectKey,
        original_file_name: values.originalFileName ?? file.name,
        allow_substitute: values.allowSubstitute
      });
      setPreview(result);
      if (result.warnings?.length) {
        toast.message(result.warnings[0]);
      } else {
        toast.success('Prévia gerada. Confira os totais e as linhas antes de importar.');
      }
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    } finally {
      setPreviewing(false);
    }
  });

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const handleImportFile = (file: File | undefined) => {
    if (!file) return;
    if (!validInvoiceFile(file)) {
      toast.error('Envie um arquivo TXT (fatura VIVO) ou PDF (recebido, parse ainda não disponível).', {
        position: 'bottom-right',
        duration: 3000
      });
      return;
    }
    if (file.size > MAX_IMPORT_FILE_BYTES) {
      toast.error('O arquivo excede 256 MB.', { position: 'bottom-right', duration: 3000 });
      return;
    }
    setImportFile(file);
    setPreview(null);
    setUploadedKey(null);
    form.setValue('originalFileName', file.name);
  };

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    handleImportFile(event.target.files?.[0]);
  };

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    handleImportFile(event.dataTransfer.files?.[0]);
  };

  const resetImportFile = () => {
    setImportFile(null);
    setPreview(null);
    setUploadedKey(null);
    form.setValue('originalFileName', '');
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const getImportFileIcon = () => {
    if (!importFile) return <File />;
    const ext = importFile.name.split('.').pop()?.toLowerCase() ?? '';
    if (ext === 'txt') {
      return <FileText className="text-foreground h-5 w-5" aria-hidden={true} />;
    }
    return <File className="text-foreground h-5 w-5" aria-hidden={true} />;
  };

  const providers = providersQuery.data?.items ?? [];
  const selectedProviderId = form.watch('providerId');
  const openProcessingMonths = (processingMonthsQuery.data?.items ?? []).filter(
    (m: ListProcessingMonthResponse) =>
      m.status === 'open' &&
      (!selectedProviderId || m.provider_id === selectedProviderId)
  );

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          setImportFile(null);
          setPreview(null);
          setUploadedKey(null);
          form.setValue('originalFileName', '');
          if (fileInputRef.current) fileInputRef.current.value = '';
        }
        onOpenChange(next);
      }}
    >
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
          <a href="#invoice-import-month" className="skip-link">
            Pular para o formulário
          </a>
          <SheetTitle>Importar fatura</SheetTitle>
          <SheetDescription>
            Selecione o mês de processamento da importação, a operadora e,
            opcionalmente, o arquivo. Com arquivo anexado, o envio usa URL
            pré-assinada da API e depois registra a solicitação. A empresa
            contratante segue o conteúdo do arquivo (011D) no processamento.
            Configure <code className="text-xs">VITE_STORAGE_BUCKET_NAME</code>{' '}
            com o nome do bucket no R2.
          </SheetDescription>
        </SheetHeader>

        <form className="flex min-h-0 flex-1 flex-col" onSubmit={onSubmit}>
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-6">
            <Controller
              control={form.control}
              name="processingMonthId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="invoice-import-month">Mês de processamento</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={field.onChange}
                    disabled={processingMonthsQuery.isPending}
                  >
                    <SelectTrigger
                      id="invoice-import-month"
                      className="border-input bg-background w-full max-w-none rounded-xl border"
                      aria-invalid={fieldState.invalid}
                      aria-describedby={fieldState.invalid ? 'invoice-import-month-error' : undefined}
                    >
                      <SelectValue placeholder="Selecione">
                        {openProcessingMonths.find(
                          (m: ListProcessingMonthResponse) => m.id === field.value
                        )?.display_name ?? 'Selecione'}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {openProcessingMonths.map(
                          (m: ListProcessingMonthResponse) => (
                            <SelectItem key={m.id} value={m.id}>
                              {m.display_name}
                            </SelectItem>
                          )
                        )}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  {fieldState.invalid ? (
                    <FieldError id="invoice-import-month-error" errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="providerId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="invoice-import-provider">Operadora</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={field.onChange}
                    disabled={providersQuery.isPending}
                  >
                    <SelectTrigger
                      id="invoice-import-provider"
                      className="border-input bg-background w-full max-w-none rounded-xl border"
                      aria-invalid={fieldState.invalid}
                      aria-describedby={fieldState.invalid ? 'invoice-import-provider-error' : undefined}
                    >
                      <SelectValue placeholder="Selecione">
                        {providers.find(
                          (op: ListProvidersResponse) =>
                            String(op.id) === String(field.value)
                        )?.name ?? 'Selecione'}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {providers.map((op: ListProvidersResponse) => (
                          <SelectItem key={op.id} value={op.id}>
                            {op.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  {fieldState.invalid ? (
                    <FieldError id="invoice-import-provider-error" errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <div className="w-full">
              <FieldLabel htmlFor="invoice-import-file-input">
                Arquivo (opcional p/ nome e chave)
              </FieldLabel>

              <div
                className="border-input mt-2 flex justify-center rounded-md border border-dashed px-4 py-10 focus-visible:ring-ring focus-visible:ring-3 outline-none"
                onDragOver={(e) => e.preventDefault()}
                onDrop={handleDrop}
                role="button"
                tabIndex={0}
                aria-label="Área para enviar arquivo da fatura"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    fileInputRef.current?.click();
                  }
                }}
              >
                <div>
                  <File
                    className="text-muted-foreground mx-auto h-12 w-12"
                    aria-hidden={true}
                  />
                  <div className="text-muted-foreground mt-2 flex flex-wrap justify-center text-sm leading-6">
                    <p>Arraste e solte ou</p>
                    <label
                      htmlFor="invoice-import-file-input"
                      className="text-primary relative cursor-pointer rounded-sm px-1 font-medium hover:underline hover:underline-offset-4"
                    >
                      <span>escolha um arquivo</span>
                      <input
                        id="invoice-import-file-input"
                        name="invoice-import-file"
                        type="file"
                        className="sr-only"
                        accept={INVOICE_IMPORT_ACCEPT}
                        onChange={handleFileChange}
                        ref={fileInputRef}
                      />
                    </label>
                    <p className="text-pretty">para anexar</p>
                  </div>
                </div>
              </div>

              <p className="text-muted-foreground mt-2 text-xs leading-5 text-pretty sm:flex sm:items-center sm:justify-between">
                <span>Tipos aceitos: TXT (preferencial) ou PDF (recebido, parse ainda não disponível).</span>
                <span className="pl-1 sm:pl-0">Tamanho máx.: 256 MB</span>
              </p>

              {importFile ? (
                <Card className="bg-muted relative mt-4 gap-4 p-4 shadow-none">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className="text-muted-foreground hover:text-foreground absolute top-1 right-1"
                    aria-label="Remover arquivo"
                    onClick={resetImportFile}
                  >
                    <X className="h-5 w-5 shrink-0" aria-hidden={true} />
                  </Button>

                  <div className="flex items-center space-x-2.5">
                    <span className="bg-background ring-border flex h-10 w-10 shrink-0 items-center justify-center rounded-sm shadow-sm ring-1 ring-inset">
                      {getImportFileIcon()}
                    </span>
                    <div className="min-w-0">
                      <p className="text-foreground truncate text-xs font-medium text-pretty">
                        {importFile.name}
                      </p>
                      <p className="text-muted-foreground mt-0.5 text-xs text-pretty">
                        {formatFileSize(importFile.size)}
                      </p>
                    </div>
                  </div>
                </Card>
              ) : null}
            </div>

            <Controller
              control={form.control}
              name="allowSubstitute"
              render={({ field }) => (
                <Field>
                  <label htmlFor="invoice-import-allow-substitute" className="flex cursor-pointer items-start gap-2 text-sm">
                    <input
                      id="invoice-import-allow-substitute"
                      type="checkbox"
                      className="mt-1 size-4 accent-primary"
                      checked={field.value}
                      onChange={(e) => field.onChange(e.target.checked)}
                    />
                    <span>
                      <span className="font-medium">Fatura substituta</span>
                      <span className="text-muted-foreground mt-0.5 block text-xs">
                        A fatura anterior permanece no histórico, marcada como substituída. Não apaga o original.
                      </span>
                    </span>
                  </label>
                </Field>
              )}
            />
            {preview ? (
              <div className="rounded-xl border p-4 text-sm">
                <p className="font-medium">Prévia da importação</p>
                <dl className="mt-2 grid grid-cols-2 gap-2">
                  <div>
                    <dt className="text-muted-foreground text-xs">Conta</dt>
                    <dd>{preview.summary.account_number || '—'}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground text-xs">Referência</dt>
                    <dd>{preview.summary.invoice_number || '—'}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground text-xs">Total</dt>
                    <dd>{formatMoney(preview.summary.total_amount)}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground text-xs">Linhas</dt>
                    <dd>
                      {preview.summary.lines_count} ({preview.summary.known_lines} conhecidas,{' '}
                      {preview.summary.unknown_lines} novas)
                    </dd>
                  </div>
                </dl>
                {preview.file_sha256 ? (
                  <p className="text-muted-foreground mt-2 font-mono text-xs break-all">
                    SHA-256 {preview.file_sha256}
                  </p>
                ) : null}
                {(preview.warnings ?? []).map((w) => (
                  <p key={w} className="text-destructive mt-1 text-xs">
                    {w}
                  </p>
                ))}
                {!preview.is_valid ? (
                  <p className="text-destructive mt-2 text-xs">
                    Prévia inválida — ajuste o arquivo ou marque fatura substituta se for duplicata.
                  </p>
                ) : null}
              </div>
            ) : null}
          </div>

          <SheetFooter className="gap-4 border-t pt-6 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>
              Cancelar
            </SheetClose>
            {importFile ? (
              <Button type="button" variant="secondary" disabled={previewing} onClick={() => void handlePreview()}>
                {previewing ? 'Pré-visualizando…' : 'Pré-visualizar'}
              </Button>
            ) : null}
            <Button type="submit" disabled={Boolean(preview && !preview.is_valid)}>
              {preview ? 'Confirmar importação' : 'Enviar'}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
