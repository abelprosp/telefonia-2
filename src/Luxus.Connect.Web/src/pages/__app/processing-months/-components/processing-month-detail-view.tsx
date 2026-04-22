import { useState } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { format, parseISO } from 'date-fns';
import { ExternalLink, Lock } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { z } from 'zod';

import type { GetProcessingMonthResponse } from '@/api';
import {
  getV1ProcessingMonthsQueryKey,
  processingMonthsControllerGetByIdQueryKey,
  usePostV1ProcessingMonthsIdClose,
  usePostV1ProcessingMonthsIdCloseContingency
} from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatProcessingMonthStatus } from '@/lib/format';

export type ProcessingMonthListSearch = {
  page: number;
  pageSize: number;
};

type ProcessingMonthDetailViewProps = {
  month: GetProcessingMonthResponse;
  listSearch: ProcessingMonthListSearch;
  providerName?: string;
};

function isProcessingMonthOpen(status: string | number | undefined): boolean {
  if (status === undefined || status === null) {
    return false;
  }
  if (typeof status === 'number') {
    return status === 0;
  }
  const s = String(status).trim().toLowerCase();
  return s === 'open' || s === '0';
}

function formatClosedAt(raw: string | null | undefined): string {
  if (!raw) {
    return '—';
  }
  try {
    return format(parseISO(raw), 'dd/MM/yyyy HH:mm');
  } catch {
    return raw;
  }
}

const contingencySchema = z.object({
  justification: z
    .string()
    .min(5, 'Informe uma justificativa (mínimo 5 caracteres)')
    .max(2000, 'Texto muito longo')
});

type ContingencyForm = z.infer<typeof contingencySchema>;

export function ProcessingMonthDetailView({
  month,
  listSearch,
  providerName
}: ProcessingMonthDetailViewProps) {
  const queryClient = useQueryClient();
  const [closeSheetOpen, setCloseSheetOpen] = useState(false);
  const [contingencySheetOpen, setContingencySheetOpen] = useState(false);

  const open = isProcessingMonthOpen(month.status);

  const contingencyForm = useForm<ContingencyForm>({
    resolver: zodResolver(contingencySchema),
    defaultValues: { justification: '' }
  });

  const closeMutation = usePostV1ProcessingMonthsIdClose({
    mutation: {
      onSuccess: async () => {
        toast.success('Mês fechado.');
        setCloseSheetOpen(false);
        await queryClient.invalidateQueries({
          queryKey: getV1ProcessingMonthsQueryKey()
        });
        await queryClient.invalidateQueries({
          queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const contingencyMutation = usePostV1ProcessingMonthsIdCloseContingency({
    mutation: {
      onSuccess: async () => {
        toast.success('Mês fechado em contingência.');
        setContingencySheetOpen(false);
        contingencyForm.reset({ justification: '' });
        await queryClient.invalidateQueries({
          queryKey: getV1ProcessingMonthsQueryKey()
        });
        await queryClient.invalidateQueries({
          queryKey: processingMonthsControllerGetByIdQueryKey(month.id)
        });
      },
      onError: (e) => {
        toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
      }
    }
  });

  const invoicesSearch = {
    page: 1,
    pageSize: listSearch.pageSize,
    processingMonthId: month.id
  };

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h3 className="text-foreground text-lg font-semibold">
            {month.display_name}
          </h3>
          <p className="text-muted-foreground mt-1 text-sm">
            {providerName ?? month.provider_id}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button
            nativeButton={false}
            variant="outline"
            size="sm"
            render={
              <Link
                to="/invoices"
                search={invoicesSearch}
                className="inline-flex items-center gap-2"
              >
                <ExternalLink className="size-4" />
                Faturas deste mês
              </Link>
            }
          />
          {open ? (
            <>
              <Button
                type="button"
                variant="default"
                size="sm"
                onClick={() => setCloseSheetOpen(true)}
              >
                <Lock className="size-4" />
                Fechar mês
              </Button>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => setContingencySheetOpen(true)}
              >
                Fecho em contingência
              </Button>
            </>
          ) : null}
        </div>
      </div>

      <Separator />

      <dl className="grid max-w-2xl grid-cols-1 gap-6 sm:grid-cols-2">
        <div>
          <dt className="text-muted-foreground text-sm">Situação</dt>
          <dd className="mt-1 font-medium">
            {formatProcessingMonthStatus(month.status)}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Competência (API)</dt>
          <dd className="mt-1 font-medium tabular-nums">
            {month.month.toString().padStart(2, '0')}/{month.year}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Fechamento</dt>
          <dd className="mt-1">{formatClosedAt(month.closed_at)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Fechado por</dt>
          <dd className="mt-1">{month.closed_by ?? '—'}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground text-sm">Contingência</dt>
          <dd className="mt-1">
            {month.closed_in_contingency ? 'Sim' : 'Não'}
          </dd>
        </div>
        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-sm">
            Justificativa de contingência
          </dt>
          <dd className="mt-1 whitespace-pre-wrap">
            {month.contingency_justification ?? '—'}
          </dd>
        </div>
      </dl>

      <Sheet open={closeSheetOpen} onOpenChange={setCloseSheetOpen}>
        <SheetContent className="flex w-full flex-col sm:max-w-md">
          <SheetHeader>
            <SheetTitle>Fechar mês</SheetTitle>
            <SheetDescription>
              Após o fechamento, operações mutáveis ligadas a este mês ficam
              bloqueadas conforme as regras do sistema. Esta ação não pode ser
              desfeita aqui.
            </SheetDescription>
          </SheetHeader>
          <SheetFooter className="mt-auto gap-2 border-t pt-4 sm:justify-end">
            <SheetClose render={<Button type="button" variant="outline" />}>
              Cancelar
            </SheetClose>
            <Button
              type="button"
              disabled={closeMutation.isPending}
              onClick={() => closeMutation.mutate({ id: month.id })}
            >
              {closeMutation.isPending ? 'Fechando…' : 'Confirmar fechamento'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Sheet
        open={contingencySheetOpen}
        onOpenChange={(next) => {
          if (!next) {
            contingencyForm.reset({ justification: '' });
          }
          setContingencySheetOpen(next);
        }}
      >
        <SheetContent side="right" className="flex w-full flex-col sm:max-w-lg">
          <SheetHeader>
            <SheetTitle>Fechamento em contingência</SheetTitle>
            <SheetDescription>
              Registre a justificativa do fecho administrativo em contingência.
            </SheetDescription>
          </SheetHeader>

          <form
            className="flex min-h-0 flex-1 flex-col gap-4"
            onSubmit={contingencyForm.handleSubmit((values) =>
              contingencyMutation.mutate({
                id: month.id,
                data: { justification: values.justification.trim() }
              })
            )}
          >
            <Controller
              control={contingencyForm.control}
              name="justification"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="pm-contingency-justification">
                    Justificativa
                  </FieldLabel>
                  <Textarea
                    id="pm-contingency-justification"
                    className="border-input bg-background min-h-32 rounded-xl border"
                    placeholder="Descreva o motivo do fecho em contingência…"
                    {...field}
                  />
                  {fieldState.invalid ? (
                    <FieldError errors={[fieldState.error]} />
                  ) : null}
                </Field>
              )}
            />

            <SheetFooter className="mt-auto gap-2 border-t pt-4 sm:justify-end">
              <SheetClose render={<Button type="button" variant="outline" />}>
                Cancelar
              </SheetClose>
              <Button type="submit" disabled={contingencyMutation.isPending}>
                {contingencyMutation.isPending
                  ? 'Enviando…'
                  : 'Confirmar fecho em contingência'}
              </Button>
            </SheetFooter>
          </form>
        </SheetContent>
      </Sheet>
    </div>
  );
}
