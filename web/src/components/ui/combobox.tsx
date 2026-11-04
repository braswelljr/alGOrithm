'use client';

import { useState } from 'react';
import { CheckIcon, ChevronsUpDownIcon } from 'lucide-react';
import {
  Command,
  CommandEmpty,
  CommandInput,
  CommandItem,
  CommandList
} from '@/components/ui/command';
import { Popover, PopoverPopup, PopoverTrigger } from '@/components/ui/popover';
import { controlSizeByData } from '@/lib/control';
import { cn } from '@/lib/utils';

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
          'flex w-full min-w-0 items-center justify-between gap-1.5 rounded-lg border py-2 text-left shadow-xs shadow-black/5 transition-[color,background-color,border-color,box-shadow] outline-none',
          'border-border bg-card text-fg hover:bg-bg',
          'focus-visible:border-control focus-visible:ring-fg/10 focus-visible:ring-3',
          'disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
          // Sizing: the shared control scale (lib/control.ts).
          controlSizeByData,
          className
        )}
      >
        <span className={cn('min-w-0 truncate', !selected && 'text-muted')}>
          {selected ? selected.label : placeholder}
        </span>
        <ChevronsUpDownIcon
          aria-hidden
          className="text-muted size-4 shrink-0"
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
          <CommandList className="scroll-quiet max-h-72 overflow-y-auto">
            {(o: ComboboxOption) => (
              <CommandItem
                key={o.value || '__none'}
                value={o}
                disabled={o.disabled}
                onClick={() => {
                  onChange(o.value);
                  setOpen(false);
                }}
                className="flex min-h-8 cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-sm sm:min-h-7 pointer-coarse:min-h-11"
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
                    <span className="text-muted block truncate text-xs">{o.hint}</span>
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
