import * as React from 'react';
import type { RowData } from '@tanstack/react-table';
import { Search, X } from 'lucide-react';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput
} from '@/components/ui/input-group';
/**
 * niko-table, created by Semir N. (Semkoo, https://github.com/Semkoo) with AI assistance.
 *
 * Before reporting anything: please check the changelog first.
 *  - In-repo: ./CHANGELOG.md
 *  - Docs site: https://niko-table.com/changelog
 *
 * Found a bug or have a fix? Open an issue or PR on GitHub so other
 * users (and future LLMs reading this code) benefit:
 * https://github.com/Semkoo/niko-table-registry
 */
import type { AppReactTable } from '@/lib/table';
import { cn } from '@/lib/utils';

export interface TableSearchFilterProps<TData extends RowData> extends Omit<
  React.ComponentProps<typeof InputGroup>,
  'onChange' | 'value'
> {
  table: AppReactTable<TData>;
  className?: string;
  placeholder?: string;
  showClearButton?: boolean;
  onChange?: (value: string) => void;
  value?: string;
  /**
   * The accessible name of the search box.
   *
   * Defaults to the placeholder, which is almost always the more specific of
   * the two. A screen has more than one of these - the standards workbench has
   * a Standards search and a criteria search - and two controls both announcing
   * "Search table" cannot be told apart by anyone navigating by name.
   */
  label?: string;
  /**
   * Debounce ms before pushing the typed value into table state. The input
   * reflects keystrokes immediately; only the call to `table.setGlobalFilter`
   * (and the supplied `onChange`) is delayed.
   *
   * **Server-driven search wants this ON**, at 250-400ms. The note that used to
   * be here said the opposite, on the reasoning that "the network request is
   * already the natural rate limiter". It is not: nothing rate-limits a request
   * per keystroke, so typing "library" opens seven, and which of them answers
   * last is a property of the network rather than of what was typed.
   *
   * Client-side filtering over a large fully-loaded set wants it on too, for
   * the original reason: each keystroke re-walks the row model synchronously.
   * Leave it off only for a small client-side table, where a filter pass is
   * cheaper than the timer.
   *
   * Only applies in uncontrolled mode (when `value` is not supplied). In
   * controlled mode the consumer owns the value and should debounce there.
   *
   * @default 0
   */
  debounceMs?: number;
}

export function TableSearchFilter<TData extends RowData>({
  table,
  className,
  // `sm` (h-8.5), not the group's `default` (h-10). A table toolbar is a row of
  // controls at the app's control height - the facet buttons and the view menu
  // beside it are `size="sm"` - and a search box a size taller than everything
  // it sits with is the one thing on the row that does not line up.
  size = 'sm',
  placeholder = 'Search...',
  showClearButton = true,
  onChange,
  value,
  label,
  debounceMs = 0,
  ...props
}: TableSearchFilterProps<TData>) {
  const isControlled = value !== undefined;

  // The live table value. Read through `table.state` so the component re-renders
  // when the filter changes from anywhere: the URL, a Clear-all, a back button.
  const tableGlobalFilter = table.state.globalFilter;
  const globalFilterValue = typeof tableGlobalFilter === 'string' ? tableGlobalFilter : '';

  // Debounce only applies uncontrolled; the consumer owns rate-limiting when it
  // owns the value.
  const debounceEnabled = !isControlled && debounceMs > 0;

  // What the input shows while a flush is pending, so keystrokes render at
  // 60fps even though the table mutation is delayed.
  const [pendingValue, setPendingValue] = React.useState<string>(globalFilterValue);

  const debounceTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  // The last value this component pushed into the table.
  //
  // # Why the echo has to be recognised
  //
  // Pushing a value comes back as a change to `globalFilterValue`, and there is
  // nothing in that value to say whether it came from this input or from
  // somewhere else. The effect below has to adopt an outside change - a back
  // button, a Clear all - and must not adopt its own.
  //
  // It used to adopt both, and the failure is reachable by typing normally.
  // Flush "a" at 300ms; the write is asynchronous (ours goes through the URL),
  // so the re-render carrying `globalFilterValue = "a"` can arrive *after* the
  // reader has typed "ab" and armed a new timer. The effect then cancelled that
  // timer and rewound the box to "a": the "b" disappeared from the input as it
  // was typed and never reached the table. Intermittent, dependent on typing
  // speed, and invisible in any test that types synchronously.
  const lastPushedRef = React.useRef<string>(globalFilterValue);

  const push = React.useCallback(
    (next: string) => {
      lastPushedRef.current = next;
      table.setGlobalFilter(next);
      onChange?.(next);
    },
    [table, onChange]
  );

  // Adopt a value that changed outside this input, and only such a value.
  React.useEffect(() => {
    if (!debounceEnabled) return;
    // The echo of our own write. The input already shows it, and a pending
    // timer holds something newer.
    if (globalFilterValue === lastPushedRef.current) return;

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
      debounceTimerRef.current = null;
    }
    lastPushedRef.current = globalFilterValue;
    setPendingValue(globalFilterValue);
  }, [globalFilterValue, debounceEnabled]);

  // Never write to a torn-down table.
  React.useEffect(
    () => () => {
      if (debounceTimerRef.current) clearTimeout(debounceTimerRef.current);
    },
    []
  );

  const currentValue = isControlled ? value : debounceEnabled ? pendingValue : globalFilterValue;

  const handleClear = React.useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
      debounceTimerRef.current = null;
    }
    // Clearing is immediate whether or not the box debounces: the reader has
    // stated an intention rather than typed a character, and waiting 300ms to
    // act on a button press reads as a dropped click.
    if (debounceEnabled) setPendingValue('');
    push('');
  }, [push, debounceEnabled]);

  const handleChange = React.useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = event.target.value;

      if (!debounceEnabled) {
        push(newValue);
        return;
      }

      setPendingValue(newValue);
      if (debounceTimerRef.current) clearTimeout(debounceTimerRef.current);
      debounceTimerRef.current = setTimeout(() => {
        debounceTimerRef.current = null;
        push(newValue);
      }, debounceMs);
    },
    [push, debounceEnabled, debounceMs]
  );

  const accessibleName = label ?? placeholder;
  const hasValue = currentValue.length > 0;

  return (
    <InputGroup
      {...props}
      size={size}
      // `InputGroup` already sets `relative flex w-full items-center`, so those
      // were being declared twice. `flex-1` is not set here either: it made the
      // box eat a whole toolbar, which every caller then undid with a
      // `max-w-*`. How wide a search box should be is the toolbar's decision.
      className={cn(className)}
      role="search"
      // A landmark with no name is a landmark nobody can navigate to.
      aria-label={accessibleName}
    >
      <InputGroupAddon>
        <Search
          className="text-stone-600 dark:text-stone-400"
          aria-hidden="true"
        />
      </InputGroupAddon>
      <InputGroupInput
        // `search` rather than `text`: it is what the control is, it lets a
        // browser offer Escape-to-clear, and it keeps the field out of the
        // autofill heuristics a bare text input attracts.
        //
        // WebKit draws its own clear glyph on one, which would sit beside the
        // button below and give the reader two of them. Suppressed here rather
        // than globally, because this is the only input in the app that asks
        // for the type and then supplies its own control.
        type="search"
        placeholder={placeholder}
        value={currentValue}
        onChange={handleChange}
        aria-label={accessibleName}
        className="[&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none"
      />
      {hasValue && showClearButton && (
        <InputGroupAddon align="inline-end">
          {/*
            `InputGroupButton`, not a `Button` positioned over the input.

            The addons are flex siblings of the input - `order-first` and
            `order-last` - and the group derives the input's padding from which
            of them are present. The old markup took the clear button out of
            flow with `absolute right-1` and then hand-wrote `pl-9 pr-9` on the
            input to leave room for both, which overrode the padding the group
            had already worked out and left the text sitting away from its own
            search icon.
          */}
          <InputGroupButton
            size="icon-xs"
            onClick={handleClear}
            aria-label={`Clear ${accessibleName.toLowerCase()}`}
          >
            <X aria-hidden="true" />
          </InputGroupButton>
        </InputGroupAddon>
      )}
    </InputGroup>
  );
}

/**
 * @required displayName is required for auto feature detection
 * @see src/components/niko-table/config/feature-detection.ts
 */
TableSearchFilter.displayName = 'TableSearchFilter';
