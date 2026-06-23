import { Link } from '@tanstack/react-router';
import { Wallet } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  formatFinancialStatus,
  formatMoney,
  useCreatePayableFromInvoice
} from '@/lib/financial-api';

type InvoiceFinancialActionsProps = {
  invoiceId: string;
  totalAmount: number;
  accountPayableId?: string | null;
  accountPayableStatus?: string | null;
  canManageFinance?: boolean;
};

export function InvoiceFinancialActions({
  invoiceId,
  totalAmount,
  accountPayableId,
  accountPayableStatus,
  canManageFinance = true
}: InvoiceFinancialActionsProps) {
  const createPayable = useCreatePayableFromInvoice();

  if (!canManageFinance) {
    return null;
  }

  if (accountPayableId) {
    return (
      <div className="flex flex-wrap items-center gap-3">
        <span className="text-muted-foreground text-sm">
          Conta a pagar vinculada
          {accountPayableStatus
            ? ` · ${formatFinancialStatus(accountPayableStatus)}`
            : ''}
        </span>
        <Button
          nativeButton={false}
          variant="outline"
          size="sm"
          render={
            <Link
              to="/finance/payables"
              search={{ page: 1, pageSize: 10 }}
            />
          }
        >
          <Wallet className="size-4" />
          Ver no financeiro
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-3">
      <span className="text-muted-foreground text-sm">
        Valor da fatura: {formatMoney(totalAmount)} — gere a conta a pagar para
        refletir no financeiro.
      </span>
      <Button
        size="sm"
        disabled={createPayable.isPending}
        onClick={() =>
          createPayable.mutate(invoiceId, {
            onSuccess: () =>
              toast.success('Conta a pagar criada a partir da fatura.'),
            onError: (e) =>
              toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
          })
        }
      >
        <Wallet className="size-4" />
        Gerar conta a pagar
      </Button>
    </div>
  );
}
