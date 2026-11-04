'use client';

import type { ReactNode } from 'react';
import Link from 'next/link';
import type { AlgorithmApi, Catalog, Entry, Input } from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryState } from 'nuqs';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { WasmGate } from '@/components/shared/wasm-gate';
import { RunPanel } from '@/components/run/run-panel';
import { docsHref, isStructure } from '@/lib/entries';
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
      <div className="flex flex-wrap items-end justify-between gap-4">
        <EntryPicker
          catalog={catalog}
          value={entry?.id ?? ''}
          onChange={(v) => void setId(v)}
        />
        {entry ? (
          <p className="max-w-xl text-sm text-muted">
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
        {entry ? (
          <EntryBadges
            entry={entry}
            links
          />
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
