'use client';

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from '@braswelljr/algorithm-ui/components/breadcrumb';
import { TLDRCard } from '@braswelljr/algorithm-ui/algorithm/tldr-card';
import { useMemo, type ReactNode } from 'react';
import Link from 'next/link';
import { buttonVariants } from '@braswelljr/algorithm-ui/components/button';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { CodeView } from '@braswelljr/algorithm-visualizer';
import {
  fallbackSnippet,
  type AlgorithmApi,
  type Catalog,
  type Entry
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryState } from 'nuqs';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { RealWorldCode } from '@/components/docs/real-world-code';
import { WasmGate } from '@/components/shared/wasm-gate';
import { CopyButton } from '@/components/run/copy-button';
import { categoryTitle, entryHref } from '@/lib/entries';
import { inputWords } from '@/lib/input-words';
import { repoUrl } from '@/config/site';

function Snippet({ title, code }: { title: string; code: string }) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between gap-2">
        <Heading
          level={3}
          className="text-sm font-semibold"
        >
          {title}
        </Heading>
        <CopyButton text={() => code}>Copy</CopyButton>
      </div>
      <pre className="overflow-x-auto rounded-md border border-border bg-subtle p-3 font-mono text-xs">
        <code>{code}</code>
      </pre>
    </div>
  );
}

function EntryDoc({ api, catalog, entry }: { api: AlgorithmApi; catalog: Catalog; entry: Entry }) {
  const source = useMemo(() => api.source(entry.id), [api, entry.id]);
  const snippet = useMemo(() => {
    const input = entry.example;
    const s = api.snippet(entry.id, input);
    return s.ok ? s.value : fallbackSnippet(entry, input);
  }, [api, entry]);
  const category = catalog.categories.find((c) => c.id === entry.category);
  const file = source.ok ? source.value.file : entry.file;
  return (
    <article className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink render={<Link href="/docs/" />}>Docs</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbLink render={<Link href={`/docs/#${entry.category}`} />}>
                {category ? categoryTitle(category) : entry.category}
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{entry.name}</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
        <Heading
          level={1}
          className="text-2xl font-semibold"
        >
          {entry.name}
        </Heading>
        <p className="max-w-3xl">{entry.summary}</p>
        <EntryBadges
          entry={entry}
          links
        />
        {entry.notes ? (
          <TLDRCard
            notes={entry.notes}
            time={entry.time}
            space={entry.space}
            why={entry.why}
            className="mt-2 max-w-3xl"
          />
        ) : null}
        <p>
          <Link
            href={entryHref(entry)}
            className="inline-flex h-9 items-center rounded-md bg-accent px-4 text-sm font-medium text-accent-fg hover:opacity-90"
          >
            Try it{' '}
            {entry.input === 'ops' ? 'in the playground' : entry.traced ? 'step by step' : ''}
          </Link>{' '}
          <Link
            href={`/quiz/?entry=${entry.id}`}
            className={buttonVariants({ variant: 'secondary', className: 'ml-2' })}
          >
            Quiz me
          </Link>
        </p>
      </header>

      <section aria-labelledby="facts">
        <Heading
          level={2}
          id="facts"
          className="sr-only"
        >
          At a glance
        </Heading>
        <dl className="grid gap-x-6 gap-y-2 text-sm sm:grid-cols-[max-content_1fr]">
          <dt className="text-muted">Time</dt>
          <dd className="font-mono">{entry.time}</dd>
          {entry.why ? (
            <>
              <dt className="text-muted">Why</dt>
              <dd>
                {entry.why}{' '}
                <Link
                  className="underline"
                  href={`/big-o/?entry=${entry.id}`}
                >
                  Measure it
                </Link>
              </dd>
            </>
          ) : null}
          <dt className="text-muted">Space</dt>
          <dd className="font-mono">{entry.space}</dd>
          <dt className="text-muted">Takes</dt>
          <dd>{inputWords[entry.input]}</dd>
          {entry.params.length ? (
            <>
              <dt className="text-muted">Options</dt>
              <dd>{entry.params.join(', ')}</dd>
            </>
          ) : null}
          {entry.ops.length ? (
            <>
              <dt className="text-muted">Operations</dt>
              <dd className="font-mono">{entry.ops.join(', ')}</dd>
            </>
          ) : null}
          {entry.limits ? (
            <>
              <dt className="text-muted">Limits</dt>
              <dd>{entry.limits}</dd>
            </>
          ) : null}
          {entry.options.length ? (
            <>
              <dt className="text-muted">Settings</dt>
              <dd>
                {entry.options
                  .map((o) => {
                    const range = o.min !== undefined ? `, ${o.min} to ${o.max ?? ''}` : '';
                    return `${o.label || o.name} (default ${o.default}${range})`;
                  })
                  .join('; ')}
              </dd>
            </>
          ) : null}
          <dt className="text-muted">Animated</dt>
          <dd>
            {entry.traced
              ? 'yes, every step'
              : entry.input === 'ops'
                ? 'by operation'
                : 'no: the result only'}
          </dd>
          {entry.import ? (
            <>
              <dt className="text-muted">Go package</dt>
              <dd className="font-mono">
                {entry.import}
                {entry.func ? ` (${entry.package}.${entry.func})` : ''}
              </dd>
            </>
          ) : null}
        </dl>
      </section>

      {source.ok && (source.value.doc || source.value.packageDoc) ? (
        <section
          aria-labelledby="how"
          className="flex flex-col gap-2"
        >
          <Heading
            level={2}
            id="how"
            className="text-lg font-semibold"
          >
            How it works
          </Heading>
          {source.value.doc ? (
            <p className="max-w-3xl whitespace-pre-wrap">{source.value.doc}</p>
          ) : null}
          {source.value.packageDoc ? (
            <details>
              <summary className="cursor-pointer text-sm font-medium">
                About the {entry.package || 'package'} package
              </summary>
              <p className="mt-2 max-w-3xl text-sm whitespace-pre-wrap">
                {source.value.packageDoc}
              </p>
            </details>
          ) : null}
        </section>
      ) : null}

      <section
        aria-labelledby="use"
        className="grid gap-4 lg:grid-cols-2"
      >
        <Heading
          level={2}
          id="use"
          className="text-lg font-semibold lg:col-span-2"
        >
          Use it
        </Heading>
        <Snippet
          title="From the command line"
          code={snippet.cli}
        />
        <Snippet
          title="From Go"
          code={snippet.go}
        />
      </section>

      <RealWorldCode
        api={api}
        entry={entry}
      />

      <section
        aria-labelledby="code"
        className="flex flex-col gap-2"
      >
        <Heading
          level={2}
          id="code"
          className="text-lg font-semibold"
        >
          The code
        </Heading>
        {source.ok ? (
          <div className="av-root">
            <CodeView
              code={source.value.code}
              file={source.value.file}
            />
          </div>
        ) : (
          <p className="text-sm text-muted">The source is not in this build of the algorithms.</p>
        )}
        {file ? (
          <p className="text-sm">
            <a
              className="underline"
              href={`${repoUrl}/blob/main/pkg/${file}`}
            >
              {file} on GitHub
            </a>
          </p>
        ) : null}
      </section>
    </article>
  );
}

function DocsEntry({
  api,
  catalog,
  fixedId
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  fixedId?: string;
}) {
  const [queryId] = useQueryState('id', parseAsString);
  const id = fixedId ?? queryId;
  const entry = catalog.entries.find((e) => e.id === id);
  if (!entry) {
    return (
      <p role="alert">
        There is no entry called &quot;{id ?? ''}&quot;.{' '}
        <Link
          className="underline"
          href="/docs/"
        >
          See them all
        </Link>
        .
      </p>
    );
  }
  return (
    <EntryDoc
      api={api}
      catalog={catalog}
      entry={entry}
    />
  );
}

/**
 * One entry's documentation: the entry named by id, or by the ?id= query when
 * there is none. loading is shown until the wasm build is ready.
 */
export function DocsEntryApp({ id, loading }: { id?: string; loading?: ReactNode } = {}) {
  return (
    <WasmGate loading={loading}>
      {(api, catalog) => (
        <DocsEntry
          api={api}
          catalog={catalog}
          fixedId={id}
        />
      )}
    </WasmGate>
  );
}
