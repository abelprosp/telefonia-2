import { useState } from 'react';

import { Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

import type { GetProviderPlanServiceResponse } from '@/api';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Field, FieldLabel } from '@/components/ui/field';
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  useCreateProviderPlanService,
  useDeleteProviderPlanService
} from '@/lib/providers-api';

function formatBrl(value: number | string | null | undefined) {
  if (value === null || value === undefined) {
    return '—';
  }
  const n = typeof value === 'string' ? Number(value) : value;
  if (!Number.isFinite(n)) {
    return '—';
  }
  return n.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

function formatServiceType(type?: string | null) {
  switch (type?.toLowerCase()) {
    case 'subscription':
      return 'Assinatura';
    case 'data':
      return 'Dados / Internet';
    case 'sms':
      return 'SMS';
    case 'roaming':
      return 'Roaming';
    case 'other':
    default:
      return 'Outros';
  }
}

function formatApplicationType(app?: string | null) {
  switch (app?.toLowerCase()) {
    case 'luxus_customer':
      return 'Luxus → Cliente';
    case 'customer_end_user':
      return 'Cliente → Usuário Final';
    case 'both':
    default:
      return 'Ambas as Perspectivas';
  }
}

function formatAvailabilityRule(rule?: string | null) {
  switch (rule?.toLowerCase()) {
    case 'customer_exclusive':
      return 'Exclusivo por Cliente';
    case 'global':
    default:
      return 'Global';
  }
}

type Props = {
  providerId: string;
  planId: string;
  services: GetProviderPlanServiceResponse[];
};

export function ProviderPlanServices({ providerId, planId, services }: Props) {
  const [sheetOpen, setSheetOpen] = useState(false);
  const [name, setName] = useState('');
  const [invoiceName, setInvoiceName] = useState('');
  const [serviceType, setServiceType] = useState('subscription');
  const [price, setPrice] = useState('');
  const [recurring, setRecurring] = useState(true);
  const [applicationType, setApplicationType] = useState('both');
  const [availabilityRule, setAvailabilityRule] = useState('global');

  const createMutation = useCreateProviderPlanService(providerId, planId);
  const deleteMutation = useDeleteProviderPlanService(providerId, planId);

  const resetForm = () => {
    setName('');
    setInvoiceName('');
    setServiceType('subscription');
    setPrice('');
    setRecurring(true);
    setApplicationType('both');
    setAvailabilityRule('global');
  };

  const handleCreate = () => {
    if (!name.trim()) {
      toast.error('Informe o nome do serviço.');
      return;
    }
    let parsedPrice: number | null = null;
    if (price.trim()) {
      const normalized = price.trim().replace(/\./g, '').replace(',', '.');
      const n = Number(normalized);
      if (Number.isFinite(n) && n >= 0) {
        parsedPrice = n;
      } else {
        toast.error('Informe um valor de preço válido.');
        return;
      }
    }

    createMutation.mutate(
      {
        name: name.trim(),
        invoice_name: invoiceName.trim() || null,
        service_type: serviceType,
        recurring,
        price: parsedPrice,
        application_type: applicationType,
        availability_rule: availabilityRule
      },
      {
        onSuccess: () => {
          toast.success('Serviço adicionado ao portfólio do plano.');
          resetForm();
          setSheetOpen(false);
        },
        onError: (err) => {
          toast.error(
            isApiHttpError(err) ? err.message : getErrorMessage(err)
          );
        }
      }
    );
  };

  const handleDelete = (serviceId: string, svcName: string) => {
    if (!confirm(`Deseja desativar o serviço "${svcName}" deste plano?`)) {
      return;
    }
    deleteMutation.mutate(serviceId, {
      onSuccess: () => {
        toast.success('Serviço desativado.');
      },
      onError: (err) => {
        toast.error(
          isApiHttpError(err) ? err.message : getErrorMessage(err)
        );
      }
    });
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Portfólio de Serviços do Plano (§6.1)
        </span>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => setSheetOpen(true)}
        >
          <Plus className="mr-1.5 size-3.5" />
          Adicionar Serviço
        </Button>
      </div>

      {services.length === 0 ? (
        <p className="text-muted-foreground py-2 text-sm">
          Nenhum serviço cadastrado neste plano.
        </p>
      ) : (
        <div className="overflow-x-auto rounded-md border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome em Sistema</TableHead>
                <TableHead>Nome em Fatura</TableHead>
                <TableHead>Tipo</TableHead>
                <TableHead>Aplicação</TableHead>
                <TableHead>Disponibilidade</TableHead>
                <TableHead>Preço Padrão</TableHead>
                <TableHead>Recorrente</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-16 text-right">Ação</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {services.map((row) => (
                <TableRow key={row.id}>
                  <TableCell className="font-medium">{row.name}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {row.invoice_name ?? '—'}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" size="sm">
                      {formatServiceType(row.service_type)}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs">
                    {formatApplicationType(row.application_type)}
                  </TableCell>
                  <TableCell className="text-xs">
                    {formatAvailabilityRule(row.availability_rule)}
                  </TableCell>
                  <TableCell>{formatBrl(row.price)}</TableCell>
                  <TableCell>{row.recurring ? 'Sim' : 'Não'}</TableCell>
                  <TableCell>
                    <Badge
                      variant={row.active ? 'success-light' : 'secondary'}
                      size="sm"
                    >
                      {row.active ? 'Ativo' : 'Inativo'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    {row.active ? (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="text-muted-foreground hover:text-destructive size-8 p-0"
                        onClick={() => handleDelete(row.id, row.name)}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    ) : null}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent className="flex flex-col gap-6 sm:max-w-lg">
          <SheetHeader>
            <SheetTitle>Novo Serviço do Portfólio</SheetTitle>
            <SheetDescription>
              Cadastre um serviço com atributos de conciliação de fatura, aplicação e disponibilidade (§6.1).
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-col gap-4">
            <Field>
              <FieldLabel>Nome do Serviço (em sistema)</FieldLabel>
              <Input
                placeholder="Ex: Assinatura Smart Empresas 10GB"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </Field>

            <Field>
              <FieldLabel>Nome em Fatura (opcional para conciliação)</FieldLabel>
              <Input
                placeholder="Ex: PACOTE SMART EMPRESAS"
                value={invoiceName}
                onChange={(e) => setInvoiceName(e.target.value)}
              />
            </Field>

            <div className="grid grid-cols-2 gap-3">
              <Field>
                <FieldLabel>Tipo de Serviço</FieldLabel>
                <Select value={serviceType} onValueChange={(v) => v && setServiceType(v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="subscription">Assinatura</SelectItem>
                    <SelectItem value="data">Dados / Internet</SelectItem>
                    <SelectItem value="sms">SMS</SelectItem>
                    <SelectItem value="roaming">Roaming</SelectItem>
                    <SelectItem value="other">Outros</SelectItem>
                  </SelectContent>
                </Select>
              </Field>

              <Field>
                <FieldLabel>Preço Padrão (R$)</FieldLabel>
                <Input
                  placeholder="0,00"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                />
              </Field>
            </div>

            <Field>
              <FieldLabel>Definição de Aplicação</FieldLabel>
              <Select value={applicationType} onValueChange={(v) => v && setApplicationType(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="both">Ambas (Luxus→Cliente e Cliente→Usuário)</SelectItem>
                  <SelectItem value="luxus_customer">Apenas Luxus → Cliente</SelectItem>
                  <SelectItem value="customer_end_user">Apenas Cliente → Usuário Final</SelectItem>
                </SelectContent>
              </Select>
            </Field>

            <Field>
              <FieldLabel>Regra de Disponibilidade</FieldLabel>
              <Select value={availabilityRule} onValueChange={(v) => v && setAvailabilityRule(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="global">Global (disponível para todos os clientes)</SelectItem>
                  <SelectItem value="customer_exclusive">Exclusivo por cliente</SelectItem>
                </SelectContent>
              </Select>
            </Field>

            <Field>
              <FieldLabel>Cobrança Recorrente</FieldLabel>
              <Select
                value={recurring ? 'true' : 'false'}
                onValueChange={(v) => setRecurring(v === 'true')}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">Sim (mensal)</SelectItem>
                  <SelectItem value="false">Não (avulso / pontual)</SelectItem>
                </SelectContent>
              </Select>
            </Field>
          </div>

          <SheetFooter className="mt-auto flex justify-end gap-2">
            <SheetClose render={<Button variant="outline" type="button">Cancelar</Button>} />
            <Button
              type="button"
              disabled={createMutation.isPending}
              onClick={handleCreate}
            >
              {createMutation.isPending ? 'Salvando...' : 'Cadastrar Serviço'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
