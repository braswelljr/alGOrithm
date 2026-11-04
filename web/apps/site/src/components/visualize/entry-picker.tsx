'use client';

import { useId } from 'react';
import type { Catalog } from '@braswelljr/algorithm-visualizer/core';
import { Select } from '@/components/ui/select';
import { categoriesWithEntries, categoryTitle, isStructure } from '@/lib/entries';

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
      <Select
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="sm:w-96"
      >
        {categoriesWithEntries(catalog).map(({ category, entries }) => {
          const runnable = entries.filter((e) => !isStructure(e));
          return runnable.length ? (
            <optgroup
              key={category.id}
              label={categoryTitle(category)}
            >
              {runnable.map((e) => (
                <option
                  key={e.id}
                  value={e.id}
                >
                  {e.name}
                </option>
              ))}
            </optgroup>
          ) : null;
        })}
      </Select>
    </div>
  );
}
