import { useState } from 'react';

import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { usePartnerCustomers } from '@/lib/partner-api';
import {
  type CreateSaleLineItemInput,
  formatLineItemType,
  formatMoney,
  usePartnerConfirmSale,
  usePartnerContractTemplates,
  usePartnerCreateSale
} from '@/lib/sales-api';

export const Route = createFileRoute('/__app/partner/commercial-sales/new/')({
  component: PartnerNewSalePage
});

type DraftItem = CreateSaleLineItemInput & { key: string };
const NONE = '__none__';

const emptyItem = (): DraftItem => ({
  key: crypto.randomUUID(),
  line_item_type: 'other',
  description: '',
  quantity: 1,
  unit_price: 0
});

function PartnerNewSalePage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [customerId, setCustomerId] = useState('');
  const [templateId, setTemplateId] = useState(NONE);
  const [notes, setNotes] = useState('');
  const [items, setItems] = useState<DraftItem[]>([emptyItem()]);

  const customersQuery = usePartnerCustomers({ page_index: 0, page_size: 500 });
  const templatesQuery = usePartnerContractTemplates();
  const createMutation = usePartnerCreateSale();
  const confirmMutation = usePartnerConfirmSale();

  const total = items.reduce((sum, i) => sum + i.quantity * i.unit_price, 0);

  const updateItem = (key: string, patch: Partial<DraftItem>) => {
    setItems((prev) => prev.map((i) => (i.key === key ? { ...i, ...patch } : i)));
  };

  const handleFinish = async (confirm: boolean) => {
    if (!customerId) {
      toast.error('Selecione o cliente.');
      return;
    }
    const validItems = items.filter((i) => i.description.trim());
    if (validItems.length === 0) {
      toast.error('Adicione ao menos um item.');
      return;
    }
    try {
      const sale = await createMutation.mutateAsync({
        customer_id: customerId,
        contract_template_id: templateId !== NONE ? templateId : undefined,
        notes: notes.trim() || undefined,
        items: validItems.map(({ key: _k, ...rest }) => rest)
      });
      if (confirm) {
        const confirmed = await confirmMutation.mutateAsync(sale.id);
        toast.success('Venda confirmada.');
        void navigate({ to: '/partner/commercial-sales/$saleId', params: { saleId: confirmed.id } });
      } else {
        toast.success('Rascunho salvo.');
        void navigate({ to: '/partner/commercial-sales/$saleId', params: { saleId: sale.id } });
      }
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Parceiro', to: '/partner' },
        { label: 'Vendas', to: '/partner/commercial-sales' },
        { label: 'Nova venda' }
      ]}
    >
      <div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
        <h1 className="text-2xl font-semibold">Nova venda</h1>

        {step === 1 && (
          <Card>
            <CardHeader>
              <CardTitle>Cliente e contrato</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Cliente</Label>
                <Select value={customerId} onValueChange={(v) => setCustomerId(v ?? '')}>
                  <SelectTrigger>
                    <SelectValue placeholder="Selecione" />
                  </SelectTrigger>
                  <SelectContent>
                    {(customersQuery.data?.items ?? []).map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Contrato</Label>
                <Select value={templateId} onValueChange={(v) => setTemplateId(v ?? NONE)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={NONE}>Sem contrato</SelectItem>
                    {(templatesQuery.data?.items ?? []).map((t) => (
                      <SelectItem key={t.id} value={t.id}>
                        {t.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Observações" />
              <Button onClick={() => setStep(2)} disabled={!customerId}>
                Próximo
              </Button>
            </CardContent>
          </Card>
        )}

        {step === 2 && (
          <Card>
            <CardHeader>
              <CardTitle>Itens</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {items.map((item) => (
                <div key={item.key} className="space-y-2 rounded border p-3">
                  <Select
                    value={item.line_item_type}
                    onValueChange={(v) =>
                      updateItem(item.key, { line_item_type: v as DraftItem['line_item_type'] })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="phone_line">Linha</SelectItem>
                      <SelectItem value="device">Aparelho</SelectItem>
                      <SelectItem value="other">Outro</SelectItem>
                    </SelectContent>
                  </Select>
                  <Input
                    placeholder="Descrição"
                    value={item.description}
                    onChange={(e) => updateItem(item.key, { description: e.target.value })}
                  />
                  <div className="grid grid-cols-2 gap-2">
                    <Input
                      type="number"
                      min={1}
                      value={item.quantity}
                      onChange={(e) => updateItem(item.key, { quantity: Number(e.target.value) || 1 })}
                    />
                    <Input
                      type="number"
                      min={0}
                      step={0.01}
                      value={item.unit_price}
                      onChange={(e) => updateItem(item.key, { unit_price: Number(e.target.value) || 0 })}
                    />
                  </div>
                </div>
              ))}
              <Button variant="outline" onClick={() => setItems((p) => [...p, emptyItem()])}>
                + Item
              </Button>
              <p className="font-medium">Total: {formatMoney(total)}</p>
              <div className="flex gap-2">
                <Button variant="outline" onClick={() => setStep(1)}>
                  Voltar
                </Button>
                <Button onClick={() => setStep(3)}>Revisar</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {step === 3 && (
          <Card>
            <CardHeader>
              <CardTitle>Confirmar</CardTitle>
              <CardDescription>
                {items.filter((i) => i.description.trim()).map((i) => (
                  <div key={i.key}>
                    {formatLineItemType(i.line_item_type)} — {i.description}
                  </div>
                ))}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex gap-2">
              <Button variant="outline" onClick={() => setStep(2)}>
                Voltar
              </Button>
              <Button variant="secondary" onClick={() => void handleFinish(false)}>
                Rascunho
              </Button>
              <Button onClick={() => void handleFinish(true)}>Finalizar</Button>
            </CardContent>
          </Card>
        )}
      </div>
    </PageWrapper>
  );
}
