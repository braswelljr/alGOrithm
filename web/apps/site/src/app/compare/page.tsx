import { Suspense } from 'react';
import type { Metadata } from 'next';
import { CompareApp } from '@/components/compare/compare-app';
import { PageHeading } from '@/components/shared/page-header';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Compare',
  description:
    'Run several sorts on the same data and compare their comparisons, swaps, writes and time.',
  path: '/compare/'
});

export default function ComparePage() {
  return (
    <>
      <PageHeading title="Compare sorts">
        The same data through every sort you pick: how many comparisons, swaps and writes each
        makes, and how long it takes here.
      </PageHeading>
      <Suspense>
        <CompareApp />
      </Suspense>
    </>
  );
}
