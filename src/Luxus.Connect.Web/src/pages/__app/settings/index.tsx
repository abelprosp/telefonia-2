import { createFileRoute, Link } from '@tanstack/react-router';
import {
  Building2,
  CheckCircle2,
  CreditCard,
  Globe,
  LogOut,
  ShieldCheck,
  UserCog,
  Users
} from 'lucide-react';
import { useAuth } from 'react-oidc-context';

import { PageWrapper } from '@/components/page-wrapper';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { roleLabel, useAuthRoles } from '@/lib/auth-roles';

export const Route = createFileRoute('/__app/settings/')({
  component: SettingsPage
});

function SettingsPage() {
  const { user, removeUser, signoutSilent } = useAuth();
  const authRoles = useAuthRoles();

  const onSignout = async () => {
    await signoutSilent();
    removeUser();
  };

  const displayName = user?.profile.name ?? user?.profile.preferred_username ?? 'Usuário';
  const email = user?.profile.email ?? '—';
  const username = user?.profile.preferred_username ?? '—';
  const avatar = user?.profile.picture;
  const initials = displayName
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  return (
    <PageWrapper
      breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Perfil e Configurações' }]}
    >
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-8 p-6 md:p-8">
        {/* Cabeçalho */}
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">Perfil e Configurações</h1>
          <p className="text-muted-foreground text-sm">
            Gerencie suas informações de acesso, dados da empresa e parâmetros do sistema.
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-3">
          {/* Card: Meu Perfil */}
          <div className="bg-card text-card-foreground flex flex-col justify-between rounded-2xl border p-6 shadow-xs md:col-span-1">
            <div className="flex flex-col items-center text-center">
              <Avatar className="size-20 border-2 shadow-xs">
                <AvatarImage src={avatar} alt={displayName} />
                <AvatarFallback className="text-lg font-semibold">{initials}</AvatarFallback>
              </Avatar>

              <h2 className="mt-4 text-lg font-semibold">{displayName}</h2>
              <p className="text-muted-foreground text-xs">{email}</p>

              <div className="mt-3 flex flex-wrap justify-center gap-1.5">
                <Badge variant="outline" className="gap-1 border-primary/30 bg-primary/5 text-primary">
                  <ShieldCheck className="size-3" />
                  {roleLabel(authRoles)}
                </Badge>
                <Badge variant="secondary" className="gap-1 text-xs">
                  <CheckCircle2 className="size-3 text-green-500" />
                  Ativo
                </Badge>
              </div>

              <div className="mt-6 flex w-full flex-col gap-2.5 border-t pt-4 text-left text-xs">
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">Usuário (login)</span>
                  <span className="font-medium">{username}</span>
                </div>
                <div className="flex justify-between py-1">
                  <span className="text-muted-foreground">Autenticação</span>
                  <span className="font-medium">Keycloak SSO</span>
                </div>
              </div>
            </div>

            <Button
              variant="outline"
              onClick={onSignout}
              className="mt-6 w-full gap-2 text-destructive hover:bg-destructive/10 hover:text-destructive"
            >
              <LogOut className="size-4" />
              Encerrar sessão
            </Button>
          </div>

          {/* Coluna Direita: Configurações do Sistema e Empresa */}
          <div className="flex flex-col gap-6 md:col-span-2">
            {/* Card: Dados da Empresa */}
            <div className="bg-card text-card-foreground rounded-2xl border p-6 shadow-xs">
              <div className="flex items-center gap-3 border-b pb-4">
                <div className="flex size-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
                  <Building2 className="size-5" />
                </div>
                <div>
                  <h3 className="text-base font-semibold">Luxus Telefonia Ltda</h3>
                  <p className="text-muted-foreground text-xs">Dados cadastrais da organização</p>
                </div>
              </div>

              <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 text-sm">
                <div>
                  <span className="text-muted-foreground text-xs">CNPJ</span>
                  <p className="font-medium">11.309.896/0001-01</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Domínio do Sistema</span>
                  <p className="font-medium flex items-center gap-1">
                    <Globe className="size-3.5 text-muted-foreground" />
                    telefonia.redobrai.online
                  </p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Sistema</span>
                  <p className="font-medium">Luxus Connect — Gestão de Telefonia</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Ambiente</span>
                  <p className="font-medium text-green-600 dark:text-green-400">● Produção (VPS)</p>
                </div>
              </div>
            </div>

            {/* Card: Integração Bancária Sicredi */}
            <div className="bg-card text-card-foreground rounded-2xl border p-6 shadow-xs">
              <div className="flex items-center justify-between border-b pb-4">
                <div className="flex items-center gap-3">
                  <div className="flex size-9 items-center justify-center rounded-xl bg-green-500/10 text-green-600 dark:text-green-400">
                    <CreditCard className="size-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-semibold">Integração Sicredi Cobrança & PIX</h3>
                    <p className="text-muted-foreground text-xs">Emissão de Boletos Híbridos com QR Code PIX</p>
                  </div>
                </div>
                <Badge className="bg-green-600 text-white hover:bg-green-600">Conectado</Badge>
              </div>

              <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3 text-sm">
                <div>
                  <span className="text-muted-foreground text-xs">Cooperativa / Agência</span>
                  <p className="font-medium">0179</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Posto de Atendimento</span>
                  <p className="font-medium">14</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Código do Beneficiário</span>
                  <p className="font-medium">01048</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Conta Corrente</span>
                  <p className="font-medium">04133-5</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Tipo de Cobrança</span>
                  <p className="font-medium">Híbrido (Boleto + PIX)</p>
                </div>
                <div>
                  <span className="text-muted-foreground text-xs">Status da API</span>
                  <p className="font-medium text-green-600 dark:text-green-400">Ativa em Produção</p>
                </div>
              </div>
            </div>

            {/* Card: Gestão de Usuários (Acesso Rápido) */}
            {authRoles.canManageUsers && (
              <div className="bg-card text-card-foreground flex items-center justify-between rounded-2xl border p-6 shadow-xs">
                <div className="flex items-center gap-3">
                  <div className="flex size-9 items-center justify-center rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400">
                    <Users className="size-5" />
                  </div>
                  <div>
                    <h3 className="text-base font-semibold">Gestão de Usuários e Acessos</h3>
                    <p className="text-muted-foreground text-xs">
                      Crie e gerencie contas de funcionários, financeiro e parceiros.
                    </p>
                  </div>
                </div>
                <Button variant="outline" render={<Link to="/users" />} className="gap-2">
                  <UserCog className="size-4" />
                  Gerenciar
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </PageWrapper>
  );
}
