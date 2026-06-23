import {
  addDays,
  addMonths,
  addSeconds,
  addYears,
  endOfDay,
  format,
  isAfter,
  isBefore,
  isEqual,
  isToday,
  isValid,
  isWeekend,
  startOfDay
} from 'date-fns';

declare global {
  export interface Date {
    addYears(years: number): Date;
    addMonths(months: number): Date;
    addDays(days: number): Date;
    addSeconds(seconds: number): Date;
    startOfDay(): Date;
    endOfDay(): Date;
    isToday(): boolean;
    isWeekend(): boolean;
    isEqual(date: Date): boolean;
    isBefore(date: Date): boolean;
    isBefore(date: Date): boolean;
    isBetween(from: Date, to: Date): boolean;
    isAfter(date: Date): boolean;
    isSameOrBefore(date: Date): boolean;
    isSameOrAfter(date: Date): boolean;
    format(format?: string): string;
  }
}

Date.prototype.format = function (formatStr?: string): string {
  if (!isValid(this)) {
    return '—';
  }

  if (formatStr) {
    return format(this, formatStr);
  }

  return format(this, 'yyyy-MM-dd HH:mm:ss');
};

Date.prototype.addYears = function (years: number): Date {
  if (!years) {
    return this;
  }

  return addYears(this, years);
};

Date.prototype.addMonths = function (months: number): Date {
  if (!months) {
    return this;
  }

  return addMonths(this, months);
};

Date.prototype.addDays = function (days: number): Date {
  if (!days) {
    return this;
  }

  return addDays(this, days);
};

Date.prototype.addSeconds = function (seconds: number): Date {
  if (!seconds) {
    return this;
  }

  return addSeconds(this, seconds);
};

Date.prototype.startOfDay = function (): Date {
  return startOfDay(this);
};

Date.prototype.endOfDay = function (): Date {
  return endOfDay(this);
};

Date.prototype.isToday = function (): boolean {
  return isToday(this);
};

Date.prototype.isWeekend = function (): boolean {
  return isWeekend(this);
};

Date.prototype.isEqual = function (date: Date): boolean {
  return isEqual(this, date);
};

Date.prototype.isBefore = function (date: Date): boolean {
  return isBefore(this, date);
};

Date.prototype.isBetween = function (from: Date, to: Date): boolean {
  return this.isAfter(from) && this.isBefore(to);
};

Date.prototype.isAfter = function (date: Date): boolean {
  return isAfter(this, date);
};

Date.prototype.isSameOrBefore = function (date: Date): boolean {
  return this.isEqual(date) || this.isBefore(date);
};

Date.prototype.isSameOrAfter = function (date: Date): boolean {
  return this.isEqual(date) || this.isAfter(date);
};
