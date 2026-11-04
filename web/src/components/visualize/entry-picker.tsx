'use client';

import { useId } from 'react';
import { Combobox } from '@/components/ui/combobox';
import { categoriesWithEntries, categoryTitle, isStructure } from '@/lib/entries';
import type { Catalog } from '@/visualizer/core';

/** Picks an entry to visualize, grouped by category, from the catalog. */
export function EntryPicker({
  catalog,
  value,
  onChange
}: {
  catalog: Catalog;
  value: string;
  onChange: (id: string) => void;
}) {
  const id = useId();
  return (
    <div className="flex flex-col gap-1.5">
      <label
        htmlFor={id}
        className="text-sm font-medium"
      >
        Algorithm
      </label>
      <Combobox
        id={id}
        value={value}
        onChange={onChange}
        className="sm:w-96"
        placeholder="Choose an entry"
        searchPlaceholder="Search entries"
        // One list to type into, each entry carrying its category, where the
        // native select grouped them under headings.
        options={categoriesWithEntries(catalog).flatMap(({ category, entries }) =>
          entries
            .filter((e) => !isStructure(e))
            .map((e) => ({ value: e.id, label: e.name, hint: categoryTitle(category) }))
        )}
      />
    </div>
  );
}
