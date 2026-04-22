declare global {
  interface Number {
    format(
      locales?: Intl.LocalesArgument,
      minimumIntegerDigits?: number
    ): string | undefined;

    toCurrency(
      locales?: Intl.LocalesArgument,
      currency?: string
    ): string | undefined;
  }
}

Number.prototype.format = function (
  locales: Intl.LocalesArgument = 'pt-BR',
  minimumIntegerDigits: number = 2
) {
  return new Intl.NumberFormat(locales, {
    minimumFractionDigits: minimumIntegerDigits,
    useGrouping: false
  }).format(+this);
};

Number.prototype.toCurrency = function (
  locales: Intl.LocalesArgument = 'pt-BR',
  currency: string = 'BRL'
) {
  const num = +this;

  return new Intl.NumberFormat(locales, {
    style: 'currency',
    currency: currency
  }).format(num);
};
