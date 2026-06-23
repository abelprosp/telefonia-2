import {
  Bell,
  ChevronDown,
  HelpCircle,
  LogOut,
  Search,
  UserCircle
} from 'lucide-react';
import { useAuth } from 'react-oidc-context';

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { roleLabel, useAuthRoles } from '@/lib/auth-roles';

export const AppTopBar = () => {
  const { user, removeUser, signoutSilent } = useAuth();
  const authRoles = useAuthRoles();

  const onSignout = async () => {
    await signoutSilent();
    removeUser();
  };

  const displayName = user?.profile.name ?? 'Usuário';
  const email = user?.profile.email ?? '';
  const avatar = user?.profile.picture;
  const initials = displayName
    .split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  return (
    <header className="bg-background/80 sticky top-0 z-20 flex h-16 shrink-0 items-center gap-3 border-b px-4 backdrop-blur md:px-6">
      <SidebarTrigger className="-ml-1 md:hidden" />

      <div className="relative mx-auto hidden w-full max-w-xl md:block">
        <Search className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />
        <Input
          placeholder="Pesquisar..."
          className="bg-muted/40 h-10 rounded-full border-0 pr-14 pl-10 shadow-none"
        />
        <kbd className="text-muted-foreground pointer-events-none absolute top-1/2 right-3 hidden -translate-y-1/2 items-center gap-1 rounded-md border bg-background px-1.5 py-0.5 text-[10px] font-medium sm:flex">
          ⌘ K
        </kbd>
      </div>

      <div className="ml-auto flex items-center gap-1">
        <Button variant="ghost" size="icon" className="rounded-full">
          <Bell className="size-4" />
        </Button>
        <Button variant="ghost" size="icon" className="rounded-full">
          <HelpCircle className="size-4" />
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                variant="ghost"
                className="h-10 gap-2 rounded-full px-2 hover:bg-muted"
              />
            }
          >
            <Avatar className="size-8">
              <AvatarImage src={avatar} alt={displayName} />
              <AvatarFallback>{initials}</AvatarFallback>
            </Avatar>
            <div className="hidden text-left text-sm leading-tight md:grid">
              <span className="max-w-32 truncate font-medium">{displayName}</span>
              <span className="text-muted-foreground max-w-32 truncate text-xs">
                {roleLabel(authRoles)}
              </span>
            </div>
            <ChevronDown className="text-muted-foreground hidden size-4 md:block" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="min-w-56 rounded-xl">
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col gap-0.5">
                <span className="font-medium">{displayName}</span>
                <span className="text-muted-foreground text-xs">{email}</span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <UserCircle />
              Conta
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onSignout} className="text-primary">
              <LogOut />
              Sair
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
};
