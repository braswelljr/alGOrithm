'use client';

import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { Catalog } from '@braswelljr/algorithm-visualizer/core';
import { useFilters } from '@/hooks/use-filters';
import { categoryTitle, topicsWithCategories } from '@/lib/entries';
import { filterCatalog, noFilters } from '@/lib/filters';
import { EntryCard } from './entry-card';
import { FilterBar, NoMatches } from './filter-bar';

/**
 * Every entry, by topic and category, with filters. The catalog comes from the
 * build (buildCatalogSummary), so the list is in the HTML and the page needs
 * no wasm until the reader runs something.
 */
export function CatalogView({ catalog: all }: { catalog: Catalog }) {
  const [filters, setFilters] = useFilters();
  const catalog = filterCatalog(all, filters);
  return (
    <div className="flex flex-col gap-12">
      <FilterBar
        catalog={all}
        filters={filters}
        onChange={setFilters}
        shown={catalog.entries.length}
      />
      {catalog.entries.length === 0 ? <NoMatches onClear={() => setFilters(noFilters)} /> : null}
      {topicsWithCategories(catalog).map(({ topic, groups }) => (
        <section
          key={topic}
          aria-labelledby={`topic-${topic}`}
          className="flex flex-col gap-8"
        >
          <Heading
            level={2}
            id={`topic-${topic}`}
            className="border-b border-border pb-2 text-2xl font-semibold"
          >
            {topic}
          </Heading>
          {groups.map(({ category, entries }) => (
            <section
              key={category.id}
              aria-labelledby={`cat-${category.id}`}
            >
              <Heading
                level={3}
                id={`cat-${category.id}`}
                className="text-lg font-semibold"
              >
                {categoryTitle(category)}
              </Heading>
              {category.summary ? (
                <p className="mt-1 text-sm text-muted">{category.summary}</p>
              ) : null}
              <ul className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {entries.map((e) => (
                  <li key={e.id}>
                    <EntryCard entry={e} />
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </section>
      ))}
      <p className="text-xs text-muted">
        {all.entries.length} entries · version {all.version}
      </p>
    </div>
  );
}
