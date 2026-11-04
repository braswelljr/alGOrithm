/**
 * The data table's public surface.
 *
 * A barrel rather than deep imports, so a screen composing a table names one
 * module and the internal folder layout (core, filters, components, grid) stays
 * the library's own business.
 *
 * The pieces compose: `DataTableRoot` owns the table instance and the context,
 * `DataTable` is the scrolling viewport, and the head, body, empty state and
 * skeleton sit inside it. A screen assembles the arrangement it needs from
 * these rather than through a wrapper: the wrapper that used to live in
 * `components/tables` assembled one arrangement for every screen, so a screen
 * wanting pinning, a sticky head or its own empty state had to leave it
 * anyway.
 */

// Structure: the table itself.
export { DataTable } from '@/components/data-table/core/data-table';
export type { DataTableContainerProps } from '@/components/data-table/core/data-table';
export { DataTableRoot } from '@/components/data-table/core/data-table-root';
// The bar above a register: search left, everything that acts on it right.
export {
  DataTableToolbar,
  TOOLBAR_CONTROL_SIZING
} from '@/components/data-table/components/data-table-toolbar';
export type { DataTableConfig } from '@/components/data-table/core/data-table-root';
export {
  DataTableBody,
  DataTableEmptyBody,
  DataTableHeader,
  DataTableSkeleton
} from '@/components/data-table/core/data-table-structure';
export { useDataTable } from '@/components/data-table/core/data-table-context';

// Empty state composition, for a table that wants more than a sentence.
export {
  DataTableEmptyActions,
  DataTableEmptyDescription,
  DataTableEmptyFilteredMessage,
  DataTableEmptyIcon,
  DataTableEmptyMessage,
  DataTableEmptyState,
  DataTableEmptyTitle
} from '@/components/data-table/components/data-table-empty-state';

// Column controls. A header cell is composed from these: the container gives
// the column to the pieces inside it, so a title, its sort menu and its
// hide/pin controls each read the same column without being handed it.
export {
  DataTableColumnHeader,
  DataTableColumnHeaderRoot
} from '@/components/data-table/components/data-table-column-header';
export { DataTableColumnActions } from '@/components/data-table/components/data-table-column-actions';
export {
  DataTableColumnSortMenu,
  DataTableColumnSortOptions
} from '@/components/data-table/components/data-table-column-sort';
export { DataTableColumnTitle } from '@/components/data-table/components/data-table-column-title';
export { DataTableColumnHideMenu } from '@/components/data-table/components/data-table-column-hide';
export { DataTableColumnPinMenu } from '@/components/data-table/components/data-table-column-pin';
export { DataTableViewMenu } from '@/components/data-table/components/data-table-view-menu';
// The same menu for a toolbar that holds the table itself rather than sitting
// inside `DataTableRoot`.
export { TableViewMenu } from '@/components/data-table/filters/table-view-menu';
export { DataTableColumnResize } from '@/components/data-table/components/data-table-column-resize';
export { DataTableColumnAutoFit } from '@/components/data-table/components/data-table-column-auto-fit';

// Toolbar and pagination.
export { DataTablePagination } from '@/components/data-table/components/data-table-pagination';
export { DataTablePaginationControls } from '@/components/data-table/components/data-table-pagination-controls';
export { DataTableSelectionBar } from '@/components/data-table/components/data-table-selection-bar';
export { DataTableSearchFilter } from '@/components/data-table/components/data-table-search-filter';
export { DataTableFacetedFilter } from '@/components/data-table/components/data-table-faceted-filter';
// The same filter given its column and options directly, for a toolbar that
// sits outside DataTableRoot or whose options and counts come from the server
// rather than from the rows on screen.
export {
  TableFacetedFilter,
  type TableFacetedFilterProps
} from '@/components/data-table/filters/table-faceted-filter';
// A facet bound to a table column: the column filter is the selection and its
// faceted values are the counts.
export { TableFacet } from '@/components/data-table/filters/table-facet';
export { DataTableExportButton } from '@/components/data-table/components/data-table-export-button';
// The reset that puts filters, search, sort and hidden columns back to their
// defaults. Takes a table directly, so it needs no DataTableRoot around it and
// can sit in a table's header beside the refresh button.
export { DataTableClearFilter } from '@/components/data-table/components/data-table-clear-filter';
export { TableClearFilter } from '@/components/data-table/filters/table-clear-filter';
export type { TableClearFilterProps } from '@/components/data-table/filters/table-clear-filter';

// Rows.
export {
  DataTableRowMenuScope,
  RowMenuGroup,
  RowMenuItem,
  RowMenuSeparator,
  useDataTableRow
} from '@/components/data-table/components/data-table-row-menu';
export { DataTableRowContextMenu } from '@/components/data-table/components/data-table-row-context-menu';

export type { DataTableColumnDef, Option } from '@/components/data-table/types';
