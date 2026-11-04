'use client';

import type { ReactNode } from 'react';
import Link from 'next/link';
import { parseAsString, useQueryState } from 'nuqs';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { RunPanel } from '@/components/run/run-panel';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Heading } from '@/components/ui/heading';
import { docsHref, isStructure } from '@/lib/entries';
import type { AlgorithmApi, Catalog, Entry, Input } from '@/visualizer/core';
import { ArrayInput } from './array-input';
import { EntryPicker } from './entry-picker';
import { GridInput } from './grid-input';
import { InputForm } from './input-form';

function InputFor({
  api,
  catalog,
  entry,
  children
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  entry: Entry;
  children: (input: Input | null) => ReactNode;
}) {
  switch (entry.input) {
    case 'ints':
    case 'sorted-ints':
    case 'any-ints':
      return (
        <ArrayInput
          api={api}
          catalog={catalog}
          entry={entry}
        >
          {children}
        </ArrayInput>
      );
    case 'grid':
      return (
        <GridInput
          api={api}
          catalog={catalog}
          entry={entry}
        >
          {children}
        </GridInput>
      );
    case 'ops':
      return (
        <p>
          {entry.name} is a data structure:{' '}
          <Link
            className="underline"
            href={`/structures/?id=${entry.id}`}
          >
            operate it in the playground
          </Link>
          .
        </p>
      );
    default:
      return <InputForm entry={entry}>{children}</InputForm>;
  }
}

function Visualize({ api, catalog }: { api: AlgorithmApi; catalog: Catalog }) {
  const runnable = catalog.entries.filter((e) => !isStructure(e));
  const [id, setId] = useQueryState('id', parseAsString.withDefault(runnable[0]?.id ?? ''));
  const entry = catalog.entries.find((e) => e.id === id);
  return (
    <div className="flex flex-col gap-6">
      {entry ? (
        // The entry, named boldly above everything else, a slow band of light
        // passing over it; still text under reduced motion.
        <header
          data-tour="entry-title"
          className="flex flex-col gap-2"
        >
          <Heading
            level={2}
            className="font-heading text-3xl font-bold tracking-tight text-balance sm:text-4xl"
          >
            <span className="shimmer text-fg motion-reduce:shimmer-none [animation-duration:3.5s]">
              {entry.name}
            </span>
          </Heading>
          <EntryBadges
            entry={entry}
            links
          />
        </header>
      ) : null}
      <div
        data-tour="entry-picker"
        className="flex flex-wrap items-end justify-between gap-4"
      >
        <EntryPicker
          catalog={catalog}
          value={entry?.id ?? ''}
          onChange={(v) => void setId(v)}
        />
        {entry ? (
          <p className="text-muted text-sm">
            {entry.summary}{' '}
            <span className="font-mono text-xs">
              time {entry.time} · space {entry.space}
            </span>{' '}
            {entry.why ? <span className="block pt-1 text-xs">{entry.why}</span> : null}
            <Link
              className="underline"
              href={docsHref(entry.id)}
            >
              Docs
            </Link>
          </p>
        ) : null}
      </div>
      {entry ? (
        <InputFor
          key={entry.id}
          api={api}
          catalog={catalog}
          entry={entry}
        >
          {(input) =>
            input ? (
              <RunPanel
                key={entry.id}
                api={api}
                entry={entry}
                input={input}
              />
            ) : null
          }
        </InputFor>
      ) : (
        <p role="alert">There is no algorithm called &quot;{id}&quot;. Pick one above.</p>
      )}
    </div>
  );
}

export function VisualizeApp() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Visualize
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
