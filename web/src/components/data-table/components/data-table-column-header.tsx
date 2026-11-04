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
import { stableContext } from '@/lib/stable-context';
import type { AppColumn } from '@/lib/table';
import { cn } from '@/lib/utils';

// ============================================================================
// CONTEXT
// ============================================================================

interface TableColumnHeaderContextValue<TData extends RowData, TValue> {
  column: AppColumn<TData, TValue>;
}

/**
 * `stableContext` rather than `createContext`, because this context's provider
 * and its consumers live in different modules *and* are reached by different
 * import specifiers: a column definition takes `DataTableColumnTitle` from the
 * `@/components/ui` barrel while `core/data-table-structure.tsx` takes
 * `DataTableColumnHeaderRoot` from this file directly. Evaluate this module
 * twice - which Vite does whenever anything it imports is invalidated - and the
 * provider publishes to one context object while the title reads the other,
 * gets the default, and throws "useColumnHeaderContext must be used within
 * DataTableColumnHeaderRoot" from inside a header root that is right there.
 *
 * Keyed on `globalThis`, so the second evaluation finds the first one's object.
 */
const TableColumnHeaderContext = stableContext<
  TableColumnHeaderContextValue<RowData, unknown> | undefined
>('data-table-column-header', undefined);

export function useColumnHeaderContext<TData extends RowData, TValue>(
  required: true
): TableColumnHeaderContextValue<TData, TValue>;
export function useColumnHeaderContext<TData extends RowData, TValue>(
  required: false
): TableColumnHeaderContextValue<TData, TValue> | undefined;
export function useColumnHeaderContext<TData extends RowData, TValue>(required = true) {
  const context = React.useContext(TableColumnHeaderContext) as
    TableColumnHeaderContextValue<TData, TValue> | undefined;

  if (required && !context) {
    throw new Error('useColumnHeaderContext must be used within DataTableColumnHeaderRoot');
  }
  return context;
}

// ============================================================================
// CONTEXT PROVIDER
// ============================================================================

/**
 * Provider for column header context.
 * Used internally by DataTableHeader to provide context to composable header components.
 */
export function DataTableColumnHeaderRoot<TData extends RowData, TValue>({
  column,
  children
}: {
  column: AppColumn<TData, TValue>;
  children: React.ReactNode;
}) {
  // Memoize so context subscribers only re-render when `column` identity changes.
  const contextValue = React.useMemo(
    () => ({ column }) as TableColumnHeaderContextValue<RowData, unknown>,
    [column]
  );
  return (
    <TableColumnHeaderContext.Provider value={contextValue}>
      {children}
    </TableColumnHeaderContext.Provider>
  );
}

// ============================================================================
// ROOT COMPONENT
// ============================================================================

export type DataTableColumnHeaderProps = React.HTMLAttributes<HTMLDivElement>;

/**
 * Composable Column Header container.
 */
export function DataTableColumnHeader({
  className,
  children,
  ...props
}: DataTableColumnHeaderProps) {
  return (
    <div
      className={cn(
        // `min-w-0` lets the header shrink below its content so the title's
        // `truncate` engages. Without it, a narrow (resized/auto-fit) column's
        // label overflows into the next header cell instead of ellipsizing.
        'group flex w-full min-w-0 items-center justify-between gap-1',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

DataTableColumnHeaderRoot.displayName = 'DataTableColumnHeaderRoot';
DataTableColumnHeader.displayName = 'DataTableColumnHeader';
