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
import { useGetV1Customers } from '@/api';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  type CreateSaleLineItemInput,
  formatLineItemType,
  formatMoney,
  useConfirmSale,
  useContractTemplates,
  useCreateSale
} from '@/lib/sales-api';

const NONE_TEMPLATE = '__none__';

export const Route = createFileRoute('/__app/sales/new/')({
  component: NewSalePage
});

type DraftItem = CreateSaleLineItemInput & { key: string };

const emptyItem = (): DraftItem => ({
  key: crypto.randomUUID(),
  line_item_type: 'other',
  description: '',
  quantity: 1,
  unit_price: 0
});

function NewSalePage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [customerId, setCustomerId] = useState('');
  const [templateId, setTemplateId] = useState(NONE_TEMPLATE);
  const [notes, setNotes] = useState('');
  const [items, setItems] = useState<DraftItem[]>([emptyItem()]);

  const customersQuery = useGetV1Customers({ page_index: 0, page_size: 500 });
  const templatesQuery = useContractTemplates(true);
  const createMutation = useCreateSale();
  const confirmMutation = useConfirmSale();

  const total = items.reduce((sum, i) => sum + i.quantity * i.unit_price, 0);

  const updateItem = (key: string, patch: Partial<DraftItem>) => {
    setItems((prev) => prev.map((i) => (i.key === key ? { ...i, ...patch } : i)));
  };

  const handleFinish = async (confirm: boolean) => {
    if (!customerId) {
      toast.error('Selecione o cliente.');
      return;
    }
    const validItems = items.filter((i) => i.description.trim() && i.unit_price >= 0);
    if (validItems.length === 0) {
      toast.error('Adicione ao menos um item à venda.');
      return;
    }

    try {
      const sale = await createMutation.mutateAsync({
        customer_id: customerId,
        contract_template_id: templateId !== NONE_TEMPLATE ? templateId : undefined,
        notes: notes.trim() || undefined,
        items: validItems.map(({ key: _k, ...rest }) => rest)
      });

      if (confirm) {
        const confirmed = await confirmMutation.mutateAsync(sale.id);
        toast.success('Venda confirmada e contrato gerado.');
        void navigate({ to: '/sales/$saleId', params: { saleId: confirmed.id } });
      } else {
        toast.success('Rascunho de venda salvo.');
        void navigate({ to: '/sales/$saleId', params: { saleId: sale.id } });
      }
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Vendas', to: '/sales' },
        { label: 'Nova venda' }
      ]}
    >
      <div className="mx-auto flex max-w-3xl flex-col gap-6">
        <div>
          <h1 className="text-2xl font-semibold">Nova venda</h1>
          <p className="text-muted-foreground text-sm">Passo {step} de 3</p>
        </div>

        {step === 1 && (
          <Card>
            <CardHeader>
              <CardTitle>Cliente e contrato</CardTitle>
              <CardDescription>Selecione o cliente e o modelo de contrato para preenchimento automático.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Cliente</Label>
                <Select value={customerId} onValueChange={(v) => setCustomerId(v ?? '')}>
                  <SelectTrigger>
                    <SelectValue placeholder="Selecione o cliente" />
                  </SelectTrigger>
                  <SelectContent>
                    {(customersQuery.data?.items ?? []).map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name} — {c.cpf_cnpj}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Modelo de contrato (opcional)</Label>
                <Select value={templateId} onValueChange={(v) => setTemplateId(v ?? NONE_TEMPLATE)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Sem contrato automático" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={NONE_TEMPLATE}>Sem contrato</SelectItem>
                    {(templatesQuery.data?.items ?? []).map((t) => (
                      <SelectItem key={t.id} value={t.id}>
                        {t.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Observações</Label>
                <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={3} />
              </div>
              <Button onClick={() => setStep(2)} disabled={!customerId}>
                Próximo: itens
              </Button>
            </CardContent>
          </Card>
        )}

        {step === 2 && (
          <Card>
            <CardHeader>
              <CardTitle>Itens da venda</CardTitle>
              <CardDescription>Linhas telefônicas, aparelhos ou outros produtos/serviços.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {items.map((item) => (
                <div key={item.key} className="grid gap-3 rounded-lg border p-4 md:grid-cols-2">
                  <div className="space-y-2 md:col-span-2">
                    <Label>Tipo</Label>
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
                        <SelectItem value="phone_line">Linha telefônica</SelectItem>
                        <SelectItem value="device">Aparelho</SelectItem>
                        <SelectItem value="other">Outro</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2 md:col-span-2">
                    <Label>Descrição</Label>
                    <Input
                      value={item.description}
                      onChange={(e) => updateItem(item.key, { description: e.target.value })}
                      placeholder="Ex.: Linha móvel 11 99999-9999"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Quantidade</Label>
                    <Input
                      type="number"
                      min={1}
                      step={1}
                      value={item.quantity}
                      onChange={(e) => updateItem(item.key, { quantity: Number(e.target.value) || 1 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Preço unitário</Label>
                    <Input
                      type="number"
                      min={0}
                      step={0.01}
                      value={item.unit_price}
                      onChange={(e) => updateItem(item.key, { unit_price: Number(e.target.value) || 0 })}
                    />
                  </div>
                  {items.length > 1 && (
                    <div className="md:col-span-2">
                      <Button variant="outline" size="sm" onClick={() => setItems((p) => p.filter((i) => i.key !== item.key))}>
                        Remover item
                      </Button>
                    </div>
                  )}
                </div>
              ))}
              <Button variant="outline" onClick={() => setItems((p) => [...p, emptyItem()])}>
                Adicionar item
              </Button>
              <p className="text-sm font-medium">Total: {formatMoney(total)}</p>
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
              <CardTitle>Revisão</CardTitle>
              <CardDescription>Confirme os dados antes de fechar a venda.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <ul className="space-y-1 text-sm">
                {items
                  .filter((i) => i.description.trim())
                  .map((i) => (
                    <li key={i.key}>
                      {formatLineItemType(i.line_item_type)} — {i.description} × {i.quantity} ={' '}
                      {formatMoney(i.quantity * i.unit_price)}
                    </li>
                  ))}
              </ul>
              <p className="font-semibold">Total: {formatMoney(total)}</p>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" onClick={() => setStep(2)}>
                  Voltar
                </Button>
                <Button
                  variant="secondary"
                  disabled={createMutation.isPending}
                  onClick={() => void handleFinish(false)}
                >
                  Salvar rascunho
                </Button>
                <Button
                  disabled={createMutation.isPending || confirmMutation.isPending}
                  onClick={() => void handleFinish(true)}
                >
                  Finalizar venda
                </Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </PageWrapper>
  );
}
