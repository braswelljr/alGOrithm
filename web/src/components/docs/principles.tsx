'use client';

import { useMemo } from 'react';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Heading } from '@/components/ui/heading';
import { LinesView } from '@/visualizer';
import type { AlgorithmApi, Entry } from '@/visualizer/core';

/**
 * The principles a category's explanatory entries demonstrate (for data
 * structures: LIFO, also called FILO; FIFO, also called LILO; priority
 * order), run on their examples. An entry counts as explanatory when it is
 * neither animated nor operated by hand and answers in lines.
 */
export function Principles({ api, entries }: { api: AlgorithmApi; entries: Entry[] }) {
  const runs = useMemo(
    () =>
      entries
        .filter((e) => !e.traced && e.input !== 'ops')
        .map((e) => ({ entry: e, result: api.run(e.id, e.example, 0) }))
        .filter((r) => r.result.ok && r.result.value.output.lines.length > 0),
    [api, entries]
  );
  if (runs.length < 1) return null;
  return (
    <section
      aria-labelledby="principles"
      className="border-accent/40 bg-accent/5 mt-4 rounded-lg border p-4"
    >
      <Heading
        level={3}
        id="principles"
        className="font-semibold"
      >
        Principles
      </Heading>
      <p className="text-muted mt-1 text-sm">
        LIFO (last in, first out, also called FILO): a stack. FIFO (first in, first out, also called
        LILO): a queue. Priority order: a heap. The same values, put in and taken out by each rule:
      </p>
      {runs.map(({ entry, result }) =>
        result.ok ? (
          <div
            key={entry.id}
            className="av-root mt-3"
          >
            <p className="text-sm font-medium">{entry.name}</p>
            <LinesView lines={result.value.output.lines} />
          </div>
        ) : null
      )}
    </section>
  );
}

/**
 * Principles with the wasm loaded first. The summary catalog a page gets at
 * build time has no examples, so the full entries come from the wasm catalog.
 */
export function LivePrinciples({ entries }: { entries: Entry[] }) {
  return (
    <WasmGate>
      {(api, catalog) => {
        const ids = new Set(entries.map((e) => e.id));
        return (
          <Principles
            api={api}
            entries={catalog.entries.filter((e) => ids.has(e.id))}
          />
        );
      }}
    </WasmGate>
  );
}
