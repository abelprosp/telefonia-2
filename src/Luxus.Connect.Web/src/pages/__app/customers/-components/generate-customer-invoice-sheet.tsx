import { useEffect, useState } from 'react';

import { useNavigate } from '@tanstack/react-router';
import { FileStack, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  const previewQuery = useManualBillingPreview(open ? [customerId] : undefined);
  const generateMutation = useGenerateCustomerBillingDocument();

  const [issueDate, setIssueDate] = useState(todayISO());
  const [dueDate, setDueDate] = useState(todayISO());
  const [description, setDescription] = useState('Mensalidade telefonia');
  const [amount, setAmount] = useState('');

  const previewItem = previewQuery.data?.items.find((i) => i.customer_id === customerId);

  useEffect(() => {
    if (!open) return;
    if (previewItem && previewItem.monthly_amount > 0) {
      setAmount(String(previewItem.monthly_amount));
    }
  }, [open, previewItem]);

  const handleGenerate = () => {
    const parsedAmount = amount.trim() ? Number(amount) : undefined;
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
    (previewItem != null &&
      !previewItem.eligible &&
      previewItem.skip_reason === 'no_active_lines') ||
    (previewItem != null &&
      !previewItem.eligible &&
      previewItem.skip_reason === 'no_monthly_amount' &&
      !amount.trim());

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>Gerar fatura</SheetTitle>
          <SheetDescription>
            Cria fatura com boleto Sicredi (código de barras + QR PIX) para {customerName}.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-4 px-4">
          {previewQuery.isPending && (
            <p className="text-muted-foreground flex items-center gap-2 text-sm">
              <Loader2 className="size-4 animate-spin" />
              Calculando valor sugerido…
            </p>
          )}
          {previewItem && !previewItem.eligible && (
            <p className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              {previewItem.skip_reason === 'no_monthly_amount'
                ? 'Vincule linhas/aparelhos com valor mensal ou informe o valor abaixo.'
                : previewItem.skip_reason === 'no_active_lines'
                  ? 'Vincule ao menos uma linha ou aparelho ao cliente.'
                  : 'Cliente não está pronto para faturamento.'}
            </p>
          )}
          {previewItem && (
            <p className="text-muted-foreground text-sm">
              {previewItem.monthly_amount > 0
                ? `Valor sugerido: ${formatMoney(previewItem.monthly_amount)} (${previewItem.line_count} linha(s), ${previewItem.device_count ?? 0} aparelho(s))`
                : 'Informe o valor da fatura abaixo.'}
              {!previewItem.billing_email && (
                <span className="mt-1 block text-amber-700">
                  E-mail não cadastrado — você pode gerar e baixar a fatura mesmo assim.
                </span>
              )}
            </p>
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
        </div>
        <SheetFooter>
          <Button onClick={handleGenerate} disabled={generateDisabled}>
            <FileStack className="mr-2 size-4" />
            {generateMutation.isPending ? 'Gerando…' : 'Gerar fatura + boleto'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
