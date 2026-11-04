import { Suspense } from 'react';
import type { Metadata } from 'next';
import Link from 'next/link';
import { GrowthChart } from '@/components/big-o/growth-chart';
import { MeasurePanel } from '@/components/big-o/measure-panel';
import { PageHeading } from '@/components/shared/page-header';
import { Card } from '@/components/ui/card';
import { Heading } from '@/components/ui/heading';
import { repoUrl } from '@/config/site';
import { pageMetadata } from '@/lib/metadata';

export const metadata: Metadata = pageMetadata({
  title: 'Big O',
  description:
    'What Big O notation means, why it drops constants, how the classes grow, and how to measure an algorithm to check its stated cost.',
  path: '/big-o/'
});

const ideas = [
  {
    title: 'It counts steps, not seconds',
    body: 'Big O describes how the number of steps grows with the size of the input, n. Seconds depend on the machine; the growth does not.'
  },
  {
    title: 'Only the fastest-growing term matters',
    body: 'n^2/2 + 3n + 7 is O(n^2): at n = 1,000 the n^2 term is over 99% of the work, and doubling n roughly quadruples it whatever the constants are.'
  },
  {
    title: 'Best, average and worst are different questions',
    body: 'Insertion sort is O(n) on sorted values and O(n^2) on random ones; quick sort is O(n log n) on average and O(n^2) when every pivot is the smallest.'
  },
  {
    title: 'Doubling n is the test',
    body: 'Double the input and watch the steps: about the same is O(1), plus one is O(log n), twice is O(n), a bit over twice is O(n log n), four times is O(n^2), eight is O(n^3).'
  }
];

export default function BigOPage() {
  return (
    <>
      <PageHeading title="Big O">
        How the work an algorithm does grows with its input, why that matters more than speed on one
        machine, and how to check a stated cost by measuring it.
      </PageHeading>
      <section
        aria-labelledby="ideas"
        className="mb-10"
      >
        <Heading
          level={2}
          id="ideas"
          className="mb-3 text-lg font-semibold"
        >
          Four ideas
        </Heading>
        <ul className="grid gap-4 sm:grid-cols-2">
          {ideas.map((i) => (
            <Card
              size="md"
              render={<li />}
              key={i.title}
              className="px-(--card-spacing)"
            >
              <Heading
                level={3}
                className="font-medium"
              >
                {i.title}
              </Heading>
              <p className="text-muted mt-1 text-sm">{i.body}</p>
            </Card>
          ))}
        </ul>
      </section>
      <section
        aria-labelledby="growth"
        className="mb-10"
      >
        <Heading
          level={2}
          id="growth"
          className="mb-3 text-lg font-semibold"
        >
          How the classes grow
        </Heading>
        <GrowthChart />
      </section>
      <section
        aria-labelledby="measure"
        className="mb-10"
      >
        <Heading
          level={2}
          id="measure"
          className="mb-1 text-lg font-semibold"
        >
          Measure it
        </Heading>
        <p className="text-muted mb-4 text-sm">
          Run an entry at growing sizes and count the steps it reports. The fit is the class whose
          curve explains the counts best. Counts are exact and repeatable, unlike time. Try bubble
          sort on sorted, then on random values, or tree sort on sorted values.
        </p>
        <Suspense>
          <MeasurePanel />
        </Suspense>
      </section>
      <p className="text-muted text-sm">
        The full guide, with every claim tied to an entry you can run:{' '}
        <a
          className="underline"
          href={`${repoUrl}/blob/main/docs/guides/big-o.md`}
        >
          docs/guides/big-o.md
        </a>
        . Every entry&apos;s page says where its cost comes from; start with the{' '}
        <Link
          className="underline"
          href="/"
        >
          catalog
        </Link>
        .
      </p>
    </>
  );
}
