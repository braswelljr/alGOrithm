import { Suspense } from 'react';
import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { StructuresApp } from '@/components/structures/structures-app';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Data structures',
  description:
    'Push, pop, insert and remove by hand and watch stacks, queues, heaps, trees, tries, hash tables and more change.',
  path: '/structures/'
});

export default function StructuresPage() {
  return (
    <>
      <PageHeading title="Data structure playground">
        Choose a structure, start it with some values, and apply operations one at a time. Step back
        through them to see each change.
      </PageHeading>
      <Suspense>
        <StructuresApp />
      </Suspense>
    </>
  );
}
