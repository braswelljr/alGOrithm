import type { RowData } from '@tanstack/react-table';
import { Check, CircleHelp, Pin, PinOff } from 'lucide-react';
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
import type { AppColumn } from '@/lib/table';
import { cn } from '@/lib/utils';

/**
 * Dropdown menu items for pinning a column.
 * Use inside a DropdownMenuContent or as a child of TableColumnActions.
 *
 * @example
 * ```tsx
 * // Inside TableColumnActions
 * <TableColumnActions column={column}>
 *   <TableColumnPinOptions column={column} />
 * </TableColumnActions>
 * ```
 */
export function TableColumnPinOptions<TData extends RowData, TValue>({
  column,
  withSeparator = true
}: {
  column: AppColumn<TData, TValue>;
  /** Whether to render a separator before the options. Defaults to true. */
  withSeparator?: boolean;
}) {
  const canPin = column.getCanPin();
  const isPinned = column.getIsPinned();

  if (!canPin) return null;

  return (
    <>
      {withSeparator && <DropdownMenuSeparator />}
      <DropdownMenuLabel className="flex items-center justify-between text-sm font-normal text-stone-600 dark:text-stone-400">
        <span>Column Pin</span>
        <Tooltip>
          <TooltipTrigger render={<CircleHelp className="size-3.5 cursor-help" />} />
          <TooltipContent side="right">Pin column to left or right side</TooltipContent>
        </Tooltip>
      </DropdownMenuLabel>
      <DropdownMenuItem
        onClick={() => column.pin('start')}
        className={cn(
          'flex items-center',
          isPinned === 'start' &&
            'bg-stone-100 text-stone-900 dark:bg-stone-900 dark:text-stone-100'
        )}
      >
        <Pin className="me-2 size-4 -rotate-45" />
        <span className="flex-1">Pin to Left</span>
        {isPinned === 'start' && <Check className="ms-2 size-4" />}
      </DropdownMenuItem>
      <DropdownMenuItem
        onClick={() => column.pin('end')}
        className={cn(
          'flex items-center',
          isPinned === 'end' && 'bg-stone-100 text-stone-900 dark:bg-stone-900 dark:text-stone-100'
        )}
      >
        <Pin className="me-2 size-4 rotate-45" />
        <span className="flex-1">Pin to Right</span>
        {isPinned === 'end' && <Check className="ms-2 size-4" />}
      </DropdownMenuItem>
      {isPinned && (
        <DropdownMenuItem
          onClick={() => column.pin(false)}
          className="flex items-center"
        >
          <PinOff className="me-2 size-4" />
          <span className="flex-1">Unpin</span>
        </DropdownMenuItem>
      )}
    </>
  );
}

/**
 * Standalone dropdown menu for pinning a column.
 * Shows a pin button that opens a dropdown with pin options.
 *
 * @example
 * ```tsx
 * // Standalone usage
 * <TableColumnPinMenu column={column} />
 * ```
 */
export function TableColumnPinMenu<TData extends RowData, TValue>({
  column,
  className
}: {
  column: AppColumn<TData, TValue>;
  className?: string;
}) {
  const canPin = column.getCanPin();
  const isPinned = column.getIsPinned();

  if (!canPin) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            rounded="full"
            variant="ghost"
            size="icon"
            className={cn(
              'size-7 transition-opacity group-hover:opacity-100 dark:text-stone-400',
              isPinned ? 'text-primary opacity-100' : 'opacity-0',
              className
            )}
          >
            <Pin className="size-4" />
            <span className="sr-only">Pin column</span>
          </Button>
        }
      />
      <DropdownMenuContent
        align="end"
        className="w-48"
      >
        <DropdownMenuLabel className="flex items-center justify-between text-sm font-normal text-stone-600 dark:text-stone-400">
          <span>Column Pin</span>
          <Tooltip>
            <TooltipTrigger render={<CircleHelp className="size-3.5 cursor-help" />} />
            <TooltipContent side="right">Pin column to left or right side</TooltipContent>
          </Tooltip>
        </DropdownMenuLabel>
        <DropdownMenuItem
          onClick={() => column.pin('start')}
          className={cn(
            'flex items-center',
            isPinned === 'start' &&
              'bg-stone-100 text-stone-900 dark:bg-stone-900 dark:text-stone-100'
          )}
        >
          <Pin className="me-2 size-4 -rotate-45" />
          <span className="flex-1">Pin to Left</span>
          {isPinned === 'start' && <Check className="ms-2 size-4" />}
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => column.pin('end')}
          className={cn(
            'flex items-center',
            isPinned === 'end' &&
              'bg-stone-100 text-stone-900 dark:bg-stone-900 dark:text-stone-100'
          )}
        >
          <Pin className="me-2 size-4 rotate-45" />
          <span className="flex-1">Pin to Right</span>
          {isPinned === 'end' && <Check className="ms-2 size-4" />}
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => column.pin(false)}
          className="flex items-center"
        >
          <PinOff className="me-2 size-4" />
          <span className="flex-1">Unpin</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/** @deprecated Use `TableColumnPinMenu` instead */
export const TableColumnPin = TableColumnPinMenu;

TableColumnPinOptions.displayName = 'TableColumnPinOptions';
TableColumnPinMenu.displayName = 'TableColumnPinMenu';
