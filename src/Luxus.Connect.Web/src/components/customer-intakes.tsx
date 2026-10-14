import { useState } from 'react';

import { Link } from '@tanstack/react-router';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { getErrorMessage } from '@/lib/api-error';
import { client } from '@/lib/client';
import {
  useCustomerIntakes,
  type CustomerIntake,
  type ImportedSale
} from '@/lib/customer-intake-api';
import { formatMoney } from '@/lib/financial-api';

const fields: [keyof ImportedSale, string][] = [
  ['id', 'ID na origem'],
  ['protocol', 'Protocolo da venda'],
  ['taskProtocol', 'Protocolo da solicitação'],
  ['soldAt', 'Data da venda'],
  ['concludedAt', 'Conclusão na origem'],
  ['value', 'Valor da venda'],
  ['commissionValue', 'Comissão em reais'],
  ['operatorName', 'Operadora'],
  ['planName', 'Plano'],
  ['lineNumber', 'Número da linha'],
  ['simType', 'Tipo de chip'],
  ['isVirginChip', 'Chip virgem'],
  ['iccid', 'ICCID'],
  ['deviceImei', 'IMEI'],
  ['deviceEid', 'EID'],
  ['devicePlatform', 'Plataforma do aparelho'],
  ['isPortability', 'Portabilidade'],
  ['portabilityNumber', 'Número a portar'],
  ['donorOperator', 'Operadora doadora'],
  ['isUpgrade', 'Upgrade'],
  ['upgradeOfProtocol', 'Protocolo da venda anterior'],
  ['contractFormat', 'Formato do contrato'],
  ['partnerName', 'Parceiro'],
  ['branchName', 'Filial'],
  ['notes', 'Observações da venda']
];

function display(sale: ImportedSale, key: keyof ImportedSale) {
  const value = sale[key];
  if (value === null || value === undefined || value === '') return '—';
  if (typeof value === 'boolean') return value ? 'Sim' : 'Não';
  if (key === 'value' || key === 'commissionValue')
    return formatMoney(Number(value));
  if (key === 'soldAt' || key === 'concludedAt')
    return new Date(String(value)).toLocaleString('pt-BR');
  return String(value);
}

export function ImportedSaleDetails({
  intake,
  showSaleLink = false
}: {
  intake: CustomerIntake;
  showSaleLink?: boolean;
}) {
  const [downloading, setDownloading] = useState<string | null>(null);
  const download = async (attachmentId: string) => {
    setDownloading(attachmentId);
    try {
      const { data } = await client<{ url: string }>({
        url: `/v1/customers/${intake.customer_id}/attachments/${attachmentId}/download`,
        method: 'GET'
      });
      const anchor = document.createElement('a');
      anchor.href = data.url;
      anchor.target = '_blank';
      anchor.rel = 'noopener noreferrer';
      anchor.click();
    } catch (error) {
      toast.error(getErrorMessage(error));
    } finally {
      setDownloading(null);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Venda importada · {intake.sale.protocol}</CardTitle>
        <p className="text-muted-foreground text-sm">
          Origem: {intake.source_system} · Recebida em{' '}
          {new Date(intake.imported_at).toLocaleString('pt-BR')}
        </p>
        {showSaleLink && (
          <Button
            nativeButton={false}
            variant="outline"
            className="w-fit"
            render={
              <Link to="/sales/$saleId" params={{ saleId: intake.sale_id }} />
            }
          >
            Abrir venda
          </Button>
        )}
      </CardHeader>
      <CardContent className="space-y-5">
        <p className="text-muted-foreground text-sm">
          Dados comerciais informados pela origem. A importação não confirma
          pagamento, assinatura nem ativação da linha.
        </p>
        <dl className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {fields.map(([key, label]) => (
            <div key={key} className="min-w-0">
              <dt className="text-muted-foreground text-xs">{label}</dt>
              <dd className="mt-1 text-sm break-words whitespace-pre-wrap">
                {display(intake.sale, key)}
              </dd>
            </div>
          ))}
        </dl>
        <details className="rounded-md border p-3">
          <summary className="cursor-pointer text-sm font-medium">
            Dados do cliente recebidos nesta venda
          </summary>
          <dl className="mt-3 grid gap-3 text-sm sm:grid-cols-2">
            <div>
              <dt className="text-muted-foreground">Nome / razão social</dt>
              <dd>{intake.client.legalName || intake.client.name}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Documento</dt>
              <dd>
                {intake.client.documentType}: {intake.client.document}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">
                Identificação na origem / RG
              </dt>
              <dd>
                {intake.client.id || '—'} / {intake.client.rg || '—'}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Contato</dt>
              <dd>
                {intake.client.email || '—'} · {intake.client.phone || '—'}
              </dd>
            </div>
            <div className="sm:col-span-2">
              <dt className="text-muted-foreground">Endereço recebido</dt>
              <dd>
                {[
                  intake.client.address,
                  intake.client.addressNumber,
                  intake.client.complement,
                  intake.client.neighborhood,
                  intake.client.city,
                  intake.client.state,
                  intake.client.zipCode
                ]
                  .filter(Boolean)
                  .join(', ') || '—'}
              </dd>
            </div>
          </dl>
        </details>
        {intake.attachments.length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium">Documentos recebidos</p>
            {intake.attachments.map((a) => (
              <div
                key={a.attachment_id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-md border p-3 text-sm"
              >
                <span className="min-w-0 break-all">
                  {a.name} · {a.type} · {(a.size_bytes / 1024).toFixed(1)} KB
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={downloading !== null}
                  onClick={() => void download(a.attachment_id)}
                >
                  {downloading === a.attachment_id ? 'Preparando…' : 'Baixar'}
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function CustomerIntakes({ customerId }: { customerId: string }) {
  const query = useCustomerIntakes(customerId);
  if (query.isPending)
    return (
      <p className="text-muted-foreground text-sm" role="status">
        Carregando vendas importadas…
      </p>
    );
  if (query.isError)
    return (
      <div role="alert" className="space-y-2 text-sm">
        <p>{getErrorMessage(query.error)}</p>
        <Button variant="outline" onClick={() => void query.refetch()}>
          Tentar novamente
        </Button>
      </div>
    );
  if (!query.data.length)
    return (
      <p className="text-muted-foreground text-sm">
        Nenhuma venda recebida de outro sistema.
      </p>
    );
  return (
    <div className="space-y-4">
      <p className="text-muted-foreground text-sm">
        Últimas 100 vendas importadas deste cliente.
      </p>
      {query.data.map((intake) => (
        <ImportedSaleDetails key={intake.id} intake={intake} showSaleLink />
      ))}
    </div>
  );
}
