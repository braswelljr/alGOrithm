import { Suspense } from 'react';
import type { Metadata } from 'next';
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
      <PageHeading title="Tags">
        What each tag means, and every algorithm, data structure and simulation that carries it.
      </PageHeading>
      <Suspense>
        <TagsIndex catalog={catalog} />
      </Suspense>
    </>
  );
}
