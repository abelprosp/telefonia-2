import { format, isValid, parse, parseISO } from 'date-fns';

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

String.prototype.toDate = function (formatStr?: string): Date | undefined {
  const value = String(this).trim();
  if (!value) {
    return undefined;
  }

  let date: Date;
  if (!formatStr) {
    date = parseISO(value);
  } else {
    date = parse(value, formatStr, new Date());
    if (!isValid(date)) {
      date = parseISO(value);
    }
  }

  return isValid(date) ? date : undefined;
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
