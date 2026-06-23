import {
  type ChangeEvent,
  type DragEvent,
  useEffect,
  useRef,
  useState
} from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { File, FileText, X } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import type { ListProcessingMonthResponse, ListProvidersResponse } from '@/api';
import {
  getV1ProviderInvoicesQueryKey,
  useGetV1ProcessingMonths,
  useGetV1Providers,
  usePostV1ProviderInvoices
} from '@/api';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Progress } from '@/components/ui/progress';
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
import {
  buildInvoiceStorageObjectKey,
  uploadFileFromPresignedUrl
} from '@/lib/invoice-import-upload';
import { cn } from '@/lib/utils';

const MAX_IMPORT_FILE_BYTES = 256 * 1024 * 1024;

const INVOICE_IMPORT_ACCEPT = '.txt';

const validInvoiceFile = (file: File) => {
  const ext = file.name.split('.').pop()?.toLowerCase() ?? '';
  if (ext !== 'txt') {
    return false;
  }
  const mime = file.type;
  if (!mime) {
    return true;
  }
  return mime === 'text/plain';
};

const formSchema = z.object({
  providerId: z.string().min(1, 'Selecione a operadora'),
  processingMonthId: z.string().min(1, 'Selecione o mês de processamento'),
  originalFileName: z.string().optional()
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
};

export function InvoiceImportSheet({
  open,
  onOpenChange,
  preferredProviderId = '',
  preferredProcessingMonthId = ''
}: InvoiceImportSheetProps) {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [importFile, setImportFile] = useState<File | null>(null);

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
      originalFileName: ''
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
      onSuccess: (_data) => {
        toast.success(
          'Importação solicitada. O processamento ocorre em segundo plano.'
        );
        void queryClient.invalidateQueries({
          queryKey: getV1ProviderInvoicesQueryKey()
        });
        onOpenChange(false);
        form.reset({
          providerId: '',
          processingMonthId: '',
          originalFileName: ''
        });
        setImportFile(null);
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const file = importFile;
      const storageObjectKey = file
        ? buildInvoiceStorageObjectKey(file, values.providerId)
        : `manual/${values.providerId}/${crypto.randomUUID()}`;

      if (file) {
        if (!defaultBucket.trim()) {
          toast.error(
            'Defina VITE_STORAGE_BUCKET_NAME no ambiente para enviar o arquivo.'
          );
          return;
        }
        await uploadFileFromPresignedUrl(file, defaultBucket, storageObjectKey);
      }

      importMutation.mutate({
        data: {
          provider_id: values.providerId,
          processing_month_id: values.processingMonthId,
          storage_bucket: defaultBucket,
          storage_object_key: storageObjectKey,
          original_file_name: values.originalFileName ?? null
        }
      });
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  });

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) {
      return '0 Bytes';
    }
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const handleImportFile = (file: File | undefined) => {
    if (!file) {
      return;
    }
    if (!validInvoiceFile(file)) {
      toast.error('Envie um arquivo TXT.', {
        position: 'bottom-right',
        duration: 3000
      });
      return;
    }
    if (file.size > MAX_IMPORT_FILE_BYTES) {
      toast.error('O arquivo excede 256 MB.', {
        position: 'bottom-right',
        duration: 3000
      });
      return;
    }
    setImportFile(file);
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
    form.setValue('originalFileName', '');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const getImportFileIcon = () => {
    if (!importFile) {
      return <File />;
    }
    const ext = importFile.name.split('.').pop()?.toLowerCase() ?? '';
    if (ext === 'txt') {
      return (
        <FileText className="text-foreground h-5 w-5" aria-hidden={true} />
      );
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
          form.setValue('originalFileName', '');
          if (fileInputRef.current) {
            fileInputRef.current.value = '';
          }
        }
        onOpenChange(next);
      }}
    >
      <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
        <SheetHeader>
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
                  <FieldLabel>Mês de processamento</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={field.onChange}
                    disabled={processingMonthsQuery.isPending}
                  >
                    <SelectTrigger className="border-input bg-background w-full max-w-none rounded-xl border">
                      <SelectValue placeholder="Selecione">
                        {openProcessingMonths.find(
                          (m: ListProcessingMonthResponse) =>
                            m.id === field.value
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
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="providerId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Operadora</FieldLabel>
                  <Select
                    value={field.value || ''}
                    onValueChange={field.onChange}
                    disabled={providersQuery.isPending}
                  >
                    <SelectTrigger className="border-input bg-background w-full max-w-none rounded-xl border">
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
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <div className="w-full">
              <FieldLabel htmlFor="invoice-import-file-input">
                Arquivo (opcional p/ nome e chave)
              </FieldLabel>

              <div
                className="border-input mt-2 flex justify-center rounded-md border border-dashed px-4 py-10"
                onDragOver={(e) => e.preventDefault()}
                onDrop={handleDrop}
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
                <span>Tipos aceitos: TXT.</span>
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
                    disabled={importMutation.isPending}
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

                  <div className="flex items-center space-x-3">
                    <Progress
                      value={importMutation.isPending ? 100 : 0}
                      className={cn(
                        'h-1.5',
                        importMutation.isPending && 'opacity-90'
                      )}
                    />
                    <span className="text-muted-foreground shrink-0 text-xs tabular-nums">
                      {importMutation.isPending ? '…' : '0%'}
                    </span>
                  </div>
                </Card>
              ) : null}
            </div>
          </div>

          <SheetFooter className="gap-4 border-t pt-6 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button type="submit" disabled={importMutation.isPending}>
              {importMutation.isPending ? 'Enviando…' : 'Enviar'}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}
