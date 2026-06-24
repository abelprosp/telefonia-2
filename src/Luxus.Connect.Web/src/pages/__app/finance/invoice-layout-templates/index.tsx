import { createFileRoute, Link } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Palette, Plus } from 'lucide-react';

import { DataTable } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { useInvoiceLayoutTemplates } from '@/lib/invoice-layout-api';
import type { InvoiceLayoutTemplate } from '@/lib/invoice-layout/types';

export const Route = createFileRoute('/__app/finance/invoice-layout-templates/')({
  component: InvoiceLayoutTemplatesPage
});

function InvoiceLayoutTemplatesPage() {
  const listQuery = useInvoiceLayoutTemplates({ page_index: 0, page_size: 100 });

  const columns: ColumnDef<InvoiceLayoutTemplate>[] = [
    { accessorKey: 'name', header: 'Nome' },
    { accessorKey: 'code', header: 'Código' },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Button variant="ghost" size="sm" render={<Link to="/finance/invoice-layout-templates/$id" params={{ id: row.original.id }} />}>
          Editar layout
        </Button>
      )
    }
  ];

  if (listQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Layouts de fatura' }]}>
        <ListPageSkeleton pageSize={6} columns={[{ header: 'Nome', cell: 'text' }, { header: 'Código', cell: 'text' }]} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper breadcrumbs={[{ label: 'Financeiro', to: '/finance' }, { label: 'Layouts de fatura' }]}>
      <div className="flex flex-col gap-4">
        <ListPageHeader
          title="Editor de layouts de fatura"
          description="Personalize logo, cores e seções do detalhamento da fatura enviada ao cliente"
          action={
            <Button render={<Link to="/finance/invoice-layout-templates/$id" params={{ id: 'new' }} />}>
              <Plus className="mr-2 size-4" />
              Novo layout
            </Button>
          }
        />
        {(listQuery.data?.items ?? []).length === 0 ? (
          <div className="dashboard-card flex flex-col items-center gap-3 p-10 text-center">
            <Palette className="text-muted-foreground size-10" />
            <p className="text-muted-foreground text-sm">Nenhum layout cadastrado ainda.</p>
            <Button render={<Link to="/finance/invoice-layout-templates/$id" params={{ id: 'new' }} />}>Criar primeiro layout</Button>
          </div>
        ) : (
          <DataTable columns={columns} data={listQuery.data?.items ?? []} />
        )}
      </div>
    </PageWrapper>
  );
}
