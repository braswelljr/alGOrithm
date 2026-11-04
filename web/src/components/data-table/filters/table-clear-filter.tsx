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
import * as React from 'react';
import type { RowData } from '@tanstack/react-table';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { AppReactTable } from '@/lib/table';
import { cn } from '@/lib/utils';

export interface TableClearFilterProps<TData extends RowData> {
  table: AppReactTable<TData>;
  className?: string;
  variant?: React.ComponentProps<typeof Button>['variant'];
  size?: React.ComponentProps<typeof Button>['size'];
  showIcon?: boolean;
  children?: React.ReactNode;
  /**
   * Enable resetting column filters
   * @default true
   */
  enableResetColumnFilters?: boolean;
  /**
   * Enable resetting global filter (search)
   * @default true
   */
  enableResetGlobalFilter?: boolean;
  /**
   * Enable resetting sorting
   * @default true
   */
  enableResetSorting?: boolean;
  /**
   * Enable resetting hidden columns
   * @default true
   */
  enableResetColumnVisibility?: boolean;
  /**
   * Replaces the table resets above with the caller's own.
   *
   * `table.resetX()` restores `initialState`, which is the right answer only for
   * a table that owns its state. On these screens the state lives in the query
   * string, where the defaults are the parsers' and some of what the button
   * clears (facet selections) is not table state at all. A page that keeps its
   * view in the URL passes a handler that clears those keys instead, and every
   * parser falls back to its own default in one write.
   */
  onClear?: () => void;
  /**
   * Whether the caller holds filter state the table cannot see.
   *
   * The button hides itself when there is nothing to reset, and it decides that
   * from the table: its column filters, its global filter, its sorting. A
   * register that also keeps a date interval in the URL has state the table has
   * never heard of, so a view narrowed only by dates showed no Clear control at
   * all and the range could not be undone except by reopening the calendar and
   * emptying it by hand.
   *
   * Pass `true` while any of that state is set. It only affects whether the
   * button appears; clearing it is `onClear`'s job.
   */
  active?: boolean;
}

/**
 * Core clear filter button component that accepts a table prop directly.
 * Use this when you want to manage the table instance yourself.
 *
 * Automatically hides when there are no active filters to clear.
 *
 * @example
 * ```tsx
 * const table = useReactTable({ ... })
 * <TableClearFilter table={table} />
 * ```
 */
export function TableClearFilter<TData extends RowData>({
  table,
  className,
  variant = 'primary',
  size = 'icon-sm',
  showIcon = true,
  children = null,
  enableResetColumnFilters = true,
  enableResetGlobalFilter = true,
  enableResetSorting = true,
  enableResetColumnVisibility = true,
  onClear,
  active = false
}: TableClearFilterProps<TData>) {
  // Read state directly - should be reactive via table re-renders
  const state = table.state;
  const hasActiveFilters = state.columnFilters.length > 0;
  const hasGlobalFilter = Boolean(state.globalFilter);
  const hasSorting = state.sorting.length > 0;
  const hasHiddenColumns = Object.values(state.columnVisibility ?? {}).some(
    (isVisible) => !isVisible
  );

  // Only check for states that are meant to be reset
  const hasAnythingToReset =
    active ||
    (enableResetColumnFilters && hasActiveFilters) ||
    (enableResetGlobalFilter && hasGlobalFilter) ||
    (enableResetSorting && hasSorting) ||
    (enableResetColumnVisibility && hasHiddenColumns);

  const handleClearAll = React.useCallback(() => {
    if (onClear) {
      onClear();
      return;
    }
    if (enableResetColumnFilters) {
      table.resetColumnFilters();
    }
    if (enableResetGlobalFilter) {
      table.setGlobalFilter('');
    }
    if (enableResetSorting) {
      table.resetSorting();
    }
    if (enableResetColumnVisibility) {
      table.resetColumnVisibility();
    }
  }, [
    onClear,
    table,
    enableResetColumnFilters,
    enableResetGlobalFilter,
    enableResetSorting,
    enableResetColumnVisibility
  ]);

  if (!hasAnythingToReset) {
    return null;
  }

  // `children={null}` asks for the icon on its own, which is how the table
  // toolbars render it beside the refresh button.
  const iconOnly = children === null;

  const label = 'Reset filters and columns';

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            variant={variant}
            size={iconOnly ? 'icon-sm' : size}
            onClick={handleClearAll}
            // One name. It carried three - `aria-label`, `title` and an
            // `sr-only` span - which a screen reader resolves to one anyway
            // while `title` also raised a second, native tooltip a beat after
            // this one.
            aria-label={label}
            className={cn(!iconOnly && 'h-7', className)}
          />
        }
      >
        {showIcon && (
          <X
            aria-hidden="true"
            className={cn('size-4')}
          />
        )}
        {iconOnly ? null : children || 'Reset'}
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

TableClearFilter.displayName = 'TableClearFilter';
