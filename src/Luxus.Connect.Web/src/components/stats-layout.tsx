import type { ReactNode } from 'react';

import { Link } from '@tanstack/react-router';
import type { LucideIcon } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import { cn } from '@/lib/utils';

export type StatsTileVariant =
  | 'default'
  | 'blue'
  | 'green'
  | 'red'
  | 'amber'
  | 'purple';

const tileVariantClass: Record<
  StatsTileVariant,
  { surface: string; icon: string }
> = {
  blue: {
    surface:
      'rounded-xl border border-blue-200 bg-blue-50 dark:border-blue-400/30 dark:bg-blue-500/15',
    icon: 'text-blue-500'
  },
  green: {
    surface:
      'rounded-xl border border-green-600/30 bg-green-50 dark:border-green-500/30 dark:bg-green-500/15',
    icon: 'text-green-600'
  },
  red: {
    surface:
      'rounded-xl border border-red-200 bg-red-50 dark:border-red-400/30 dark:bg-red-400/15',
    icon: 'text-red-500'
  },
  amber: {
    surface:
      'rounded-xl border border-amber-600/30 bg-amber-50 dark:border-amber-500/30 dark:bg-amber-500/15',
    icon: 'text-amber-600'
  },
  purple: {
    surface:
      'rounded-xl border border-purple-200 bg-purple-50 dark:border-purple-400/30 dark:bg-purple-400/15',
    icon: 'text-purple-500'
  },
  default: {
    surface: '',
    icon: ''
  }
};

export type StatsHeroProps = {
  title: ReactNode;
  subtitle?: ReactNode;
  align?: 'center' | 'left';
  className?: string;
};

export const StatsHero = ({
  title,
  subtitle,
  align = 'center',
  className
}: StatsHeroProps) => {
  const alignTitle =
    align === 'center'
      ? 'text-balance text-center font-semibold text-3xl tracking-tight sm:text-4xl md:text-5xl'
      : 'text-balance text-left font-semibold text-2xl tracking-tight sm:text-3xl';

  const alignSub =
    align === 'center'
      ? 'mt-4 text-center text-lg text-muted-foreground sm:text-xl md:text-2xl'
      : 'mt-2 text-left text-sm text-muted-foreground sm:text-base';

  return (
    <div className={cn(className)}>
      <h2 className={alignTitle}>{title}</h2>
      {subtitle ? <p className={alignSub}>{subtitle}</p> : null}
    </div>
  );
};

export const StatsMetricGrid = ({
  children,
  className
}: {
  children: ReactNode;
  className?: string;
}) => {
  return (
    <div
      className={cn(
        'grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3',
        className
      )}
    >
      {children}
    </div>
  );
};

export type StatsMetricCardProps = {
  variant: StatsTileVariant;
  icon: LucideIcon;
  value: string | number;
  title: string;
  description?: string;
  featured?: boolean;
  footer?: ReactNode;
  to?: string;
  search?: Record<string, unknown>;
};

export const StatsMetricCard = ({
  variant,
  icon,
  value,
  title,
  description,
  featured,
  footer,
  to,
  search
}: StatsMetricCardProps) => {
  const v = tileVariantClass[variant];
  const Icon = icon;
  const boxClass = cn(
    v.surface,
    'p-6 py-7',
    featured && 'row-span-2 flex flex-col overflow-hidden pb-0',
    to &&
      'transition-opacity hover:opacity-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
  );

  const content = (
    <>
      <Icon className={cn('mb-7 h-10 w-10', v.icon)} />
      <span className="text-5xl font-semibold tabular-nums">{value}</span>
      <p className={cn('mt-4 text-lg', featured && 'mb-2')}>{title}</p>
      {description ? (
        <p className="text-muted-foreground mt-1 text-sm leading-snug">
          {description}
        </p>
      ) : null}
      {footer}
    </>
  );

  if (to) {
    return (
      <Link
        to={to}
        {...(search !== undefined ? { search } : {})}
        className={cn(boxClass, 'block')}
      >
        {content}
      </Link>
    );
  }

  return <div className={boxClass}>{content}</div>;
};

export type StatsPanelProps = {
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
  variant?: StatsTileVariant;
};

/**
 * Painel no estilo dos tiles de stats (borda + fundo suave), para listas no dashboard.
 */
export const StatsPanel = ({
  title,
  description,
  children,
  variant = 'default',
  className
}: StatsPanelProps) => {
  const v = tileVariantClass[variant];
  return (
    <Card className={cn(v.surface, 'overflow-hidden', className)}>
      <CardHeader>
        <CardTitle className="text-lg font-semibold">{title}</CardTitle>
        {description && (
          <CardDescription className="text-md">{description}</CardDescription>
        )}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
};
