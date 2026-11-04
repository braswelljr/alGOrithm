import { Suspense } from 'react';
import type { Metadata } from 'next';
import { GameArena } from '@/components/arena/arena';
import { PageHeading } from '@/components/shared/page-header';
import { Skeleton } from '@/components/ui/skeleton';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'The arena',
  description:
    'Games played against the real algorithms: Snake with a pathfinder autopilot, a swap race against every sort, and the keyboard shortcuts as a scored round.',
  path: '/shortcuts/'
});

export default function ArenaPage() {
  return (
    <>
      {/* The title is in the top bar; the console says the rest. */}
      <PageHeading title="The arena" />
      <Suspense fallback={<Skeleton className="h-[34rem] w-full rounded-3xl" />}>
        <GameArena />
      </Suspense>
    </>
  );
}
