import { useEffect, useState } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import { Plus, Trash2 } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { withMask } from 'use-mask-input';
import { z } from 'zod';

import { useGetV1Providers, useProvidersControllerGetById } from '@/api';
import { ConfirmDialog } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
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
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { useBulkCreateStockPhoneLines } from '@/lib/stock-api';

const formSchema = z.object({
  providerId: z.string().min(1, 'Selecione a operadora'),
  providerAccountNumber: z.string().min(1, 'Informe o número da conta'),
  providerPlanId: z.string().min(1, 'Selecione o plano')
});

type FormValues = z.infer<typeof formSchema>;

type StockLineBulkCreateSheetProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
};

export function StockLineBulkCreateSheet({
  open,
  onOpenChange,
  onSuccess
}: StockLineBulkCreateSheetProps) {
  const bulkMutation = useBulkCreateStockPhoneLines();
  const providersQuery = useGetV1Providers({ page_index: 0, page_size: 500 });
  const [numbers, setNumbers] = useState<string[]>(['']);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [pendingValues, setPendingValues] = useState<FormValues | null>(null);
  const [lastSummary, setLastSummary] = useState<string | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      providerId: '',
      providerAccountNumber: '',
      providerPlanId: ''
    }
  });

  const providerId = form.watch('providerId');
  const providerDetailQuery = useProvidersControllerGetById(providerId, {
    query: { enabled: Boolean(providerId) }
  });

  useEffect(() => {
    form.setValue('providerPlanId', '');
  }, [providerId, form]);

  useEffect(() => {
    if (!open) {
      form.reset();
      setNumbers(['']);
      setLastSummary(null);
      setConfirmOpen(false);
      setPendingValues(null);
    }
  }, [open, form]);

  const plans = providerDetailQuery.data?.plans ?? [];
  const filledCount = numbers.filter((n) => n.replace(/\D/g, '').length >= 8).length;

  const requestSubmit = form.handleSubmit((values) => {
    if (filledCount === 0) {
      toast.error('Informe ao menos um número válido.');
      return;
    }
    setPendingValues(values);
    setConfirmOpen(true);
  });

  const runBulk = async () => {
    if (!pendingValues) return;
    try {
      const data = await bulkMutation.mutateAsync({
        provider_id: pendingValues.providerId,
        provider_account_number: pendingValues.providerAccountNumber.trim(),
        provider_plan_id: pendingValues.providerPlanId,
        lines: numbers
          .map((n) => n.trim())
          .filter(Boolean)
          .map((number) => ({ number }))
      });
      const summary = `Cadastradas: ${data.created}. Ignoradas: ${data.ignored}. Rejeitadas: ${data.rejected}.`;
      setLastSummary(summary);
      toast.success(summary);
      setConfirmOpen(false);
      if (data.created > 0) onSuccess?.();
      if (data.rejected === 0 && data.ignored === 0) onOpenChange(false);
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="overflow-y-auto sm:max-w-lg">
          <SheetHeader>
            <SheetTitle>Cadastro em lote de linhas</SheetTitle>
            <SheetDescription>
              Defina operadora, conta e plano compartilhados e adicione múltiplos números. Cada
              linha é validada individualmente; inválidas não são gravadas.
            </SheetDescription>
          </SheetHeader>

          <form onSubmit={requestSubmit} className="flex flex-col gap-4 px-4">
            <Controller
              control={form.control}
              name="providerId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Operadora (compartilhada)</FieldLabel>
                  <Select value={field.value} onValueChange={(v) => field.onChange(v ?? '')}>
                    <SelectTrigger>
                      <SelectValue placeholder="Selecione" />
                    </SelectTrigger>
                    <SelectContent>
                      {(providersQuery.data?.items ?? []).map((p) => (
                        <SelectItem key={p.id} value={p.id}>
                          {p.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="providerAccountNumber"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Conta (compartilhada)</FieldLabel>
                  <Input {...field} />
                  {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
                </Field>
              )}
            />

            <Controller
              control={form.control}
              name="providerPlanId"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel>Plano (compartilhado)</FieldLabel>
                  <Select
                    value={field.value}
                    onValueChange={(v) => field.onChange(v ?? '')}
                    disabled={!providerId}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Selecione o plano" />
                    </SelectTrigger>
                    <SelectContent>
                      {plans.map((p) => (
                        <SelectItem key={p.id} value={p.id}>
                          {p.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
                </Field>
              )}
            />

            <div className="space-y-2">
              <FieldLabel>Números</FieldLabel>
              {numbers.map((value, index) => (
                <div key={index} className="flex gap-2">
                  <Input
                    value={value}
                    ref={withMask('(99) 99999-9999')}
                    onChange={(e) => {
                      const next = [...numbers];
                      next[index] = e.target.value;
                      setNumbers(next);
                    }}
                    placeholder="(51) 99999-8888"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    disabled={numbers.length <= 1}
                    onClick={() => setNumbers(numbers.filter((_, i) => i !== index))}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setNumbers([...numbers, ''])}
              >
                <Plus className="size-4" />
                Adicionar linha
              </Button>
            </div>

            {lastSummary ? (
              <p className="bg-muted rounded-md px-3 py-2 text-sm">{lastSummary}</p>
            ) : null}

            <SheetFooter className="px-0">
              <SheetClose render={<Button type="button" variant="outline" />}>
                Fechar
              </SheetClose>
              <Button type="submit" disabled={bulkMutation.isPending}>
                Validar e cadastrar ({filledCount})
              </Button>
            </SheetFooter>
          </form>
        </SheetContent>
      </Sheet>

      <ConfirmDialog
        open={confirmOpen}
        title="Confirmar cadastro em lote"
        description={`Serão processadas ${filledCount} linha(s) com a operadora/conta/plano selecionados. Duplicadas no lote ou no banco serão ignoradas; inválidas serão rejeitadas sem gravação silenciosa.`}
        confirmLabel="Cadastrar lote"
        loading={bulkMutation.isPending}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={runBulk}
      />
    </>
  );
}
