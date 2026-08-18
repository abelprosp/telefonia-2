import { useEffect, useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import {
  FileText,
  FileUp,
  MapPin,
  Plus,
  Trash2,
  Users,
  Zap
} from 'lucide-react';
import { toast } from 'sonner';

import { DataTable } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  CONTRACT_PLACEHOLDERS,
  type ContractTemplate,
  type SignerConfig,
  useContractTemplate,
  useContractTemplates,
  useCreateContractTemplate,
  useUpdateContractTemplate
} from '@/lib/sales-api';

export const Route = createFileRoute('/__app/contract-templates/')({
  component: ContractTemplatesPage
});

const SKELETON_COLUMNS = [
  { header: 'Nome', cell: 'text' as const },
  { header: 'Código', cell: 'text' as const },
  { header: 'Tipo', cell: 'text' as const },
  { header: 'Signatários', cell: 'text' as const },
  { header: 'Ativo', cell: 'text' as const }
];

function ContractTemplatesPage() {
  const listQuery = useContractTemplates(false);
  const createMutation = useCreateContractTemplate();
  const updateMutation = useUpdateContractTemplate();

  const [sheetOpen, setSheetOpen] = useState(false);
  const [editing, setEditing] = useState<ContractTemplate | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const editingQuery = useContractTemplate(editingId ?? '');

  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [body, setBody] = useState('');
  const [pdfBaseUrl, setPdfBaseUrl] = useState('');
  const [pdfFileName, setPdfFileName] = useState('');
  const [signers, setSigners] = useState<SignerConfig[]>([]);
  const [active, setActive] = useState(true);
  const [activeTab, setActiveTab] = useState<'pdf' | 'html'>('pdf');

  useEffect(() => {
    if (editingQuery.data) {
      setBody(editingQuery.data.body_template ?? '');
      setPdfBaseUrl(editingQuery.data.pdf_base_url ?? '');
      setSigners(editingQuery.data.signers_config ?? []);
      if (editingQuery.data.pdf_base_url) {
        setActiveTab('pdf');
      } else {
        setActiveTab('html');
      }
    }
  }, [editingQuery.data]);

  const openCreate = () => {
    setEditing(null);
    setEditingId(null);
    setName('');
    setCode('');
    setPdfBaseUrl('');
    setPdfFileName('');
    setSigners([
      {
        role: 'cliente',
        label: 'Cliente / Representante Legal',
        page: 1,
        x: 50,
        y: 85,
        require_email: true,
        require_phone: true
      }
    ]);
    setBody(`<h1>Contrato de Prestação de Serviços</h1>
<p>Cliente: {{customer.name}} ({{customer.document}})</p>
<p>Endereço: {{customer.address}}</p>
<p>Valor total: {{sale.total_amount}}</p>
{{sale.items_table}}`);
    setActive(true);
    setActiveTab('pdf');
    setSheetOpen(true);
  };

  const openEdit = (row: ContractTemplate) => {
    setEditing(row);
    setEditingId(row.id);
    setName(row.name);
    setCode(row.code);
    setActive(row.active);
    setPdfBaseUrl(row.pdf_base_url ?? '');
    setPdfFileName(row.pdf_base_url ? 'modelo-contrato.pdf' : '');
    setSigners(row.signers_config ?? []);
    setBody('');
    setSheetOpen(true);
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.type !== 'application/pdf') {
      toast.error('Selecione um arquivo PDF válido.');
      return;
    }
    setPdfFileName(file.name);
    const reader = new FileReader();
    reader.onload = () => {
      const base64 = reader.result as string;
      setPdfBaseUrl(base64);
      toast.success('Arquivo PDF carregado no modelo.');
    };
    reader.readAsDataURL(file);
  };

  const addSigner = () => {
    setSigners((prev) => [
      ...prev,
      {
        role: `signatario_${prev.length + 1}`,
        label: `Signatário ${prev.length + 1}`,
        page: 1,
        x: 50,
        y: 85,
        require_email: true,
        require_phone: false
      }
    ]);
  };

  const updateSigner = (index: number, patch: Partial<SignerConfig>) => {
    setSigners((prev) =>
      prev.map((s, i) => (i === index ? { ...s, ...patch } : s))
    );
  };

  const removeSigner = (index: number) => {
    setSigners((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSave = () => {
    if (!name.trim() || !code.trim()) {
      toast.error('Preencha nome e código do template.');
      return;
    }
    if (!body.trim() && !pdfBaseUrl.trim()) {
      toast.error('Envie um PDF Modelo ou preencha o corpo do template.');
      return;
    }

    const payload = {
      name,
      code,
      body_template: body,
      pdf_base_url: pdfBaseUrl || null,
      signers_config: signers,
      active
    };

    if (editing) {
      updateMutation.mutate(
        { id: editing.id, ...payload },
        {
          onSuccess: () => {
            toast.success('Template de contrato atualizado com sucesso.');
            setSheetOpen(false);
            void listQuery.refetch();
          },
          onError: (e) =>
            toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
        }
      );
    } else {
      createMutation.mutate(payload, {
        onSuccess: () => {
          toast.success('Template de contrato criado com sucesso.');
          setSheetOpen(false);
          void listQuery.refetch();
        },
        onError: (e) =>
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      });
    }
  };

  const columns = useMemo<ColumnDef<ContractTemplate>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Nome do Template',
        cell: ({ row }) => (
          <div className="flex flex-col">
            <span className="font-semibold text-foreground">
              {row.original.name}
            </span>
            <span className="text-muted-foreground text-xs font-mono">
              {row.original.code}
            </span>
          </div>
        )
      },
      {
        id: 'type',
        header: 'Formato',
        cell: ({ row }) =>
          row.original.pdf_base_url ? (
            <Badge variant="outline" className="gap-1 border-primary/30 text-primary">
              <FileText className="size-3" /> PDF Base
            </Badge>
          ) : (
            <Badge variant="secondary" className="gap-1">
              HTML / Texto
            </Badge>
          )
      },
      {
        id: 'signers',
        header: 'Assinaturas (ZapSign)',
        cell: ({ row }) => {
          const count = row.original.signers_config?.length ?? 0;
          return count > 0 ? (
            <Badge variant="outline" className="gap-1 border-emerald-300 text-emerald-700 dark:text-emerald-400">
              <Zap className="size-3 fill-current" /> {count} {count === 1 ? 'signatário' : 'signatários'}
            </Badge>
          ) : (
            <span className="text-muted-foreground text-xs">Manual</span>
          );
        }
      },
      {
        accessorKey: 'active',
        header: 'Status',
        cell: ({ row }) => (
          <span
            className={
              row.original.active
                ? 'inline-flex items-center rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
                : 'text-muted-foreground inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs font-semibold'
            }
          >
            {row.original.active ? 'Ativo' : 'Inativo'}
          </span>
        )
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => openEdit(row.original)}
          >
            Editar
          </Button>
        )
      }
    ],
    []
  );

  if (listQuery.isPending) {
    return (
      <PageWrapper
        breadcrumbs={[
          { label: 'Início', to: '/' },
          { label: 'Templates de contrato' }
        ]}
      >
        <ListPageSkeleton pageSize={10} columns={SKELETON_COLUMNS} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Início', to: '/' },
        { label: 'Templates de contrato' }
      ]}
    >
      <div className="flex flex-col gap-6 p-6">
        <ListPageHeader
          title="Templates de Contrato"
          description="Gerencie modelos de contrato em PDF com mapeamento de assinaturas via ZapSign ou impressão manual."
          action={
            <Button onClick={openCreate} className="gap-1.5">
              <Plus className="size-4" />
              Novo Template
            </Button>
          }
        />

        <div className="rounded-3xl border bg-card p-4 shadow-xs">
          <DataTable
            columns={columns}
            data={listQuery.data?.items ?? []}
          />
        </div>
      </div>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent className="flex w-full flex-col sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle>
              {editing ? 'Editar Template de Contrato' : 'Novo Template de Contrato'}
            </SheetTitle>
          </SheetHeader>

          <div className="flex flex-col gap-5 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="tpl-name">Nome do Template *</Label>
                <Input
                  id="tpl-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Ex: Contrato de Prestação Telefonia Móvel"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="tpl-code">Código Único *</Label>
                <Input
                  id="tpl-code"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="Ex: contrato-telefonia-padrao"
                />
              </div>
            </div>

            {/* Alternador de Modo */}
            <div className="flex rounded-xl border bg-muted/40 p-1">
              <button
                type="button"
                onClick={() => setActiveTab('pdf')}
                className={`flex flex-1 items-center justify-center gap-1.5 rounded-lg py-1.5 text-xs font-semibold transition-all ${
                  activeTab === 'pdf'
                    ? 'bg-card text-foreground shadow-xs'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <FileUp className="size-3.5" /> Modelo PDF & ZapSign
              </button>
              <button
                type="button"
                onClick={() => setActiveTab('html')}
                className={`flex flex-1 items-center justify-center gap-1.5 rounded-lg py-1.5 text-xs font-semibold transition-all ${
                  activeTab === 'html'
                    ? 'bg-card text-foreground shadow-xs'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <FileText className="size-3.5" /> Editor HTML / Texto
              </button>
            </div>

            {activeTab === 'pdf' ? (
              <div className="space-y-5 pt-1">
                <div className="rounded-2xl border-2 border-dashed p-6 text-center">
                  <FileUp className="text-muted-foreground mx-auto size-8" />
                  <p className="mt-2 text-sm font-semibold text-foreground">
                    {pdfFileName ? `Arquivo: ${pdfFileName}` : 'Upload do PDF Base / Modelo'}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    Suba o modelo de contrato em formato .pdf
                  </p>
                  <label className="mt-4 inline-block cursor-pointer">
                    <span className="bg-primary text-primary-foreground inline-flex items-center rounded-lg px-3 py-1.5 text-xs font-semibold hover:opacity-90">
                      Selecionar Arquivo PDF
                    </span>
                    <input
                      type="file"
                      accept="application/pdf"
                      className="hidden"
                      onChange={handleFileUpload}
                    />
                  </label>
                </div>

                {/* Editor de Signatários e Coordenadas de Assinatura */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5">
                      <Users className="text-primary size-4" />
                      <Label className="text-sm font-bold text-foreground">
                        Posicionamento das Assinaturas (ZapSign)
                      </Label>
                    </div>
                    <Button variant="outline" size="sm" onClick={addSigner} className="gap-1 text-xs">
                      <Plus className="size-3" /> Adicionar Signatário
                    </Button>
                  </div>

                  {signers.length === 0 ? (
                    <div className="text-muted-foreground rounded-xl border border-dashed p-4 text-center text-xs">
                      Nenhum signatário configurado. Adicione ao menos um para envio via ZapSign.
                    </div>
                  ) : (
                    signers.map((signer, index) => (
                      <div
                        key={index}
                        className="space-y-3 rounded-2xl border bg-muted/20 p-4"
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-xs font-bold text-foreground">
                            Signatário #{index + 1}: {signer.label}
                          </span>
                          <button
                            type="button"
                            onClick={() => removeSigner(index)}
                            className="text-muted-foreground hover:text-destructive"
                          >
                            <Trash2 className="size-4" />
                          </button>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <Label className="text-xs">Rótulo / Papel</Label>
                            <Input
                              value={signer.label}
                              onChange={(e) =>
                                updateSigner(index, { label: e.target.value })
                              }
                              placeholder="Ex: Titular, Testemunha"
                              className="h-8 text-xs"
                            />
                          </div>
                          <div>
                            <Label className="text-xs">Página no PDF</Label>
                            <Input
                              type="number"
                              min={1}
                              value={signer.page}
                              onChange={(e) =>
                                updateSigner(index, {
                                  page: parseInt(e.target.value, 10) || 1
                                })
                              }
                              className="h-8 text-xs"
                            />
                          </div>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <Label className="text-xs flex items-center gap-1">
                              <MapPin className="size-3" /> Posição X (%)
                            </Label>
                            <Input
                              type="number"
                              min={0}
                              max={100}
                              value={signer.x}
                              onChange={(e) =>
                                updateSigner(index, {
                                  x: parseFloat(e.target.value) || 0
                                })
                              }
                              placeholder="50% (Centro)"
                              className="h-8 text-xs"
                            />
                          </div>
                          <div>
                            <Label className="text-xs flex items-center gap-1">
                              <MapPin className="size-3" /> Posição Y (%)
                            </Label>
                            <Input
                              type="number"
                              min={0}
                              max={100}
                              value={signer.y}
                              onChange={(e) =>
                                updateSigner(index, {
                                  y: parseFloat(e.target.value) || 0
                                })
                              }
                              placeholder="85% (Rodapé)"
                              className="h-8 text-xs"
                            />
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            ) : (
              <div className="space-y-4 pt-1">
                <div className="space-y-2">
                  <Label htmlFor="tpl-body">Corpo do Template (HTML / Texto)</Label>
                  <Textarea
                    id="tpl-body"
                    value={body}
                    onChange={(e) => setBody(e.target.value)}
                    rows={12}
                    className="font-mono text-xs"
                    placeholder="Escreva o contrato utilizando HTML ou texto com tags dinâmicas..."
                  />
                </div>

                <div>
                  <Label className="text-xs font-semibold text-muted-foreground">
                    Tags Dinâmicas Disponíveis:
                  </Label>
                  <div className="mt-2 flex flex-wrap gap-1.5">
                    {CONTRACT_PLACEHOLDERS.map((p) => (
                      <button
                        key={p}
                        type="button"
                        onClick={() => {
                          setBody((prev) => prev + ` ${p}`);
                          toast.success(`Tag ${p} inserida.`);
                        }}
                        className="rounded-md border bg-muted/40 px-2 py-0.5 font-mono text-[11px] hover:bg-muted"
                      >
                        {p}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            <div className="flex items-center gap-2 pt-2">
              <input
                type="checkbox"
                id="tpl-active"
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
                className="size-4 rounded border-gray-300"
              />
              <Label htmlFor="tpl-active" className="cursor-pointer text-sm font-medium">
                Template Ativo para Emissão
              </Label>
            </div>
          </div>

          <SheetFooter className="mt-auto border-t pt-4">
            <Button variant="outline" onClick={() => setSheetOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleSave}>Salvar Template</Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
