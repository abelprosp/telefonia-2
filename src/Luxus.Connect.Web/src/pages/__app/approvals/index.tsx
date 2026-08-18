import { createFileRoute } from '@tanstack/react-router';
import { toast } from 'sonner';

import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { useApprovals, useApproveRequest, useRejectRequest } from '@/lib/ops-api';

export const Route = createFileRoute('/__app/approvals/')({
  component: ApprovalsPage
});

function ApprovalsPage() {
  const query = useApprovals();
  const approveMutation = useApproveRequest();
  const rejectMutation = useRejectRequest();
  const items = query.data ?? [];

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Aprovações' }]}>
      <ListPageHeader
        title="Aprovação em dois níveis"
        description="Quem solicita não aprova a si mesmo. Master, admin e financeiro participam como usuários distintos. Consolidação, reabertura, cobrança forçada, lote, desconto alto, cancelamento retroativo e alteração de ciclo."
      />
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Ação</TableHead>
            <TableHead>Entidade</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Solicitante</TableHead>
            <TableHead>Decisão</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((a) => (
            <TableRow key={a.id}>
              <TableCell>{a.action_type}</TableCell>
              <TableCell>{a.entity_id}</TableCell>
              <TableCell>{a.status}</TableCell>
              <TableCell>{a.requester_user_id}</TableCell>
              <TableCell className="space-x-2">
                {(a.status === 'pending_first' || a.status === 'pending_second') && (
                  <>
                    <Button
                      type="button"
                      size="sm"
                      onClick={() =>
                        approveMutation.mutate(a.id, {
                          onSuccess: () => toast.success('Aprovação registrada.'),
                          onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                        })
                      }
                    >
                      Aprovar
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() =>
                        rejectMutation.mutate(
                          { id: a.id, reason: 'Recusado na tela de aprovações' },
                          {
                            onSuccess: () => toast.success('Pedido recusado.'),
                            onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                          }
                        )
                      }
                    >
                      Recusar
                    </Button>
                  </>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </PageWrapper>
  );
}
