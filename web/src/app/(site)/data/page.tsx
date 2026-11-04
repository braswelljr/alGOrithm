import { Suspense } from 'react';
import type { Metadata } from 'next';
import { DataApp } from '@/components/data/data-app';
import { PageHeading } from '@/components/shared/page-header';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Your data',
  description:
    'Paste values, open a CSV or JSON file, fetch one from a link, or generate a dataset, then run any algorithm on it.',
  path: '/data/'
});

export default function DataPage() {
  return (
    <>
      <PageHeading title="Your data">
        Paste values, open a CSV or JSON file, load one from a link, or generate a dataset. Your
        browser reads it; this site never receives it.
      </PageHeading>
      <Suspense>
        <DataApp />
      </Suspense>
    </>
  );
}
