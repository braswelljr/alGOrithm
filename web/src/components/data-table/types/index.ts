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
import { type CellData, type RowData, type TableFeatures } from '@tanstack/react-table';
import {
  FILTER_OPERATORS,
  FILTER_VARIANTS,
  JOIN_OPERATORS
} from '@/components/data-table/lib/constants';
// ============================================================================
// FILTER TYPES
// ============================================================================

import type {
  FilterOperator as _FilterOperator,
  FilterVariant as _FilterVariant,
  JoinOperator as _JoinOperator
} from '@/components/data-table/lib/constants';
import type { AppColumnDef, AppReactTable, AppRow } from '@/lib/table';

// ============================================================================
// TANSTACK REACT-TABLE MODULE AUGMENTATION
// ============================================================================
declare module '@tanstack/react-table' {
  interface ColumnMeta<
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    TFeatures extends TableFeatures,
    TData extends RowData,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    TValue extends CellData = CellData
  > {
    // Display
    label?: string;
    placeholder?: string;
    /**
     * Which column absorbs the leftover row width when column resizing is on
     * (needs `<DataTableColumnResize />`). Filling is ON BY DEFAULT, the first
     * non-pinned data column flexes automatically, so you rarely set this.
     *
     * - `true` overrides the default to flex THIS column instead (use when the
     *   first column isn't the one that should grow).
     * - `false` opts this column out of being auto-picked.
     *
     * The flex column renders with no width and absorbs the surplus, so every
     * other column keeps its `size` and a trailing actions/menu column pins to
     * the right edge. Pure layout, never touches `columnSizing`, so no
     * persistence side effects. Turn filling off for a whole table with
     * `TableMeta.disableFlexFill`.
     */
    flex?: boolean;

    // Filtering
    variant?: FilterVariant;
    options?: Option[];
    range?: [number, number];
    /**
     * Automatically generate options for select/multiSelect columns if not provided.
     * When true and no static `options` exist, generation logic (wrappers / hooks) may supply them.
     */
    autoOptions?: boolean;
    /** Whether to automatically rename option labels using formatLabel. When false, uses raw value as label. */
    autoOptionsFormat?: boolean;
    /**
     * Per-column label formatter applied when options are auto-derived from
     * row data. Receives the stringified row value and returns the display
     * label. Wins over `autoOptionsFormat` when present. Ignored when the
     * caller passes explicit `options`.
     */
    formatOptionLabel?: (value: string) => string;
    /** Per-column override for showing counts (falls back to wrapper prop). */
    showCounts?: boolean;
    /** Per-column override for using filtered rows for counts (falls back to wrapper prop). */
    dynamicCounts?: boolean;
    /** Merge strategy override: preserve | augment | replace (falls back to wrapper prop). */
    mergeStrategy?: 'preserve' | 'augment' | 'replace';

    // Formatting
    unit?: string;
    icon?: React.ComponentType<{ className?: string }>;

    // Row Expansion
    expandedContent?: (row: TData) => React.ReactNode;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface TableMeta<TFeatures extends TableFeatures, TData extends RowData> {
    joinOperator?: JoinOperator;
    hasIndividualJoinOperators?: boolean;
    /**
     * Turn off default flex fill for the whole table, so columns size to their
     * own widths and the table scrolls horizontally instead of stretching to
     * fill. Use for wide, many-column tables meant to scroll.
     *
     * Set it statically via `<DataTableRoot meta={{ disableFlexFill: true }}>`
     *, it's read when table options are built, so toggling it at runtime
     * alone won't re-apply until another table state change rebuilds them.
     */
    disableFlexFill?: boolean;
  }
}

// ============================================================================
// CORE TYPES
// ============================================================================

export interface Option {
  label: string;
  value: string;
  count?: number;
  icon?: React.ComponentType<{ className?: string }>;
}

export type FilterVariant = _FilterVariant;
export type FilterOperator = _FilterOperator;
export type JoinOperator = _JoinOperator;

/**
 * Extended column filter with additional metadata
 */
export interface ExtendedColumnFilter<TData extends RowData> {
  id: Extract<keyof TData, string>;
  value: string | string[];
  variant: FilterVariant;
  operator: FilterOperator;
  filterId: string;
  joinOperator?: JoinOperator; // Individual join operator for each filter
  // You can extend with additional properties if needed
}

/** Global filter type */
export type GlobalFilter = string | Record<string, unknown>;

/**
 * Extended column sort (for URL state management)
 */
export interface ExtendedColumnSort<TData extends RowData> {
  id: Extract<keyof TData, string>;
  desc: boolean;
  // You can extend with additional properties if needed
}

/**
 * Query keys for URL state management
 */
export interface QueryKeys {
  page?: string;
  perPage?: string;
  sort?: string;
  filters?: string;
  joinOperator?: string;
  // Additional keys can be added as needed
}

// ============================================================================
// COLUMN DEFINITION
// ============================================================================

/**
 * Extended column definition for data table
 * Inherits all TanStack Table ColumnDef properties
 */
export type DataTableColumnDef<TData extends RowData, TValue = unknown> = AppColumnDef<
  TData,
  TValue
> & {
  // You can extend with additional properties if needed
};

// ============================================================================
// ROW TYPES
// ============================================================================

/**
 * Data table row type
 * Alias for TanStack Table Row
 */
export type DataTableRow<TData extends RowData> = AppRow<TData> & {
  // You can extend with additional properties if needed
};

export type DataTableInstance<TData extends RowData> = AppReactTable<TData> & {
  // You can extend with additional properties if needed
};

// ============================================================================
// CONVENIENCE TYPE HELPERS
// ============================================================================

/**
 * Convenience type for accessing constant values with better type safety
 */
export type JoinOperatorValues = typeof JOIN_OPERATORS;
export type FilterOperatorValues = typeof FILTER_OPERATORS;
export type FilterVariantValues = typeof FILTER_VARIANTS;

/**
 * Utility type to get the literal values from constant objects
 */
export type ValueOf<T> = T[keyof T];
