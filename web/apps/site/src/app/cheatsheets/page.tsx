import type { Metadata } from 'next';
import { CheatSheetTable } from '@braswelljr/algorithm-ui/algorithm/cheat-sheet-table';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { PageHeading } from '@/components/shared/page-header';
import { PrintButton } from '@/components/cheatsheets/print-button';
import { buildCatalog } from '@/lib/catalog.server';
import { categoriesWithEntries, categoryTitle, docsHref } from '@/lib/entries';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Cheat sheets',
  description:
    'Every category on one page: each entry’s time and space cost and its key idea, ready to print.',
  path: '/cheatsheets/'
});

/**
 * One cheat sheet per category, generated from the catalog at build time:
 * costs and key ideas, printable (the print stylesheet drops the site around it).
 */
export default async function CheatSheetsPage() {
  const catalog = await buildCatalog();
  const groups = categoriesWithEntries(catalog);
  let topic = '';
  return (
    <>
      <PageHeading title="Cheat sheets">
        Every category on one page: what each entry costs and the idea that makes it work. Print
        them, or keep one open beside the visualizer.
      </PageHeading>
      <div className="mb-6 flex flex-wrap items-center gap-3 print:hidden">
        <PrintButton />
        <nav aria-label="Categories">
          <ul className="flex flex-wrap gap-x-3 gap-y-1 text-sm">
            {groups.map(({ category }) => (
              <li key={category.id}>
                <a
                  href={`#${category.id}`}
                  className="text-accent underline-offset-4 hover:underline"
                >
                  {categoryTitle(category)}
                </a>
              </li>
            ))}
          </ul>
        </nav>
      </div>
      <div className="flex flex-col gap-8">
        {groups.map(({ category, entries }) => {
          const heading =
            category.topic !== topic ? (
              <Heading
                level={2}
                className="mt-4 print:break-before-page"
              >
                {category.topic}
              </Heading>
            ) : null;
          topic = category.topic;
          return (
            <section
              key={category.id}
              id={category.id}
              aria-label={categoryTitle(category)}
              className="flex scroll-mt-20 flex-col gap-3"
            >
              {heading}
              <CheatSheetTable
                caption={`${categoryTitle(category)}: ${category.summary}`}
                rows={entries.map((e) => ({
                  id: e.id,
                  name: e.name,
                  href: docsHref(e.id),
                  level: e.level,
                  time: e.time,
                  space: e.space,
                  keyIdea: e.notes?.keyIdea ?? e.summary
                }))}
              />
            </section>
          );
        })}
      </div>
    </>
  );
}
