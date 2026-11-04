'use client';

import { useId, useMemo, useState, type ReactNode } from 'react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import type { Entry, InputKind, Input as RunInput } from '@braswelljr/algorithm-visualizer/core';
import { Field, Input, Textarea } from '@/components/ui/field';
import { Select } from '@/components/ui/select';
import { parseList } from '@/hooks/use-dataset';
import { useEntryInput, useEntryOptions } from '@/hooks/use-entry-input';
import { OptionFields } from './option-fields';
import { PointsPicker } from './points-picker';

/** Which fields a kind shows, and what they are called. */
interface Layout {
  text?: { label: string; multiline?: boolean; hint?: string };
  other?: { label: string };
  values?: { label: string; hint?: string; count?: number };
  target?: { label: string };
}

const layouts: Partial<Record<InputKind, Layout>> = {
  text: { text: { label: 'Text' } },
  'two-texts': { text: { label: 'First word' }, other: { label: 'Second word' } },
  pattern: { text: { label: 'Text' }, other: { label: 'Pattern' } },
  number: { values: { label: 'Number', count: 1 } },
  'two-numbers': {
    values: { label: 'Numbers', hint: 'Two whole numbers, separated by a comma.', count: 2 }
  },
  amount: {
    values: { label: 'Coins', hint: 'Whole numbers separated by commas.' },
    target: { label: 'Amount' }
  },
  processes: {
    text: {
      label: 'Processes',
      multiline: true,
      hint: 'One per line, as in the example: a name, then its numbers (arrival, burst and priority for CPU processes; period and burst for real-time tasks).'
    }
  },
  pages: {
    values: {
      label: 'Reference string',
      hint: 'The pages asked for, in order, separated by commas.'
    }
  },
  tracks: {
    values: { label: 'Requests', hint: 'The tracks asked for, in order, separated by commas.' }
  },
  partitions: {
    values: { label: 'Partition sizes', hint: 'The free blocks of memory, separated by commas.' },
    text: { label: 'Process sizes', hint: 'The memory each process asks for, separated by commas.' }
  },
  banker: { text: { label: 'Matrices', multiline: true, hint: 'In the format of the example.' } },
  graph: {
    text: {
      label: 'Edges',
      multiline: true,
      hint: 'One edge a line: two node names and an optional weight. A first line "directed" makes the edges one-way.'
    }
  },
  points: {
    text: { label: 'Points', multiline: true, hint: 'One point a line: x and y, whole numbers.' }
  },
  simulation: {}
};

/** A text field that keeps what is being typed and reports only values that parse. */
function ListField({
  id,
  label,
  hint,
  values,
  count,
  onValues
}: {
  id: string;
  label: string;
  hint?: string;
  values: number[];
  count?: number;
  onValues: (v: number[]) => void;
}) {
  const [draft, setDraft] = useState(values.join(', '));
  const parsed = parseList(draft.replace(/\s+/g, ''));
  const ok = parsed !== null && (count === undefined || parsed.length === count);
  return (
    <Field
      label={label}
      htmlFor={id}
      hint={
        ok
          ? hint
          : count
            ? `Give ${count === 1 ? 'one whole number' : `${count} whole numbers`}.`
            : 'Whole numbers separated by commas.'
      }
    >
      <Input
        id={id}
        value={draft}
        aria-invalid={!ok}
        onChange={(e) => {
          setDraft(e.target.value);
          const p = parseList(e.target.value.replace(/\s+/g, ''));
          if (p && (count === undefined || p.length === count)) onValues(p);
        }}
      />
    </Field>
  );
}

/**
 * The input for every entry that is not an array, a grid or a structure,
 * built from its kind, its params and its options, and prefilled from its
 * example. The whole input lives in the link.
 */
export function InputForm({
  entry,
  children
}: {
  entry: Entry;
  children: (input: RunInput) => ReactNode;
}) {
  const id = useId();
  const [input, setInput] = useEntryInput(entry);
  const [options, setOption] = useEntryOptions(entry);
  const layout = layouts[entry.input] ?? {};
  const set = (patch: Partial<RunInput>) => setInput({ ...input, ...patch });
  // The same object until the input or an option changes: RunView redraws on it.
  const full: RunInput = useMemo(
    () => (entry.options.length ? { ...input, options } : input),
    [entry, input, options]
  );
  const wants = (p: string) => entry.params.includes(p as never);
  return (
    <div className="flex flex-col gap-6">
      <Card
        size="md"
        render={<section />}
        aria-label="Input"
        className="grid gap-3 px-(--card-spacing) sm:grid-cols-2"
      >
        {layout.text ? (
          <Field
            label={layout.text.label}
            htmlFor={`${id}-text`}
            hint={layout.text.hint}
            className={layout.other || layout.values ? '' : 'sm:col-span-2'}
          >
            {layout.text.multiline ? (
              <Textarea
                id={`${id}-text`}
                value={input.text ?? ''}
                rows={8}
                spellCheck={false}
                onChange={(e) => set({ text: e.target.value })}
              />
            ) : (
              <Input
                id={`${id}-text`}
                value={input.text ?? ''}
                maxLength={10_000}
                onChange={(e) => set({ text: e.target.value })}
              />
            )}
          </Field>
        ) : null}
        {layout.other ? (
          <Field
            label={layout.other.label}
            htmlFor={`${id}-other`}
          >
            <Input
              id={`${id}-other`}
              value={input.other ?? ''}
              maxLength={10_000}
              onChange={(e) => set({ other: e.target.value })}
            />
          </Field>
        ) : null}
        {layout.values ? (
          <ListField
            id={`${id}-values`}
            label={layout.values.label}
            hint={layout.values.hint}
            count={layout.values.count}
            values={input.values ?? []}
            onValues={(values) => set({ values })}
          />
        ) : null}
        {entry.input === 'points' ? (
          <PointsPicker
            text={input.text ?? ''}
            onText={(text) => set({ text })}
          />
        ) : null}
        {wants('start') ? (
          <Field
            label="Start node"
            htmlFor={id + '-start'}
          >
            <Input
              id={id + '-start'}
              value={input.other ?? ''}
              onChange={(e) => set({ other: e.target.value })}
            />
          </Field>
        ) : null}
        {wants('end') ? (
          <Field
            label="End node"
            htmlFor={id + '-target'}
          >
            <Input
              id={id + '-target'}
              value={input.key ?? ''}
              onChange={(e) => set({ key: e.target.value })}
            />
          </Field>
        ) : null}
        {layout.target || wants('target') || wants('amount') ? (
          <Field
            label={layout.target?.label ?? (wants('amount') ? 'Amount' : 'Target')}
            htmlFor={`${id}-target`}
          >
            <Input
              id={`${id}-target`}
              type="number"
              value={input.target ?? 0}
              onChange={(e) => {
                const n = Number(e.target.value);
                if (Number.isSafeInteger(n)) set({ target: n });
              }}
            />
          </Field>
        ) : null}
        {wants('key') ? (
          <Field
            label={entry.choices.length ? 'Callback' : 'Keyword'}
            htmlFor={`${id}-key`}
            hint={
              entry.choices.length
                ? 'The function applied to each element.'
                : 'Only its letters count.'
            }
          >
            {entry.choices.length ? (
              <Select
                id={`${id}-key`}
                value={input.key ?? entry.choices[0]}
                onChange={(e) => set({ key: e.target.value })}
              >
                {entry.choices.map((c) => (
                  <option
                    key={c}
                    value={c}
                  >
                    {c}
                  </option>
                ))}
              </Select>
            ) : (
              <Input
                id={`${id}-key`}
                value={input.key ?? ''}
                onChange={(e) => set({ key: e.target.value })}
              />
            )}
          </Field>
        ) : null}
        {wants('shift') ? (
          <Field
            label="Shift"
            htmlFor={`${id}-shift`}
          >
            <Input
              id={`${id}-shift`}
              type="number"
              value={input.shift ?? 3}
              onChange={(e) => {
                const n = Number(e.target.value);
                if (Number.isSafeInteger(n)) set({ shift: n });
              }}
            />
          </Field>
        ) : null}
        <OptionFields
          entry={entry}
          values={options}
          onChange={setOption}
        />
        {entry.limits ? (
          <p className="text-xs text-muted sm:col-span-2">Limits: {entry.limits}</p>
        ) : null}
      </Card>
      {children(full)}
    </div>
  );
}
