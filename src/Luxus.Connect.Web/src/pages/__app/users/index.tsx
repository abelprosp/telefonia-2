import { useMemo, useState } from 'react';

import { createFileRoute, Navigate } from '@tanstack/react-router';
import { type ColumnDef } from '@tanstack/react-table';
import { Building2, KeyRound, Pencil, Plus, Sparkles, UserCheck, UserCog, UserX } from 'lucide-react';
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
  { header: 'Organização', cell: 'text' as const },
  { header: 'Ações', cell: 'text' as const }
];

function UsersPage() {
  const { canManageUsers } = useAuthRoles();
  const listQuery = useOrganizationUsers();
  const createMutation = useCreateOrganizationUser();
  const updateMutation = useUpdateOrganizationUser();

  // Estado do Modal de Criação
  const [createOpen, setCreateOpen] = useState(false);
  const [createUsername, setCreateUsername] = useState('');
  const [createEmail, setCreateEmail] = useState('');
  const [createFirstName, setCreateFirstName] = useState('');
  const [createLastName, setCreateLastName] = useState('');
  const [createPassword, setCreatePassword] = useState('');
  const [createProfile, setCreateProfile] = useState<UserProfile>('employee');
  const [createOrgName, setCreateOrgName] = useState('');

  // Estado do Modal de Edição
  const [editOpen, setEditOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<OrganizationUser | null>(null);
  const [editFirstName, setEditFirstName] = useState('');
  const [editLastName, setEditLastName] = useState('');
  const [editEmail, setEditEmail] = useState('');
  const [editProfile, setEditProfile] = useState<UserProfile>('employee');
  const [editPassword, setEditPassword] = useState('');
  const [editEnabled, setEditEnabled] = useState(true);

  const handleOpenEdit = (user: OrganizationUser) => {
    setEditingUser(user);
    setEditFirstName(user.first_name || '');
    setEditLastName(user.last_name || '');
    setEditEmail(user.email || '');
    setEditProfile(user.profile);
    setEditPassword('');
    setEditEnabled(user.enabled);
    setEditOpen(true);
  };

  const handleSaveEdit = () => {
    if (!editingUser) return;

    if (editPassword && editPassword.length < 6) {
      toast.error('A nova senha deve ter no mínimo 6 caracteres.');
      return;
    }

    updateMutation.mutate(
      {
        id: editingUser.id,
        first_name: editFirstName.trim(),
        last_name: editLastName.trim(),
        email: editEmail.trim(),
        profile: editProfile,
        enabled: editEnabled,
        password: editPassword.trim() ? editPassword.trim() : undefined
      },
      {
        onSuccess: () => {
          toast.success(
            editPassword.trim()
              ? 'Dados e senha do usuário atualizados com sucesso!'
              : 'Dados do usuário atualizados com sucesso!'
          );
          setEditOpen(false);
          setEditingUser(null);
          setEditPassword('');
        },
        onError: (e) => toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
      }
    );
  };

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
        header: 'Status',
        cell: ({ row }) => (
          <Badge variant={row.original.enabled ? 'outline' : 'destructive'} className="font-normal text-xs">
            {row.original.enabled ? 'Ativo' : 'Inativo'}
          </Badge>
        )
      },
      {
        id: 'actions',
        header: 'Ações',
        cell: ({ row }) => (
          <div className="flex gap-2 justify-end">
            <Button
              size="sm"
              variant="outline"
              className="gap-1.5 text-xs"
              onClick={() => handleOpenEdit(row.original)}
            >
              <Pencil className="size-3.5" />
              Editar / Senha
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className={row.original.enabled ? 'text-destructive hover:bg-destructive/10' : 'text-green-600 hover:bg-green-50'}
              onClick={() =>
                updateMutation.mutate(
                  { id: row.original.id, enabled: !row.original.enabled },
                  {
                    onSuccess: () =>
                      toast.success(
                        row.original.enabled ? 'Usuário desativado.' : 'Usuário ativado.'
                      ),
                    onError: (e) =>
                      toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                  }
                )
              }
            >
              {row.original.enabled ? <UserX className="size-3.5" /> : <UserCheck className="size-3.5" />}
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
        username: createUsername.trim(),
        email: createEmail.trim(),
        first_name: createFirstName.trim(),
        last_name: createLastName.trim(),
        password: createPassword,
        profile: createProfile,
        organization_name:
          createProfile === 'master' && createOrgName.trim() ? createOrgName.trim() : undefined
      },
      {
        onSuccess: () => {
          toast.success(
            createProfile === 'master'
              ? 'Usuário Master criado com nova organização própria!'
              : 'Usuário cadastrado na sua organização com sucesso!'
          );
          setCreateOpen(false);
          setCreateUsername('');
          setCreateEmail('');
          setCreateFirstName('');
          setCreateLastName('');
          setCreatePassword('');
          setCreateProfile('employee');
          setCreateOrgName('');
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
          description="Gerencie colaboradores da sua organização, edite perfis e redefina senhas com segurança."
          action={
            <Button onClick={() => setCreateOpen(true)} className="gap-2">
              <Plus className="size-4" />
              Novo Usuário
            </Button>
          }
        />

        <DataTable columns={columns} data={listQuery.data ?? []} getRowId={(r) => r.id} />
      </div>

      {/* MODAL 1: CADASTRAR NOVO USUÁRIO */}
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
                  value={createFirstName}
                  placeholder="Ex: Carlos"
                  onChange={(e) => setCreateFirstName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Sobrenome</Label>
                <Input
                  value={createLastName}
                  placeholder="Ex: Silva"
                  onChange={(e) => setCreateLastName(e.target.value)}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Usuário (Login)</Label>
              <Input
                value={createUsername}
                placeholder="Ex: carlos.silva"
                onChange={(e) => setCreateUsername(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>E-mail</Label>
              <Input
                type="email"
                value={createEmail}
                placeholder="carlos@empresa.com.br"
                onChange={(e) => setCreateEmail(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Senha inicial</Label>
              <Input
                type="password"
                value={createPassword}
                placeholder="Mínimo 6 caracteres"
                onChange={(e) => setCreatePassword(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Perfil de Acesso</Label>
              <Select
                value={createProfile}
                onValueChange={(v) => setCreateProfile((v ?? 'employee') as UserProfile)}
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

            {createProfile === 'master' ? (
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
                    value={createOrgName}
                    placeholder="Ex: Telecom Brasil / Alpha Soluções"
                    className="bg-background text-sm"
                    onChange={(e) => setCreateOrgName(e.target.value)}
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

      {/* MODAL 2: EDITAR USUÁRIO & REDEFINIR SENHA */}
      <Sheet open={editOpen} onOpenChange={setEditOpen}>
        <SheetContent className="overflow-y-auto sm:max-w-lg">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2">
              <Pencil className="size-5 text-primary" />
              Editar Usuário & Senha
            </SheetTitle>
            <SheetDescription>
              Atualize os dados cadastrais, perfil de acesso ou redefina a senha de login do usuário.
            </SheetDescription>
          </SheetHeader>

          {editingUser && (
            <div className="grid gap-4 px-4 py-2">
              <div className="rounded-xl border bg-muted/30 p-3 text-xs flex items-center justify-between">
                <div>
                  <span className="text-muted-foreground block">Login (Imutável)</span>
                  <span className="font-semibold text-sm">{editingUser.username}</span>
                </div>
                <div className="text-right">
                  <span className="text-muted-foreground block">Organização</span>
                  <span className="font-semibold text-xs">{editingUser.organization_name || 'Luxus Telefonia'}</span>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label>Nome</Label>
                  <Input
                    value={editFirstName}
                    onChange={(e) => setEditFirstName(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Sobrenome</Label>
                  <Input
                    value={editLastName}
                    onChange={(e) => setEditLastName(e.target.value)}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label>E-mail</Label>
                <Input
                  type="email"
                  value={editEmail}
                  onChange={(e) => setEditEmail(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label>Perfil de Acesso</Label>
                <Select
                  value={editProfile}
                  onValueChange={(v) => setEditProfile((v ?? 'employee') as UserProfile)}
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

              <div className="space-y-2">
                <Label>Status da Conta</Label>
                <Select
                  value={editEnabled ? 'active' : 'inactive'}
                  onValueChange={(v) => setEditEnabled(v === 'active')}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">Ativo (Acesso Liberado)</SelectItem>
                    <SelectItem value="inactive">Inativo (Acesso Bloqueado)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Redefinição de Senha */}
              <div className="rounded-xl border border-amber-500/20 bg-amber-500/5 p-4 space-y-2">
                <div className="flex items-center gap-2 text-amber-700 dark:text-amber-400 font-medium text-sm">
                  <KeyRound className="size-4" />
                  Redefinir Senha de Acesso
                </div>
                <p className="text-xs text-muted-foreground">
                  Preencha uma nova senha abaixo somente se desejar alterá-la. Se deixar em branco, a senha atual continuará inalterada.
                </p>
                <Input
                  type="password"
                  value={editPassword}
                  placeholder="Digite a nova senha (mínimo 6 dígitos)"
                  className="bg-background text-sm mt-1"
                  onChange={(e) => setEditPassword(e.target.value)}
                />
              </div>
            </div>
          )}

          <SheetFooter className="mt-4">
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleSaveEdit} disabled={updateMutation.isPending} className="gap-2">
              {updateMutation.isPending ? 'Salvando...' : 'Salvar Alterações'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </PageWrapper>
  );
}
