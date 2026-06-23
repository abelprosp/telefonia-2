import { useEffect, useState } from 'react';

import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger
} from '@/components/ui/collapsible';
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem
} from '@/components/ui/sidebar';
import { cn } from '@/lib/utils';

export const NavMain = ({
  label,
  items
}: {
  label: string;
  items: {
    title: string;
    url: string;
    icon?: React.ReactNode;
    isActive?: boolean;
    badge?: string;
    items?: {
      title: string;
      url: string;
    }[];
  }[];
}) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  const isPathActive = (url: string) => {
    if (url === '/') return pathname === '/';
    return pathname === url || pathname.startsWith(`${url}/`);
  };

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="text-muted-foreground px-3 text-xs font-semibold tracking-wide uppercase">
        {label}
      </SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => {
          const active = item.isActive ?? isPathActive(item.url);

          return item.items ? (
            <NavCollapsibleItem
              key={item.title}
              item={item}
              active={active}
              isPathActive={isPathActive}
            />
          ) : (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                isActive={active}
                className={cn(active && 'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground')}
                render={<Link to={item.url} />}
              >
                {item.icon}
                <span>{item.title}</span>
                {item.badge ? (
                  <SidebarMenuBadge className="bg-primary text-primary-foreground">
                    {item.badge}
                  </SidebarMenuBadge>
                ) : null}
              </SidebarMenuButton>
            </SidebarMenuItem>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
};

function NavCollapsibleItem({
  item,
  active,
  isPathActive
}: {
  item: {
    title: string;
    url: string;
    icon?: React.ReactNode;
    badge?: string;
    items?: { title: string; url: string }[];
  };
  active: boolean;
  isPathActive: (url: string) => boolean;
}) {
  const [open, setOpen] = useState(active);

  useEffect(() => {
    if (active) {
      setOpen(true);
    }
  }, [active]);

  return (
    <Collapsible
      open={open}
      onOpenChange={setOpen}
      className="group/collapsible"
      render={<SidebarMenuItem />}
    >
      <CollapsibleTrigger
        render={
          <SidebarMenuButton
            tooltip={item.title}
            isActive={active}
            className={cn(active && 'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground')}
          />
        }
      >
        {item.icon}
        <span>{item.title}</span>
        {item.badge ? (
          <SidebarMenuBadge className="bg-primary text-primary-foreground">
            {item.badge}
          </SidebarMenuBadge>
        ) : null}
        <ChevronRight className="ml-auto transition-transform duration-200 group-data-open/collapsible:rotate-90" />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <SidebarMenuSub>
          {item.items?.map((subItem) => (
            <SidebarMenuSubItem key={subItem.title}>
              <SidebarMenuSubButton
                isActive={isPathActive(subItem.url)}
                render={<Link to={subItem.url} />}
              >
                <span>{subItem.title}</span>
              </SidebarMenuSubButton>
            </SidebarMenuSubItem>
          ))}
        </SidebarMenuSub>
      </CollapsibleContent>
    </Collapsible>
  );
}
