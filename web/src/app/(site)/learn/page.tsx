import { Suspense } from 'react';
import type { Metadata } from 'next';
import { MermaidDiagram } from '@/components/algorithm/mermaid-diagram';
import { LearnIndex } from '@/components/learn/learn-index';
import { PageHeading } from '@/components/shared/page-header';
import { buildPaths } from '@/lib/catalog.server';
import { pathsChart } from '@/lib/learn';
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
      <MermaidDiagram
        chart={pathsChart(paths)}
        label="How the learning paths build on each other: an arrow from a path to the paths that need it first."
        className="mb-6"
      />
      <Suspense>
        <LearnIndex paths={paths} />
      </Suspense>
    </>
  );
}
