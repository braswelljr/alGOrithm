import type { Metadata } from 'next';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { Catalog, Entry } from '@braswelljr/algorithm-visualizer/core';
import { DocsEntryApp } from '@/components/docs/docs-entry';
import { JsonLd } from '@/components/shared/json-ld';
import { buildCatalog } from '@/lib/catalog.server';
import { categoryTitle, docsHref } from '@/lib/entries';
import { pageMetadata } from '@/lib/metadata';
import { breadcrumbLd, entryLd } from '@/lib/structured-data';

type Props = { params: Promise<{ id: string }> };

// Every entry is a page built ahead of time; there are no others.
export const dynamicParams = false;

export async function generateStaticParams() {
  return (await buildCatalog()).entries.map((e) => ({ id: e.id }));
}

async function find(id: string): Promise<{ catalog: Catalog; entry: Entry } | undefined> {
  const catalog = await buildCatalog();
  const entry = catalog.entries.find((e) => e.id === id);
  return entry ? { catalog, entry } : undefined;
}

function categoryName(catalog: Catalog, e: Entry): string {
  const c = catalog.categories.find((x) => x.id === e.category);
  return c ? categoryTitle(c) : e.category;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const found = await find((await params).id);
  if (!found) return {};
  const { catalog, entry } = found;
  const cost = `Time ${entry.time}, space ${entry.space}.`;
  return {
    ...pageMetadata({
      title: `${entry.name} (${categoryName(catalog, entry)})`,
      description: `${entry.summary} ${cost} Run it on your own data and watch every step.`,
      path: docsHref(entry.id),
      images: [`${docsHref(entry.id)}og.png`]
    }),
    keywords: [entry.name, categoryName(catalog, entry), ...entry.tags, 'algorithm visualizer']
  };
}

/**
 * What a reader and a search engine see before the wasm build runs: the
 * entry's name, summary and cost, from the catalog at build time. The full
 * page replaces it once the algorithms load.
 */
function Prerendered({ catalog, entry }: { catalog: Catalog; entry: Entry }) {
  const category = categoryName(catalog, entry);
  return (
    <article className="flex flex-col gap-4">
      <p className="text-sm text-muted">
        <Link
          className="underline"
          href="/docs/"
        >
          Docs
        </Link>{' '}
        / {category}
      </p>
      <Heading
        level={1}
        className="text-2xl font-semibold"
      >
        {entry.name}
      </Heading>
      <p className="max-w-3xl">{entry.summary}</p>
      <dl className="grid max-w-xl grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-muted">Time</dt>
        <dd className="font-mono">{entry.time}</dd>
        <dt className="text-muted">Space</dt>
        <dd className="font-mono">{entry.space}</dd>
        {entry.level ? (
          <>
            <dt className="text-muted">Level</dt>
            <dd>{entry.level}</dd>
          </>
        ) : null}
        {entry.tags.length ? (
          <>
            <dt className="text-muted">Tags</dt>
            <dd>{entry.tags.join(', ')}</dd>
          </>
        ) : null}
      </dl>
      {entry.why ? <p className="max-w-3xl text-sm text-muted">{entry.why}</p> : null}
      <p
        role="status"
        className="text-sm text-muted"
      >
        Loading the interactive version...
      </p>
    </article>
  );
}

export default async function DocsEntryStaticPage({ params }: Props) {
  const found = await find((await params).id);
  if (!found) notFound();
  const { catalog, entry } = found;
  const path = docsHref(entry.id);
  return (
    <>
      <JsonLd data={entryLd(entry, categoryName(catalog, entry), path)} />
      <JsonLd
        data={breadcrumbLd([
          { name: 'Home', path: '/' },
          { name: 'Docs', path: '/docs/' },
          { name: entry.name, path }
        ])}
      />
      <DocsEntryApp
        id={entry.id}
        loading={
          <Prerendered
            catalog={catalog}
            entry={entry}
          />
        }
      />
    </>
  );
}
