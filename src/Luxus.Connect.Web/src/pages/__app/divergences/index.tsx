import { useState } from 'react';

import { createFileRoute } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useAuth } from 'react-oidc-context';

import { ListPageHeader } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  useAssignDivergence,
  useCommentDivergence,
  useDivergences,
  useResolveDivergence
} from '@/lib/ops-api';

export const Route = createFileRoute('/__app/divergences/')({
  component: DivergencesPage
});

function DivergencesPage() {
  const query = useDivergences();
  const resolveMutation = useResolveDivergence();
  const assignMutation = useAssignDivergence();
  const commentMutation = useCommentDivergence();
  const auth = useAuth();
  const [comment, setComment] = useState('');
  const [activeId, setActiveId] = useState('');
  const items = query.data ?? [];
  const userId = auth.user?.profile.sub ?? '';

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Divergências' }]}>
      <ListPageHeader
        title="Centro de divergências"
        description="Pendências persistentes: tipo, gravidade, impacto, idade, causa, evidência e resolução. Atribuir, comentar, resolver e reabrir."
      />
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Tipo</TableHead>
            <TableHead>Gravidade</TableHead>
            <TableHead>Linha</TableHead>
            <TableHead>Descrição / causa</TableHead>
            <TableHead>Impacto</TableHead>
            <TableHead>Idade (h)</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Ação</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((d) => (
            <TableRow key={d.id}>
              <TableCell>{d.divergence_type}</TableCell>
              <TableCell>
                {d.severity}
                {d.severity === 'HIGH' ? ' (alta)' : d.severity === 'MEDIUM' ? ' (média)' : ' (baixa)'}
              </TableCell>
              <TableCell>{d.phone_number ?? '—'}</TableCell>
              <TableCell>
                <p>{d.description}</p>
                {d.recommended_action ? (
                  <p className="text-xs">Ação recomendada: {d.recommended_action}</p>
                ) : null}
              </TableCell>
              <TableCell>{d.financial_impact ?? 0}</TableCell>
              <TableCell>{d.age_hours ?? 0}</TableCell>
              <TableCell>{d.status}</TableCell>
              <TableCell className="space-x-2 whitespace-nowrap">
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  disabled={!userId}
                  onClick={() =>
                    assignMutation.mutate(
                      { id: d.id, owner_user_id: userId },
                      {
                        onSuccess: () => toast.success('Divergência atribuída a você.'),
                        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  Assumir
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    resolveMutation.mutate(
                      { id: d.id, action: 'resolve', notes: 'Resolvido no centro de divergências' },
                      {
                        onSuccess: () => toast.success('Divergência resolvida.'),
                        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  Resolver
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    resolveMutation.mutate(
                      { id: d.id, action: 'reopen' },
                      {
                        onSuccess: () => toast.success('Reaberta.'),
                        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      }
                    )
                  }
                >
                  Reabrir
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <form
        className="mt-6 flex max-w-xl flex-wrap items-end gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          if (!activeId) {
            toast.error('Informe o ID da divergência para comentar.');
            return;
          }
          commentMutation.mutate(
            { id: activeId, body: comment },
            {
              onSuccess: () => {
                toast.success('Comentário registrado.');
                setComment('');
              },
              onError: (err) => toast.error(isApiHttpError(err) ? err.message : getErrorMessage(err))
            }
          );
        }}
      >
        <div>
          <Label htmlFor="div-id">ID da divergência</Label>
          <Input id="div-id" value={activeId} onChange={(e) => setActiveId(e.target.value)} />
        </div>
        <div className="flex-1">
          <Label htmlFor="div-comment">Comentário</Label>
          <Input id="div-comment" value={comment} onChange={(e) => setComment(e.target.value)} required />
        </div>
        <Button type="submit">Comentar</Button>
      </form>
    </PageWrapper>
  );
}
