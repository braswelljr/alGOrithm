// A facet bound to a table column.
//
// The column is the state: its filter value is what the facet shows as
// selected, and its faceted unique values are the counts. That is the whole
// component, and it exists because the alternative was six lines of reading and
// writing column state repeated beside every facet in every table header, where
// one of them eventually reads `undefined` differently from the rest.
//
// For a table filtered server-side, pass counts from `meta.facets` to
// `FacetFilter` directly instead: the numbers there are computed over the whole
// result set rather than the page in hand.

import type { RowData } from '@tanstack/react-table';
import { FacetFilter, type FacetFilterOption } from '@/components/ui/facet-filter';
import type { AppColumn } from '@/lib/table';

export function TableFacet<TData extends RowData, TValue>({
  column,
  title,
  options,
  valueType = 'string'
}: {
  /** Undefined when the table was rendered without this column, in which case
   *  nothing renders: a filter for a column nobody can see is a dead control. */
  column?: AppColumn<TData, TValue>;
  title: string;
  options: readonly FacetFilterOption[];
  /**
   * What the column stores. A facet always deals in strings, so a boolean
   * column is offered as "true" and "false" and converted back on the way in.
   * Without this each caller writes the same two-line conversion beside its
   * facet, and the third copy is where one of them forgets it.
   */
  valueType?: 'string' | 'boolean';
}) {
  if (!column) return null;

  const current = column.getFilterValue();
  const value = Array.isArray(current) ? current.map(String) : [];
  const counts = Object.fromEntries(
    Array.from(column.getFacetedUniqueValues() ?? []).map(([key, count]) => [String(key), count])
  );

  return (
    <FacetFilter
      title={title}
      options={options}
      value={value}
      counts={counts}
      // Clearing the last value removes the filter rather than leaving an empty
      // array behind, which the table would treat as "match nothing".
      onValueChange={(next) => {
        if (next.length < 1) {
          column.setFilterValue(undefined);
          return;
        }
        column.setFilterValue(
          valueType === 'boolean' ? next.map((entry) => entry === 'true') : next
        );
      }}
    />
  );
}
