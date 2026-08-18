import { useEffect, useMemo, type ComponentProps, type ReactNode } from 'react';

import { Link, useLocation, useNavigate } from '@tanstack/react-router';
import {
  Building2,
  ClipboardCheck,
  ClipboardList,
  FileText,
  Layers,
  LayoutDashboard,
  LifeBuoy,
  LogOut,
  Phone,
  Receipt,
  ShoppingCart,
  Settings,
  Sparkles,
  TrendingUp,
  UserCog,
  Users,
  Wallet
} from 'lucide-react';
import { useAuth } from 'react-oidc-context';

import { NavMain } from '@/components/nav-main';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail
} from '@/components/ui/sidebar';
import { useAuthRoles } from '@/lib/auth-roles';
import { performLogout } from '@/lib/auth-actions';
import { useWhitelabel } from '@/providers/whitelabel-provider';

type MenuItem = {
  title: string;
  url: string;
  icon: ReactNode;
  items?: { title: string; url: string }[];
};

const operationalMenuItems: MenuItem[] = [
  {
    title: 'Dashboard',
    url: '/',
    icon: <LayoutDashboard />
  },
  {
    title: 'Cadastros',
    url: '/customers',
    icon: <Layers />,
    items: [
      { title: 'Clientes', url: '/customers' },
      { title: 'Operadoras', url: '/providers' },
      { title: 'Linhas telefônicas', url: '/phone-lines' },
      { title: 'Estoque de linhas', url: '/stock' },
      { title: 'Estoque de aparelhos', url: '/stock/devices' }
    ]
  },
  {
    title: 'Faturamento',
    url: '/invoices',
    icon: <FileText />,
    items: [
      { title: 'Faturas', url: '/invoices' },
      { title: 'Meses de processamento', url: '/processing-months' },
      { title: 'Ciclos de faturamento', url: '/billing-cycles' },
      { title: 'Termos de excedente', url: '/exceedance-terms' }
    ]
  },
  {
    title: 'Operação avançada',
    url: '/tickets',
    icon: <LifeBuoy />,
    items: [
      { title: 'Tickets', url: '/tickets' },
      { title: 'Divergências', url: '/divergences' },
      { title: 'Aprovações', url: '/approvals' }
    ]
  },
  {
    title: 'Vendas',
    url: '/sales',
    icon: <ShoppingCart />,
    items: [
      { title: 'Vendas', url: '/sales' },
      { title: 'Templates de contrato', url: '/contract-templates' }
    ]
  },
  {
    title: 'Relatórios',
    url: '/reports/transition-pending',
    icon: <TrendingUp />,
    items: [
      { title: 'Movimentação de linhas', url: '/reports/transition-pending' },
      { title: 'Solicitações de parceiros', url: '/line-requests' },
      { title: 'Resumo financeiro', url: '/finance' },
      { title: 'Rentabilidade', url: '/finance' }
    ]
  }
];

const financialMenuItem: MenuItem = {
  title: 'Financeiro',
  url: '/finance',
  icon: <Wallet />,
  items: [
    { title: 'Visão geral', url: '/finance' },
    { title: 'Contas a pagar', url: '/finance/payables' },
    { title: 'Contas a receber', url: '/finance/receivables' },
    { title: 'Faturas para envio', url: '/finance/customer-invoices' },
    { title: 'Inadimplentes', url: '/finance/collections' },
    { title: 'Templates de e-mail', url: '/finance/invoice-email-templates' },
    { title: 'Layouts de fatura', url: '/finance/invoice-layout-templates' },
    { title: 'Vendas de parceiros', url: '/finance/partner-sales' }
  ]
};

const usersMenuItem: MenuItem = {
  title: 'Usuários',
  url: '/users',
  icon: <UserCog />
};

const adminProductItems: MenuItem[] = [
  { title: 'Clientes', url: '/customers', icon: <Users /> },
  { title: 'Operadoras', url: '/providers', icon: <Building2 /> },
  { title: 'Linhas', url: '/phone-lines', icon: <Phone /> },
  { title: 'Faturas', url: '/invoices', icon: <Receipt /> }
];

const portalMenuItems: MenuItem[] = [
  { title: 'Meu portal', url: '/portal', icon: <LayoutDashboard /> }
];

const partnerMenuItems: MenuItem[] = [
  { title: 'Resumo', url: '/partner', icon: <LayoutDashboard /> },
  { title: 'Vendas', url: '/partner/commercial-sales', icon: <ShoppingCart /> },
  { title: 'Financeiro', url: '/partner/financial', icon: <Wallet /> },
  { title: 'Clientes', url: '/partner/customers', icon: <Users /> },
  { title: 'Linhas', url: '/partner/phone-lines', icon: <Phone /> },
  { title: 'Solicitações', url: '/partner/requests', icon: <ClipboardList /> }
];

function buildStaffMenu(
  canAccessOperations: boolean,
  canAccessFinance: boolean,
  canManageUsers: boolean
): MenuItem[] {
  const items: MenuItem[] = [];

  if (canAccessOperations) {
    items.push(...operationalMenuItems);
  }

  if (canAccessFinance) {
    items.push(financialMenuItem);
  }

  if (canManageUsers) {
    items.push(usersMenuItem);
  }

  if (!canAccessOperations && canAccessFinance) {
    return [financialMenuItem];
  }

  return items;
}

export const LayoutSidebar = ({ ...props }: ComponentProps<typeof Sidebar>) => {
  const auth = useAuth();
  const { user } = auth;
  const { settings: whitelabel } = useWhitelabel();
  const {
    isMaster,
    isPartnerOnly,
    isCustomerPortal,
    canAccessOperations,
    canAccessFinance,
    canManageUsers
  } = useAuthRoles();
  const location = useLocation();
  const navigate = useNavigate();

  const menuItems = useMemo(() => {
    if (isCustomerPortal) {
      return portalMenuItems;
    }
    if (isPartnerOnly) {
      return partnerMenuItems;
    }
    return buildStaffMenu(canAccessOperations, canAccessFinance, canManageUsers);
  }, [isCustomerPortal, isPartnerOnly, canAccessOperations, canAccessFinance, canManageUsers]);

  useEffect(() => {
    const path = location.pathname;

    if (isCustomerPortal) {
      if (!path.startsWith('/portal') && path !== '/settings') {
        void navigate({ to: '/portal' });
      }
      return;
    }

    if (isPartnerOnly) {
      if (!path.startsWith('/partner') && path !== '/settings') {
        void navigate({ to: '/partner' });
      }
      return;
    }

    if (canAccessFinance && !canAccessOperations) {
      if (!path.startsWith('/finance') && path !== '/settings') {
        void navigate({ to: '/finance' });
      }
      return;
    }

    if (canAccessOperations && !canAccessFinance) {
      if (path.startsWith('/finance') || path.startsWith('/users')) {
        void navigate({ to: '/' });
      }
    }
  }, [
    isCustomerPortal,
    isPartnerOnly,
    canAccessOperations,
    canAccessFinance,
    location.pathname,
    navigate
  ]);

  if (!user) {
    return null;
  }

  const onSignout = async () => {
    await performLogout(auth);
  };

  const homeTo = isCustomerPortal
    ? '/portal'
    : isPartnerOnly
      ? '/partner'
      : canAccessFinance && !canAccessOperations
        ? '/finance'
        : '/';

  const portalLabel = isCustomerPortal
    ? 'Portal do cliente'
    : isPartnerOnly
      ? 'Portal do parceiro'
      : canAccessFinance && !canAccessOperations
        ? 'Financeiro'
        : (whitelabel.app_slogan || 'Gestão de telefonia');

  const appDisplayName = whitelabel.app_name || 'Luxus.Connect';

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader className="px-3 py-4">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="hover:bg-transparent"
              render={<Link to={homeTo} />}
            >
              <div className="bg-primary text-primary-foreground flex aspect-square size-9 items-center justify-center rounded-xl shadow-sm overflow-hidden">
                {whitelabel.logo_url ? (
                  <img
                    src={whitelabel.logo_url}
                    alt={appDisplayName}
                    className="size-full object-contain p-0.5"
                    onError={(e) => {
                      // Fallback se a imagem quebrar
                      (e.target as HTMLElement).style.display = 'none';
                    }}
                  />
                ) : (
                  <Phone className="size-4" />
                )}
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate text-base font-semibold">{appDisplayName}</span>
                <span className="text-muted-foreground truncate text-xs">{portalLabel}</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>


      <SidebarContent className="gap-0">
        <NavMain label={isPartnerOnly ? 'Parceiro' : 'Navegação'} items={menuItems} />
        {isMaster && !isPartnerOnly ? (
          <NavMain label="Cadastros rápidos" items={adminProductItems} />
        ) : null}
        {canAccessOperations && !isPartnerOnly ? (
          <NavMain
            label="Operações"
            items={[
              {
                title: 'Solicitações',
                url: '/line-requests',
                icon: <ClipboardCheck />
              }
            ]}
          />
        ) : null}

        {/* Card Inferior de Faturamento */}
        {!isCustomerPortal ? (
          <div className="p-3">
            <div className="flex flex-col gap-2 rounded-2xl border bg-muted/40 p-4 text-foreground shadow-xs">
              <div className="flex items-center gap-2">
                <div className="bg-primary text-primary-foreground flex size-7 items-center justify-center rounded-lg">
                  <Sparkles className="size-3.5" />
                </div>
                <span className="text-xs font-bold tracking-tight uppercase">Ciclo Vigente</span>
              </div>
              <p className="text-muted-foreground text-[11px] leading-snug">
                Acompanhe e concilie faturas de operadoras em aberto.
              </p>
              <Link
                to="/invoices"
                search={{ page: 1, pageSize: 10, processingMonthId: undefined }}
                className="bg-primary text-primary-foreground hover:bg-primary/90 mt-1 flex items-center justify-center rounded-xl py-2 text-xs font-semibold shadow-xs transition-colors"
              >
                Faturamento
              </Link>
            </div>
          </div>
        ) : null}
      </SidebarContent>

      <SidebarFooter className="gap-2 border-t p-3">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton render={<Link to="/settings" />}>
              <Settings className="size-4" />
              <span>Configurações</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton
              onClick={onSignout}
              className="text-destructive hover:text-destructive"
            >
              <LogOut className="size-4" />
              <span>Sair</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
};
