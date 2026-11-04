import * as React from 'react';
import type { RowData } from '@tanstack/react-table';
import { Check, PlusCircle, XCircle } from 'lucide-react';
import {
  FILTER_OPERATORS,
  FILTER_VARIANTS,
  JOIN_OPERATORS
} from '@/components/data-table/lib/constants';
import type { ExtendedColumnFilter, Option } from '@/components/data-table/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandCollection,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator
} from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Separator } from '@/components/ui/separator';
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
 * Table faceted filter component
 * @description A faceted filter component for DataTable that allows users to filter data based on multiple selectable options. It supports both single and multiple selection modes.
 */

import type { AppColumn } from '@/lib/table';
import { cn } from '@/lib/utils';

export interface TableFacetedFilterProps<TData extends RowData, TValue> {
  column?: AppColumn<TData, TValue>;
  title?: string;
  options: Option[];
  multiple?: boolean;
  /**
   * Callback fired when filter value changes
   * Useful for server-side filtering or external state management
   */
  onValueChange?: (value: string[] | undefined) => void;
  /**
   * Optional custom trigger element
   */
  /** Rendered *as* the trigger, so it must be a single element the
   *  trigger can attach its props and ref to. */
  trigger?: React.ReactElement;
}

export function useTableFacetedFilter<TData extends RowData, TValue = unknown>({
  column,
  onValueChange,
  multiple
}: {
  column?: AppColumn<TData, TValue>;
  onValueChange?: (value: string[] | undefined) => void;
  multiple?: boolean;
}) {
  const columnFilterValue = column?.getFilterValue();

  /**
   * Whether this facet takes more than one value.
   *
   * Read from the column when the caller says nothing, because the column
   * already declares it: `meta.variant: "multiSelect"`. No call site in this
   * app passed `multiple` - all 28 of them - so every facet ran the
   * single-select branch and picking a second option silently replaced the
   * first, while the column meta, the `IN` operator and the repeated-param
   * facet contract all said the opposite. Declaring it in one place per column
   * beats repeating it at every call.
   */
  const isMultiple = multiple ?? column?.columnDef.meta?.variant === FILTER_VARIANTS.MULTI_SELECT;

  // Handle both ExtendedColumnFilter format (new) and legacy array format
  const selectedValues = React.useMemo(() => {
    // Handle ExtendedColumnFilter format (from filter menu or new faceted filter)
    if (
      columnFilterValue &&
      typeof columnFilterValue === 'object' &&
      !Array.isArray(columnFilterValue) &&
      'value' in columnFilterValue
    ) {
      const filterValue = (columnFilterValue as ExtendedColumnFilter<TData>).value;
      return new Set(Array.isArray(filterValue) ? filterValue : [String(filterValue)]);
    }
    // Handle legacy array format (backward compatibility)
    return new Set(Array.isArray(columnFilterValue) ? columnFilterValue : []);
  }, [columnFilterValue]);

  const onItemSelect = React.useCallback(
    (option: Option, isSelected: boolean) => {
      if (!column) return;

      if (isMultiple) {
        const newSelectedValues = new Set(selectedValues);
        if (isSelected) {
          newSelectedValues.delete(option.value);
        } else {
          newSelectedValues.add(option.value);
        }
        const filterValues = Array.from(newSelectedValues);

        if (filterValues.length < 1) {
          column.setFilterValue(undefined);
          onValueChange?.(undefined);
        } else {
          // Create ExtendedColumnFilter format for interoperability with filter menu
          // FORCE variant to multiSelect when using IN operator to ensure it shows up in the menu
          const extendedFilter: ExtendedColumnFilter<TData> = {
            id: column.id as Extract<keyof TData, string>,
            value: filterValues,
            variant: FILTER_VARIANTS.MULTI_SELECT,
            operator: FILTER_OPERATORS.IN,
            filterId: `faceted-${column.id}`,
            joinOperator: JOIN_OPERATORS.AND
          };
          column.setFilterValue(extendedFilter);
          onValueChange?.(filterValues);
        }
      } else {
        // Single selection
        if (isSelected) {
          column.setFilterValue(undefined);
          onValueChange?.(undefined);
        } else {
          // Create ExtendedColumnFilter format for single selection
          // Use EQUAL operator for single select
          const extendedFilter: ExtendedColumnFilter<TData> = {
            id: column.id as Extract<keyof TData, string>,
            value: option.value, // Single value, not array
            variant: FILTER_VARIANTS.SELECT,
            operator: FILTER_OPERATORS.EQ,
            filterId: `faceted-${column.id}`,
            joinOperator: JOIN_OPERATORS.AND
          };
          column.setFilterValue(extendedFilter);
          onValueChange?.([option.value]);
        }
      }
    },
    [column, isMultiple, selectedValues, onValueChange]
  );

  const onReset = React.useCallback(
    (event?: React.MouseEvent) => {
      event?.stopPropagation();
      column?.setFilterValue(undefined);
      onValueChange?.(undefined);
    },
    [column, onValueChange]
  );

  return {
    selectedValues,
    onItemSelect,
    onReset,
    /** The resolved value, so a caller that omitted `multiple` behaves as one. */
    isMultiple
  };
}

export function TableFacetedFilter<TData extends RowData, TValue>({
  column,
  title,
  options = [],
  multiple,
  onValueChange,
  trigger
}: TableFacetedFilterProps<TData, TValue>) {
  const [open, setOpen] = React.useState(false);

  const { selectedValues, onItemSelect, onReset, isMultiple } = useTableFacetedFilter({
    column,
    onValueChange,
    multiple
  });

  // Close only a single-select popover: on a multi-select facet the whole point
  // is picking a second option without reopening it. `isMultiple` rather than
  // `multiple`, or a facet that takes its setting from the column would close
  // after every tick.
  const handleItemSelect = React.useCallback(
    (option: Option, isSelected: boolean) => {
      onItemSelect(option, isSelected);
      if (!isMultiple) {
        setOpen(false);
      }
    },
    [onItemSelect, isMultiple, setOpen]
  );

  return (
    <Popover
      open={open}
      onOpenChange={setOpen}
    >
      <PopoverTrigger
        render={
          trigger || (
            <Button
              variant="ghost"
              size="sm"
              className="border-dashed"
            >
              {selectedValues?.size > 0 ? (
                <div
                  role="button"
                  aria-label={`Clear ${title} filter`}
                  tabIndex={0}
                  onClick={onReset}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      onReset(e as unknown as React.MouseEvent);
                    }
                  }}
                  className="focus-visible:ring-primary rounded-full opacity-70 transition-opacity hover:opacity-100 focus-visible:ring-1 focus-visible:outline-none"
                >
                  <XCircle className="size-4" />
                </div>
              ) : (
                <PlusCircle className="size-4" />
              )}
              {title}
              {selectedValues?.size > 0 && (
                <>
                  <Separator
                    orientation="vertical"
                    className="mx-2 h-4"
                  />
                  <Badge
                    variant="secondary"
                    className="px-1.5 font-normal lg:hidden"
                  >
                    {selectedValues.size}
                  </Badge>
                  <div className="hidden items-center gap-1 lg:flex">
                    {selectedValues.size > 2 ? (
                      <Badge
                        variant="secondary"
                        className="px-1.5 font-normal"
                      >
                        {selectedValues.size} selected
                      </Badge>
                    ) : (
                      options
                        .filter((option) => selectedValues.has(option.value))
                        .map((option) => (
                          <Badge
                            variant="secondary"
                            key={option.value}
                            className="px-1.5 font-normal"
                          >
                            {option.label}
                          </Badge>
                        ))
                    )}
                  </div>
                </>
              )}
            </Button>
          )
        }
      />
      <PopoverContent
        className="w-60 p-0"
        align="start"
      >
        <TableFacetedFilterContent
          title={title}
          options={options}
          selectedValues={selectedValues}
          onItemSelect={handleItemSelect}
          onReset={onReset}
        />
      </PopoverContent>
    </Popover>
  );
}

export function TableFacetedFilterContent({
  title,
  options,
  selectedValues,
  onItemSelect,
  onReset
}: {
  title?: string;
  options: Option[];
  selectedValues: Set<string>;
  onItemSelect: (option: Option, isSelected: boolean) => void;
  onReset: (event?: React.MouseEvent) => void;
}) {
  return (
    // `items` is the collection the primitive counts and filters. Without it it
    // has none, so the search box did nothing and "No results found." sat above
    // a full list of options.
    <Command items={options}>
      <CommandInput
        placeholder={title}
        size="sm"
      />
      <CommandList className="max-h-full">
        <CommandEmpty>No results found.</CommandEmpty>
        {/* No `items` here: the group would then render its own list rather
            than the root's filtered one, and the search would hide nothing. */}
        <CommandGroup className="max-h-75 overflow-x-hidden overflow-y-auto pe-2">
          {/* Rendered from the collection rather than mapped: the primitive
              filters what it owns, so options mapped by hand stayed on screen
              while the search hid nothing and the empty message appeared over
              them. */}
          <CommandCollection>
            {(option: Option) => {
              const isSelected = selectedValues.has(option.value);

              return (
                <CommandItem
                  key={option.value}
                  value={option.value}
                  onClick={() => onItemSelect(option, isSelected)}
                >
                  <div
                    className={cn(
                      'border-primary me-2 flex size-4 items-center justify-center rounded-sm border',
                      isSelected ? 'bg-primary text-secondary-950' : 'opacity-50 [&_svg]:invisible'
                    )}
                  >
                    <Check className="size-4" />
                  </div>
                  {option.icon && <option.icon className="me-2 size-4" />}
                  <span className="truncate">{option.label}</span>
                  {option.count !== undefined && (
                    <span className="ms-auto font-mono text-xs">{option.count}</span>
                  )}
                </CommandItem>
              );
            }}
          </CommandCollection>
        </CommandGroup>
        {selectedValues.size > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup>
              <CommandItem
                onClick={() => onReset()}
                className="justify-center text-center"
              >
                Clear filters
              </CommandItem>
            </CommandGroup>
          </>
        )}
      </CommandList>
    </Command>
  );
}

/**
 * @required displayName is required for auto feature detection
 * @see "feature-detection.ts"
 */

TableFacetedFilter.displayName = 'TableFacetedFilter';
