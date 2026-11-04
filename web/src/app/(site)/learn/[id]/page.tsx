import { Suspense } from 'react';
import type { Metadata } from 'next';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { LearnPathApp } from '@/components/learn/learn-path';
import { JsonLd } from '@/components/shared/json-ld';
import { Heading } from '@/components/ui/heading';
import { buildPaths } from '@/lib/catalog.server';
import { pathHref } from '@/lib/learn';
import { pageMetadata } from '@/lib/metadata';
import { breadcrumbLd, pathLd } from '@/lib/structured-data';
import type { LearningPath } from '@/visualizer/core';

type Props = { params: Promise<{ id: string }> };

// Every learning path is a page built ahead of time; there are no others.
export const dynamicParams = false;

export async function generateStaticParams() {
  return (await buildPaths()).map((p) => ({ id: p.id }));
}

async function find(id: string): Promise<LearningPath | undefined> {
  return (await buildPaths()).find((p) => p.id === id);
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const p = await find((await params).id);
  if (!p) return {};
  return pageMetadata({
    title: `${p.title}: a learning path`,
    description: `${p.summary} ${p.steps.length} steps, ${p.level}.`,
    path: pathHref(p.id)
  });
}

/** The path as text, readable before the wasm build runs. */
function Prerendered({ path }: { path: LearningPath }) {
  return (
    <article className="flex flex-col gap-4">
      <p className="text-muted text-sm">
        <Link
          className="underline"
          href="/learn/"
        >
          Learn
        </Link>{' '}
        / {path.title}
      </p>
      <Heading
        level={1}
        className="text-2xl font-semibold"
      >
        {path.title}
      </Heading>
      <p className="">{path.summary}</p>
      <ol className="flex list-decimal flex-col gap-2 pl-6">
        {path.steps.map((s, i) => (
          <li key={`${i}-${s.entry}`}>
            <span className="font-medium">{s.name || s.entry}</span>: {s.note}
          </li>
        ))}
      </ol>
      <p
        role="status"
        className="text-muted text-sm"
      >
        Loading the interactive version...
      </p>
    </article>
  );
}

export default async function LearnPathStaticPage({ params }: Props) {
  const p = await find((await params).id);
  if (!p) notFound();
  const href = pathHref(p.id);
  return (
    <>
      <JsonLd data={pathLd(p, href)} />
      <JsonLd
        data={breadcrumbLd([
          { name: 'Home', path: '/' },
          { name: 'Learn', path: '/learn/' },
          { name: p.title, path: href }
        ])}
      />
      <Suspense fallback={<Prerendered path={p} />}>
        <LearnPathApp
          id={p.id}
          loading={<Prerendered path={p} />}
        />
      </Suspense>
    </>
  );
}
