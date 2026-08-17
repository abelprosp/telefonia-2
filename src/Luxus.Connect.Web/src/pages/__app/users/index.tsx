import { useMemo, useState } from 'react';

import { createFileRoute, Navigate } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Building2, Plus, Sparkles, UserCog } from 'lucide-react';
import { toast } from 'sonner';

import { DataTable } from '@/components/data-table';
import { ListPageHeader, ListPageSkeleton } from '@/components/list-page';
import { PageWrapper } from '@/components/page-wrapper';
import { Badge } from '@/components/ui/badge';
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
  SheetDescription,
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
  { header: 'Perfil', cell: 'text' as const },
  { header: 'Organização', cell: 'text' as const }
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
  const [organizationName, setOrganizationName] = useState('');

  const columns = useMemo<ColumnDef<OrganizationUser>[]>(
    () => [
      { accessorKey: 'full_name', header: 'Nome' },
      { accessorKey: 'username', header: 'Usuário' },
      { accessorKey: 'email', header: 'E-mail' },
      {
        accessorKey: 'organization_name',
        header: 'Organização / Empresa',
        cell: ({ row }) => (
          <div className="flex items-center gap-1.5 font-medium text-xs">
            <Building2 className="size-3.5 text-muted-foreground" />
            <span>{row.original.organization_name || 'Luxus Telefonia'}</span>
          </div>
        )
      },
      {
        accessorKey: 'profile',
        header: 'Perfil',
        cell: ({ row }) => {
          const isMaster = row.original.profile === 'master';
          return (
            <Badge variant={isMaster ? 'default' : 'secondary'} className="font-normal text-xs">
              {profileLabel(row.original.profile)}
            </Badge>
          );
        }
      },
      {
        accessorKey: 'enabled',
        header: 'Ativo',
        cell: ({ row }) => (
          <Badge variant={row.original.enabled ? 'outline' : 'destructive'} className="font-normal text-xs">
            {row.original.enabled ? 'Ativo' : 'Inativo'}
          </Badge>
        )
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => (
          <div className="flex gap-2 justify-end">
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                updateMutation.mutate(
                  { id: row.original.id, enabled: !row.original.enabled },
                  {
                    onSuccess: () => toast.success('Status do usuário atualizado.'),
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
        profile,
        organization_name: profile === 'master' && organizationName.trim() ? organizationName.trim() : undefined
      },
      {
        onSuccess: () => {
          toast.success(
            profile === 'master'
              ? 'Usuário Master criado com nova organização própria!'
              : 'Usuário cadastrado na sua organização com sucesso!'
          );
          setCreateOpen(false);
          setUsername('');
          setEmail('');
          setFirstName('');
          setLastName('');
          setPassword('');
          setProfile('employee');
          setOrganizationName('');
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
          title="Gestão de Usuários"
          description="Gerencie os colaboradores da sua organização ou crie novos clientes Master com organizações independentes."
          action={
            <Button onClick={() => setCreateOpen(true)} className="gap-2">
              <Plus className="size-4" />
              Novo Usuário
            </Button>
          }
        />

        <DataTable columns={columns} data={listQuery.data ?? []} getRowId={(r) => r.id} />
      </div>

      <Sheet open={createOpen} onOpenChange={setCreateOpen}>
        <SheetContent className="overflow-y-auto sm:max-w-lg">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2">
              <UserCog className="size-5 text-primary" />
              Cadastrar Novo Usuário
            </SheetTitle>
            <SheetDescription>
              Preencha os dados de acesso e selecione o perfil de permissão do usuário.
            </SheetDescription>
          </SheetHeader>
          <div className="grid gap-4 px-4 py-2">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>Nome</Label>
                <Input
                  value={firstName}
                  placeholder="Ex: Carlos"
                  onChange={(e) => setFirstName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Sobrenome</Label>
                <Input
                  value={lastName}
                  placeholder="Ex: Silva"
                  onChange={(e) => setLastName(e.target.value)}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Usuário (Login)</Label>
              <Input
                value={username}
                placeholder="Ex: carlos.silva"
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>E-mail</Label>
              <Input
                type="email"
                value={email}
                placeholder="carlos@empresa.com.br"
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Senha inicial</Label>
              <Input
                type="password"
                value={password}
                placeholder="Mínimo 6 caracteres"
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Perfil de Acesso</Label>
              <Select
                value={profile}
                onValueChange={(v) => setProfile((v ?? 'employee') as UserProfile)}
              >
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

            {/* Informação sobre a Organização Multi-tenant */}
            {profile === 'master' ? (
              <div className="rounded-xl border border-primary/20 bg-primary/5 p-4 space-y-3">
                <div className="flex items-center gap-2 text-primary font-medium text-sm">
                  <Sparkles className="size-4" />
                  Nova Organização Multi-tenant
                </div>
                <p className="text-xs text-muted-foreground">
                  Este usuário terá perfil <strong>Master</strong> e receberá uma organização exclusiva,
                  com configurações de empresa, whitelabel e parâmetros de faturamento isolados.
                </p>
                <div className="space-y-1.5 pt-1">
                  <Label className="text-xs font-semibold">Nome da Nova Empresa / Organização</Label>
                  <Input
                    value={organizationName}
                    placeholder="Ex: Telecom Brasil / Alpha Soluções"
                    className="bg-background text-sm"
                    onChange={(e) => setOrganizationName(e.target.value)}
                  />
                </div>
              </div>
            ) : (
              <div className="rounded-xl border bg-muted/40 p-3 text-xs text-muted-foreground flex items-center gap-2">
                <Building2 className="size-4 text-primary shrink-0" />
                <span>
                  Este usuário pertencerá à sua organização atual (<strong>Luxus Telefonia</strong>).
                </span>
              </div>
            )}
          </div>
          <SheetFooter className="mt-4">
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending} className="gap-2">
              {createMutation.isPending ? 'Criando...' : 'Cadastrar Usuário'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
