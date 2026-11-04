'use client';

import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Toggle } from '@/components/ui/toggle';
import { facets, isFiltered, noFilters, words, type Filters } from '@/lib/filters';
import type { Catalog } from '@/visualizer/core';

const groups: { key: keyof Filters; label: string }[] = [
  { key: 'kind', label: 'Kind' },
  { key: 'level', label: 'Level' },
  { key: 'tag', label: 'Tags' }
];

/**
 * Chips to narrow the catalog by kind, level and tags. Each chip says how
 * many entries it would leave, is a toggle button (aria-pressed), and shows
 * a check mark when on, so its state is not told by color alone.
 */
export function FilterBar({
  catalog,
  filters,
  onChange,
  shown
}: {
  catalog: Catalog;
  filters: Filters;
  onChange: (next: Partial<Filters>) => void;
  shown: number;
}) {
  return (
    <Card
      size="md"
      render={<section />}
      aria-label="Filters"
      className="flex flex-col gap-3 px-(--card-spacing)"
    >
      {groups.map(({ key, label }) => (
        <fieldset
          key={key}
          className="flex flex-wrap items-center gap-2"
        >
          <legend className="mr-2 text-sm font-medium">{label}</legend>
          {facets(catalog, filters, key).map((f) => (
            <Toggle
              key={f.value}
              variant="chip"
              size="sm"
              pressed={f.active}
              title={f.meaning}
              disabled={!f.active && f.count === 0}
              onPressedChange={() =>
                onChange({
                  [key]: f.active
                    ? filters[key].filter((v) => v !== f.value)
                    : key === 'tag'
                      ? [...filters.tag, f.value]
                      : [f.value]
                })
              }
            >
              {f.active ? '✓ ' : ''}
              {words(f.value)} <span className="font-mono opacity-80">({f.count})</span>
            </Toggle>
          ))}
        </fieldset>
      ))}
      <p
        className="text-muted text-xs"
        aria-live="polite"
      >
        {shown} of {catalog.entries.length} entries.
        {isFiltered(filters) ? (
          <>
            {' '}
            <Button
              variant="link"
              className="h-auto p-0 text-xs"
              onClick={() => onChange(noFilters)}
            >
              Clear the filters
            </Button>
          </>
        ) : null}
      </p>
    </Card>
  );
}

/** What shows when the filters leave nothing. */
export function NoMatches({ onClear }: { onClear: () => void }) {
  return (
    <div className="border-border rounded-lg border border-dashed p-8 text-center">
      <p className="font-medium">No entry has all of these.</p>
      <p className="text-muted mt-1 text-sm">Remove a tag or choose another level.</p>
      <Button
        variant="link"
        className="mt-3 h-auto p-0"
        onClick={onClear}
      >
        Clear the filters
      </Button>
    </div>
  );
}
