import type { ReactNode } from 'react';

type ListPageHeaderProps = {
  title: string;
  description: string;
  action?: ReactNode;
};

export function ListPageHeader({
  title,
  description,
  action
}: ListPageHeaderProps) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h2 className="text-lg font-semibold">{title}</h2>
        <p className="text-muted-foreground text-sm">{description}</p>
      </div>
      {action ? (
        <div className="flex shrink-0 flex-wrap gap-2">{action}</div>
      ) : null}
    </div>
  );
}
