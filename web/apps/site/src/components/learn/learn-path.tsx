'use client';

import { PathCheck } from './path-check';
import type { ReactNode } from 'react';
import Link from 'next/link';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import type { AlgorithmApi, Catalog } from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryState } from 'nuqs';
import { WasmGate } from '@/components/shared/wasm-gate';
import { CopyButton } from '@/components/run/copy-button';
import { Badge } from '@/components/ui/badge';
import { docsHref } from '@/lib/entries';
import { pathHref, stepHref } from '@/lib/learn';
import { repoUrl } from '@/config/site';

function PathView({ api, catalog, id }: { api: AlgorithmApi; catalog: Catalog; id: string }) {
  const paths = api.paths();
  if (!paths.ok) {
    return (
      <p
        role="alert"
        className="text-danger"
      >
        The learning paths could not be loaded: {paths.failure.error}
      </p>
    );
  }
  const path = paths.value.find((p) => p.id === id);
  if (!path) {
    return (
      <p role="alert">
        There is no learning path called &ldquo;{id}&rdquo;.{' '}
        <Link
          className="underline"
          href="/learn/"
        >
          See them all
        </Link>
        .
      </p>
    );
  }
  const titles = new Map(paths.value.map((p) => [p.id, p.title]));
  const entries = new Map(catalog.entries.map((e) => [e.id, e]));
  return (
    <article className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <p className="text-sm text-muted">
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
        <p className="max-w-3xl text-muted">{path.summary}</p>
        <p className="flex flex-wrap items-center gap-2 text-sm text-muted">
          <Badge>{path.level}</Badge>
          <span>for {path.audience}</span>
          <span>{path.steps.length} steps</span>
        </p>
        {path.prerequisites.length > 0 ? (
          <p className="text-sm">
            Walk first:{' '}
            {path.prerequisites.map((p, i) => (
              <span key={p}>
                {i > 0 ? ', ' : null}
                <Link
                  className="underline"
                  href={pathHref(p)}
                >
                  {titles.get(p) ?? p}
                </Link>
              </span>
            ))}
          </p>
        ) : null}
        {path.guide ? (
          <p className="text-sm">
            Read alongside:{' '}
            <a
              className="underline"
              href={`${repoUrl}/blob/main/${path.guide}`}
            >
              {path.guide}
            </a>
          </p>
        ) : null}
      </header>
      <ol className="flex flex-col gap-4">
        {path.steps.map((step, i) => {
          const entry = entries.get(step.entry);
          return (
            <Card
              size="md"
              render={<li />}
              key={`${step.entry}-${i}`}
              className="px-(--card-spacing)"
            >
              <Heading
                level={2}
                className="font-semibold"
              >
                <span className="mr-2 font-mono text-sm text-muted">{i + 1}.</span>
                {step.name}
              </Heading>
              <p className="mt-1">{step.note}</p>
              <div className="mt-3 flex flex-wrap items-center gap-3 text-sm">
                {entry ? (
                  <>
                    <Link
                      className="rounded-md bg-accent px-3 py-1.5 font-medium text-accent-fg"
                      href={stepHref(entry, step)}
                    >
                      Run it
                    </Link>
                    <Link
                      className="underline"
                      href={docsHref(entry.id)}
                    >
                      How it works
                    </Link>
                  </>
                ) : (
                  <span className="text-muted">This entry is not in this build.</span>
                )}
                <CopyButton text={() => step.command}>Copy the command</CopyButton>
              </div>
            </Card>
          );
        })}
      </ol>
      <PathCheck
        id={path.id}
        title={path.title}
      />
    </article>
  );
}

/**
 * One learning path: the one named by id, or by `?id=` in the link when there
 * is none. loading is shown until the wasm build is ready.
 */
export function LearnPathApp({ id: fixedId, loading }: { id?: string; loading?: ReactNode } = {}) {
  const [queryId] = useQueryState('id', parseAsString.withDefault(''));
  const id = fixedId ?? queryId;
  return (
    <WasmGate loading={loading}>
      {(api, catalog) => (
        <PathView
          api={api}
          catalog={catalog}
          id={id}
        />
      )}
    </WasmGate>
  );
}
