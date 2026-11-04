import { Suspense } from 'react';
import type { Metadata } from 'next';
import { DocsIndex } from '@/components/docs/docs-index';
import { PageHeading } from '@/components/shared/page-header';
import { buildCatalogSummary } from '@/lib/catalog.server';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Documentation',
  description:
    'Every algorithm and data structure: what it does, what it costs, how to use it from the command line and from Go, and its source.',
  path: '/docs/'
});

export default async function DocsPage() {
  const catalog = await buildCatalogSummary();
  return (
    <>
      <PageHeading title="Documentation">
        Every algorithm and data structure in the library, by package: what it does, what it costs,
        how to call it, and its source.
      </PageHeading>
      <Suspense>
        <DocsIndex catalog={catalog} />
      </Suspense>
    </>
  );
}
