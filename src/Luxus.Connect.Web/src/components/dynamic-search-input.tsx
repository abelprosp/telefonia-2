import { useEffect, useState } from 'react';

import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

type DynamicSearchInputProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
  className?: string;
  'aria-label'?: string;
};

/** Debounced search input for server-side list filters (stock / customers / lines / invoices). */
export function DynamicSearchInput({
  value,
  onChange,
  placeholder = 'Pesquisar…',
  debounceMs = 300,
  className,
  'aria-label': ariaLabel = 'Pesquisar'
}: DynamicSearchInputProps) {
  const [local, setLocal] = useState(value);

  useEffect(() => {
    setLocal(value);
  }, [value]);

  useEffect(() => {
    const t = window.setTimeout(() => {
      if (local !== value) onChange(local);
    }, debounceMs);
    return () => window.clearTimeout(t);
  }, [local, debounceMs, onChange, value]);

  return (
    <Input
      value={local}
      onChange={(e) => setLocal(e.target.value)}
      placeholder={placeholder}
      aria-label={ariaLabel}
      className={cn('max-w-sm', className)}
    />
  );
}
