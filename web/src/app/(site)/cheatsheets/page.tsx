import { Suspense } from 'react';
import type { Metadata } from 'next';
import { CheatSheetStatic } from '@/components/cheatsheets/cheat-sheet-static';
import { CheatSheetTable, type CheatRow } from '@/components/cheatsheets/cheat-sheet-table';
import { PageHeading } from '@/components/shared/page-header';
import { buildCatalog } from '@/lib/catalog.server';
import { categoryTitle } from '@/lib/entries';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Cheat sheets',
  description:
    'Every entry’s time and space cost and its key idea, in one table you can search, filter, sort and export.',
  path: '/cheatsheets/'
});

/**
 * Every entry's cost and key idea as one data-table, built from the catalog
 * at build time so the page ships its rows in the HTML.
 */
export default async function CheatSheetsPage() {
  const catalog = await buildCatalog();
  const categories = new Map(catalog.categories.map((c) => [c.id, c]));
  const rows: CheatRow[] = catalog.entries.map((e) => {
    const c = categories.get(e.category);
    return {
      id: e.id,
      name: e.name,
      topic: c?.topic ?? '',
      category: c ? categoryTitle(c) : e.category,
      level: e.level,
      time: e.time,
      space: e.space,
      keyIdea: e.notes?.keyIdea ?? e.summary
    };
  });
  return (
    <>
      <PageHeading title="Cheat sheets">
        What each entry costs and the idea that makes it work, in one table: search it, filter it by
        domain, category or level, sort any column, or export it as CSV.
      </PageHeading>
      <Suspense fallback={<CheatSheetStatic rows={rows} />}>
        <CheatSheetTable rows={rows} />
      </Suspense>
    </>
  );
}
