import { Suspense } from 'react';
import type { Metadata } from 'next';
import { PageHeading } from '@/components/shared/page-header';
import { SystemDesignHub } from '@/components/system-design/system-design-hub';
import { Skeleton } from '@/components/ui/skeleton';
import { repoUrl } from '@/config/site';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'System design',
  description:
    'Case studies you can run: a URL shortener, a Dynamo-style key-value store, a news feed, chat, payments and more, with their requirements, estimates, trade-offs, failure modes and a live architecture.',
  path: '/system-design/'
});

export default function SystemDesignPage() {
  return (
    <>
      {/* The title is in the top bar; the intro sits inside the hub, above the
          study, so the case studies can fill the screen. */}
      <PageHeading title="System design" />
      <Suspense fallback={<Skeleton className="h-96 w-full" />}>
        <SystemDesignHub
          intro={
            <p>
              Size the problem, draw the boxes, then watch requests flow through them. Every design
              here is a real simulation you can change: add a cache, take a node down, turn on
              idempotency keys, and see what happens. The approach, step by step, is in the{' '}
              <a
                className="underline underline-offset-4"
                href={`${repoUrl}/blob/main/docs/guides/system-design.md`}
              >
                system design guide
              </a>
              .
            </p>
          }
        />
      </Suspense>
    </>
  );
}
