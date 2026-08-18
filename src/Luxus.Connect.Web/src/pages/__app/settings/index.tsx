import { useEffect, useState } from 'react';
import { createFileRoute, Link } from '@tanstack/react-router';
import {
  Building2,
  CheckCircle2,
  CreditCard,
  Eye,
  EyeOff,
  Globe,
  KeyRound,
  Loader2,
  LogOut,
  Palette,
  Phone,
  Save,
  Settings,
  ShieldCheck,
  Sparkles,
  Webhook,
  UserCheck,
  UserCircle,
  UserCog,
  Users
} from 'lucide-react';

import { useAuth } from 'react-oidc-context';
import { toast } from 'sonner';

import {
  useOrganizationSettingsQuery,
  useUpdateCompanySettingsMutation,
  useUpdateSystemSettingsMutation,
  useUpdateUserProfileMutation,
  useUpdateWhitelabelSettingsMutation,
  useUserProfileQuery,
  type CompanySettings,
  type SystemSettings,
  type WhitelabelSettings
} from '@/api/settings-api';
import { PageWrapper } from '@/components/page-wrapper';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { getErrorMessage, isApiHttpError } from '@/lib/api-error';
import { performLogout } from '@/lib/auth-actions';
import { roleLabel, useAuthRoles } from '@/lib/auth-roles';
import { useExportOrganizationData } from '@/lib/ops-api';
import { WebhooksSettingsPanel } from './-components/webhooks-settings-panel';

export const Route = createFileRoute('/__app/settings/')({
  component: SettingsPage
});

type TabKey = 'profile' | 'company' | 'whitelabel' | 'system' | 'users' | 'webhooks';

const COLOR_PRESETS = [
  { name: 'Esmeralda / Teal', color: '#0f766e' },
  { name: 'Azul Real', color: '#2563eb' },
  { name: 'Índigo', color: '#4f46e5' },
  { name: 'Safira Escura', color: '#1e3a8a' },
  { name: 'Grafite / Slate', color: '#334155' },
  { name: 'Vinho / Rose', color: '#be123c' },
  { name: 'Âmbar / Laranja', color: '#d97706' }
];

function SettingsPage() {
  const auth = useAuth();
  const { user } = auth;
  const authRoles = useAuthRoles();

  const [activeTab, setActiveTab] = useState<TabKey>('profile');

  // Queries
  const { data: userProfile, refetch: refetchProfile } = useUserProfileQuery();
  const { data: orgSettings } = useOrganizationSettingsQuery();


  // Mutations
  const updateProfileMutation = useUpdateUserProfileMutation();
  const updateCompanyMutation = useUpdateCompanySettingsMutation();
  const updateWhitelabelMutation = useUpdateWhitelabelSettingsMutation();
  const updateSystemMutation = useUpdateSystemSettingsMutation();
  const exportOrg = useExportOrganizationData();

  // Form States - Perfil
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');

  // Form States - Senha
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  // Form States - Empresa
  const [companyForm, setCompanyForm] = useState<CompanySettings>({
    company_name: '',
    trading_name: '',
    cnpj: '',
    state_registration: '',
    email: '',
    phone: '',
    website: '',
    zip_code: '',
    street: '',
    number: '',
    complement: '',
    neighborhood: '',
    city: '',
    state: ''
  });

  // Form States - Whitelabel
  const [whitelabelForm, setWhitelabelForm] = useState<WhitelabelSettings>({
    app_name: '',
    app_slogan: '',
    logo_url: '',
    dark_logo_url: '',
    favicon_url: '',
    primary_color: '#0f766e',
    support_email: '',
    support_phone: '',
    footer_text: ''
  });

  // Form States - Sistema
  const [systemForm, setSystemForm] = useState<SystemSettings>({
    default_due_day: 10,
    late_fee_percentage: 2.0,
    interest_rate_monthly: 1.0,
    days_before_due_reminder: 3,
    days_after_due_reminder: 2,
    auto_send_invoice_email: true,
    auto_send_collection_reminder: false,
    prorata_divisor: 30
  });

  // Carregar dados iniciais de perfil
  useEffect(() => {
    if (userProfile) {
      setFirstName(userProfile.first_name || '');
      setLastName(userProfile.last_name || '');
      setEmail(userProfile.email || '');
    } else if (user?.profile) {
      const nameParts = (user.profile.name || '').split(' ');
      setFirstName(nameParts[0] || '');
      setLastName(nameParts.slice(1).join(' ') || '');
      setEmail(user.profile.email || '');
    }
  }, [userProfile, user]);

  // Carregar dados de configurações da organização
  useEffect(() => {
    if (orgSettings) {
      if (orgSettings.company) {
        setCompanyForm({ ...orgSettings.company });
      }
      if (orgSettings.whitelabel) {
        setWhitelabelForm({ ...orgSettings.whitelabel });
      }
      if (orgSettings.system) {
        setSystemForm({ ...orgSettings.system });
      }
    }
  }, [orgSettings]);

  // Logout seguro
  const onSignout = async () => {
    toast.info('Encerrando sessão...');
    await performLogout(auth);
  };

  // Handlers de Salvar
  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!firstName.trim()) {
      toast.error('O primeiro nome é obrigatório.');
      return;
    }
    try {
      await updateProfileMutation.mutateAsync({
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        email: email.trim()
      });
      toast.success('Perfil atualizado com sucesso!');
      refetchProfile();
    } catch (err: any) {
      toast.error(err?.message || 'Falha ao atualizar perfil.');
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPassword) {
      toast.error('Informe a nova senha.');
      return;
    }
    if (newPassword.length < 6) {
      toast.error('A senha deve ter pelo menos 6 caracteres.');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('A confirmação de senha não confere.');
      return;
    }
    try {
      await updateProfileMutation.mutateAsync({
        new_password: newPassword
      });
      toast.success('Senha alterada com sucesso!');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      toast.error(err?.message || 'Falha ao alterar senha.');
    }
  };

  const handleSaveCompany = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateCompanyMutation.mutateAsync(companyForm);
      toast.success('Dados da empresa atualizados com sucesso!');
    } catch (err: any) {
      toast.error(err?.message || 'Falha ao salvar dados da empresa.');
    }
  };

  const handleSaveWhitelabel = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateWhitelabelMutation.mutateAsync(whitelabelForm);
      toast.success('Configurações de Whitelabel atualizadas com sucesso!');
    } catch (err: any) {
      toast.error(err?.message || 'Falha ao salvar Whitelabel.');
    }
  };

  const handleSaveSystem = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateSystemMutation.mutateAsync(systemForm);
      toast.success('Parâmetros do sistema atualizados com sucesso!');
    } catch (err: any) {
      toast.error(err?.message || 'Falha ao salvar parâmetros.');
    }
  };

  const displayName = userProfile?.full_name || user?.profile.name || user?.profile.preferred_username || 'Usuário';
  const displayEmail = userProfile?.email || user?.profile.email || '—';
  const username = userProfile?.username || user?.profile.preferred_username || '—';
  const avatar = user?.profile.picture;
  const initials = displayName
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  return (
    <PageWrapper breadcrumbs={[{ label: 'Início', to: '/' }, { label: 'Perfil e Configurações' }]}>
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 p-4 md:p-8">
        {/* Cabeçalho */}
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold tracking-tight">Perfil e Configurações</h1>
            <p className="text-muted-foreground text-sm">
              Gerencie sua conta, informações da empresa, identidade visual (Whitelabel) e parâmetros do sistema.
            </p>
          </div>

          <Button
            variant="outline"
            onClick={onSignout}
            className="self-start sm:self-auto gap-2 text-destructive hover:bg-destructive/10 hover:text-destructive border-destructive/20"
          >
            <LogOut className="size-4" />
            Sair do Sistema
          </Button>
        </div>

        {/* Barra de Abas (Tabs) */}
        <div className="flex overflow-x-auto border-b pb-px gap-2 no-scrollbar">
          <button
            type="button"
            onClick={() => setActiveTab('profile')}
            className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
              activeTab === 'profile'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <UserCircle className="size-4" />
            Meu Perfil
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('company')}
            className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
              activeTab === 'company'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Building2 className="size-4" />
            Dados da Empresa
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('whitelabel')}
            className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
              activeTab === 'whitelabel'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Palette className="size-4" />
            Whitelabel & Marca
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('system')}
            className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
              activeTab === 'system'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Settings className="size-4" />
            Parâmetros & Integrações
          </button>

          {authRoles.canManageUsers && (
            <button
              type="button"
              onClick={() => setActiveTab('users')}
              className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
                activeTab === 'users'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              <Users className="size-4" />
              Gestão de Usuários
            </button>
          )}
          {authRoles.canManageUsers && (
            <button
              type="button"
              onClick={() => setActiveTab('webhooks')}
              className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors whitespace-nowrap ${
                activeTab === 'webhooks'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              <Webhook className="size-4" />
              Webhooks
            </button>
          )}
        </div>

        {/* CONTEÚDO DAS ABAS */}

        {/* 1. ABA MEU PERFIL */}
        {activeTab === 'profile' && (
          <div className="grid gap-6 md:grid-cols-3">
            {/* Coluna Esquerda: Card de Identificação */}
            <div className="md:col-span-1">
              <Card className="rounded-2xl border shadow-xs">
                <CardContent className="flex flex-col items-center p-6 text-center">
                  <Avatar className="size-24 border-4 border-background shadow-md">
                    <AvatarImage src={avatar} alt={displayName} />
                    <AvatarFallback className="text-2xl font-bold bg-primary/10 text-primary">
                      {initials}
                    </AvatarFallback>
                  </Avatar>

                  <h2 className="mt-4 text-lg font-bold">{displayName}</h2>
                  <p className="text-muted-foreground text-xs">{displayEmail}</p>

                  <div className="mt-3 flex flex-wrap justify-center gap-2">
                    <Badge variant="outline" className="gap-1 border-primary/30 bg-primary/5 text-primary">
                      <ShieldCheck className="size-3.5" />
                      {roleLabel(authRoles)}
                    </Badge>
                    <Badge variant="secondary" className="gap-1 text-xs">
                      <CheckCircle2 className="size-3.5 text-green-500" />
                      Conta Ativa
                    </Badge>
                  </div>

                  <div className="mt-6 flex w-full flex-col gap-3 border-t pt-4 text-left text-xs">
                    <div className="flex justify-between py-1">
                      <span className="text-muted-foreground">Login (Username)</span>
                      <span className="font-semibold">{username}</span>
                    </div>
                    <div className="flex justify-between py-1">
                      <span className="text-muted-foreground">Autenticação</span>
                      <span className="font-semibold">Keycloak SSO / OIDC</span>
                    </div>
                  </div>

                  <Button
                    variant="outline"
                    onClick={onSignout}
                    className="mt-6 w-full gap-2 text-destructive hover:bg-destructive/10 hover:text-destructive border-destructive/20"
                  >
                    <LogOut className="size-4" />
                    Encerrar Sessão
                  </Button>
                </CardContent>
              </Card>
            </div>

            {/* Coluna Direita: Formulários de Edição de Perfil e Senha */}
            <div className="flex flex-col gap-6 md:col-span-2">
              {/* Form: Dados Pessoais */}
              <Card className="rounded-2xl border shadow-xs">
                <form onSubmit={handleSaveProfile}>
                  <CardHeader className="pb-4">
                    <CardTitle className="text-lg flex items-center gap-2">
                      <UserCheck className="size-5 text-primary" />
                      Editar Informações Pessoais
                    </CardTitle>
                    <CardDescription>
                      Atualize seu nome, sobrenome e e-mail de contato cadastrados no sistema.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="first_name">Primeiro Nome</Label>
                        <Input
                          id="first_name"
                          value={firstName}
                          onChange={(e) => setFirstName(e.target.value)}
                          placeholder="Ex: Carlos"
                          required
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="last_name">Sobrenome</Label>
                        <Input
                          id="last_name"
                          value={lastName}
                          onChange={(e) => setLastName(e.target.value)}
                          placeholder="Ex: Silva"
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="email">E-mail Principal</Label>
                      <Input
                        id="email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="usuario@empresa.com.br"
                        required
                      />
                    </div>
                  </CardContent>
                  <CardFooter className="flex justify-end border-t pt-4">
                    <Button type="submit" disabled={updateProfileMutation.isPending} className="gap-2">
                      {updateProfileMutation.isPending ? (
                        <Loader2 className="size-4 animate-spin" />
                      ) : (
                        <Save className="size-4" />
                      )}
                      Salvar Alterações
                    </Button>
                  </CardFooter>
                </form>
              </Card>

              {/* Form: Alterar Senha */}
              <Card className="rounded-2xl border shadow-xs">
                <form onSubmit={handleChangePassword}>
                  <CardHeader className="pb-4">
                    <CardTitle className="text-lg flex items-center gap-2">
                      <KeyRound className="size-5 text-primary" />
                      Alterar Senha de Acesso
                    </CardTitle>
                    <CardDescription>
                      Para sua segurança, utilize uma senha forte com no mínimo 6 caracteres.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="new_password">Nova Senha</Label>
                        <div className="relative">
                          <Input
                            id="new_password"
                            type={showPassword ? 'text' : 'password'}
                            value={newPassword}
                            onChange={(e) => setNewPassword(e.target.value)}
                            placeholder="Mínimo 6 caracteres"
                            className="pr-10"
                          />
                          <button
                            type="button"
                            onClick={() => setShowPassword(!showPassword)}
                            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                          >
                            {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                          </button>
                        </div>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="confirm_password">Confirmar Nova Senha</Label>
                        <Input
                          id="confirm_password"
                          type={showPassword ? 'text' : 'password'}
                          value={confirmPassword}
                          onChange={(e) => setConfirmPassword(e.target.value)}
                          placeholder="Repita a nova senha"
                        />
                      </div>
                    </div>
                  </CardContent>
                  <CardFooter className="flex justify-end border-t pt-4">
                    <Button
                      type="submit"
                      disabled={updateProfileMutation.isPending || !newPassword}
                      variant="secondary"
                      className="gap-2"
                    >
                      {updateProfileMutation.isPending ? (
                        <Loader2 className="size-4 animate-spin" />
                      ) : (
                        <KeyRound className="size-4" />
                      )}
                      Atualizar Senha
                    </Button>
                  </CardFooter>
                </form>
              </Card>
            </div>
          </div>
        )}

        {/* 2. ABA DADOS DA EMPRESA */}
        {activeTab === 'company' && (
          <form onSubmit={handleSaveCompany} className="flex flex-col gap-6">
            <Card className="rounded-2xl border shadow-xs">
              <CardHeader className="pb-4">
                <CardTitle className="text-lg flex items-center gap-2">
                  <Building2 className="size-5 text-primary" />
                  Informações Cadastrais da Empresa
                </CardTitle>
                <CardDescription>
                  Dados jurídicos e fiscais da organização utilizados na emissão de faturas e contratos.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="company_name">Razão Social</Label>
                    <Input
                      id="company_name"
                      value={companyForm.company_name}
                      onChange={(e) => setCompanyForm({ ...companyForm, company_name: e.target.value })}
                      placeholder="Ex: Luxus Telefonia Ltda"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="trading_name">Nome Fantasia</Label>
                    <Input
                      id="trading_name"
                      value={companyForm.trading_name}
                      onChange={(e) => setCompanyForm({ ...companyForm, trading_name: e.target.value })}
                      placeholder="Ex: Luxus Connect"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="cnpj">CNPJ</Label>
                    <Input
                      id="cnpj"
                      value={companyForm.cnpj}
                      onChange={(e) => setCompanyForm({ ...companyForm, cnpj: e.target.value })}
                      placeholder="00.000.000/0000-00"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="state_registration">Inscrição Estadual (IE)</Label>
                    <Input
                      id="state_registration"
                      value={companyForm.state_registration}
                      onChange={(e) => setCompanyForm({ ...companyForm, state_registration: e.target.value })}
                      placeholder="Isento ou número da IE"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  <div className="space-y-2">
                    <Label htmlFor="company_email">E-mail de Contato</Label>
                    <Input
                      id="company_email"
                      type="email"
                      value={companyForm.email}
                      onChange={(e) => setCompanyForm({ ...companyForm, email: e.target.value })}
                      placeholder="contato@empresa.com.br"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="company_phone">Telefone / WhatsApp</Label>
                    <Input
                      id="company_phone"
                      value={companyForm.phone}
                      onChange={(e) => setCompanyForm({ ...companyForm, phone: e.target.value })}
                      placeholder="(00) 00000-0000"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="company_website">Domínio / Website</Label>
                    <Input
                      id="company_website"
                      value={companyForm.website}
                      onChange={(e) => setCompanyForm({ ...companyForm, website: e.target.value })}
                      placeholder="https://telefonia.suaempresa.com"
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Endereço */}
            <Card className="rounded-2xl border shadow-xs">
              <CardHeader className="pb-4">
                <CardTitle className="text-lg flex items-center gap-2">
                  <Globe className="size-5 text-primary" />
                  Endereço da Sede
                </CardTitle>
                <CardDescription>Endereço oficial que consta nas propostas e cabeçalhos de cobrança.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="zip_code">CEP</Label>
                    <Input
                      id="zip_code"
                      value={companyForm.zip_code}
                      onChange={(e) => setCompanyForm({ ...companyForm, zip_code: e.target.value })}
                      placeholder="00000-000"
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-2">
                    <Label htmlFor="street">Logradouro / Rua</Label>
                    <Input
                      id="street"
                      value={companyForm.street}
                      onChange={(e) => setCompanyForm({ ...companyForm, street: e.target.value })}
                      placeholder="Av. Paulista"
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="number">Número</Label>
                    <Input
                      id="number"
                      value={companyForm.number}
                      onChange={(e) => setCompanyForm({ ...companyForm, number: e.target.value })}
                      placeholder="1000"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="complement">Complemento</Label>
                    <Input
                      id="complement"
                      value={companyForm.complement}
                      onChange={(e) => setCompanyForm({ ...companyForm, complement: e.target.value })}
                      placeholder="Sala 101"
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="neighborhood">Bairro</Label>
                    <Input
                      id="neighborhood"
                      value={companyForm.neighborhood}
                      onChange={(e) => setCompanyForm({ ...companyForm, neighborhood: e.target.value })}
                      placeholder="Bela Vista"
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="city">Cidade</Label>
                    <Input
                      id="city"
                      value={companyForm.city}
                      onChange={(e) => setCompanyForm({ ...companyForm, city: e.target.value })}
                      placeholder="São Paulo"
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-1">
                    <Label htmlFor="state">UF (Estado)</Label>
                    <Input
                      id="state"
                      value={companyForm.state}
                      maxLength={2}
                      onChange={(e) => setCompanyForm({ ...companyForm, state: e.target.value.toUpperCase() })}
                      placeholder="SP"
                    />
                  </div>
                </div>
              </CardContent>
              <CardFooter className="flex justify-end border-t pt-4">
                <Button type="submit" disabled={updateCompanyMutation.isPending} className="gap-2">
                  {updateCompanyMutation.isPending ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Save className="size-4" />
                  )}
                  Salvar Dados da Empresa
                </Button>
              </CardFooter>
            </Card>
          </form>
        )}

        {/* 3. ABA WHITELABEL & MARCA */}
        {activeTab === 'whitelabel' && (
          <div className="grid gap-6 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <form onSubmit={handleSaveWhitelabel} className="flex flex-col gap-6">
                <Card className="rounded-2xl border shadow-xs">
                  <CardHeader className="pb-4">
                    <CardTitle className="text-lg flex items-center gap-2">
                      <Sparkles className="size-5 text-primary" />
                      Identidade Visual & Marca (Whitelabel)
                    </CardTitle>
                    <CardDescription>
                      Personalize o nome do sistema, logotipos, cor temática e informações de suporte.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="app_name">Nome da Aplicação / Sistema</Label>
                        <Input
                          id="app_name"
                          value={whitelabelForm.app_name}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, app_name: e.target.value })}
                          placeholder="Ex: Luxus Connect"
                          required
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="app_slogan">Slogan / Subtítulo</Label>
                        <Input
                          id="app_slogan"
                          value={whitelabelForm.app_slogan}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, app_slogan: e.target.value })}
                          placeholder="Ex: Gestão Inteligente de Telefonia"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="logo_url">URL do Logotipo Principal (PNG / SVG)</Label>
                        <Input
                          id="logo_url"
                          value={whitelabelForm.logo_url}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, logo_url: e.target.value })}
                          placeholder="https://exemplo.com/logo.png"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="favicon_url">URL do Ícone / Favicon</Label>
                        <Input
                          id="favicon_url"
                          value={whitelabelForm.favicon_url}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, favicon_url: e.target.value })}
                          placeholder="https://exemplo.com/favicon.ico"
                        />
                      </div>
                    </div>

                    {/* Seletor de Cor Primária */}
                    <div className="space-y-3 pt-2">
                      <Label htmlFor="primary_color">Cor Primária da Marca</Label>
                      <div className="flex flex-wrap items-center gap-3">
                        <div className="flex items-center gap-2 rounded-xl border p-1.5 bg-muted/30">
                          <input
                            type="color"
                            id="primary_color"
                            value={whitelabelForm.primary_color || '#0f766e'}
                            onChange={(e) => setWhitelabelForm({ ...whitelabelForm, primary_color: e.target.value })}
                            className="size-8 cursor-pointer rounded-lg border-0 bg-transparent"
                          />
                          <span className="font-mono text-xs font-semibold px-2">
                            {whitelabelForm.primary_color || '#0f766e'}
                          </span>
                        </div>

                        {/* Presets */}
                        <div className="flex flex-wrap gap-1.5 items-center">
                          {COLOR_PRESETS.map((preset) => (
                            <button
                              key={preset.color}
                              type="button"
                              onClick={() => setWhitelabelForm({ ...whitelabelForm, primary_color: preset.color })}
                              title={preset.name}
                              className={`size-7 rounded-full border-2 transition-transform hover:scale-110 ${
                                whitelabelForm.primary_color === preset.color ? 'border-foreground ring-2 ring-primary/40 scale-105' : 'border-background shadow-xs'
                              }`}
                              style={{ backgroundColor: preset.color }}
                            />
                          ))}
                        </div>
                      </div>
                    </div>

                    <Separator className="my-2" />

                    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="support_email">E-mail de Suporte ao Usuário</Label>
                        <Input
                          id="support_email"
                          type="email"
                          value={whitelabelForm.support_email}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, support_email: e.target.value })}
                          placeholder="suporte@suaempresa.com.br"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="support_phone">Telefone / WhatsApp de Suporte</Label>
                        <Input
                          id="support_phone"
                          value={whitelabelForm.support_phone}
                          onChange={(e) => setWhitelabelForm({ ...whitelabelForm, support_phone: e.target.value })}
                          placeholder="(11) 99999-9999"
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="footer_text">Texto de Rodapé / Direitos Autorais</Label>
                      <Input
                        id="footer_text"
                        value={whitelabelForm.footer_text}
                        onChange={(e) => setWhitelabelForm({ ...whitelabelForm, footer_text: e.target.value })}
                        placeholder="© 2026 Minha Empresa. Todos os direitos reservados."
                      />
                    </div>
                  </CardContent>
                  <CardFooter className="flex justify-end border-t pt-4">
                    <Button type="submit" disabled={updateWhitelabelMutation.isPending} className="gap-2">
                      {updateWhitelabelMutation.isPending ? (
                        <Loader2 className="size-4 animate-spin" />
                      ) : (
                        <Save className="size-4" />
                      )}
                      Salvar Identidade Visual
                    </Button>
                  </CardFooter>
                </Card>
              </form>
            </div>

            {/* Coluna Direita: Live Preview da Marca */}
            <div className="lg:col-span-1">
              <Card className="rounded-2xl border shadow-xs sticky top-24">
                <CardHeader className="pb-3">
                  <CardTitle className="text-base flex items-center gap-2">
                    <Eye className="size-4 text-primary" />
                    Pré-visualização da Marca
                  </CardTitle>
                  <CardDescription className="text-xs">
                    Veja como sua marca será exibida no cabeçalho e na barra lateral.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Mock Sidebar Header */}
                  <div className="rounded-xl border bg-muted/40 p-4">
                    <p className="text-muted-foreground text-[10px] uppercase font-semibold tracking-wider mb-2">
                      Sidebar Header
                    </p>
                    <div className="flex items-center gap-3 rounded-xl bg-background p-2.5 border shadow-xs">
                      <div
                        className="flex size-9 shrink-0 items-center justify-center rounded-xl text-white font-bold text-sm shadow-sm overflow-hidden"
                        style={{ backgroundColor: whitelabelForm.primary_color || '#0f766e' }}
                      >
                        {whitelabelForm.logo_url ? (
                          <img
                            src={whitelabelForm.logo_url}
                            alt="Logo"
                            className="size-full object-contain p-0.5"
                            onError={(e) => {
                              (e.target as HTMLElement).style.display = 'none';
                            }}
                          />
                        ) : (
                          <Phone className="size-4" />
                        )}
                      </div>
                      <div className="flex flex-col min-w-0">
                        <span className="truncate font-semibold text-sm">
                          {whitelabelForm.app_name || 'Luxus.Connect'}
                        </span>
                        <span className="truncate text-muted-foreground text-xs">
                          {whitelabelForm.app_slogan || 'Gestão de Telefonia'}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Mock Footer / Info */}
                  <div className="rounded-xl border bg-muted/40 p-4 space-y-2 text-xs">
                    <p className="text-muted-foreground text-[10px] uppercase font-semibold tracking-wider">
                      Canais de Suporte & Rodapé
                    </p>
                    <div className="flex justify-between py-0.5">
                      <span className="text-muted-foreground">Suporte:</span>
                      <span className="font-medium truncate max-w-40">{whitelabelForm.support_email || '—'}</span>
                    </div>
                    <div className="flex justify-between py-0.5">
                      <span className="text-muted-foreground">Telefone:</span>
                      <span className="font-medium">{whitelabelForm.support_phone || '—'}</span>
                    </div>
                    <p className="text-muted-foreground text-[11px] pt-2 border-t mt-2 text-center">
                      {whitelabelForm.footer_text || '© 2026. Todos os direitos reservados.'}
                    </p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        )}

        {/* 4. ABA PARÂMETROS DO SISTEMA & INTEGRAÇÕES */}
        {activeTab === 'system' && (
          <form onSubmit={handleSaveSystem} className="flex flex-col gap-6">
            {/* Parâmetros Financeiros */}
            <Card className="rounded-2xl border shadow-xs">
              <CardHeader className="pb-4">
                <CardTitle className="text-lg flex items-center gap-2">
                  <CreditCard className="size-5 text-primary" />
                  Regras de Faturamento e Cobrança
                </CardTitle>
                <CardDescription>
                  Defina o vencimento padrão, taxas de atraso e notificações enviadas aos clientes.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  <div className="space-y-2">
                    <Label htmlFor="default_due_day">Dia de Vencimento Padrão</Label>
                    <Input
                      id="default_due_day"
                      type="number"
                      min={1}
                      max={31}
                      value={systemForm.default_due_day}
                      onChange={(e) => setSystemForm({ ...systemForm, default_due_day: Number(e.target.value) || 10 })}
                      placeholder="Ex: 10"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="late_fee_percentage">Multa por Atraso (%)</Label>
                    <Input
                      id="late_fee_percentage"
                      type="number"
                      step="0.01"
                      min={0}
                      value={systemForm.late_fee_percentage}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, late_fee_percentage: Number(e.target.value) || 0 })
                      }
                      placeholder="Ex: 2.00"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="interest_rate_monthly">Juros ao Mês (%)</Label>
                    <Input
                      id="interest_rate_monthly"
                      type="number"
                      step="0.01"
                      min={0}
                      value={systemForm.interest_rate_monthly}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, interest_rate_monthly: Number(e.target.value) || 0 })
                      }
                      placeholder="Ex: 1.00"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 pt-2">
                  <div className="space-y-2">
                    <Label htmlFor="days_before">Aviso Prévio de Vencimento (Dias)</Label>
                    <Input
                      id="days_before"
                      type="number"
                      min={0}
                      value={systemForm.days_before_due_reminder}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, days_before_due_reminder: Number(e.target.value) || 0 })
                      }
                      placeholder="Ex: 3"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="days_after">Tolerância para Lembrete de Atraso (Dias)</Label>
                    <Input
                      id="days_after"
                      type="number"
                      min={0}
                      value={systemForm.days_after_due_reminder}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, days_after_due_reminder: Number(e.target.value) || 0 })
                      }
                      placeholder="Ex: 2"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="prorata_divisor">Divisor de pró-rata (dias)</Label>
                    <Input
                      id="prorata_divisor"
                      type="number"
                      min={1}
                      max={31}
                      value={systemForm.prorata_divisor ?? 30}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, prorata_divisor: Number(e.target.value) || 30 })
                      }
                    />
                    <p className="text-muted-foreground text-xs">
                      Padrão 30. Use 28 ou 31 conforme a competência.
                    </p>
                  </div>
                </div>

                <div className="space-y-3 pt-2">
                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="auto_invoice"
                      checked={systemForm.auto_send_invoice_email}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, auto_send_invoice_email: e.target.checked })
                      }
                      className="size-4 rounded border-gray-300 text-primary focus:ring-primary"
                    />
                    <Label htmlFor="auto_invoice" className="cursor-pointer font-medium text-sm">
                      Enviar faturas por e-mail automaticamente após a emissão do lote
                    </Label>
                  </div>

                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="auto_collection"
                      checked={systemForm.auto_send_collection_reminder}
                      onChange={(e) =>
                        setSystemForm({ ...systemForm, auto_send_collection_reminder: e.target.checked })
                      }
                      className="size-4 rounded border-gray-300 text-primary focus:ring-primary"
                    />
                    <Label htmlFor="auto_collection" className="cursor-pointer font-medium text-sm">
                      Disparar réguas de cobrança e lembretes automáticos para clientes em atraso
                    </Label>
                  </div>
                </div>
              </CardContent>
              <CardFooter className="flex justify-end border-t pt-4">
                <Button type="submit" disabled={updateSystemMutation.isPending} className="gap-2">
                  {updateSystemMutation.isPending ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Save className="size-4" />
                  )}
                  Salvar Parâmetros do Sistema
                </Button>
              </CardFooter>
            </Card>

            {/* Painel da Integração Bancária Sicredi */}
            <Card className="rounded-2xl border shadow-xs">
              <CardHeader className="flex flex-row items-center justify-between pb-4">
                <div className="space-y-1">
                  <CardTitle className="text-lg flex items-center gap-2">
                    <CreditCard className="size-5 text-green-600 dark:text-green-400" />
                    Integração Bancária Sicredi (Boleto & PIX)
                  </CardTitle>
                  <CardDescription>
                    Emissão de Boletos Híbridos com QR Code PIX dinâmico e conciliação automática via Webhook.
                  </CardDescription>
                </div>
                <Badge className="bg-green-600 text-white hover:bg-green-600">Conectado & Ativo</Badge>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 text-sm">
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Cooperativa / Agência</span>
                    <span className="font-semibold text-base">0179</span>
                  </div>
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Posto de Atendimento</span>
                    <span className="font-semibold text-base">14</span>
                  </div>
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Código do Beneficiário</span>
                    <span className="font-semibold text-base">01048</span>
                  </div>
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Conta Corrente</span>
                    <span className="font-semibold text-base">04133-5</span>
                  </div>
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Tipo de Cobrança</span>
                    <span className="font-semibold text-base">Híbrido (Boleto + PIX)</span>
                  </div>
                  <div className="rounded-xl bg-muted/40 p-3">
                    <span className="text-muted-foreground text-xs block">Webhook de Notificação</span>
                    <span className="font-semibold text-base text-green-600 dark:text-green-400">Ativo em Produção</span>
                  </div>
                </div>
              </CardContent>
            </Card>
            {authRoles.isMaster ? (
              <Card className="rounded-2xl border shadow-xs">
                <CardHeader>
                  <CardTitle className="text-lg">Exportação organizacional</CardTitle>
                  <CardDescription>
                    Dump master-only com checksum SHA-256 auditado. Use apenas para backup controlado.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={exportOrg.isPending}
                    onClick={() =>
                      exportOrg.mutate(undefined, {
                        onSuccess: (res) =>
                          toast.success(
                            `Exportação gerada. SHA-256 ${res.checksum_sha256 ?? '—'}`
                          ),
                        onError: (e) =>
                          toast.error(isApiHttpError(e) ? e.message : getErrorMessage(e))
                      })
                    }
                  >
                    {exportOrg.isPending ? 'Exportando…' : 'Gerar dump (SHA-256)'}
                  </Button>
                </CardContent>
              </Card>
            ) : null}
          </form>
        )}

        {activeTab === 'webhooks' && authRoles.canManageUsers && <WebhooksSettingsPanel />}

        {/* 5. ABA GESTÃO DE USUÁRIOS */}
        {activeTab === 'users' && authRoles.canManageUsers && (
          <Card className="rounded-2xl border shadow-xs">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Users className="size-5 text-primary" />
                Gestão de Usuários e Permissões
              </CardTitle>
              <CardDescription>
                Crie novos acessos, redefina senhas de colaboradores e gerencie permissões de Administrador, Financeiro,
                Operações e Parceiros.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4 rounded-xl border bg-muted/30 p-6">
                <div>
                  <h4 className="font-semibold">Módulo de Usuários da Organização</h4>
                  <p className="text-muted-foreground text-sm">
                    Acesse o painel centralizado para gerenciar todos os colaboradores do sistema.
                  </p>
                </div>
                <Button render={<Link to="/users" />} className="gap-2 shrink-0">
                  <UserCog className="size-4" />
                  Abrir Gestão de Usuários
                </Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </PageWrapper>
  );
}
