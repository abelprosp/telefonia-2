import { useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Power, PowerOff } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';

import { DataTable, DataTablePagination } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { formatPhoneLineStatus, formatPhoneNumber } from '@/lib/format';
import {
  type PartnerPhoneLine,
  useCreatePartnerLineRequest,
  usePartnerPhoneLines
} from '@/lib/partner-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

const SKELETON_COLUMNS = [
  { header: 'Número', cell: 'text' as const },
  { header: 'Cliente', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  { header: 'Solicitar', cell: 'actionsPair' as const }
];

export const Route = createFileRoute('/__app/partner/phone-lines/')({
  validateSearch: searchSchema,
  component: PartnerPhoneLinesPage
});

function PartnerPhoneLinesPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [requestTarget, setRequestTarget] = useState<{
    line: PartnerPhoneLine;
    operation: 'activation' | 'deactivation';
  } | null>(null);
  const [justification, setJustification] = useState('');

  const listQuery = usePartnerPhoneLines({
    page_index: page - 1,
    page_size: pageSize
  });

  const createRequest = useCreatePartnerLineRequest();

  const columns = useMemo<ColumnDef<PartnerPhoneLine>[]>(
    () => [
      {
        accessorKey: 'number',
        header: 'Número',
        cell: ({ row }) =>
          formatPhoneNumber(row.original.number) ?? row.original.number
      },
      {
        accessorKey: 'customer_name',
        header: 'Cliente'
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatPhoneLineStatus(row.original.status)
      },
      {
        id: 'actions',
        header: 'Solicitar',
        cell: ({ row }) => {
          const line = row.original;
          if (!line.customer_id) {
            return '—';
          }
          return (
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  setJustification('');
                  setRequestTarget({ line, operation: 'activation' });
                }}
              >
                <Power className="size-3.5" />
                Ativar
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  setJustification('');
                  setRequestTarget({ line, operation: 'deactivation' });
                }}
              >
                <PowerOff className="size-3.5" />
                Desativar
              </Button>
            </div>
          );
        }
      }
    ],
    []
  );

  const submitRequest = () => {
    if (!requestTarget?.line.customer_id) {
      return;
    }

    createRequest.mutate(
      {
        phone_line_id: requestTarget.line.id,
        customer_id: requestTarget.line.customer_id,
        operation_type: requestTarget.operation,
        justification: justification.trim() || undefined
      },
      {
        onSuccess: () => {
          toast.success('Solicitação enviada à empresa.');
          setRequestTarget(null);
        },
        onError: (e) => {
          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e));
        }
      }
    );
  };

  const total = listQuery.data?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (listQuery.isPending) {
    return (
      <ListPageSkeleton pageSize={pageSize} columns={SKELETON_COLUMNS} />
    );
  }

  if (listQuery.error) {
    return (
      <div className="p-6 text-sm text-destructive">
        {isApiHttpError(listQuery.error)
          ? listQuery.error.message
          : getErrorMessage(listQuery.error)}
      </div>
    );
  }

  const items = listQuery.data?.items ?? [];

  return (
    <div className="flex flex-1 flex-col gap-4 p-6">
      <ListPageHeader
        title="Minhas linhas"
        description="Linhas vinculadas aos clientes da sua carteira"
      />

      <DataTable columns={columns} data={items} getRowId={(row) => row.id} />
      <DataTablePagination
        page={page}
        pageSize={pageSize}
        total={total}
        totalPages={totalPages}
        onPageChange={(next) =>
          navigate({ search: (prev) => ({ ...prev, page: next }) })
        }
        onPageSizeChange={(next) =>
          navigate({ search: (prev) => ({ ...prev, page: 1, pageSize: next }) })
        }
      />

      <Sheet
        open={!!requestTarget}
        onOpenChange={(open) => !open && setRequestTarget(null)}
      >
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>
              Solicitar{' '}
              {requestTarget?.operation === 'activation' ? 'ativação' : 'desativação'}
            </SheetTitle>
            <SheetDescription>
              A empresa receberá sua solicitação e poderá aprovar ou rejeitar a
              operação na linha{' '}
              {requestTarget
                ? formatPhoneNumber(requestTarget.line.number)
                : ''}
              .
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-2 px-4">
            <Label htmlFor="justification">Justificativa (opcional)</Label>
            <Textarea
              id="justification"
              value={justification}
              onChange={(e) => setJustification(e.target.value)}
              rows={4}
            />
          </div>
          <SheetFooter className="gap-2 sm:justify-end">
            <Button variant="outline" onClick={() => setRequestTarget(null)}>
              Cancelar
            </Button>
            <Button onClick={submitRequest} disabled={createRequest.isPending}>
              Enviar solicitação
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
