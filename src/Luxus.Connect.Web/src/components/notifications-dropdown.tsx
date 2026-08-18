import { Link } from '@tanstack/react-router';
import {
  AlertTriangle,
  Bell,
  Check,
  CheckCheck,
  CheckCircle2,
  ExternalLink,
  Info,
  ShieldAlert,
  Sparkles
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import {
  useGetNotifications,
  useMarkAllNotificationsRead,
  useMarkNotificationRead,
  type InAppNotification
} from '@/lib/notifications-api';

export const NotificationsDropdown = () => {
  const { data } = useGetNotifications();
  const markReadMutation = useMarkNotificationRead();
  const markAllReadMutation = useMarkAllNotificationsRead();

  const notifications = data?.items ?? [];
  const unreadCount = data?.unread_count ?? 0;

  const handleMarkAsRead = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    markReadMutation.mutate(id);
  };

  const handleMarkAllRead = () => {
    markAllReadMutation.mutate();
  };

  const renderIcon = (notif: InAppNotification) => {
    switch (notif.severity) {
      case 'warning':
        return <AlertTriangle className="size-4 text-amber-500" />;
      case 'critical':
        return <ShieldAlert className="size-4 text-rose-500" />;
      case 'success':
        return <CheckCircle2 className="size-4 text-emerald-500" />;
      default:
        return <Info className="size-4 text-sky-500" />;
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <button
            type="button"
            className="relative flex size-9 items-center justify-center rounded-full hover:bg-muted cursor-pointer transition-colors outline-hidden"
            aria-label="Notificações"
          />
        }
      >
        <Bell className="size-4 text-foreground" />
        {unreadCount > 0 ? (
          <span className="bg-primary text-primary-foreground absolute -top-0.5 -right-0.5 flex size-4.5 items-center justify-center rounded-full text-[10px] font-bold shadow-xs animate-in zoom-in-50">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        ) : null}
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-80 sm:w-96 rounded-2xl p-0 shadow-xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div className="flex items-center gap-2">
            <h4 className="text-sm font-bold text-foreground">Notificações</h4>
            {unreadCount > 0 ? (
              <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary">
                {unreadCount} nova{unreadCount > 1 ? 's' : ''}
              </span>
            ) : null}
          </div>

          {unreadCount > 0 ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleMarkAllRead}
              className="h-7 text-xs text-muted-foreground hover:text-foreground"
            >
              <CheckCheck className="mr-1 size-3.5" />
              Marcar todas
            </Button>
          ) : null}
        </div>

        <DropdownMenuGroup className="max-h-80 overflow-y-auto p-1 divide-y divide-border/40">
          {notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-6 text-center">
              <div className="bg-muted flex size-10 items-center justify-center rounded-full text-muted-foreground mb-2">
                <Sparkles className="size-5" />
              </div>
              <p className="text-sm font-semibold text-foreground">Tudo em dia!</p>
              <p className="text-xs text-muted-foreground mt-0.5">
                Não há notificações pendentes no momento.
              </p>
            </div>
          ) : (
            notifications.map((notif) => (
              <div
                key={notif.id}
                className={`flex items-start gap-3 p-3 transition-colors ${
                  notif.is_read ? 'opacity-70 bg-transparent' : 'bg-muted/30'
                }`}
              >
                <div className="mt-0.5 shrink-0">{renderIcon(notif)}</div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-1">
                    <p className="text-xs font-bold text-foreground truncate">
                      {notif.title}
                    </p>
                    {!notif.is_read ? (
                      <button
                        type="button"
                        onClick={(e) => handleMarkAsRead(notif.id, e)}
                        title="Marcar como lida"
                        className="text-muted-foreground hover:text-foreground p-0.5"
                      >
                        <Check className="size-3.5" />
                      </button>
                    ) : null}
                  </div>

                  <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2 leading-relaxed">
                    {notif.description}
                  </p>

                  {notif.action_url ? (
                    <div className="mt-2">
                      <Link
                        to={notif.action_url}
                        onClick={() => {
                          if (!notif.is_read) {
                            markReadMutation.mutate(notif.id);
                          }
                        }}
                        className="inline-flex items-center gap-1 text-[11px] font-semibold text-primary hover:underline"
                      >
                        <span>{notif.action_label || 'Ver detalhes'}</span>
                        <ExternalLink className="size-3" />
                      </Link>
                    </div>
                  ) : null}
                </div>
              </div>
            ))
          )}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
