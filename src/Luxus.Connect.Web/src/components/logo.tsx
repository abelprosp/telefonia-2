import { cn } from '@/lib/utils';

export const Logo = ({ className, ...props }: React.ComponentProps<'img'>) => {
  return (
    <img
      alt="logo"
      className={cn('size-7', className)}
      src="https://grupoluxus.com.br/telecom/wp-content/uploads/2019/03/cropped-luxus-favicon-192x192.png"
      {...props}
    />
  );
};
