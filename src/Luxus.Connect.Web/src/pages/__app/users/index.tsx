import { useMemo, useState } from 'react';

import { createFileRoute, Navigate } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Plus, UserCog } from 'lucide-react';
import { toast } from 'sonner';

import { DataTable } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
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
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { PROFILE_OPTIONS, profileLabel, useAuthRoles, type UserProfile } from '@/lib/auth-roles';
import {
  type OrganizationUser,
  useCreateOrganizationUser,
  useOrganizationUsers,
  useUpdateOrganizationUser
} from '@/lib/users-api';

export const Route = createFileRoute('/__app/users/')({
  component: UsersPage
});

const SKELETON_COLUMNS = [
  { header: 'Nome', cell: 'text' as const },
  { header: 'Usuário', cell: 'text' as const },
  { header: 'Perfil', cell: 'text' as const }
];

function UsersPage() {
  const { canManageUsers } = useAuthRoles();
  const listQuery = useOrganizationUsers();
  const createMutation = useCreateOrganizationUser();
  const updateMutation = useUpdateOrganizationUser();

  const [createOpen, setCreateOpen] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [password, setPassword] = useState('');
  const [profile, setProfile] = useState<UserProfile>('employee');

  const columns = useMemo<ColumnDef<OrganizationUser>[]>(
    () => [
      { accessorKey: 'full_name', header: 'Nome' },
      { accessorKey: 'username', header: 'Usuário' },
      { accessorKey: 'email', header: 'E-mail' },
      {
        accessorKey: 'profile',
        header: 'Perfil',
        cell: ({ row }) => profileLabel(row.original.profile)
      },
      {
        accessorKey: 'enabled',
        header: 'Ativo',
        cell: ({ row }) => (row.original.enabled ? 'Sim' : 'Não')
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                updateMutation.mutate(
                  { id: row.original.id, enabled: !row.original.enabled },
                  {
                    onSuccess: () => toast.success('Status atualizado.'),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  }
                )
              }
            >
              {row.original.enabled ? 'Desativar' : 'Ativar'}
            </Button>
          </div>
        )
      }
    ],
    [updateMutation]
  );

  const handleCreate = () => {
    createMutation.mutate(
      {
        username: username.trim(),
        email: email.trim(),
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        password,
        profile
      },
      {
        onSuccess: () => {
          toast.success('Usuário criado.');
          setCreateOpen(false);
          setUsername('');
          setEmail('');
          setFirstName('');
          setLastName('');
          setPassword('');
          setProfile('employee');
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

  if (!canManageUsers) {
    return <Navigate to="/" />;
  }

  if (listQuery.isPending) {
    return (
      <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Usuários' }]}>
        <ListPageSkeleton pageSize={10} columns={SKELETON_COLUMNS} />
      </PageWrapper>
    );
  }

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Usuários' }]}>
      <div className="flex flex-col gap-6 p-6">
        <ListPageHeader
          title="Usuários"
          description="Crie contas com perfil Master, Funcionário, Financeiro ou Parceiro."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus />
              Novo usuário
            </Button>
          }
        />

        <DataTable columns={columns} data={listQuery.data ?? []} getRowId={(r) => r.id} />
      </div>

      <Sheet open={createOpen} onOpenChange={setCreateOpen}>
        <SheetContent className="overflow-y-auto sm:max-w-lg">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2">
              <UserCog className="size-5" />
              Novo usuário
            </SheetTitle>
          </SheetHeader>
          <div className="grid gap-4 px-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>Nome</Label>
                <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label>Sobrenome</Label>
                <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Usuário (login)</Label>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>E-mail</Label>
              <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Senha inicial</Label>
              <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Perfil de acesso</Label>
              <Select value={profile} onValueChange={(v) => setProfile((v ?? 'employee') as UserProfile)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PROFILE_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label} — {opt.description}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <SheetFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              Criar usuário
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
