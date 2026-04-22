import { type ComponentProps } from 'react';

import {
  Building2,
  Calendar,
  CalendarDays,
  FileText,
  Layers,
  LayoutDashboard,
  PackageX,
  Phone,
  Receipt,
  Settings,
  TrendingUp,
  UserRound,
  Users
} from 'lucide-react';
import { useAuth } from 'react-oidc-context';

import { NavMain } from '@/components/nav-main';
import { NavSecondary } from '@/components/nav-secondary';
import { NavUser } from '@/components/nav-user';
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

const getOrganization = (
  organizationClaim: Record<string, { id: string; name: string[] }>
) => {
  const orgEntries = Object.entries(organizationClaim);

  if (orgEntries.length === 0) {
    throw new Error('Organization claim is missing.');
  }

  const [alias, orgMeta] = orgEntries[0];

  return { alias, name: orgMeta.name[0] };
};

const data = {
  navMain: [
    {
      title: 'Dashboard',
      url: '/',
      icon: <LayoutDashboard />,
      isActive: true
    },
    {
      title: 'Cadastros',
      url: '#',
      icon: <Layers />,
      isActive: true,
      items: [
        {
          title: 'Clientes',
          url: '/customers',
          icon: <Users />
        },
        {
          title: 'Operadoras',
          url: '/providers',
          icon: <Building2 />
        },
        {
          title: 'Linhas telefônicas',
          url: '/phone-lines',
          icon: <Phone />
        },
        {
          title: 'Estoque de linhas',
          url: '/stock',
          icon: <PackageX />
        }
      ]
    },
    {
      title: 'Faturamento',
      url: '#',
      icon: <FileText />,
      isActive: true,
      items: [
        {
          title: 'Faturas',
          url: '/invoices',
          icon: <Receipt />
        },
        {
          title: 'Meses de processamento',
          url: '/processing-months',
          icon: <CalendarDays />
        },
        {
          title: 'Ciclos de faturamento',
          url: '/billing-cycles',
          icon: <Calendar />
        }
      ]
    },
    {
      title: 'Relatórios',
      url: '#',
      icon: <TrendingUp />,
      isActive: true,
      items: [
        {
          title: 'Linhas em transição',
          url: '/reports/transition-pending'
        }
      ]
    }
  ],
  navSecondary: [
    {
      title: 'Configurações',
      url: '/settings',
      icon: <Settings />
    }
  ]
};

export const LayoutSidebar = ({ ...props }: ComponentProps<typeof Sidebar>) => {
  const { user } = useAuth();

  if (!user) {
    return null;
  }

  const useInfo = {
    name: user.profile.name!,
    email: user.profile.email!,
    avatar: user.profile.picture!,
    organization: getOrganization(user.profile['organization'] as any)
  };

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              render={
                <a href="#">
                  <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                    <UserRound className="size-4" />
                  </div>
                  <div className="ml-1 grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">
                      {useInfo.organization.name}
                    </span>
                    <span className="text-muted-foreground truncate text-xs">
                      {useInfo.organization.alias}
                    </span>
                  </div>
                </a>
              }
            ></SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={useInfo} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
};
