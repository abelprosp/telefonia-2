import { useEffect, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';

import { DataTable } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
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
  BILLING_PLACEHOLDERS,
  type InvoiceEmailTemplate,
  useCreateInvoiceEmailTemplate,
  useInvoiceEmailTemplate,
  useInvoiceEmailTemplates,
  useUpdateInvoiceEmailTemplate
} from '@/lib/billing-api';

export const Route = createFileRoute('/__app/finance/invoice-email-templates/')({
  component: InvoiceEmailTemplatesPage
});

function InvoiceEmailTemplatesPage() {
  const listQuery = useInvoiceEmailTemplates({ page_index: 0, page_size: 100 });
  const createMutation = useCreateInvoiceEmailTemplate();
  const updateMutation = useUpdateInvoiceEmailTemplate();

  const [sheetOpen, setSheetOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const editingQuery = useInvoiceEmailTemplate(editingId ?? '');

  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [kind, setKind] = useState('billing_invoice');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');

  useEffect(() => {
    if (editingQuery.data) {
      setName(editingQuery.data.name);
      setCode(editingQuery.data.code);
      setKind(editingQuery.data.kind);
      setSubject(editingQuery.data.subject_template);
      setBody(editingQuery.data.body_template_html ?? '');
    }
  }, [editingQuery.data]);

  const openCreate = () => {
    setEditingId(null);
    setName('');
    setCode('');
    setKind('billing_invoice');
    setSubject('Fatura {{invoice.number}} — {{customer.name}}');
    setBody(
      '<h2>Fatura {{invoice.number}}</h2>\n<p>Olá, <strong>{{customer.name}}</strong>.</p>\n<p>Valor: <strong>{{invoice.amount}}</strong>, vencimento <strong>{{invoice.due_date}}</strong>.</p>\n<p>{{invoice.description}}</p>'
    );
    setSheetOpen(true);
  };

  const openEdit = (row: InvoiceEmailTemplate) => {
    setEditingId(row.id);
    setSheetOpen(true);
  };

  const handleSave = async () => {
    try {
      if (editingId) {
        await updateMutation.mutateAsync({
          id: editingId,
          name,
          subject_template: subject,
          body_template_html: body,
          active: true
        });
        toast.success('Template atualizado.');
      } else {
        await createMutation.mutateAsync({
          name,
          code,
          kind,
          subject_template: subject,
          body_template_html: body,
          active: true
        });
        toast.success('Template criado.');
      }
      setSheetOpen(false);
    } catch (e) {
      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
    }
  };

  const columns: ColumnDef<InvoiceEmailTemplate>[] = [
    { accessorKey: 'name', header: 'Nome' },
    { accessorKey: 'code', header: 'Código' },
    { accessorKey: 'kind', header: 'Tipo' },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Button variant="ghost" size="sm" onClick={() => openEdit(row.original)}>
          Editar
        </Button>
      )
    }
  ];

  if (listQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Templates de e-mail' }]}>
        <ListPageSkeleton pageSize={10} columns={[{ header: 'Nome', cell: 'text' }, { header: 'Código', cell: 'text' }]} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Templates de e-mail' }]}>
      <div className="flex flex-col gap-4">
        <ListPageHeader
          title="Templates de e-mail"
          description="Modelos para faturas e cobrança de inadimplentes"
          action={
            <Button onClick={openCreate}>
              <Plus className="mr-2 size-4" />
              Novo template
            </Button>
          }
        />
        <DataTable columns={columns} data={listQuery.data?.items ?? []} />
      </div>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent className="overflow-y-auto sm:max-w-xl">
          <SheetHeader>
            <SheetTitle>{editingId ? 'Editar template' : 'Novo template'}</SheetTitle>
          </SheetHeader>
          <div className="mt-4 space-y-4">
            <div className="space-y-2">
              <Label>Nome</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            {!editingId && (
              <>
                <div className="space-y-2">
                  <Label>Código</Label>
                  <Input value={code} onChange={(e) => setCode(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label>Tipo</Label>
                  <Input value={kind} onChange={(e) => setKind(e.target.value)} placeholder="billing_invoice" />
                </div>
              </>
            )}
            <div className="space-y-2">
              <Label>Assunto</Label>
              <Input value={subject} onChange={(e) => setSubject(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Corpo HTML</Label>
              <Textarea className="min-h-[200px] font-mono text-xs" value={body} onChange={(e) => setBody(e.target.value)} />
            </div>
            <p className="text-muted-foreground text-xs">
              Placeholders: {BILLING_PLACEHOLDERS.join(', ')}
            </p>
          </div>
          <SheetFooter className="mt-6">
            <Button onClick={() => void handleSave()} disabled={createMutation.isPending || updateMutation.isPending}>
              Salvar
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
