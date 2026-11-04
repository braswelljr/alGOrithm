'use client';

import { useId, useState } from 'react';
import Link from 'next/link';
import { parseAsString, useQueryStates } from 'nuqs';
import { LuListPlus, LuX } from 'react-icons/lu';
import { EntryBadges } from '@/components/catalog/entry-badges';
import { RunPanel } from '@/components/run/run-panel';
import { Button } from '@/components/shared/button';
import { Field, Input } from '@/components/shared/field';
import { NumberInput } from '@/components/shared/number-input';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Badge } from '@/components/ui/badge';
import { Combobox } from '@/components/ui/combobox';
import { Heading } from '@/components/ui/heading';
import { InputDialog } from '@/components/visualize/input-dialog';
import { OptionFields } from '@/components/visualize/option-fields';
import { parseList } from '@/hooks/use-dataset';
import { useEntryOptions } from '@/hooks/use-entry-input';
import { docsHref, isStructure } from '@/lib/entries';
import type { AlgorithmApi, Catalog, Entry, Op } from '@/visualizer/core';

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

/** Arguments that are words or lists, not numbers: the rest take a number field. */
const TEXT_ARGS = new Set(['WORD', 'TEXT', 'VALUES']);

/**
 * The operation buttons and the arguments they take: number fields for
 * numbers (a second one when an operation takes two, A B or KEY VALUE), and a
 * text field only for an operation that takes a word or a list.
 */
function OpButtons({ entry, onOp }: { entry: Entry; onOp: (op: Op) => void }) {
  const id = useId();
  const [first, setFirst] = useState<number | null>(7);
  const [second, setSecond] = useState<number | null>(70);
  const [word, setWord] = useState('cat');
  const templates = entry.ops.map(template);
  const numbersAt = (t: ReturnType<typeof template>) =>
    t.takes.filter((a) => !TEXT_ARGS.has(a)).length;
  const twoNumbers = templates.some((t) => numbersAt(t) > 1);
  const anyNumber = templates.some((t) => numbersAt(t) > 0);
  const textArg = templates.flatMap((t) => t.takes).find((a) => TEXT_ARGS.has(a));

  // The value an operation is sent with, its arguments in order.
  const valueFor = (t: ReturnType<typeof template>): string | null => {
    const numbers = [first, second];
    let n = 0;
    const parts: string[] = [];
    for (const a of t.takes) {
      if (TEXT_ARGS.has(a)) {
        if (!word.trim()) return null;
        parts.push(word.trim());
      } else {
        const v = numbers[n++];
        if (v === null || v === undefined) return null;
        parts.push(String(v));
      }
    }
    return parts.join(' ');
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="grid gap-3 sm:grid-cols-3">
        {anyNumber ? (
          <Field
            label={twoNumbers ? 'First number' : 'Value'}
            htmlFor={`${id}-a`}
            hint={twoNumbers ? 'A, KEY, INDEX or LO.' : undefined}
          >
            <NumberInput
              id={`${id}-a`}
              value={first}
              onValueChange={setFirst}
            />
          </Field>
        ) : null}
        {twoNumbers ? (
          <Field
            label="Second number"
            htmlFor={`${id}-b`}
            hint="B, VALUE or HI, for the operations that take two."
          >
            <NumberInput
              id={`${id}-b`}
              value={second}
              onValueChange={setSecond}
            />
          </Field>
        ) : null}
        {textArg ? (
          <Field
            label={textArg === 'VALUES' ? 'Values' : textArg === 'TEXT' ? 'Text' : 'Word'}
            htmlFor={`${id}-w`}
            hint={textArg === 'VALUES' ? 'Whole numbers, separated by commas.' : undefined}
          >
            <Input
              id={`${id}-w`}
              value={word}
              onChange={(e) => setWord(e.target.value)}
            />
          </Field>
        ) : null}
      </div>
      <div
        className="flex flex-wrap gap-2"
        role="group"
        aria-label="Operations"
      >
        {templates.map((t, i) => {
          const value = t.takes.length ? valueFor(t) : '';
          return (
            <Button
              key={entry.ops[i]}
              variant="secondary"
              size="sm"
              disabled={value === null}
              onClick={() =>
                onOp(t.takes.length && value ? { name: t.name, value } : { name: t.name })
              }
            >
              {t.takes.length ? `${t.name} ${value ?? t.takes.join(' ')}` : t.name}
            </Button>
          );
        })}
      </div>
    </div>
  );
}

/** The dialog's body: the queue so far, each step removable, and the buttons that add to it. */
function OpsFields({
  entry,
  ops,
  setOps
}: {
  entry: Entry;
  ops: Op[];
  setOps: (next: Op[]) => void;
}) {
  return (
    <div className="flex flex-col gap-5">
      <OpButtons
        entry={entry}
        onOp={(op) => setOps([...ops, op])}
      />
      <section
        aria-label="Queued operations"
        className="flex flex-col gap-2"
      >
        <div className="flex flex-wrap items-center justify-between gap-2">
          <Heading
            level={3}
            className="text-sm font-semibold"
          >
            {ops.length ? `${ops.length} in order` : 'Nothing queued'}
          </Heading>
          {ops.length ? (
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setOps(ops.slice(0, -1))}
              >
                Undo the last
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setOps([])}
              >
                Clear them all
              </Button>
            </div>
          ) : null}
        </div>
        {ops.length ? (
          <ol className="flex flex-wrap gap-1.5">
            {ops.map((o, i) => (
              <li key={i}>
                <Badge
                  variant="secondary"
                  case="normal"
                  className="gap-1 pe-1 font-mono"
                >
                  <span className="text-muted">{i + 1}.</span>
                  {o.value ? `${o.name} ${o.value}` : o.name}
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => setOps(ops.filter((_, k) => k !== i))}
                    aria-label={`Take back step ${i + 1}, ${o.name}${o.value ? ` ${o.value}` : ''}`}
                    className="size-5 bg-transparent"
                  >
                    <LuX aria-hidden />
                  </Button>
                </Badge>
              </li>
            ))}
          </ol>
        ) : (
          <p className="text-muted text-sm">
            Press an operation above to add it; the run starts from the values you set.
          </p>
        )}
      </section>
    </div>
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
  const [options, , setOptions] = useEntryOptions(entry);
  const startText = q.start ?? (entry.example.values ?? [4, 2, 9]).join(', ');
  const start = parseList(startText.replace(/\s+/g, '')) ?? [];
  // No ops in the link: the example's. "~" in the link means none at all.
  const ops = q.ops === null ? (entry.example.ops ?? []) : decodeOps(q.ops);
  const input = { values: start, ops, ...(entry.options.length ? { options } : {}) };
  return (
    <div className="flex flex-col gap-6">
      {/* The structure, named boldly above everything else, as the visualizer
          names its entry; still text under reduced motion. */}
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
      <div data-tour="entry-picker">
        <Field
          label="Structure"
          htmlFor={`${id}-entry`}
          className="sm:max-w-md"
        >
          <Combobox
            id={`${id}-entry`}
            value={entry.id}
            onChange={(value) => void setQ({ id: value, ops: null })}
            options={structures.map((st) => ({ value: st.id, label: st.name }))}
          />
        </Field>
      </div>
      {/* What it starts with and its options, set in the configure dialog. */}
      <InputDialog<{ start: string; options: Record<string, number> }>
        entry={entry}
        values={{ start: startText, options }}
        summary={[
          start.length ? `${start.length} values to start` : 'an empty structure',
          ...Object.entries(entry.options.length ? options : {}).map(([k, v]) => `${k} ${v}`)
        ].join(' · ')}
        onApply={(v) => {
          void setQ({ start: v.start });
          if (entry.options.length) setOptions(v.options);
        }}
      >
        {(form) => (
          <div className="grid gap-3 sm:grid-cols-2">
            <Field
              label="Start with"
              htmlFor={`${id}-start`}
              hint="Whole numbers, separated by commas."
            >
              <Input
                id={`${id}-start`}
                value={form.watch('start')}
                aria-invalid={
                  form.watch('start').trim() !== '' &&
                  parseList(form.watch('start').replace(/\s+/g, '')) === null
                }
                onChange={(e) => form.setValue('start', e.target.value, { shouldDirty: true })}
              />
            </Field>
            <OptionFields
              entry={entry}
              values={form.watch('options')}
              onChange={(name, val) =>
                form.setValue(
                  'options',
                  { ...form.getValues('options'), [name]: val },
                  { shouldDirty: true }
                )
              }
            />
          </div>
        )}
      </InputDialog>
      <p className="text-muted text-sm">
        {entry.summary} <span className="font-mono text-xs">{entry.time}</span>{' '}
        <Link
          className="underline"
          href={docsHref(entry.id)}
        >
          Docs
        </Link>
      </p>
      {/* The operations, queued in their own dialog: the page keeps the view. */}
      <InputDialog<{ ops: Op[] }>
        entry={entry}
        values={{ ops }}
        tour="operations"
        icon={<LuListPlus aria-hidden />}
        lead="Operations:"
        action={`Apply operations to ${entry.name}`}
        title={`Operations on ${entry.name}`}
        description="Queue operations, take any back, and apply them; the run plays them in order. Nothing changes until you apply."
        summary={
          ops.length
            ? `${ops.length}, the last ${ops.length > 3 ? '... ' : ''}${ops
                .slice(-3)
                .map((o) => (o.value ? `${o.name} ${o.value}` : o.name))
                .join(', ')}`
            : 'none yet'
        }
        onApply={(v) => void setQ({ id: entry.id, ops: encodeOps(v.ops) || '~' })}
      >
        {(form) => (
          <OpsFields
            entry={entry}
            ops={form.watch('ops')}
            setOps={(next) => form.setValue('ops', next.slice(-MAX_OPS), { shouldDirty: true })}
          />
        )}
      </InputDialog>
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
