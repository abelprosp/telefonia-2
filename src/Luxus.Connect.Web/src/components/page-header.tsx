import { Fragment } from 'react';

import { Link } from '@tanstack/react-router';

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';

export type PageHeaderBreadcrumb = {
  label: string;
  to?: string;
  /** Opcional: preserva query da rota (ex.: filtro da lista). */
  search?: Record<string, unknown>;
};

type PageHeaderProps = {
  breadcrumbs: PageHeaderBreadcrumb[];
};

export const PageHeader = ({ breadcrumbs }: PageHeaderProps) => {
  if (breadcrumbs.length === 0) {
    throw new Error('AppPageShell requires at least one breadcrumb');
  }

  const lastIndex = breadcrumbs.length - 1;

  return (
    <header className="flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 my-auto data-[orientation=vertical]:h-4"
        />
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.map((crumb, i) => {
              const isLast = i === lastIndex;
              const hideOnMobile =
                breadcrumbs.length > 1 && !isLast
                  ? 'hidden md:inline-flex'
                  : undefined;

              return (
                <Fragment key={`${crumb.label}-${i}`}>
                  {i > 0 && <BreadcrumbSeparator className="hidden md:block" />}
                  <BreadcrumbItem className={hideOnMobile}>
                    {isLast ? (
                      <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                    ) : crumb.to ? (
                      <BreadcrumbLink
                        render={
                          <Link
                            to={crumb.to}
                            {...(crumb.search !== undefined
                              ? { search: crumb.search }
                              : {})}
                          />
                        }
                      >
                        {crumb.label}
                      </BreadcrumbLink>
                    ) : (
                      <span className="text-muted-foreground">
                        {crumb.label}
                      </span>
                    )}
                  </BreadcrumbItem>
                </Fragment>
              );
            })}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
    </header>
  );
};
