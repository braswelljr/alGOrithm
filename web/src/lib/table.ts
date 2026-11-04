import {
  aggregationFns,
  createExpandedRowModel,
  createFacetedMinMaxValues,
  createFacetedRowModel,
  createFacetedUniqueValues,
  createFilteredRowModel,
  createPaginatedRowModel,
  createSortedRowModel,
  filterFns,
  sortFns,
  stockFeatures,
  tableFeatures,
  type Cell,
  type Column,
  type ColumnDef,
  type ColumnVisibilityState,
  type FilterFn,
  type Header,
  type Row,
  type RowData,
  type SortingState,
  type Table,
  type TableFeatures,
  type TableOptions
} from '@tanstack/react-table';
import type { ReactTable } from '@tanstack/react-table';

/**
 * The one feature set every table in this app is built from.
 *
 * v9 makes features, row models and fn registries explicit on the table
 * options instead of implying them from which `get*RowModel` you passed. They
 * are declared once here rather than per table for two reasons: the row models
 * below are the ones the app actually uses, and `TFeatures` is part of every
 * table type, so a shared component like the pagination bar can only accept
 * tables from across the app if they all agree on it.
 *
 * The core row model is not listed: v9 always builds it.
 */
export const appTableFeatures = tableFeatures({
  ...stockFeatures,
  filterFns,
  sortFns,
  aggregationFns,
  filteredRowModel: createFilteredRowModel(),
  sortedRowModel: createSortedRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
  facetedUniqueValues: createFacetedUniqueValues(),
  // Added for the data-table tree, which uses faceted filters, min/max range
  // filters and expandable rows. Under v8 each of these was passed per table as
  // a `getFacetedRowModel()`-style option; in v9 the feature set is static, so
  // they are declared once here. Declaring one costs nothing at runtime: a row
  // model only computes when the state slice it serves is actually populated.
  facetedRowModel: createFacetedRowModel(),
  facetedMinMaxValues: createFacetedMinMaxValues(),
  expandedRowModel: createExpandedRowModel()
});

export type AppTableFeatures = typeof appTableFeatures;

/** A table instance as the core sees it, which is what components take as a prop. */
export type AppTable<TData extends RowData> = Table<AppTableFeatures, TData>;

/**
 * What `useTable` hands back. Carries the React extras (`Subscribe`,
 * `FlexRender`, `state`) that the plain `Table` above does not.
 */
export type AppReactTable<TData extends RowData> = ReactTable<AppTableFeatures, TData>;

export type AppColumn<TData extends RowData, TValue = unknown> = Column<
  AppTableFeatures,
  TData,
  TValue
>;

export type AppColumnDef<TData extends RowData, TValue = unknown> = ColumnDef<
  AppTableFeatures,
  TData,
  TValue
>;

export type AppRow<TData extends RowData> = Row<AppTableFeatures, TData>;

export type AppHeader<TData extends RowData, TValue = unknown> = Header<
  AppTableFeatures,
  TData,
  TValue
>;

export type AppCell<TData extends RowData, TValue = unknown> = Cell<
  AppTableFeatures,
  TData,
  TValue
>;

export type AppFilterFn<TData extends RowData> = FilterFn<AppTableFeatures, TData>;

export type AppTableOptions<TData extends RowData> = TableOptions<AppTableFeatures, TData>;

/**
 * The columns a visibility map hides, as a plain list of ids.
 *
 * TanStack tracks visibility as a full `{id: boolean}` map, which is the wrong
 * thing to put in a query string: it grows with every column a reader has ever
 * toggled and spells out the ones that are shown, which is the default. The URL
 * carries only what differs from the default, such as `?hidden=phone,status`, and
 * `columnVisibilityFrom` reads it back.
 */
export function hiddenColumnIds(visibility: ColumnVisibilityState): string[] {
  return Object.entries(visibility)
    .filter(([, isVisible]) => !isVisible)
    .map(([id]) => id);
}

/** The visibility map a list of hidden column ids stands for. */
export function columnVisibilityFrom(hidden: readonly string[]): ColumnVisibilityState {
  return Object.fromEntries(hidden.map((id) => [id, false]));
}

/**
 * Each sortable column mapped to the sort key its endpoint knows it by.
 *
 * Every list on these screens sorts server-side, so a header the reader can
 * click has to name a key the endpoint declares in its filter spec. Reading that
 * off the column definitions keeps it in one place: a column is sortable when it
 * does not say `enableSorting: false`, and its id is the key sent to the API
 * unless `meta.sortKey` names a different one. The alternative, a whitelist kept
 * beside each query, drifts the moment a column is added and fails silently when
 * it does, because an unknown sort key is ignored server-side and the list
 * quietly comes back in its default order.
 *
 * Which columns should be sortable at all is a separate question, settled the
 * same way everywhere: dates, amounts and names rank rows and get a toggle;
 * statuses, categories and yes/no flags only group them and are filtered
 * instead. `__tests__/unit/table-wiring.test.ts` holds the reasoning and checks
 * that every column states its answer either way.
 */
export function columnSortKeys<TData extends RowData>(
  columns: readonly AppColumnDef<TData>[]
): ReadonlyMap<string, string> {
  const keys = new Map<string, string>();
  for (const column of columns) {
    if (column.enableSorting === false) continue;
    const id = column.id ?? ('accessorKey' in column ? String(column.accessorKey) : undefined);
    if (id) keys.set(id, column.meta?.sortKey ?? id);
  }
  return keys;
}

/**
 * Render TanStack sorting state as the API's `sort` parameter.
 *
 * The contract is one or more `field:dir` terms separated by commas, so a table
 * sorted by status and then by date sends `status:asc,createdAt:desc` and the
 * server orders by both. Terms naming a column the table does not offer are
 * dropped rather than sent: a hand-edited URL is the only way to produce one,
 * and passing it through would show the list in its default order while the
 * header claimed otherwise.
 *
 * `fallback` is what the endpoint should sort by when nothing survives.
 */
export function toSortParam(
  sorting: SortingState,
  sortKeys: ReadonlyMap<string, string>,
  fallback = 'createdAt:desc'
): string {
  const terms = sorting
    .map((sort) => {
      const key = sortKeys.get(sort.id);
      return key ? `${key}:${sort.desc ? 'desc' : 'asc'}` : undefined;
    })
    .filter((term): term is string => term !== undefined);
  return terms.length > 0 ? terms.join(',') : fallback;
}

/**
 * The app's own additions to a column's and a table's `meta`.
 *
 * Declared here rather than inside `components/data-table`, which carries its
 * own augmentation for the fields the library itself reads. Both merge, so this
 * is where anything specific to these screens belongs: the library should not
 * have to know that one of our pages wants to refetch from a row action, or
 * that a column's endpoint spells its name differently.
 *
 * Optional to a fault, because `meta` is passed per column and per table, and
 * nothing supplies all of them.
 */
declare module '@tanstack/react-table' {
  interface ColumnMeta<
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    TFeatures extends TableFeatures,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    TData extends RowData,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    TValue = unknown
  > {
    /**
     * The key this column's endpoint sorts by, when it differs from the column
     * id. Read by `columnSortKeys`.
     *
     * A column id names the thing on screen; a sort key names a column in the
     * database. They are usually the same word, and this is for when they are
     * not, so a table can keep the id its filters and selection already use
     * without the API having to learn a name shaped by the UI.
     */
    sortKey?: string;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface TableMeta<TFeatures extends TableFeatures, TData extends RowData> {
    /** Re-read the list after a row action changed it. */
    refetch?: () => void;
    onDeleteAddon?: (id: string) => void;
    onEditSpot?: (id: string) => void;
    /** Sorting driven from a column header rather than the table's own state,
     *  where the page keeps sorting in the URL. Takes an updater, like the
     *  table option of the same name. */
    onSortingChange?: (
      updaterOrValue: SortingState | ((prev: SortingState) => SortingState)
    ) => void;
    /** The course a row belongs to, for row actions that link out to it. */
    courseId?: string;
  }
}
