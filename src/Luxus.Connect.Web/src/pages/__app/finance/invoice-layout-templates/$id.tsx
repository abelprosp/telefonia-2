import { useEffect, useMemo, useState } from 'react';

import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { ImagePlus, Loader2, Save } from 'lucide-react';
import { toast } from 'sonner';

import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import {
  useCreateInvoiceLayoutTemplate,
  useInvoiceLayoutTemplate,
  useUpdateInvoiceLayoutTemplate
} from '@/lib/invoice-layout-api';
import { DEFAULT_INVOICE_LAYOUT_CONFIG, SAMPLE_INVOICE_LAYOUT_DATA } from '@/lib/invoice-layout/defaults';
import { renderInvoiceLayoutHtml } from '@/lib/invoice-layout/render';
import type { InvoiceLayoutConfig } from '@/lib/invoice-layout/types';

export const Route = createFileRoute('/__app/finance/invoice-layout-templates/$id')({
  component: InvoiceLayoutEditorPage
});

const COLOR_FIELDS: { key: keyof InvoiceLayoutConfig['theme']; label: string }[] = [
  { key: 'primaryColor', label: 'Cor da marca' },
  { key: 'accentColor', label: 'Cor de destaque' },
  { key: 'borderColor', label: 'Cor das bordas' },
  { key: 'headerBackground', label: 'Fundo dos blocos' },
  { key: 'titleColor', label: 'Cor do título' },
  { key: 'textColor', label: 'Cor do texto' },
  { key: 'tableHeaderBackground', label: 'Fundo da tabela' }
];

const SECTION_FIELDS: { key: keyof InvoiceLayoutConfig['sections']; label: string }[] = [
  { key: 'userData', label: 'Dados do usuário' },
  { key: 'accountValue', label: 'Valor da conta' },
  { key: 'billingDates', label: 'Datas de faturamento' },
  { key: 'accountSummary', label: 'Resumo da conta' },
  { key: 'detailedConsumption', label: 'Consumo detalhado' }
];

function InvoiceLayoutEditorPage() {
  const { id } = Route.useParams();
  const navigate = useNavigate();
  const isNew = id === 'new';

  const detailQuery = useInvoiceLayoutTemplate(isNew ? '' : id);
  const createMutation = useCreateInvoiceLayoutTemplate();
  const updateMutation = useUpdateInvoiceLayoutTemplate();

  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [config, setConfig] = useState<InvoiceLayoutConfig>(DEFAULT_INVOICE_LAYOUT_CONFIG);

  useEffect(() => {
    if (detailQuery.data?.config_json) {
      setName(detailQuery.data.name);
      setCode(detailQuery.data.code);
      setConfig(detailQuery.data.config_json);
    }
  }, [detailQuery.data]);

  const previewHtml = useMemo(
    () => renderInvoiceLayoutHtml(config, SAMPLE_INVOICE_LAYOUT_DATA),
    [config]
  );

  const updateTheme = (key: keyof InvoiceLayoutConfig['theme'], value: string | number) => {
    setConfig((prev) => ({
      ...prev,
      theme: { ...prev.theme, [key]: value }
    }));
  };

  const updateBranding = (key: keyof InvoiceLayoutConfig['branding'], value: string) => {
    setConfig((prev) => ({
      ...prev,
      branding: { ...prev.branding, [key]: value }
    }));
  };

  const updateSection = (key: keyof InvoiceLayoutConfig['sections'], enabled: boolean) => {
    setConfig((prev) => ({
      ...prev,
      sections: {
        ...prev.sections,
        [key]: { ...prev.sections[key], enabled }
      }
    }));
  };

  const handleLogoUpload = (file: File | null) => {
    if (!file) return;
    if (file.size > 512 * 1024) {
      toast.error('Logo deve ter no máximo 512 KB.');
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        updateBranding('logoDataUrl', reader.result);
      }
    };
    reader.readAsDataURL(file);
  };

  const handleSave = async () => {
    try {
      if (isNew) {
        const created = await createMutation.mutateAsync({ name, code, config_json: config });
        toast.success('Layout criado.');
        void navigate({ to: '/finance/invoice-layout-templates/$id', params: { id: created.id } });
      } else {
        await updateMutation.mutateAsync({ id, name, config_json: config });
        toast.success('Layout salvo.');
      }
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  if (!isNew && detailQuery.isPending) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Loader2 className="text-muted-foreground size-8 animate-spin" />
      </div>
    );
  }

  return (
    <PageWrapper
      breadcrumbs={[
        { label: 'Financeiro', to: '/finance' },
        { label: 'Layouts de fatura', to: '/finance/invoice-layout-templates' },
        { label: isNew ? 'Novo layout' : name || 'Editar' }
      ]}
    >
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Editor de fatura</h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Ajuste logo, cores e seções — a pré-visualização atualiza em tempo real
          </p>
        </div>
        <Button onClick={() => void handleSave()} disabled={createMutation.isPending || updateMutation.isPending}>
          <Save className="mr-2 size-4" />
          Salvar layout
        </Button>
      </div>

      <div className="grid gap-6 xl:grid-cols-[380px_1fr]">
        <div className="space-y-4">
          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Identificação</h2>
            <div className="space-y-2">
              <Label>Nome do layout</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Detalhamento Luxus" />
            </div>
            {isNew && (
              <div className="space-y-2">
                <Label>Código</Label>
                <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="luxus-detalhamento" />
              </div>
            )}
          </div>

          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Logo e marca</h2>
            <div className="space-y-2">
              <Label>Logo</Label>
              <div className="flex items-center gap-3">
                <Button variant="outline" size="sm" type="button" onClick={() => document.getElementById('logo-upload')?.click()}>
                  <ImagePlus className="mr-2 size-4" />
                  Enviar imagem
                </Button>
                {config.branding.logoDataUrl ? (
                  <Button variant="ghost" size="sm" type="button" onClick={() => updateBranding('logoDataUrl', '')}>
                    Remover
                  </Button>
                ) : null}
              </div>
              <input
                id="logo-upload"
                type="file"
                accept="image/png,image/jpeg,image/webp,image/svg+xml"
                className="hidden"
                onChange={(e) => handleLogoUpload(e.target.files?.[0] ?? null)}
              />
              {config.branding.logoDataUrl ? (
                <img src={config.branding.logoDataUrl} alt="Logo" className="max-h-20 rounded border p-2" />
              ) : null}
            </div>
            <div className="space-y-2">
              <Label>Nome da empresa</Label>
              <Input value={config.branding.companyName} onChange={(e) => updateBranding('companyName', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Slogan</Label>
              <Input value={config.branding.tagline} onChange={(e) => updateBranding('tagline', e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Título do documento</Label>
              <Input
                value={config.branding.documentTitle}
                onChange={(e) => updateBranding('documentTitle', e.target.value)}
              />
            </div>
          </div>

          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Cores</h2>
            <div className="grid gap-3 sm:grid-cols-2">
              {COLOR_FIELDS.map((field) => (
                <div key={field.key} className="space-y-1">
                  <Label className="text-xs">{field.label}</Label>
                  <div className="flex items-center gap-2">
                    <input
                      type="color"
                      value={config.theme[field.key] as string}
                      onChange={(e) => updateTheme(field.key, e.target.value)}
                      className="h-9 w-12 cursor-pointer rounded border"
                    />
                    <Input
                      value={config.theme[field.key] as string}
                      onChange={(e) => updateTheme(field.key, e.target.value)}
                      className="font-mono text-xs"
                    />
                  </div>
                </div>
              ))}
            </div>
            <div className="space-y-2">
              <Label>Raio das bordas (px)</Label>
              <Input
                type="number"
                min={0}
                max={32}
                value={config.theme.borderRadius}
                onChange={(e) => updateTheme('borderRadius', Number(e.target.value) || 0)}
              />
            </div>
          </div>

          <div className="dashboard-card space-y-4 p-5">
            <h2 className="font-semibold">Seções visíveis</h2>
            {SECTION_FIELDS.map((section) => (
              <label key={section.key} className="flex cursor-pointer items-center justify-between gap-3">
                <span className="text-sm">{section.label}</span>
                <input
                  type="checkbox"
                  checked={config.sections[section.key].enabled}
                  onChange={(e) => updateSection(section.key, e.target.checked)}
                  className="size-4 accent-[var(--primary)]"
                />
              </label>
            ))}
          </div>
        </div>

        <div className="dashboard-card overflow-hidden p-0">
          <div className="border-b px-5 py-3">
            <h2 className="font-semibold">Pré-visualização</h2>
            <p className="text-muted-foreground text-xs">Dados de exemplo — na fatura real entram os dados do cliente</p>
          </div>
          <div className="bg-muted/30 overflow-auto p-6">
            <iframe
              className="mx-auto block min-h-[900px] w-full max-w-[860px] bg-white shadow-sm"
              title="Pré-visualização do layout"
              srcDoc={previewHtml}
            />
          </div>
        </div>
      </div>

      <Link
        to="/finance/invoice-layout-templates"
        className="text-primary mt-6 inline-block text-sm font-medium hover:underline"
      >
        Voltar à lista
      </Link>
    </PageWrapper>
  );
}
