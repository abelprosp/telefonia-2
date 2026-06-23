import { useEffect, useMemo, type ComponentProps, type ReactNode } from 'react';

import { Link, useLocation, useNavigate } from '@tanstack/react-router';
import {
  Building2,
  ClipboardCheck,
  ClipboardList,
  FileText,
  Layers,
  LayoutDashboard,
  LogOut,
  Phone,
  Receipt,
  ShoppingCart,
  Settings,
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
      { title: 'Estoque de linhas', url: '/stock' }
    ]
  },
  {
    title: 'Faturamento',
    url: '/invoices',
    icon: <FileText />,
    items: [
      { title: 'Faturas', url: '/invoices' },
      { title: 'Meses de processamento', url: '/processing-months' },
      { title: 'Ciclos de faturamento', url: '/billing-cycles' }
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
      { title: 'Linhas em transição', url: '/reports/transition-pending' },
      { title: 'Solicitações de parceiros', url: '/line-requests' }
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
  const { user, removeUser, signoutSilent } = useAuth();
  const {
    isMaster,
    isPartnerOnly,
    canAccessOperations,
    canAccessFinance,
    canManageUsers
  } = useAuthRoles();
  const location = useLocation();
  const navigate = useNavigate();

  const menuItems = useMemo(() => {
    if (isPartnerOnly) {
      return partnerMenuItems;
    }
    return buildStaffMenu(canAccessOperations, canAccessFinance, canManageUsers);
  }, [isPartnerOnly, canAccessOperations, canAccessFinance, canManageUsers]);

  useEffect(() => {
    const path = location.pathname;

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
    await signoutSilent();
    removeUser();
  };

  const homeTo = isPartnerOnly
    ? '/partner'
    : canAccessFinance && !canAccessOperations
      ? '/finance'
      : '/';

  const portalLabel = isPartnerOnly
    ? 'Portal do parceiro'
    : canAccessFinance && !canAccessOperations
      ? 'Financeiro'
      : 'Gestão de telefonia';

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
              <div className="bg-primary text-primary-foreground flex aspect-square size-9 items-center justify-center rounded-xl shadow-sm">
                <Phone className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate text-base font-semibold">Luxus.Connect</span>
                <span className="text-muted-foreground truncate text-xs">{portalLabel}</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="gap-0">
        <NavMain label={isPartnerOnly ? 'Parceiro' : 'Menu'} items={menuItems} />
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
      </SidebarContent>

      <SidebarFooter className="gap-2">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton render={<Link to="/settings" />}>
              <Settings />
              <span>Configurações</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton
              onClick={onSignout}
              className="text-primary hover:text-primary"
            >
              <LogOut />
              <span>Sair</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
};
