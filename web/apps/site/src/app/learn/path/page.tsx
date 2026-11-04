import { Suspense } from 'react';
import type { Metadata } from 'next';
import { LearnPathApp } from '@/components/learn/learn-path';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = {
  ...pageMetadata({
    title: 'Learning path',
    description: 'One learning path: the entries to run in order, and what to notice at each.',
    path: '/learn/path/'
  }),
  // One page for every id: the per-id static pages are the ones to index.
  robots: { index: false, follow: true }
};

export default function LearnPathPage() {
  return (
    <Suspense>
      <LearnPathApp />
    </Suspense>
  );
}
