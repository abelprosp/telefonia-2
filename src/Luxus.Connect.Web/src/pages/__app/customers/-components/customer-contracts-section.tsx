import { useState } from 'react';

import {
  CheckCircle2,
  Clock,
  ExternalLink,
  FileCheck,
  FileDown,
  FileText,
  FileUp,
  Plus,
  RefreshCw,
  Send,
  UploadCloud,
  Users,
  Zap
} from 'lucide-react';
import { toast } from 'sonner';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Sheet,
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
  type ContractSignerInput,
  type GeneratedContract,
  useContractTemplates,
  useCustomerGeneratedContracts,
  useGenerateCustomerContract,
  useSyncZapSignContract,
  useUploadSignedContract
} from '@/lib/sales-api';

type Props = {
  customerId: string;
  customerName: string;
  isActive: boolean;
};

export function CustomerContractsSection({
  customerId,
  customerName,
  isActive
}: Props) {
  const contractsQuery = useCustomerGeneratedContracts(customerId);
  const templatesQuery = useContractTemplates(true);
  const generateContractMutation = useGenerateCustomerContract();
  const uploadSignedMutation = useUploadSignedContract();
  const syncZapSignMutation = useSyncZapSignContract();

  const [emitModalOpen, setEmitModalOpen] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState('');
  const [signatureMethod, setSignatureMethod] = useState<'zapsign' | 'manual'>('zapsign');
  const [signers, setSigners] = useState<ContractSignerInput[]>([
    {
      name: customerName,
      email: '',
      phone: '',
      auth_mode: 'whatsapp'
    }
  ]);

  // Modal de anexo manual
  const [attachModalOpen, setAttachModalOpen] = useState(false);
  const [attachContractId, setAttachContractId] = useState<string | null>(null);
  const [signedPdfUrl, setSignedPdfUrl] = useState('');
  const [signedFileName, setSignedFileName] = useState('');
  const [signedBy, setSignedBy] = useState('');

  const handleOpenEmitModal = () => {
    const firstActive = templatesQuery.data?.items?.[0]?.id ?? '';
    setSelectedTemplateId(firstActive);
    setSignatureMethod('zapsign');
    setSigners([
      {
        name: customerName,
        email: '',
        phone: '',
        auth_mode: 'whatsapp'
      }
    ]);
    setEmitModalOpen(true);
  };

  const handleEmitContract = () => {
    if (!selectedTemplateId) {
      toast.error('Selecione um template de contrato.');
      return;
    }

    generateContractMutation.mutate(
      {
        customerId,
        contract_template_id: selectedTemplateId,
        signature_method: signatureMethod,
        signers: signatureMethod === 'zapsign' ? signers : undefined
      },
      {
        onSuccess: (data) => {
          if (signatureMethod === 'zapsign' && data.zapsign_sign_url) {
            toast.success(
              'Contrato emitido e enviado para assinatura via ZapSign!'
            );
          } else {
            toast.success(
              'Contrato gerado com sucesso para impressão manual.'
            );
          }
          setEmitModalOpen(false);
          void contractsQuery.refetch();
        },
        onError: (e) =>
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

  const handleOpenAttachModal = (contract: GeneratedContract) => {
    setAttachContractId(contract.id);
    setSignedPdfUrl('');
    setSignedFileName('');
    setSignedBy(customerName);
    setAttachModalOpen(true);
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.type !== 'application/pdf') {
      toast.error('O arquivo precisa ser em formato PDF.');
      return;
    }
    setSignedFileName(file.name);
    const reader = new FileReader();
    reader.onload = () => {
      setSignedPdfUrl(reader.result as string);
      toast.success('Arquivo PDF carregado.');
    };
    reader.readAsDataURL(file);
  };

  const handleSaveAttached = () => {
    if (!attachContractId || !signedPdfUrl || !signedBy.trim()) {
      toast.error('Selecione o arquivo PDF assinado e informe quem assinou.');
      return;
    }

    uploadSignedMutation.mutate(
      {
        customerId,
        contractId: attachContractId,
        signed_pdf_url: signedPdfUrl,
        signed_by: signedBy
      },
      {
        onSuccess: () => {
          toast.success('Contrato assinado anexado com sucesso ao cliente.');
          setAttachModalOpen(false);
          void contractsQuery.refetch();
        },
        onError: (e) =>
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

  const handleSyncZapSign = (contractId: string) => {
    syncZapSignMutation.mutate(
      { contractId, customerId },
      {
        onSuccess: (data) => {
          if (data.zapsign_status === 'signed') {
            toast.success('Contrato assinado pelo cliente via ZapSign!');
          } else {
            toast.info(`Status no ZapSign: ${data.zapsign_status ?? 'pendente'}`);
          }
          void contractsQuery.refetch();
        },
        onError: (e) =>
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

  const contracts = contractsQuery.data ?? [];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h4 className="text-base font-bold text-foreground">
            Contratos e Termos do Cliente
          </h4>
          <p className="text-muted-foreground text-xs">
            Emita contratos com envio automático via ZapSign (WhatsApp/E-mail) ou imprima para assinatura física e anexo.
          </p>
        </div>

        {isActive ? (
          <Button onClick={handleOpenEmitModal} className="gap-1.5 shadow-xs">
            <Plus className="size-4" />
            Emitir Novo Contrato
          </Button>
        ) : null}
      </div>

      {contractsQuery.isPending ? (
        <div className="rounded-2xl border border-dashed p-8 text-center text-sm text-muted-foreground">
          Carregando contratos do cliente...
        </div>
      ) : contracts.length === 0 ? (
        <div className="rounded-2xl border border-dashed bg-muted/20 p-8 text-center">
          <FileText className="mx-auto size-8 text-muted-foreground" />
          <p className="mt-2 text-sm font-semibold text-foreground">
            Nenhum contrato emitido para este cliente
          </p>
          <p className="text-muted-foreground text-xs">
            Gere contratos vinculados a este cliente utilizando os templates cadastrados.
          </p>
          {isActive ? (
            <Button
              variant="outline"
              size="sm"
              onClick={handleOpenEmitModal}
              className="mt-4 gap-1.5"
            >
              <Plus className="size-3.5" />
              Emitir Primeiro Contrato
            </Button>
          ) : null}
        </div>
      ) : (
        <div className="overflow-x-auto rounded-2xl border bg-card shadow-xs">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Template / Documento</TableHead>
                <TableHead>Método</TableHead>
                <TableHead>Status da Assinatura</TableHead>
                <TableHead>Emissão</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {contracts.map((contract) => {
                const isSigned =
                  Boolean(contract.signed_pdf_url) ||
                  contract.zapsign_status === 'signed';

                return (
                  <TableRow key={contract.id}>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-semibold text-foreground">
                          {contract.contract_template_name ?? 'Contrato de Telefonia'}
                        </span>
                        <span className="text-muted-foreground text-xs font-mono">
                          ID: {contract.id.slice(0, 8)}…
                        </span>
                      </div>
                    </TableCell>

                    <TableCell>
                      {contract.signature_method === 'zapsign' ? (
                        <Badge
                          variant="outline"
                          className="gap-1 border-emerald-300 text-emerald-700 dark:text-emerald-400"
                        >
                          <Zap className="size-3 fill-current" /> ZapSign
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="gap-1">
                          <FileText className="size-3" /> Manual
                        </Badge>
                      )}
                    </TableCell>

                    <TableCell>
                      {isSigned ? (
                        <div className="flex flex-col gap-0.5">
                          <Badge className="w-fit gap-1 bg-emerald-600 hover:bg-emerald-700">
                            <CheckCircle2 className="size-3" /> Assinado
                          </Badge>
                          {contract.signed_by ? (
                            <span className="text-muted-foreground text-[11px]">
                              Por: {contract.signed_by}
                            </span>
                          ) : null}
                        </div>
                      ) : contract.signature_method === 'zapsign' ? (
                        <div className="flex flex-col gap-0.5">
                          <Badge variant="outline" className="w-fit gap-1 border-amber-300 text-amber-700 dark:text-amber-400">
                            <Clock className="size-3" /> Aguardando ZapSign
                          </Badge>
                          {contract.zapsign_sign_url ? (
                            <a
                              href={contract.zapsign_sign_url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-primary hover:underline flex items-center gap-1 text-[11px] font-medium"
                            >
                              Link de assinatura <ExternalLink className="size-2.5" />
                            </a>
                          ) : null}
                        </div>
                      ) : (
                        <Badge variant="outline" className="w-fit gap-1">
                          <Clock className="size-3" /> Aguardando Impressão / Anexo
                        </Badge>
                      )}
                    </TableCell>

                    <TableCell className="text-xs text-muted-foreground">
                      {new Date(contract.created_at).toLocaleDateString('pt-BR')}
                    </TableCell>

                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        {/* Se assinado, permite visualizar o PDF assinado */}
                        {isSigned && contract.signed_pdf_url ? (
                          <a
                            href={contract.signed_pdf_url}
                            target="_blank"
                            rel="noreferrer"
                            className="inline-flex items-center gap-1 rounded-lg border border-emerald-300 px-2.5 py-1 text-xs font-semibold text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/50"
                          >
                            <FileCheck className="size-3.5" /> Ver Assinado
                          </a>
                        ) : null}

                        {/* Baixar PDF modelo/gerado para impressão manual */}
                        {contract.pdf_url ? (
                          <a
                            href={contract.pdf_url}
                            target="_blank"
                            rel="noreferrer"
                            download
                            className="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-semibold text-foreground hover:bg-muted"
                          >
                            <FileDown className="size-3.5" /> Baixar PDF
                          </a>
                        ) : null}

                        {/* Sincronizar ZapSign */}
                        {contract.signature_method === 'zapsign' && !isSigned ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleSyncZapSign(contract.id)}
                            disabled={syncZapSignMutation.isPending}
                            className="gap-1 text-xs"
                          >
                            <RefreshCw className="size-3.5" /> Sincronizar
                          </Button>
                        ) : null}

                        {/* Anexo manual do contrato assinado */}
                        {!isSigned ? (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => handleOpenAttachModal(contract)}
                            className="gap-1 text-xs"
                          >
                            <FileUp className="size-3.5" /> Anexar Assinado
                          </Button>
                        ) : null}
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Modal / Sheet Emitir Novo Contrato */}
      <Sheet open={emitModalOpen} onOpenChange={setEmitModalOpen}>
        <SheetContent className="flex w-full flex-col sm:max-w-xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Emitir Contrato para o Cliente</SheetTitle>
            <SheetDescription>
              Selecione o template de contrato e a modalidade de assinatura (Eletrônica via ZapSign ou Manual para impressão física).
            </SheetDescription>
          </SheetHeader>

          <div className="space-y-4 py-3">
            <div className="space-y-2">
              <Label>Template de Contrato *</Label>
              <Select
                value={selectedTemplateId}
                onValueChange={(val) => setSelectedTemplateId(val ?? '')}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Selecione um template..." />
                </SelectTrigger>
                <SelectContent>
                  {templatesQuery.data?.items.map((t) => (
                    <SelectItem key={t.id} value={t.id}>
                      {t.name} ({t.pdf_base_url ? 'PDF' : 'HTML'})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Forma de Assinatura *</Label>
              <div className="grid grid-cols-2 gap-3">
                <div
                  className={`flex cursor-pointer items-start gap-3 rounded-2xl border p-4 transition-all ${
                    signatureMethod === 'zapsign'
                      ? 'border-primary bg-primary/5 ring-1 ring-primary'
                      : 'hover:bg-muted/50'
                  }`}
                  onClick={() => setSignatureMethod('zapsign')}
                >
                  <input
                    type="radio"
                    name="signatureMethod"
                    checked={signatureMethod === 'zapsign'}
                    onChange={() => setSignatureMethod('zapsign')}
                    className="mt-1"
                  />
                  <div className="space-y-1">
                    <Label className="cursor-pointer font-bold text-foreground flex items-center gap-1">
                      <Zap className="size-3.5 fill-primary text-primary" /> ZapSign
                    </Label>
                    <p className="text-muted-foreground text-xs leading-relaxed">
                      Envio digital para assinatura eletrônica por WhatsApp ou E-mail com validade jurídica.
                    </p>
                  </div>
                </div>

                <div
                  className={`flex cursor-pointer items-start gap-3 rounded-2xl border p-4 transition-all ${
                    signatureMethod === 'manual'
                      ? 'border-primary bg-primary/5 ring-1 ring-primary'
                      : 'hover:bg-muted/50'
                  }`}
                  onClick={() => setSignatureMethod('manual')}
                >
                  <input
                    type="radio"
                    name="signatureMethod"
                    checked={signatureMethod === 'manual'}
                    onChange={() => setSignatureMethod('manual')}
                    className="mt-1"
                  />
                  <div className="space-y-1">
                    <Label className="cursor-pointer font-bold text-foreground flex items-center gap-1">
                      <FileText className="size-3.5 text-foreground" /> Manual / Físico
                    </Label>
                    <p className="text-muted-foreground text-xs leading-relaxed">
                      Gera o PDF preenchido para impressão e posterior anexo do documento assinado.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Configuração de Signatários no ZapSign */}
            {signatureMethod === 'zapsign' ? (
              <div className="space-y-3 rounded-2xl border bg-muted/20 p-4">
                <div className="flex items-center gap-1.5">
                  <Users className="size-4 text-primary" />
                  <Label className="text-xs font-bold text-foreground">
                    Dados do Signatário para Envio
                  </Label>
                </div>

                <div className="space-y-3">
                  <div>
                    <Label className="text-xs">Nome Completo</Label>
                    <Input
                      value={signers[0]?.name ?? ''}
                      onChange={(e) =>
                        setSigners((prev) => [
                          { ...prev[0], name: e.target.value }
                        ])
                      }
                      placeholder="Nome do titular"
                      className="h-8 text-xs"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <Label className="text-xs">E-mail (opcional)</Label>
                      <Input
                        type="email"
                        value={signers[0]?.email ?? ''}
                        onChange={(e) =>
                          setSigners((prev) => [
                            { ...prev[0], email: e.target.value }
                          ])
                        }
                        placeholder="cliente@email.com"
                        className="h-8 text-xs"
                      />
                    </div>
                    <div>
                      <Label className="text-xs">WhatsApp / Celular</Label>
                      <Input
                        value={signers[0]?.phone ?? ''}
                        onChange={(e) =>
                          setSigners((prev) => [
                            { ...prev[0], phone: e.target.value }
                          ])
                        }
                        placeholder="(11) 99999-9999"
                        className="h-8 text-xs"
                      />
                    </div>
                  </div>
                </div>
              </div>
            ) : null}
          </div>

          <SheetFooter className="mt-auto border-t pt-4">
            <Button
              variant="outline"
              onClick={() => setEmitModalOpen(false)}
            >
              Cancelar
            </Button>
            <Button
              onClick={handleEmitContract}
              disabled={generateContractMutation.isPending}
              className="gap-1.5"
            >
              <Send className="size-4" />
              {signatureMethod === 'zapsign' ? 'Emitir e Enviar ZapSign' : 'Gerar Contrato'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      {/* Modal / Sheet Anexar Contrato Assinado Manualmente */}
      <Sheet open={attachModalOpen} onOpenChange={setAttachModalOpen}>
        <SheetContent className="flex w-full flex-col sm:max-w-md overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Anexar Contrato Assinado</SheetTitle>
            <SheetDescription>
              Suba a via assinada e digitalizada do contrato em formato PDF para arquivar no cadastro do cliente.
            </SheetDescription>
          </SheetHeader>

          <div className="space-y-4 py-3">
            <div className="rounded-2xl border-2 border-dashed p-6 text-center">
              <UploadCloud className="text-muted-foreground mx-auto size-8" />
              <p className="mt-2 text-sm font-semibold text-foreground">
                {signedFileName ? signedFileName : 'Selecione o PDF Assinado'}
              </p>
              <p className="text-muted-foreground text-xs">
                Apenas arquivos em formato .pdf
              </p>
              <label className="mt-4 inline-block cursor-pointer">
                <span className="bg-primary text-primary-foreground inline-flex items-center rounded-lg px-3 py-1.5 text-xs font-semibold hover:opacity-90">
                  Escolher Arquivo PDF
                </span>
                <input
                  type="file"
                  accept="application/pdf"
                  className="hidden"
                  onChange={handleFileUpload}
                />
              </label>
            </div>

            <div className="space-y-2">
              <Label>Nome do Signatário / Responsável</Label>
              <Input
                value={signedBy}
                onChange={(e) => setSignedBy(e.target.value)}
                placeholder="Ex: João da Silva"
              />
            </div>
          </div>

          <SheetFooter className="mt-auto border-t pt-4">
            <Button
              variant="outline"
              onClick={() => setAttachModalOpen(false)}
            >
              Cancelar
            </Button>
            <Button
              onClick={handleSaveAttached}
              disabled={uploadSignedMutation.isPending}
              className="gap-1.5"
            >
              <FileCheck className="size-4" />
              Salvar Contrato Assinado
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
