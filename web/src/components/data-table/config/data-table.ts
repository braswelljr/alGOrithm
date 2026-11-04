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

/**
 * @file Data Table Configuration
 * Defines runtime configuration values for the Data Table component.
 * This includes filter operators, sort icons, and other constants.
 
 */

import type { ComponentType } from 'react';
import {
  ArrowDown01,
  ArrowDown10,
  ArrowDownAZ,
  ArrowDownZA,
  ArrowUpDown,
  Calendar,
  Check,
  X as XIcon
} from 'lucide-react';
import {
  FILTER_OPERATORS,
  FILTER_VARIANTS,
  JOIN_OPERATORS,
  type FilterOperator,
  type FilterVariant,
  type JoinOperator
} from '@/components/data-table/lib/constants';

export type SortIconVariant = FilterVariant;

/** A sort icon: any component that takes a `className`, whichever icon set it
 *  comes from. */
type SortIcon = ComponentType<{ className?: string }>;

interface SortIcons {
  asc: SortIcon;
  desc: SortIcon;
  unsorted: SortIcon;
}

interface SortLabels {
  asc: string;
  desc: string;
}

export const SORT_ICONS: Record<SortIconVariant, SortIcons> = {
  [FILTER_VARIANTS.TEXT]: {
    asc: ArrowDownAZ,
    desc: ArrowDownZA,
    unsorted: ArrowUpDown
  },
  [FILTER_VARIANTS.NUMBER]: {
    asc: ArrowDown01,
    desc: ArrowDown10,
    unsorted: ArrowUpDown
  },
  [FILTER_VARIANTS.RANGE]: {
    asc: ArrowDown01,
    desc: ArrowDown10,
    unsorted: ArrowUpDown
  },
  [FILTER_VARIANTS.DATE]: {
    asc: ArrowUpDown,
    desc: ArrowUpDown,
    unsorted: Calendar
  },
  [FILTER_VARIANTS.DATE_RANGE]: {
    asc: ArrowUpDown,
    desc: ArrowUpDown,
    unsorted: Calendar
  },
  [FILTER_VARIANTS.BOOLEAN]: {
    asc: XIcon, // False First
    desc: Check, // True First
    unsorted: ArrowUpDown
  },
  [FILTER_VARIANTS.SELECT]: {
    asc: ArrowDownAZ,
    desc: ArrowDownZA,
    unsorted: ArrowUpDown
  },
  [FILTER_VARIANTS.MULTI_SELECT]: {
    asc: ArrowDownAZ,
    desc: ArrowDownZA,
    unsorted: ArrowUpDown
  }
};

export const SORT_LABELS: Record<SortIconVariant, SortLabels> = {
  [FILTER_VARIANTS.TEXT]: {
    asc: 'Asc',
    desc: 'Desc'
  },
  [FILTER_VARIANTS.NUMBER]: {
    asc: 'Low to High',
    desc: 'High to Low'
  },
  [FILTER_VARIANTS.RANGE]: {
    asc: 'Low to High',
    desc: 'High to Low'
  },
  [FILTER_VARIANTS.DATE]: {
    asc: 'Oldest First',
    desc: 'Newest First'
  },
  [FILTER_VARIANTS.DATE_RANGE]: {
    asc: 'Oldest First',
    desc: 'Newest First'
  },
  [FILTER_VARIANTS.BOOLEAN]: {
    asc: 'False First',
    desc: 'True First'
  },
  [FILTER_VARIANTS.SELECT]: {
    asc: 'Asc',
    desc: 'Desc'
  },
  [FILTER_VARIANTS.MULTI_SELECT]: {
    asc: 'Asc',
    desc: 'Desc'
  }
};

/**
 * @credit Adapted from React Table's default config
 * @see https://react-table.tanstack.com/docs/overview
 */

export const dataTableConfig = {
  debounceMs: 300,
  throttleMs: 50,
  textOperators: [
    { label: 'Contains', value: FILTER_OPERATORS.ILIKE },
    { label: 'Does not contain', value: FILTER_OPERATORS.NOT_ILIKE },
    { label: 'Is', value: FILTER_OPERATORS.EQ },
    { label: 'Is not', value: FILTER_OPERATORS.NEQ },
    { label: 'Is empty', value: FILTER_OPERATORS.EMPTY },
    { label: 'Is not empty', value: FILTER_OPERATORS.NOT_EMPTY }
  ] satisfies { label: string; value: FilterOperator }[],
  numericOperators: [
    { label: 'Is', value: FILTER_OPERATORS.EQ },
    { label: 'Is not', value: FILTER_OPERATORS.NEQ },
    { label: 'Is less than', value: FILTER_OPERATORS.LT },
    {
      label: 'Is less than or equal to',
      value: FILTER_OPERATORS.LTE
    },
    { label: 'Is greater than', value: FILTER_OPERATORS.GT },
    {
      label: 'Is greater than or equal to',
      value: FILTER_OPERATORS.GTE
    },
    { label: 'Is between', value: FILTER_OPERATORS.BETWEEN },
    { label: 'Is empty', value: FILTER_OPERATORS.EMPTY },
    { label: 'Is not empty', value: FILTER_OPERATORS.NOT_EMPTY }
  ] satisfies { label: string; value: FilterOperator }[],
  dateOperators: [
    { label: 'Is', value: FILTER_OPERATORS.EQ },
    { label: 'Is not', value: FILTER_OPERATORS.NEQ },
    { label: 'Is before', value: FILTER_OPERATORS.LT },
    { label: 'Is after', value: FILTER_OPERATORS.GT },
    { label: 'Is on or before', value: FILTER_OPERATORS.LTE },
    { label: 'Is on or after', value: FILTER_OPERATORS.GTE },
    { label: 'Is between', value: FILTER_OPERATORS.BETWEEN },
    // FILTER_OPERATORS.RELATIVE hidden until its filter case ships in
    // `lib/filter-functions.ts`. The enum stays in the catalogue for
    // server-side consumers, re-add this entry when wiring the logic.
    { label: 'Is empty', value: FILTER_OPERATORS.EMPTY },
    { label: 'Is not empty', value: FILTER_OPERATORS.NOT_EMPTY }
  ] satisfies { label: string; value: FilterOperator }[],
  selectOperators: [
    { label: 'Is', value: FILTER_OPERATORS.EQ },
    { label: 'Is not', value: FILTER_OPERATORS.NEQ },
    { label: 'Is empty', value: FILTER_OPERATORS.EMPTY },
    { label: 'Is not empty', value: FILTER_OPERATORS.NOT_EMPTY }
  ] satisfies { label: string; value: FilterOperator }[],
  multiSelectOperators: [
    { label: 'Has any of', value: FILTER_OPERATORS.IN },
    { label: 'Has none of', value: FILTER_OPERATORS.NOT_IN },
    { label: 'Is empty', value: FILTER_OPERATORS.EMPTY },
    { label: 'Is not empty', value: FILTER_OPERATORS.NOT_EMPTY }
  ] satisfies { label: string; value: FilterOperator }[],
  booleanOperators: [
    { label: 'Is', value: FILTER_OPERATORS.EQ },
    { label: 'Is not', value: FILTER_OPERATORS.NEQ }
  ] satisfies { label: string; value: FilterOperator }[],
  sortOrders: [
    { label: 'Asc', value: 'asc' as const },
    { label: 'Desc', value: 'desc' as const }
  ],
  filterVariants: [
    FILTER_VARIANTS.TEXT,
    FILTER_VARIANTS.NUMBER,
    FILTER_VARIANTS.RANGE,
    FILTER_VARIANTS.DATE,
    FILTER_VARIANTS.DATE_RANGE,
    FILTER_VARIANTS.BOOLEAN,
    FILTER_VARIANTS.SELECT,
    FILTER_VARIANTS.MULTI_SELECT
  ] satisfies FilterVariant[],
  operators: [
    FILTER_OPERATORS.ILIKE,
    FILTER_OPERATORS.NOT_ILIKE,
    FILTER_OPERATORS.EQ,
    FILTER_OPERATORS.NEQ,
    FILTER_OPERATORS.IN,
    FILTER_OPERATORS.NOT_IN,
    FILTER_OPERATORS.EMPTY,
    FILTER_OPERATORS.NOT_EMPTY,
    FILTER_OPERATORS.LT,
    FILTER_OPERATORS.LTE,
    FILTER_OPERATORS.GT,
    FILTER_OPERATORS.GTE,
    FILTER_OPERATORS.BETWEEN,
    FILTER_OPERATORS.RELATIVE
  ] satisfies FilterOperator[],
  joinOperators: [JOIN_OPERATORS.AND, JOIN_OPERATORS.OR] satisfies JoinOperator[]
} as const;

export type DataTableConfig = typeof dataTableConfig;
