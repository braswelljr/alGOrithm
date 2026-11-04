'use client';

import dynamic from 'next/dynamic';
import Link from 'next/link';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { Table } from '@braswelljr/algorithm-ui/components/table';
import type { Catalog } from '@braswelljr/algorithm-visualizer/core';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { FilterBar, NoMatches } from '@/components/catalog/filter-bar';
import { useFilters } from '@/hooks/use-filters';
import {
  categoriesWithEntries,
  categoryTitle,
  docsHref,
  topicsWithCategories
} from '@/lib/entries';
import { filterCatalog, isFiltered, noFilters } from '@/lib/filters';

// The principles panel runs entries through the wasm and draws them with the
// visualizer's views, so it loads only when the structures section shows it.
const LivePrinciples = dynamic(() => import('./principles').then((m) => m.LivePrinciples), {
  ssr: false
});

/**
 * The documentation of every algorithm and data structure, by category,
 * generated from the catalog: it cannot list something the code does not have.
 */
export function DocsIndex({ catalog: all }: { catalog: Catalog }) {
  const [filters, setFilters] = useFilters();
  const catalog = filterCatalog(all, filters);
  return (
    <div className="flex flex-col gap-10">
      <p className="text-sm">
        <Link
          className="underline"
          href="/docs/tags/"
        >
          Every tag and what it means
        </Link>
      </p>
      <FilterBar
        catalog={all}
        filters={filters}
        onChange={setFilters}
        shown={catalog.entries.length}
      />
      {catalog.entries.length === 0 ? <NoMatches onClear={() => setFilters(noFilters)} /> : null}
      <nav aria-label="Categories">
        <ul className="flex flex-wrap gap-2">
          {categoriesWithEntries(catalog).map(({ category, entries }) => (
            <li key={category.id}>
              <a
                href={`#${category.id}`}
                className="rounded-md border border-border px-3 py-1 text-sm hover:border-accent"
              >
                {categoryTitle(category)} ({entries.length})
              </a>
            </li>
          ))}
        </ul>
      </nav>
      {topicsWithCategories(catalog).map(({ topic, groups }) => (
        <div
          key={topic}
          className="flex flex-col gap-10"
        >
          <Heading
            level={2}
            className="border-b border-border pb-2 text-2xl font-semibold"
          >
            {topic}
          </Heading>
          {groups.map(({ category, entries }) => (
            <section
              key={category.id}
              id={category.id}
              aria-labelledby={`h-${category.id}`}
              className="scroll-mt-6"
            >
              <Heading
                level={3}
                id={`h-${category.id}`}
                className="text-xl font-semibold"
              >
                {categoryTitle(category)}
              </Heading>
              {category.import ? (
                <p className="mt-1 font-mono text-xs text-muted">
                  import &quot;{category.import}&quot;
                </p>
              ) : null}
              {category.doc || category.summary ? (
                <p className="mt-2 max-w-3xl text-sm whitespace-pre-wrap">
                  {category.doc || category.summary}
                </p>
              ) : null}
              <Table className="mt-4">
                <caption className="sr-only">
                  {categoryTitle(category)}: every entry and its cost
                </caption>
                <thead>
                  <tr className="border-b border-border text-left text-muted">
                    <th
                      scope="col"
                      className="py-2 pr-4"
                    >
                      Name
                    </th>
                    <th
                      scope="col"
                      className="py-2 pr-4"
                    >
                      What it does
                    </th>
                    <th
                      scope="col"
                      className="py-2 pr-4"
                    >
                      Time
                    </th>
                    <th
                      scope="col"
                      className="py-2"
                    >
                      Space
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map((e) => (
                    <tr
                      key={e.id}
                      className="border-b border-border align-top last:border-0"
                    >
                      <th
                        scope="row"
                        className="py-2 pr-4 text-left font-medium"
                      >
                        <Link
                          className="underline decoration-border underline-offset-2 hover:decoration-accent"
                          href={docsHref(e.id)}
                        >
                          {e.name}
                        </Link>
                      </th>
                      <td className="py-2 pr-4 text-muted">
                        {e.summary}
                        <span className="mt-1 block">
                          <EntryBadges entry={e} />
                        </span>
                      </td>
                      <td className="py-2 pr-4 font-mono text-xs">{e.time}</td>
                      <td className="py-2 font-mono text-xs">{e.space}</td>
                    </tr>
                  ))}
                </tbody>
              </Table>
              {category.id === 'structures' && !isFiltered(filters) ? (
                <LivePrinciples entries={entries} />
              ) : null}
            </section>
          ))}
        </div>
      ))}
    </div>
  );
}
