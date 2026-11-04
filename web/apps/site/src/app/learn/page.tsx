import { Suspense } from 'react';
import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { LearnIndex } from '@/components/learn/learn-index';
import { buildPaths } from '@/lib/catalog.server';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Learn',
  description:
    'Learning paths: ordered walks through the algorithms, data structures and operating-system simulations, with what to notice at each step.',
  path: '/learn/'
});

export default async function LearnPage() {
  const paths = await buildPaths();
  return (
    <>
      <PageHeading title="Learn">
        Ordered walks through the catalog. Each step runs one entry on an input chosen to show
        something, and says what to watch for.
      </PageHeading>
      <Suspense>
        <LearnIndex paths={paths} />
      </Suspense>
    </>
  );
}
