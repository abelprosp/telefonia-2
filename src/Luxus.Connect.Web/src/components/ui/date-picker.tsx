'use client';

import {
  forwardRef,
  useCallback,
  useId,
  useMemo,
  useState,
  type ComponentProps
} from 'react';
import * as React from 'react';

import { format, isValid, parse } from 'date-fns';
import { enUS, ptBR } from 'date-fns/locale';
import { CalendarIcon } from 'lucide-react';
import type { DropdownProps } from 'react-day-picker';
import {
  Controller,
  type Control,
  type FieldPath,
  type FieldValues
} from 'react-hook-form';
import { withMask } from 'use-mask-input';

import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import {
  Combobox,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxPopup
} from '@/components/ui/combobox';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput
} from '@/components/ui/input-group';
import { Popover, PopoverPopup, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';

const DISPLAY_FORMAT = 'dd/MM/yyyy';

type DatePickerProps = {
  value?: Date | string | null;
  onChange?: (date: Date | undefined) => void;
  onBlur?: () => void;
  id?: string;
  name?: string;
  placeholder?: string;
  disabled?: boolean;
  locale?: string;
  className?: string;
  calendarProps?: Omit<
    ComponentProps<typeof Calendar>,
    'mode' | 'selected' | 'onSelect' | 'month' | 'onMonthChange'
  >;
};

type RhfDatePickerProps<TFieldValues extends FieldValues> = Omit<
  DatePickerProps,
  'value' | 'onChange'
> & {
  control: Control<TFieldValues>;
  name: FieldPath<TFieldValues>;
  mapValue?: (value: unknown) => Date | undefined;
  mapDate?: (date: Date | undefined) => unknown;
};

const isValidDate = (value: Date) => !Number.isNaN(value.getTime());

const normalizeDate = (value: DatePickerProps['value']): Date | undefined => {
  if (value === null || value === undefined || value === '') {
    return undefined;
  }

  if (value instanceof Date) {
    return isValidDate(value) ? value : undefined;
  }

  const parsed = new Date(value);
  return isValidDate(parsed) ? parsed : undefined;
};

function parseInputToDate(raw: string): Date | undefined {
  const s = raw.trim();
  if (!s) {
    return undefined;
  }

  const dmy = parse(s, DISPLAY_FORMAT, new Date());
  if (isValid(dmy) && s.length >= 10) {
    return dmy;
  }

  const iso = parse(s, 'yyyy-MM-dd', new Date());
  if (isValid(iso)) {
    return iso;
  }

  const fallback = new Date(s);
  return isValidDate(fallback) ? fallback : undefined;
}

interface CalendarDropdownItem {
  disabled?: boolean;
  label: string;
  value: string;
}

function CalendarDropdown(props: DropdownProps) {
  const { options, value, onChange, 'aria-label': ariaLabel } = props;

  const items: CalendarDropdownItem[] =
    options?.map((option) => ({
      disabled: option.disabled,
      label: option.label,
      value: option.value.toString()
    })) ?? [];

  const selectedItem = items.find((item) => item.value === value?.toString());

  const handleValueChange = (newValue: CalendarDropdownItem | null) => {
    if (onChange && newValue) {
      const syntheticEvent = {
        target: { value: newValue.value }
      } as React.ChangeEvent<HTMLSelectElement>;
      onChange(syntheticEvent);
    }
  };

  return (
    <Combobox
      aria-label={ariaLabel}
      autoHighlight
      items={items}
      onValueChange={handleValueChange}
      value={selectedItem}
    >
      <ComboboxInput
        className="**:[input]:w-0 **:[input]:flex-1"
        onFocus={(e: React.FocusEvent<HTMLInputElement>) =>
          e.currentTarget.select()
        }
      />
      <ComboboxPopup aria-label={ariaLabel}>
        <ComboboxEmpty>No items found.</ComboboxEmpty>
        <ComboboxList>
          {(item: CalendarDropdownItem) => (
            <ComboboxItem
              disabled={item.disabled}
              key={item.value}
              value={item}
            >
              {item.label}
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxPopup>
    </Combobox>
  );
}

const DatePicker = forwardRef<HTMLInputElement, DatePickerProps>(
  function DatePicker(
    {
      value,
      onChange,
      onBlur,
      id,
      name,
      placeholder = 'Selecione uma data',
      disabled,
      locale = 'pt-BR',
      className,
      calendarProps
    },
    ref
  ) {
    const autoId = useId();
    const inputId = id ?? autoId;
    const [open, setOpen] = useState(false);

    const selectedDate = useMemo(() => normalizeDate(value), [value]);

    const [inputValue, setInputValue] = useState(() =>
      selectedDate ? format(selectedDate, DISPLAY_FORMAT) : ''
    );

    const [month, setMonth] = useState<Date>(() => selectedDate ?? new Date());

    const selectedTimestamp = selectedDate?.getTime() ?? null;
    const [prevSelectedTimestamp, setPrevSelectedTimestamp] = useState<
      number | null
    >(selectedTimestamp);
    if (prevSelectedTimestamp !== selectedTimestamp) {
      setPrevSelectedTimestamp(selectedTimestamp);
      setInputValue(selectedDate ? format(selectedDate, DISPLAY_FORMAT) : '');
      if (selectedDate) {
        setMonth(selectedDate);
      }
    }

    const maskRef = useMemo(
      () =>
        withMask('99/99/9999', {
          placeholder: '_',
          showMaskOnHover: false
        }),
      []
    );

    const setInputRefs = useCallback(
      (node: HTMLInputElement | null) => {
        maskRef(node);
        if (typeof ref === 'function') {
          ref(node);
        } else if (ref) {
          ref.current = node;
        }
      },
      [maskRef, ref]
    );

    const dayPickerLocale = locale.toLowerCase().startsWith('pt') ? ptBR : enUS;

    const {
      components: userCalendarComponents,
      className: calendarClassName,
      ...restCalendarProps
    } = calendarProps ?? {};

    const handleCalendarSelect = (date: Date | undefined) => {
      onChange?.(date);
      if (date) {
        setInputValue(format(date, DISPLAY_FORMAT));
        setMonth(date);
      } else {
        setInputValue('');
      }
      setOpen(false);
    };

    const commitInput = (raw: string) => {
      const parsed = parseInputToDate(raw);
      if (!raw.trim()) {
        onChange?.(undefined);
        setInputValue('');
        return;
      }
      if (parsed) {
        onChange?.(parsed);
        setInputValue(format(parsed, DISPLAY_FORMAT));
        setMonth(parsed);
      } else if (selectedDate) {
        setInputValue(format(selectedDate, DISPLAY_FORMAT));
      } else {
        setInputValue('');
        onChange?.(undefined);
      }
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const next = e.target.value;
      setInputValue(next);

      if (!next.trim()) {
        onChange?.(undefined);
        return;
      }

      const parsed = parseInputToDate(next);
      if (parsed) {
        onChange?.(parsed);
        setMonth(parsed);
      }
    };

    const handleInputBlur = (e: React.FocusEvent<HTMLInputElement>) => {
      commitInput(e.target.value);
      onBlur?.();
    };

    return (
      <div className="w-full">
        <Popover onOpenChange={setOpen} open={open}>
          <InputGroup className={cn('w-full', className)}>
            <InputGroupInput
              ref={setInputRefs}
              aria-label={placeholder}
              autoComplete="off"
              disabled={disabled}
              id={inputId}
              inputMode="numeric"
              name={name}
              onBlur={handleInputBlur}
              onChange={handleInputChange}
              onClick={(e) => e.stopPropagation()}
              placeholder={placeholder}
              type="text"
              value={inputValue}
            />
            <InputGroupAddon align="inline-end">
              <PopoverTrigger
                aria-label="Abrir calendário"
                disabled={disabled}
                render={
                  <Button
                    disabled={disabled}
                    size="icon-xs"
                    type="button"
                    variant="ghost"
                  />
                }
              >
                <CalendarIcon aria-hidden="true" />
              </PopoverTrigger>
            </InputGroupAddon>
          </InputGroup>
          <PopoverPopup align="start" alignOffset={-4} sideOffset={8}>
            <Calendar
              {...restCalendarProps}
              captionLayout="dropdown"
              className={calendarClassName}
              components={{
                ...userCalendarComponents,
                Dropdown: userCalendarComponents?.Dropdown ?? CalendarDropdown
              }}
              disabled={disabled}
              endMonth={new Date(2100, 11)}
              locale={dayPickerLocale}
              mode="single"
              month={month}
              onMonthChange={setMonth}
              onSelect={handleCalendarSelect}
              selected={selectedDate}
              startMonth={new Date(1900, 0)}
            />
          </PopoverPopup>
        </Popover>
      </div>
    );
  }
);

function RhfDatePicker<TFieldValues extends FieldValues>({
  control,
  name,
  mapValue,
  mapDate,
  ...props
}: RhfDatePickerProps<TFieldValues>) {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <DatePicker
          {...props}
          name={field.name}
          onBlur={field.onBlur}
          ref={field.ref}
          value={mapValue ? mapValue(field.value) : normalizeDate(field.value)}
          onChange={(date) => field.onChange(mapDate ? mapDate(date) : date)}
        />
      )}
    />
  );
}

export {
  DatePicker,
  RhfDatePicker,
  type DatePickerProps,
  type RhfDatePickerProps
};
