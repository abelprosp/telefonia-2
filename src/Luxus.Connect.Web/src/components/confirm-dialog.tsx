import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

type ConfirmDialogProps = {
  open: boolean;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
  /** When set, user must type this exact phrase to enable confirm (mass/irreversible ops). */
  requirePhrase?: string;
  loading?: boolean;
  onConfirm: () => void | Promise<void>;
  onCancel: () => void;
};

export function ConfirmDialog(props: ConfirmDialogProps) {
  if (!props.open) return null;
  // Remount while open so the typed phrase resets each time the dialog appears.
  return <ConfirmDialogOpen key="confirm-open" {...props} />;
}

function ConfirmDialogOpen({
  title,
  description,
  confirmLabel = 'Confirmar',
  cancelLabel = 'Cancelar',
  destructive = false,
  requirePhrase,
  loading = false,
  onConfirm,
  onCancel
}: ConfirmDialogProps) {
  const [phrase, setPhrase] = useState('');
  const phraseOk = !requirePhrase || phrase.trim() === requirePhrase;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      role="presentation"
      onClick={onCancel}
    >
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        aria-describedby="confirm-dialog-desc"
        className={cn(
          'bg-background border-border w-full max-w-md rounded-lg border p-5 shadow-lg'
        )}
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="confirm-dialog-title" className="text-base font-semibold">
          {title}
        </h2>
        <p id="confirm-dialog-desc" className="text-muted-foreground mt-2 text-sm whitespace-pre-wrap">
          {description}
        </p>
        {requirePhrase ? (
          <div className="mt-4 space-y-2">
            <p className="text-muted-foreground text-xs">
              Digite <span className="font-mono font-medium">{requirePhrase}</span> para confirmar.
            </p>
            <Input
              value={phrase}
              onChange={(e) => setPhrase(e.target.value)}
              autoFocus
              aria-label="Confirmação reforçada"
            />
          </div>
        ) : null}
        <div className="mt-5 flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            type="button"
            variant={destructive ? 'destructive' : 'default'}
            disabled={!phraseOk || loading}
            onClick={() => void onConfirm()}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
