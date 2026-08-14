import { CheckCircle2, FileText, Loader2, RefreshCw, X } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';

export type ImportProgressStage =
  | 'idle'
  | 'presigning'
  | 'uploading'
  | 'registering'
  | 'processing'   // polling backend: pending | processing
  | 'done'
  | 'error';

export type ImportProgressState = {
  stage: ImportProgressStage;
  progress: number;
  fileName: string;
  errorMessage?: string;
};

const STAGE_LABELS: Record<ImportProgressStage, string> = {
  idle: '',
  presigning: 'Obtendo URL segura…',
  uploading: 'Enviando arquivo…',
  registering: 'Registrando solicitação…',
  processing: 'Processando arquivo…',
  done: 'Importação concluída! A fatura já está na lista.',
  error: 'Erro na importação'
};

// Steps shown during active upload/register phases
const UPLOAD_STEPS: ImportProgressStage[] = ['presigning', 'uploading', 'registering'];

type Props = {
  state: ImportProgressState;
  onDismiss: () => void;
  onRefresh?: () => void;
};

export function InvoiceImportProgressBanner({ state, onDismiss, onRefresh }: Props) {
  if (state.stage === 'idle') return null;

  const isUploadPhase = UPLOAD_STEPS.includes(state.stage);
  const isProcessing = state.stage === 'processing';
  const isActive = isUploadPhase || isProcessing;
  const isDone = state.stage === 'done';
  const isError = state.stage === 'error';

  return (
    <div
      className={cn(
        'relative flex flex-col gap-3 rounded-xl border px-5 py-4 shadow-sm transition-all duration-300',
        isDone && 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950/30',
        isError && 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950/30',
        isActive && 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950/30'
      )}
      role="status"
      aria-live="polite"
    >
      {/* Dismiss / Refresh buttons */}
      <div className="absolute right-2 top-2 flex gap-1">
        {isDone && onRefresh && (
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="h-6 w-6 text-green-600 hover:bg-green-100"
            onClick={onRefresh}
            aria-label="Recarregar lista"
            title="Recarregar lista"
          >
            <RefreshCw className="size-3.5" />
          </Button>
        )}
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className={cn(
            'h-6 w-6 text-muted-foreground hover:text-foreground',
            isDone && 'text-green-600 hover:bg-green-100 hover:text-green-700',
            isError && 'text-red-500 hover:bg-red-100 hover:text-red-600',
            isActive && 'text-blue-500 hover:bg-blue-100 hover:text-blue-700'
          )}
          onClick={onDismiss}
          aria-label={isActive ? 'Cancelar acompanhamento' : 'Fechar'}
          title={isActive ? 'Cancelar acompanhamento' : 'Fechar'}
        >
          <X className="size-3.5" />
        </Button>
      </div>

      {/* Header */}
      <div className="flex items-center gap-3">
        <span
          className={cn(
            'flex h-9 w-9 shrink-0 items-center justify-center rounded-lg',
            isDone && 'bg-green-100 text-green-600',
            isError && 'bg-red-100 text-red-500',
            isActive && 'bg-blue-100 text-blue-600'
          )}
        >
          {isActive && <Loader2 className="size-4 animate-spin" />}
          {isDone && <CheckCircle2 className="size-4" />}
          {isError && <X className="size-4" />}
        </span>

        <div className="min-w-0 flex-1">
          <p
            className={cn(
              'text-sm font-semibold',
              isDone && 'text-green-800',
              isError && 'text-red-700',
              isActive && 'text-blue-800'
            )}
          >
            {STAGE_LABELS[state.stage]}
          </p>
          {state.fileName && (
            <p
              className={cn(
                'mt-0.5 flex items-center gap-1 truncate text-xs',
                isDone && 'text-green-600',
                isError && 'text-red-500',
                isActive && 'text-blue-600'
              )}
            >
              <FileText className="size-3 shrink-0" />
              {state.fileName}
            </p>
          )}
          {isError && state.errorMessage && (
            <p className="mt-0.5 text-xs text-red-600">{state.errorMessage}</p>
          )}
        </div>

        <span
          className={cn(
            'shrink-0 text-sm font-bold tabular-nums',
            isDone && 'text-green-700',
            isError && 'text-red-600',
            isActive && 'text-blue-700'
          )}
        >
          {isProcessing ? '…' : `${state.progress}%`}
        </span>
      </div>

      {/* Progress bar */}
      {!isError && (
        <Progress
          value={isProcessing ? null : state.progress}
          className={cn(
            'h-2 transition-all duration-500',
            isDone && '[&>div]:bg-green-500',
            isProcessing && 'animate-pulse [&>div]:bg-blue-400',
            isUploadPhase && '[&>div]:bg-blue-500'
          )}
        />
      )}

      {/* Step indicators — upload phase */}
      {isUploadPhase && (
        <div className="flex items-center gap-4">
          {UPLOAD_STEPS.map((s, i) => {
            const stepDone = UPLOAD_STEPS.indexOf(state.stage) > i;
            const stepActive = state.stage === s;
            return (
              <div key={s} className="flex items-center gap-1.5">
                <span
                  className={cn(
                    'flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold',
                    stepDone && 'bg-blue-500 text-white',
                    stepActive && 'bg-blue-200 text-blue-700 ring-2 ring-blue-400',
                    !stepDone && !stepActive && 'bg-blue-100 text-blue-400'
                  )}
                >
                  {stepDone ? '✓' : i + 1}
                </span>
                <span
                  className={cn(
                    'text-xs',
                    stepDone && 'text-blue-600',
                    stepActive && 'font-medium text-blue-700',
                    !stepDone && !stepActive && 'text-blue-400'
                  )}
                >
                  {s === 'presigning' ? 'URL segura' : s === 'uploading' ? 'Upload' : 'Registro'}
                </span>
              </div>
            );
          })}
        </div>
      )}

      {/* Processing hint */}
      {isProcessing && (
        <div className="flex items-center justify-between gap-2 pt-1 text-xs text-blue-600">
          <span>
            O backend está parseando o arquivo TXT da operadora e gravando as linhas no banco de dados…
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-7 shrink-0 text-xs border-blue-300 bg-white/80 hover:bg-white text-blue-700 hover:text-blue-800"
            onClick={onDismiss}
          >
            Cancelar
          </Button>
        </div>
      )}
    </div>
  );
}
