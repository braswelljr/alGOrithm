import type { RowData } from '@tanstack/react-table';
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
import { useDataTable } from '@/components/data-table/core/data-table-context';
import {
  TableViewMenu,
  type TableViewMenuProps
} from '@/components/data-table/filters/table-view-menu';

type DataTableViewMenuProps<TData extends RowData> = Omit<TableViewMenuProps<TData>, 'table'>;

export function DataTableViewMenu<TData extends RowData>(props: DataTableViewMenuProps<TData>) {
  const { table } = useDataTable<TData>();
  return (
    <TableViewMenu
      table={table}
      {...props}
    />
  );
}

/**
 * @required displayName is required for auto feature detection
 * @see "feature-detection.ts"
 */

DataTableViewMenu.displayName = 'DataTableViewMenu';
