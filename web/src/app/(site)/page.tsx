import { Suspense } from 'react';
import type { Metadata } from 'next';
import { CatalogView } from '@/components/catalog/catalog-view';
import { PageHeading } from '@/components/shared/page-header';
import { siteDescription, siteName, siteTagline } from '@/config/site';
import { buildCatalogSummary } from '@/lib/catalog.server';
import { pageMetadata } from '@/lib/metadata';

const home = pageMetadata({
  title: 'Algorithms and data structures, step by step',
  description: siteDescription,
  path: '/'
});

export const metadata: Metadata = {
  ...home,
  title: { absolute: `${siteName}: ${siteTagline}` }
};

export default async function Home() {
  const catalog = await buildCatalogSummary();
  return (
    <>
      <PageHeading title="Algorithms and data structures, step by step">
        Run the real Go code on your own data and watch every comparison, swap, step and link. It
        runs as WebAssembly in your browser; nothing you load is sent anywhere.
      </PageHeading>
      <Suspense>
        <CatalogView catalog={catalog} />
      </Suspense>
    </>
  );
}
