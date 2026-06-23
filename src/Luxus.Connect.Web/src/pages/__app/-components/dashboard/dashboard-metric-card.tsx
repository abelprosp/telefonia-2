import type { LucideIcon } from 'lucide-react';
import { Link } from '@tanstack/react-router';

export type DashboardMetricCardProps = {
  title: string;
  value: string | number;
  icon: LucideIcon;
  to?: string;
  search?: Record<string, unknown>;
};

export const DashboardMetricCard = ({
  title,
  value,
  icon: Icon,
  to,
  search
}: DashboardMetricCardProps) => {
  const content = (
    <div className="dashboard-card group relative overflow-hidden p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-3">
          <p className="text-muted-foreground text-sm font-medium">{title}</p>
          <p className="text-3xl font-semibold tracking-tight tabular-nums">
            {value}
          </p>
        </div>
        <div className="bg-primary/10 text-primary flex size-10 shrink-0 items-center justify-center rounded-xl">
          <Icon className="size-5" />
        </div>
      </div>
    </div>
  );

  if (to) {
    return (
      <Link
        to={to}
        {...(search !== undefined ? { search } : {})}
        className="block transition-transform hover:-translate-y-0.5"
      >
        {content}
      </Link>
    );
  }

  return content;
};
