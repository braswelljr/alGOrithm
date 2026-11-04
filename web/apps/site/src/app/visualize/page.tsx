import { Suspense } from 'react';
import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { VisualizeApp } from '@/components/visualize/visualize-app';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Visualize',
  description:
    'Watch a sort, a search, a pathfinder, a dynamic program or a string search run on your data, step by step.',
  path: '/visualize/'
});

export default function VisualizePage() {
  return (
    <>
      <PageHeading title="Visualize">
        Pick an algorithm and its input, then play, pause, step or scrub. The link keeps everything,
        so you can share the exact step you are on.
      </PageHeading>
      <Suspense>
        <VisualizeApp />
      </Suspense>
    </>
  );
}
