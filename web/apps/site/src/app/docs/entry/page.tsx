import { Suspense } from 'react';
import type { Metadata } from 'next';
import { DocsEntryApp } from '@/components/docs/docs-entry';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Documentation',
    description:
      'One algorithm or data structure: what it does, what it costs, how to use it, and its source.',
    path: '/docs/entry/'
  }),
  // One page for every id: the per-id static pages are the ones to index.
  robots: { index: false, follow: true }
};

export default function DocsEntryPage() {
  return (
    <Suspense>
      <DocsEntryApp />
    </Suspense>
  );
}
