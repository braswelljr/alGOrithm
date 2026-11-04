'use client';

import { useMemo, type ReactNode } from 'react';
import Link from 'next/link';
import { ChevronDown } from 'lucide-react';
import { parseAsString, useQueryState } from 'nuqs';
import { TLDRCard } from '@/components/algorithm/tldr-card';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { GoDoc } from '@/components/docs/go-doc';
import { RealWorldCode } from '@/components/docs/real-world-code';
import { PageHeading } from '@/components/shared/page-header';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Button } from '@/components/ui/button';
import { CodeSample } from '@/components/ui/code-block';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Heading } from '@/components/ui/heading';
import { repoUrl } from '@/config/site';
import { entryHref } from '@/lib/entries';
import { fallbackSnippet, type AlgorithmApi, type Catalog, type Entry } from '@/visualizer/core';

/**
 * One way to run the entry, as a full-width collapsible: the header row names
 * it, and opening it shows the code, highlighted, with its copy button.
 */
function Snippet({
  title,
  code,
  language,
  defaultOpen = false
}: {
  title: string;
  code: string;
  language: string;
  defaultOpen?: boolean;
}) {
  return (
    <Collapsible
      defaultOpen={defaultOpen}
      className="w-full"
    >
      <CollapsibleTrigger
        render={
          <Button
            variant="ghost"
            className="group w-full justify-between"
          />
        }
      >
        <span className="font-semibold">{title}</span>
        <ChevronDown
          aria-hidden
          className="size-4 transition-transform group-data-panel-open:rotate-180 motion-reduce:transition-none"
        />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <CodeSample
          title={language === 'go' ? 'main.go' : 'Shell'}
          language={language}
          code={code}
          className="mt-2"
        />
      </CollapsibleContent>
    </Collapsible>
  );
}

function EntryDoc({ api, catalog, entry }: { api: AlgorithmApi; catalog: Catalog; entry: Entry }) {
  const source = useMemo(() => api.source(entry.id), [api, entry.id]);
  const snippet = useMemo(() => {
    const input = entry.example;
    const s = api.snippet(entry.id, input);
    return s.ok ? s.value : fallbackSnippet(entry, input);
  }, [api, entry]);
  const file = source.ok ? source.value.file : entry.file;
  return (
    <article className="flex w-full flex-col gap-6">
      <header className="flex w-full flex-col gap-3">
        <PageHeading title={entry.name}>{entry.summary}</PageHeading>
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
            className="w-full"
          />
        ) : null}
        <div
          data-tour="entry-actions"
          className="flex flex-wrap items-center gap-2"
        >
          <Button
            variant="primary"
            render={<Link href={entryHref(entry)} />}
          >
            Try it{' '}
            {entry.input === 'ops' ? 'in the playground' : entry.traced ? 'step by step' : ''}
          </Button>
          <Button
            variant="secondary"
            render={<Link href={`/quiz/?entry=${entry.id}`} />}
          >
            Quiz me
          </Button>
          {entry.traced ? (
            <Button
              variant="ghost"
              render={<Link href={`/big-o/?entry=${entry.id}`} />}
            >
              Measure its cost
            </Button>
          ) : null}
        </div>
      </header>

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
          {source.value.doc ? <GoDoc text={source.value.doc} /> : null}
          {source.value.packageDoc ? (
            <Collapsible className="w-full">
              <CollapsibleTrigger
                render={
                  <Button
                    variant="ghost"
                    className="group w-full justify-between"
                  />
                }
              >
                <span className="font-medium">About the {entry.package || 'package'} package</span>
                <ChevronDown
                  aria-hidden
                  className="size-4 transition-transform group-data-panel-open:rotate-180 motion-reduce:transition-none"
                />
              </CollapsibleTrigger>
              <CollapsibleContent>
                <GoDoc
                  text={source.value.packageDoc}
                  className="mt-2 text-sm"
                />
              </CollapsibleContent>
            </Collapsible>
          ) : null}
        </section>
      ) : null}

      <section
        aria-labelledby="use"
        className="flex w-full flex-col gap-2"
      >
        <Heading
          level={2}
          id="use"
          className="text-lg font-semibold"
        >
          Use it
        </Heading>
        <Snippet
          title="From the command line"
          code={snippet.cli}
          language="bash"
          defaultOpen
        />
        <Snippet
          title="From Go"
          code={snippet.go}
          language="go"
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
          <CodeSample
            title={source.value.file}
            language="go"
            code={source.value.code}
            lineNumbers
          />
        ) : (
          <p className="text-muted text-sm">The source is not in this build of the algorithms.</p>
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
