import { type ReactNode } from 'react';

import { PageHeader, type PageHeaderBreadcrumb } from './page-header';

type PageWrapperProps = {
  breadcrumbs: PageHeaderBreadcrumb[];
  children?: ReactNode;
};

export const PageWrapper = ({ breadcrumbs, children }: PageWrapperProps) => {
  return (
    <>
      <PageHeader breadcrumbs={breadcrumbs} />
      <div className="@container/main flex flex-1 flex-col gap-4 p-6 md:gap-6 md:p-6">
        {children}
      </div>
    </>
  );
};
