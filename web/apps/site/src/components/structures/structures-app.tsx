'use client';

import { useId, useState } from 'react';
import Link from 'next/link';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import type { AlgorithmApi, Catalog, Entry, Op } from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryStates } from 'nuqs';
import { WasmGate } from '@/components/shared/wasm-gate';
import { RunPanel } from '@/components/run/run-panel';
import { Button } from '@/components/ui/button';
import { Field, Input } from '@/components/ui/field';
import { Select } from '@/components/ui/select';
import { OptionFields } from '@/components/visualize/option-fields';
import { parseList } from '@/hooks/use-dataset';
import { useEntryOptions } from '@/hooks/use-entry-input';
import { docsHref, isStructure } from '@/lib/entries';

/** The most operations kept in a link. */
const MAX_OPS = 200;

/** "push VALUE" -> { name: "push", takes: ["VALUE"] }. */
function template(op: string) {
  const [name, ...takes] = op.split(/\s+/);
  return { name, takes };
}

/** Operations written as "push 4,pop,put 5 50". */
export function encodeOps(ops: Op[]): string {
  return ops.map((o) => (o.value ? `${o.name} ${o.value}` : o.name)).join(',');
}

export function decodeOps(s: string): Op[] {
  return s
    .split(',')
    .map((part) => part.trim())
    .filter((part) => part !== '' && part !== '~')
    .slice(0, MAX_OPS)
    .map((part) => {
      const [name, ...rest] = part.split(/\s+/);
      return rest.length ? { name, value: rest.join(' ') } : { name };
    });
}

function OpButtons({ entry, onOp }: { entry: Entry; onOp: (op: Op) => void }) {
  const id = useId();
  const [value, setValue] = useState('7');
  return (
    <form
      className="flex flex-col gap-3"
      onSubmit={(e) => e.preventDefault()}
    >
      <Field
        label="Value"
        htmlFor={id}
        hint="Used by the operations that take one. Two numbers (A B, KEY VALUE) go with a space between."
        className="sm:w-80"
      >
        <Input
          id={id}
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
      </Field>
      <div
        className="flex flex-wrap gap-2"
        role="group"
        aria-label="Operations"
      >
        {entry.ops.map((op) => {
          const t = template(op);
          const needs = t.takes.length > 0;
          return (
            <Button
              key={op}
              variant="secondary"
              size="sm"
              disabled={needs && value.trim() === ''}
              onClick={() => onOp(needs ? { name: t.name, value: value.trim() } : { name: t.name })}
            >
              {needs ? `${t.name} ${value.trim() || t.takes.join(' ')}` : t.name}
            </Button>
          );
        })}
      </div>
    </form>
  );
}

const queryParams = { id: parseAsString, start: parseAsString, ops: parseAsString };

function Structures({ api, catalog }: { api: AlgorithmApi; catalog: Catalog }) {
  const structures = catalog.entries.filter(isStructure);
  const [q] = useQueryStates(queryParams);
  const entry = structures.find((e) => e.id === q.id) ?? structures[0];
  if (!entry) return <p>This build has no data structures.</p>;
  return (
    <StructurePlayground
      key={entry.id}
      api={api}
      entry={entry}
      structures={structures}
    />
  );
}

function StructurePlayground({
  api,
  entry,
  structures
}: {
  api: AlgorithmApi;
  entry: Entry;
  structures: Entry[];
}) {
  const id = useId();
  const [q, setQ] = useQueryStates(queryParams, { history: 'replace' });
  const [options, setOption] = useEntryOptions(entry);
  const startText = q.start ?? (entry.example.values ?? [4, 2, 9]).join(', ');
  const start = parseList(startText.replace(/\s+/g, '')) ?? [];
  // No ops in the link: the example's. "~" in the link means none at all.
  const ops = q.ops === null ? (entry.example.ops ?? []) : decodeOps(q.ops);
  const input = { values: start, ops, ...(entry.options.length ? { options } : {}) };
  return (
    <div className="flex flex-col gap-6">
      <Card
        size="md"
        render={<section />}
        aria-label="The structure"
        className="flex flex-col gap-4 px-(--card-spacing)"
      >
        <div className="grid gap-3 sm:grid-cols-[2fr_2fr_1fr] sm:items-end">
          <Field
            label="Structure"
            htmlFor={`${id}-entry`}
          >
            <Select
              id={`${id}-entry`}
              value={entry.id}
              onChange={(e) => void setQ({ id: e.target.value, ops: null })}
            >
              {structures.map((s) => (
                <option
                  key={s.id}
                  value={s.id}
                >
                  {s.name}
                </option>
              ))}
            </Select>
          </Field>
          <Field
            label="Start with"
            htmlFor={`${id}-start`}
            hint="Whole numbers, separated by commas."
          >
            <Input
              id={`${id}-start`}
              value={startText}
              onChange={(e) => void setQ({ start: e.target.value })}
            />
          </Field>
          <OptionFields
            entry={entry}
            values={options}
            onChange={setOption}
          />
        </div>
        <p className="text-sm text-muted">
          {entry.summary} <span className="font-mono text-xs">{entry.time}</span>{' '}
          <Link
            className="underline"
            href={docsHref(entry.id)}
          >
            Docs
          </Link>
        </p>
        <OpButtons
          entry={entry}
          onOp={(op) => void setQ({ id: entry.id, ops: encodeOps([...ops, op].slice(-MAX_OPS)) })}
        />
        {ops.length ? (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => void setQ({ ops: encodeOps(ops.slice(0, -1)) || '~' })}
            >
              Undo the last operation
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => void setQ({ ops: '~' })}
            >
              Clear the operations
            </Button>
          </div>
        ) : null}
      </Card>
      <RunPanel
        key={entry.id}
        api={api}
        entry={entry}
        input={input}
        followEnd
      />
    </div>
  );
}

export function StructuresApp() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Structures
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
