import { useMemo, useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
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
import { formatPhoneNumber } from '@/lib/format';
import {
  type PhoneLineOperationRequest,
  useAdminLineRequests,
  useReviewLineRequest
} from '@/lib/partner-api';

const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  pageSize: z.number().int().min(5).max(100).catch(10)
});

const SKELETON_COLUMNS = [
  { header: 'Linha', cell: 'text' as const },
  { header: 'Cliente', cell: 'text' as const },
  { header: 'Operação', cell: 'text' as const },
  { header: 'Status', cell: 'text' as const },
  { header: 'Ações', cell: 'actionsPair' as const }
];

export const Route = createFileRoute('/__app/line-requests/')({
  validateSearch: searchSchema,
  component: LineRequestsAdminPage
});

function LineRequestsAdminPage() {
  const { page, pageSize } = Route.useSearch();
  const navigate = Route.useNavigate();
  const [reviewTarget, setReviewTarget] = useState<{
    request: PhoneLineOperationRequest;
    status: 'approved' | 'rejected';
  } | null>(null);
  const [adminNotes, setAdminNotes] = useState('');

  const listQuery = useAdminLineRequests({
    page_index: page - 1,
    page_size: pageSize
  });

  const reviewMutation = useReviewLineRequest();

  const columns = useMemo<ColumnDef<PhoneLineOperationRequest>[]>(
    () => [
      {
        accessorKey: 'phone_line_number',
        header: 'Linha',
        cell: ({ row }) =>
          formatPhoneNumber(row.original.phone_line_number) ??
          row.original.phone_line_number
      },
      {
        accessorKey: 'customer_name',
        header: 'Cliente'
      },
      {
        accessorKey: 'operation_type',
        header: 'Operação',
        cell: ({ row }) =>
          row.original.operation_type === 'activation' ? 'Ativação' : 'Desativação'
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => formatStatus(row.original.status)
      },
      {
        id: 'actions',
        header: 'Ações',
        cell: ({ row }) => {
          if (row.original.status !== 'pending') {
            return '—';
          }
          return (
            <div className="flex gap-2">
              <Button
                size="sm"
                onClick={() => {
                  setAdminNotes('');
                  setReviewTarget({ request: row.original, status: 'approved' });
                }}
              >
                Aprovar
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  setAdminNotes('');
                  setReviewTarget({ request: row.original, status: 'rejected' });
                }}
              >
                Rejeitar
              </Button>
            </div>
          );
        }
      }
    ],
    []
  );

  const submitReview = () => {
    if (!reviewTarget) {
      return;
    }

    reviewMutation.mutate(
      {
        id: reviewTarget.request.id,
        status: reviewTarget.status,
        admin_notes: adminNotes.trim() || undefined
      },
      {
        onSuccess: () => {
          toast.success(
            reviewTarget.status === 'approved'
              ? 'Solicitação aprovada.'
              : 'Solicitação rejeitada.'
          );
          setReviewTarget(null);
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
        title="Solicitações de parceiros"
        description="Aprove ou rejeite pedidos de ativação e desativação de linhas"
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
        open={!!reviewTarget}
        onOpenChange={(open) => !open && setReviewTarget(null)}
      >
        <SheetContent side="right" className="sm:max-w-md">
          <SheetHeader>
            <SheetTitle>
              {reviewTarget?.status === 'approved' ? 'Aprovar' : 'Rejeitar'}{' '}
              solicitação
            </SheetTitle>
            <SheetDescription>
              Linha{' '}
              {reviewTarget
                ? formatPhoneNumber(reviewTarget.request.phone_line_number)
                : ''}{' '}
              — cliente {reviewTarget?.request.customer_name}
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-2 px-4">
            <Label htmlFor="admin-notes">Observações (opcional)</Label>
            <Textarea
              id="admin-notes"
              value={adminNotes}
              onChange={(e) => setAdminNotes(e.target.value)}
              rows={4}
            />
          </div>
          <SheetFooter className="gap-2 sm:justify-end">
            <Button variant="outline" onClick={() => setReviewTarget(null)}>
              Cancelar
            </Button>
            <Button onClick={submitReview} disabled={reviewMutation.isPending}>
              Confirmar
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}

function formatStatus(status: string) {
  switch (status) {
    case 'pending':
      return 'Pendente';
    case 'approved':
      return 'Aprovada';
    case 'rejected':
      return 'Rejeitada';
    case 'cancelled':
      return 'Cancelada';
    default:
      return status;
  }
}
