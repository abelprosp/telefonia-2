import { HeadContent, Outlet, createRootRoute } from '@tanstack/react-router';

const RootComponent = () => {
  return (
    <>
      <HeadContent />
      <Outlet />
    </>
  );
};

export const Route = createRootRoute({
  component: RootComponent
});
