import { format, parse, parseISO } from 'date-fns';

declare global {
  interface String {
    toDate(format?: string): Date | undefined;
    toCurrency(
      locales?: Intl.LocalesArgument,
      currency?: string
    ): string | undefined;
    formatAsDate(format?: string): string;
  }
}

String.prototype.toDate = function (format?: string): Date | undefined {
  if (!format) {
    return parseISO(String(this));
  }

  return parse(String(this), format, new Date());
};

String.prototype.toCurrency = function (
  locales: Intl.LocalesArgument = 'pt-BR',
  currency: string = 'BRL'
) {
  const num = +this;

  if (isNaN(num)) {
    return String(this);
  }

  return new Intl.NumberFormat(locales, {
    style: 'currency',
    currency: currency
  }).format(num);
};

String.prototype.formatAsDate = function (formatStr?: string): string {
  if (formatStr) {
    return format(String(this), formatStr);
  }

  return format(String(this), 'yyyy-MM-dd HH:mm:ss');
};
