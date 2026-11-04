'use client';

import { useState } from 'react';
import {
  Command,
  CommandEmpty,
  CommandInput,
  CommandItem,
  CommandList
} from '@braswelljr/algorithm-ui/components/command';
import { Popover, PopoverPopup, PopoverTrigger } from '@braswelljr/algorithm-ui/components/popover';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { CheckIcon, ChevronsUpDownIcon } from 'lucide-react';

/** One choice: what is sent, what is read, and a second line if it helps. */
export type ComboboxOption = {
  value: string;
  label: string;
  hint?: string;
  disabled?: boolean;
};

/**
 * Choose one of a list by typing to narrow it: a trigger that reads like a
 * select, opening the Command search in a popover. Keyboard: Enter, Space or
 * the arrows open it, typing filters, the arrows move, Enter chooses, Escape
 * closes and gives focus back. Works for a handful of choices and for a
 * register of hundreds.
 */
export function Combobox({
  id,
  value,
  onChange,
  options,
  placeholder = 'Choose',
  searchPlaceholder = 'Search',
  emptyText = 'Nothing matches',
  disabled = false,
  invalid = false,
  size = 'md',
  className,
  'aria-label': ariaLabel,
  'aria-describedby': ariaDescribedBy,
  'aria-labelledby': ariaLabelledBy
}: {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  options: readonly ComboboxOption[];
  /** Read on the trigger while nothing is chosen. */
  placeholder?: string;
  searchPlaceholder?: string;
  emptyText?: string;
  disabled?: boolean;
  invalid?: boolean;
  size?: 'sm' | 'default' | 'md';
  className?: string;
  'aria-label'?: string;
  'aria-describedby'?: string;
  'aria-labelledby'?: string;
}) {
  const [open, setOpen] = useState(false);
  const selected = options.find((o) => o.value === value);
  return (
    <Popover
      open={open}
      onOpenChange={setOpen}
    >
      <PopoverTrigger
        id={id}
        disabled={disabled}
        aria-label={ariaLabel}
        aria-labelledby={ariaLabelledBy}
        aria-describedby={ariaDescribedBy}
        aria-invalid={invalid || undefined}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        data-slot="combobox-trigger"
        data-size={size}
        className={cn(
          'flex w-full min-w-0 items-center justify-between gap-1.5 rounded-lg border py-2 pr-2 pl-3 text-left text-sm shadow-xs shadow-black/5 transition-[color,background-color,border-color,box-shadow] outline-none',
          'border-border bg-card text-fg hover:bg-bg',
          'focus-visible:border-control focus-visible:ring-3 focus-visible:ring-fg/10',
          'disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
          'data-[size=default]:h-9 data-[size=md]:h-10 data-[size=sm]:h-8',
          className
        )}
      >
        <span className={cn('min-w-0 truncate', !selected && 'text-muted')}>
          {selected ? selected.label : placeholder}
        </span>
        <ChevronsUpDownIcon
          aria-hidden
          className="size-4 shrink-0 text-muted"
        />
      </PopoverTrigger>
      <PopoverPopup
        align="start"
        className="w-(--anchor-width) min-w-64 p-0"
      >
        <Command
          items={options as ComboboxOption[]}
          itemToStringValue={(o) => {
            const opt = o as ComboboxOption;
            return opt.hint ? `${opt.label} ${opt.hint}` : opt.label;
          }}
        >
          <CommandInput
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
          />
          <CommandEmpty>{emptyText}</CommandEmpty>
          <CommandList className="max-h-72 overflow-y-auto scroll-quiet">
            {(o: ComboboxOption) => (
              <CommandItem
                key={o.value || '__none'}
                value={o}
                disabled={o.disabled}
                onClick={() => {
                  onChange(o.value);
                  setOpen(false);
                }}
                className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-sm"
              >
                <CheckIcon
                  aria-hidden
                  className={cn(
                    'mt-0.5 size-4 shrink-0',
                    o.value === value ? 'opacity-100' : 'opacity-0'
                  )}
                />
                <span className="min-w-0">
                  <span className="block truncate">{o.label}</span>
                  {o.hint ? (
                    <span className="block truncate text-xs text-muted">{o.hint}</span>
                  ) : null}
                </span>
              </CommandItem>
            )}
          </CommandList>
        </Command>
      </PopoverPopup>
    </Popover>
  );
}
