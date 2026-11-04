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
import React from 'react';
import type { RowData } from '@tanstack/react-table';
import { Check, CircleHelp } from 'lucide-react';
import { SORT_ICONS, SORT_LABELS } from '@/components/data-table/config/data-table';
import type { SortIconVariant } from '@/components/data-table/config/data-table';
import { useDataTable } from '@/components/data-table/core/data-table-context';
import { FILTER_VARIANTS } from '@/components/data-table/lib/constants';
import type { FilterVariant } from '@/components/data-table/lib/constants';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { AppColumn, AppReactTable } from '@/lib/table';
import { cn } from '@/lib/utils';

/**
 * Sort options menu items for composition inside TableColumnActions.
 *
 * @example
 * ```tsx
 * <TableColumnActions>
 *   <TableColumnSortOptions column={column} />
 * </TableColumnActions>
 * ```
 */
export function TableColumnSortOptions<TData extends RowData, TValue>({
  column,
  table: propTable,
  variant: propVariant,
  withSeparator = true
}: {
  column: AppColumn<TData, TValue>;
  table?: AppReactTable<TData>;
  variant?: SortIconVariant;
  /** Whether to render a separator before the options. @default true */
  withSeparator?: boolean;
}) {
  const context = useDataTable<TData>();
  const table = propTable || context.table;
  const sortState = column.getIsSorted();

  const variant: FilterVariant =
    propVariant || column.columnDef.meta?.variant || FILTER_VARIANTS.TEXT;

  const icons = SORT_ICONS[variant] || SORT_ICONS[FILTER_VARIANTS.TEXT];
  const labels = SORT_LABELS[variant] || SORT_LABELS[FILTER_VARIANTS.TEXT];

  const sortIndex = column.getSortIndex();
  const isMultiSort = table && table.state.sorting.length > 1;
  const showSortBadge = isMultiSort && sortIndex !== -1;

  /**
   * Use a ref for immediate synchronous access to shift key state.
   * React state updates are batched and async, which can cause race conditions
   * when the dropdown closes - the keyup event might reset state before
   * handleSort reads it. A ref provides synchronous access.
   */
  const isShiftPressedRef = React.useRef(false);
  const [isShiftPressed, setIsShiftPressed] = React.useState(false);

  React.useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Shift') {
        const isDown = e.type === 'keydown';
        isShiftPressedRef.current = isDown;
        setIsShiftPressed(isDown);
      }
    };
    window.addEventListener('keydown', handleKey, { capture: true });
    window.addEventListener('keyup', handleKey, { capture: true });
    return () => {
      window.removeEventListener('keydown', handleKey, { capture: true });
      window.removeEventListener('keyup', handleKey, { capture: true });
    };
  }, []);

  const handleSort = (
    direction: 'asc' | 'desc' | false,
    e:
      | React.MouseEvent
      | React.KeyboardEvent
      | React.SyntheticEvent // Base UI menu items emit wrapped synthetic events
      | Event
      | {
          detail?: { originalEvent?: { shiftKey?: boolean } };
          shiftKey?: boolean;
          nativeEvent?: { shiftKey?: boolean };
        }
  ) => {
    // Detect Shift across event sources (ref → state → native event → Radix CustomEvent.detail).
    const isMulti =
      isShiftPressedRef.current ||
      isShiftPressed ||
      ('shiftKey' in e && !!e.shiftKey) ||
      (e as { detail?: { originalEvent?: { shiftKey?: boolean } } }).detail?.originalEvent
        ?.shiftKey ||
      (e as { nativeEvent?: { shiftKey?: boolean } }).nativeEvent?.shiftKey;

    if (direction === false) {
      column.clearSorting();
    } else {
      const isDesc = direction === 'desc';
      const canMultiSort = column.getCanMultiSort();

      /**
       * @see https://tanstack.com/table/v8/docs/guide/sorting#multi-sorting
       * When using toggleSorting explicitly, we must manually pass the multi-sort flag.
       */
      column.toggleSorting(isDesc, canMultiSort ? isMulti : false);
    }
  };

  return (
    <>
      {withSeparator && <DropdownMenuSeparator />}
      <DropdownMenuLabel className="flex items-center justify-between text-sm font-normal text-stone-600 dark:text-stone-400">
        <div className="flex items-center gap-2">
          <span>Column Sort</span>
          {showSortBadge && (
            <Tooltip>
              <TooltipTrigger
                render={
                  <span className="bg-primary text-secondary-950 flex size-4 cursor-help items-center justify-center rounded-full text-sm font-medium">
                    {sortIndex + 1}
                  </span>
                }
              />
              <TooltipContent side="right">
                Sort priority (order in which columns are sorted)
              </TooltipContent>
            </Tooltip>
          )}
        </div>
        <Tooltip>
          <TooltipTrigger render={<CircleHelp className="size-3.5 cursor-help" />} />
          <TooltipContent side="right">
            TIP: Hold &apos;shift&apos; key to enable multi sort
          </TooltipContent>
        </Tooltip>
      </DropdownMenuLabel>
      <DropdownMenuItem
        onClick={(e) => handleSort('asc', e)}
        aria-current={sortState === 'asc' ? 'true' : undefined}
        data-active={sortState === 'asc' ? '' : undefined}
        className={cn(
          'flex items-center',
          sortState === 'asc' &&
            'bg-primary/10 dark:bg-primary/15 font-medium text-stone-900 dark:text-stone-100'
        )}
      >
        <icons.asc className="me-2 size-4 text-stone-600/70 dark:text-stone-400/70" />
        <span className="flex-1">{labels.asc}</span>
        {sortState === 'asc' && <Check className="ms-2 size-4" />}
      </DropdownMenuItem>
      <DropdownMenuItem
        onClick={(e) => handleSort('desc', e)}
        aria-current={sortState === 'desc' ? 'true' : undefined}
        data-active={sortState === 'desc' ? '' : undefined}
        className={cn(
          'flex items-center',
          sortState === 'desc' &&
            'bg-primary/10 dark:bg-primary/15 font-medium text-stone-900 dark:text-stone-100'
        )}
      >
        <icons.desc className="me-2 size-4 text-stone-600/70 dark:text-stone-400/70" />
        <span className="flex-1">{labels.desc}</span>
        {sortState === 'desc' && <Check className="ms-2 size-4" />}
      </DropdownMenuItem>
      {sortState && (
        <DropdownMenuItem onClick={() => column.clearSorting()}>
          <icons.unsorted className="me-2 size-4 text-stone-600/70 dark:text-stone-400/70" />
          Clear Sort
        </DropdownMenuItem>
      )}
    </>
  );
}

/**
 * Standalone dropdown menu for sorting.
 *
 * For composition inside TableColumnActions, use TableColumnSortOptions instead.
 *
 * @example
 * ```tsx
 * // Standalone
 * <TableColumnSortMenu column={column} table={table} />
 *
 * // Composed
 * <TableColumnActions>
 *   <TableColumnSortOptions column={column} />
 * </TableColumnActions>
 * ```
 */
export function TableColumnSortMenu<TData extends RowData, TValue>({
  column,
  table: propTable,
  variant: propVariant,
  className
}: {
  column: AppColumn<TData, TValue>;
  table?: AppReactTable<TData>;
  variant?: SortIconVariant;
  className?: string;
}) {
  const context = useDataTable<TData>();
  const table = propTable || context.table;
  const canSort = column.getCanSort();
  const sortState = column.getIsSorted();

  const variant: FilterVariant =
    propVariant || column.columnDef.meta?.variant || FILTER_VARIANTS.TEXT;

  const icons = SORT_ICONS[variant] || SORT_ICONS[FILTER_VARIANTS.TEXT];

  if (!canSort) return null;

  const SortIcon =
    sortState === 'asc' ? icons.asc : sortState === 'desc' ? icons.desc : icons.unsorted;

  const sortIndex = column.getSortIndex();
  const isMultiSort = table && table.state.sorting.length > 1;
  const showSortBadge = isMultiSort && sortIndex !== -1;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            rounded="full"
            variant="ghost"
            size="icon"
            className={cn(
              'size-7 transition-opacity dark:text-stone-400',
              sortState && 'text-primary',
              className
            )}
          >
            <div className="relative flex items-center justify-center">
              <SortIcon className="size-4" />
              {showSortBadge && (
                <span className="bg-primary text-secondary-950 absolute -end-2 -top-1 flex size-3 items-center justify-center rounded-full text-xs">
                  {sortIndex + 1}
                </span>
              )}
            </div>
            <span className="sr-only">Sort column</span>
          </Button>
        }
      />
      <DropdownMenuContent
        align="end"
        className="w-48"
      >
        <TableColumnSortOptions
          column={column}
          table={table}
          variant={variant}
          withSeparator={false}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

TableColumnSortOptions.displayName = 'TableColumnSortOptions';
TableColumnSortMenu.displayName = 'TableColumnSortMenu';
