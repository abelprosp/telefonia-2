import { useEffect, useRef, useState } from 'react';

import { useNavigate } from '@tanstack/react-router';
import { FileStack, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Progress } from '@/components/ui/progress';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  useGenerateCustomerBillingDocument,
  useManualBillingPreview
} from '@/lib/billing-api';
import { formatMoney, todayISO } from '@/lib/financial-api';

type GenerateCustomerInvoiceSheetProps = {
  customerId: string;
  customerName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function GenerateCustomerInvoiceSheet({
  customerId,
  customerName,
  open,
  onOpenChange
}: GenerateCustomerInvoiceSheetProps) {
  const navigate = useNavigate();
  const previewQuery = useManualBillingPreview(
    open ? [customerId] : undefined,
    open
  );
  const generateMutation = useGenerateCustomerBillingDocument();

  const [issueDate, setIssueDate] = useState(todayISO());
  const [dueDate, setDueDate] = useState(todayISO());
  const [description, setDescription] = useState('Mensalidade telefonia');
  const [amount, setAmount] = useState('');

  // Progress animation while generating
  const [genProgress, setGenProgress] = useState(0);
  const [genStage, setGenStage] = useState(0);
  const genStages = ['Calculando valores…', 'Gerando boleto…', 'Finalizando…'];
  const genIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (generateMutation.isPending) {
      setGenProgress(5);
      setGenStage(0);
      let elapsed = 0;
      genIntervalRef.current = setInterval(() => {
        elapsed += 1;
        // Simulate progress: ramps up to ~90% over ~10 seconds, then stalls
        setGenProgress((prev) => {
          if (prev >= 90) return prev;
          const increment = Math.max(1, Math.round((90 - prev) * 0.12));
          return Math.min(90, prev + increment);
        });
        // Cycle through stage labels every 3s
        setGenStage(Math.min(genStages.length - 1, Math.floor(elapsed / 3)));
      }, 1000);
    } else {
      if (genIntervalRef.current) {
        clearInterval(genIntervalRef.current);
        genIntervalRef.current = null;
      }
      if (genProgress > 0) {
        setGenProgress(100);
        setTimeout(() => setGenProgress(0), 700);
      }
    }
    return () => {
      if (genIntervalRef.current) clearInterval(genIntervalRef.current);
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [generateMutation.isPending]);

  const previewItems = (previewQuery.data?.items ?? []).filter(
    (i) => i.customer_id === customerId
  );
  const previewItem = previewItems[0];
  const eligibleItems = previewItems.filter((i) => i.eligible);
  const multipleGroups = previewItems.length > 1;

  useEffect(() => {
    if (!open) return;
    if (!multipleGroups && previewItem && previewItem.monthly_amount > 0) {
      setAmount(String(previewItem.monthly_amount));
    }
  }, [open, previewItem, multipleGroups]);

  const handleGenerate = () => {
    const parsedAmount =
      !multipleGroups && amount.trim() ? Number(amount) : undefined;
    if (parsedAmount !== undefined && (Number.isNaN(parsedAmount) || parsedAmount <= 0)) {
      toast.error('Informe um valor válido.');
      return;
    }
    generateMutation.mutate(
      {
        customerId,
        issue_date: issueDate,
        due_date: dueDate,
        description: description.trim() || 'Mensalidade telefonia',
        ...(parsedAmount !== undefined ? { amount: parsedAmount } : {})
      },
      {
        onSuccess: (data) => {
          toast.success(data.message);
          onOpenChange(false);
          if ((data.document_ids?.length ?? 0) > 1 || (data.created_count ?? 1) > 1) {
            void navigate({
              to: '/finance/customer-invoices',
              search: { page: 1, pageSize: 10 }
            });
            return;
          }
          void navigate({
            to: '/finance/customer-invoices/$id',
            params: { id: data.id }
          });
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

  const generateDisabled =
    generateMutation.isPending ||
    (previewItems.length > 0 &&
      eligibleItems.length === 0 &&
      previewItems.every((i) => i.skip_reason === 'no_active_lines')) ||
    (!multipleGroups &&
      previewItem != null &&
      !previewItem.eligible &&
      previewItem.skip_reason === 'no_monthly_amount' &&
      !amount.trim());

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>Gerar fatura</SheetTitle>
          <SheetDescription>
            Cria fatura com boleto Sicredi (um boleto por linha Normal; Titular+dependentes no mesmo boleto) para {customerName}.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-4 px-4">
          {previewQuery.isPending && (
            <p className="text-muted-foreground flex items-center gap-2 text-sm">
              <Loader2 className="size-4 animate-spin" />
              Calculando valor sugerido…
            </p>
          )}
          {previewItem && !previewItem.eligible && !multipleGroups && (
            <p className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              {previewItem.skip_reason === 'no_monthly_amount'
                ? 'Vincule linhas/aparelhos com valor mensal ou informe o valor abaixo.'
                : previewItem.skip_reason === 'no_active_lines'
                  ? 'Vincule ao menos uma linha ou aparelho ao cliente.'
                  : 'Cliente não está pronto para faturamento.'}
            </p>
          )}
          {previewItems.length > 0 && (
            <div className="text-muted-foreground space-y-1 text-sm">
              {multipleGroups ? (
                <ul className="list-disc space-y-1 pl-4">
                  {previewItems.map((item) => (
                    <li key={item.billing_group_id || item.customer_id}>
                      {item.group_label || item.phone_line_number || 'Grupo'} —{' '}
                      {item.eligible ? formatMoney(item.monthly_amount) : 'não elegível'}
                    </li>
                  ))}
                </ul>
              ) : previewItem && previewItem.monthly_amount > 0 ? (
                <p>
                  Valor sugerido: {formatMoney(previewItem.monthly_amount)} (
                  {previewItem.line_count} linha(s), {previewItem.device_count ?? 0} aparelho(s)
                  {previewItem.group_label ? ` · ${previewItem.group_label}` : ''})
                </p>
              ) : (
                <p>Informe o valor da fatura abaixo.</p>
              )}
              {previewItem && !previewItem.billing_email && (
                <span className="mt-1 block text-amber-700">
                  E-mail não cadastrado — você pode gerar e baixar a fatura mesmo assim.
                </span>
              )}
            </div>
          )}
          <div>
            <Label>Descrição</Label>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label>Emissão</Label>
              <Input type="date" value={issueDate} onChange={(e) => setIssueDate(e.target.value)} />
            </div>
            <div>
              <Label>Vencimento</Label>
              <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
            </div>
          </div>
          {!multipleGroups ? (
            <div>
              <Label>Valor (R$)</Label>
              <Input
                type="number"
                step="0.01"
                min="0"
                placeholder="Valor da fatura"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
          ) : null}
        </div>
        <SheetFooter className="flex-col gap-3">
          {generateMutation.isPending || genProgress > 0 ? (
            <div className="w-full space-y-1.5 px-1">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground text-xs">
                  {generateMutation.isPending ? genStages[genStage] : 'Concluído'}
                </span>
                <span className="text-muted-foreground text-xs tabular-nums font-medium">
                  {genProgress === 100 ? '100%' : `${genProgress}%`}
                </span>
              </div>
              <Progress
                value={genProgress}
                className="h-1.5 transition-all duration-500"
              />
            </div>
          ) : null}
          <Button onClick={handleGenerate} disabled={generateDisabled} className="w-full sm:w-auto">
            <FileStack className="mr-2 size-4" />
            {generateMutation.isPending
              ? 'Gerando…'
              : multipleGroups
                ? `Gerar ${eligibleItems.length || previewItems.length} boleto(s)`
                : 'Gerar fatura + boleto'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
