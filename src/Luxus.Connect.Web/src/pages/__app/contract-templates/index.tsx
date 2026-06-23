import { useEffect, useMemo, useState } from 'react';

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
  CONTRACT_PLACEHOLDERS,
  type ContractTemplate,
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
  const [active, setActive] = useState(true);

  useEffect(() => {
    if (editingQuery.data?.body_template) {
      setBody(editingQuery.data.body_template);
    }
  }, [editingQuery.data?.body_template]);

  const openCreate = () => {
    setEditing(null);
    setEditingId(null);
    setName('');
    setCode('');
    setBody(`<h1>Contrato</h1>
<p>Cliente: {{customer.name}} ({{customer.document}})</p>
<p>Valor total: {{sale.total_amount}}</p>
{{sale.items_table}}`);
    setActive(true);
    setSheetOpen(true);
  };

  const openEdit = (row: ContractTemplate) => {
    setEditing(row);
    setEditingId(row.id);
    setName(row.name);
    setCode(row.code);
    setActive(row.active);
    setBody('');
    setSheetOpen(true);
  };

  const handleSave = () => {
    if (!name.trim() || !code.trim() || !body.trim()) {
      toast.error('Preencha nome, código e corpo do template.');
      return;
    }
    if (editing) {
      updateMutation.mutate(
        { id: editing.id, name, code, body_template: body, active },
        {
          onSuccess: () => {
            toast.success('Template atualizado.');
            setSheetOpen(false);
            void listQuery.refetch();
          },
          onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
        }
      );
    } else {
      createMutation.mutate(
        { name, code, body_template: body, active },
        {
          onSuccess: () => {
            toast.success('Template criado.');
            setSheetOpen(false);
            void listQuery.refetch();
          },
          onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
        }
      );
    }
  };

  const columns = useMemo<ColumnDef<ContractTemplate>[]>(
    () => [
      { accessorKey: 'name', header: 'Nome' },
      { accessorKey: 'code', header: 'Código' },
      {
        accessorKey: 'active',
        header: 'Ativo',
        cell: ({ row }) => (row.original.active ? 'Sim' : 'Não')
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <Button variant="link" size="sm" onClick={() => openEdit(row.original)}>
            Editar
          </Button>
        )
      }
    ],
    []
  );

  if (listQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Templates de contrato' }]}>
        <ListPageSkeleton pageSize={10} columns={SKELETON_COLUMNS} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Templates de contrato' }]}>
      <div className="flex flex-col gap-6 p-6">
        <ListPageHeader
          title="Templates de contrato"
          description="Personalize modelos com placeholders preenchidos automaticamente após a venda."
          action={
            <Button onClick={openCreate}>
              <Plus />
              Novo template
            </Button>
          }
        />

        <DataTable columns={columns} data={listQuery.data?.items ?? []} getRowId={(r) => r.id} />
      </div>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent className="overflow-y-auto sm:max-w-2xl">
          <SheetHeader>
            <SheetTitle>{editing ? 'Editar template' : 'Novo template'}</SheetTitle>
          </SheetHeader>
          <div className="space-y-4 px-4">
            <div className="space-y-2">
              <Label>Nome</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Código</Label>
              <Input value={code} onChange={(e) => setCode(e.target.value)} disabled={Boolean(editing)} />
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} />
              Ativo
            </label>
            <div className="space-y-2">
              <Label>Corpo (HTML)</Label>
              <Textarea value={body} onChange={(e) => setBody(e.target.value)} rows={12} className="font-mono text-xs" />
              <p className="text-muted-foreground text-xs">
                Placeholders: {CONTRACT_PLACEHOLDERS.join(', ')}
              </p>
            </div>
          </div>
          <SheetFooter>
            <Button variant="outline" onClick={() => setSheetOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleSave} disabled={createMutation.isPending || updateMutation.isPending}>
              Salvar
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
