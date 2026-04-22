interface PageLoaderProps {
  label?: string;
  variant?: 'spinner' | 'skeleton';
}

export const PageLoader = ({
  label = 'Carregando...',
  variant = 'spinner'
}: PageLoaderProps) => {
  if (variant === 'skeleton') {
    return (
      <div className="flex min-h-[400px] w-full flex-col gap-4 p-10">
        <div className="bg-muted h-8 w-48 animate-pulse rounded-md" />
        <div className="bg-muted h-10 w-full max-w-md animate-pulse rounded-md" />
        <div className="border-input overflow-hidden rounded-lg border">
          <div className="bg-muted h-12 w-full animate-pulse" />
          {Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              className="border-border flex h-14 w-full animate-pulse border-t"
              style={{ animationDelay: `${i * 50}ms` }}
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-4">
      <div className="border-primary/10 border-t-primary border-b-primary size-12 animate-spin rounded-full border-[3px]" />
      {label && <span className="text-muted-foreground text-sm">{label}</span>}
    </div>
  );
};
