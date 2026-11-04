import { Suspense } from 'react';
import type { Metadata } from 'next';
import { TagsExplorer } from '@/components/docs/tags-explorer';
import { TagsIndex } from '@/components/docs/tags-index';
import { PageHeading } from '@/components/shared/page-header';
import { buildCatalogSummary } from '@/lib/catalog.server';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Tags',
  description: 'Every tag of the catalog, what it means, and the entries that carry it.',
  path: '/docs/tags/'
});

export default async function TagsPage() {
  const catalog = await buildCatalogSummary();
  return (
    <>
      {/* The title is in the top bar; the intro sits beside the tag. */}
      <PageHeading title="Tags" />
      {/* The full list is the exported HTML (for search engines and readers
          without JavaScript); the list and its tag take over in the browser. */}
      <Suspense fallback={<TagsIndex catalog={catalog} />}>
        <TagsExplorer
          catalog={catalog}
          intro={
            <p>
              What each tag means, and every algorithm, data structure and simulation that carries
              it.
            </p>
          }
        />
      </Suspense>
    </>
  );
}
