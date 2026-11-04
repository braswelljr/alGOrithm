'use client';

import { useId, useMemo, type ReactNode } from 'react';
import { parseAsInteger, useQueryState } from 'nuqs';
import type { UseFormReturn } from 'react-hook-form';
import { DataPanel } from '@/components/data/data-panel';
import { Field } from '@/components/shared/field';
import { NumberInput } from '@/components/shared/number-input';
import { useDataset, type Dataset } from '@/hooks/use-dataset';
import { useEntryOptions } from '@/hooks/use-entry-input';
import type { AlgorithmApi, Catalog, Entry, Input as RunInput } from '@/visualizer/core';
import { InputDialog } from './input-dialog';
import { OptionFields } from './option-fields';

/** Above this many values the bars are too thin to follow. */
export const ANIMATION_LIMIT = 300;

/**
 * Where a run's values come from, as the form holds it: a generator's shape,
 * size and seed, or a list of values (pasted, from a file, an address, the
 * example).
 */
type Source = { kind: string; size: number; seed: number } | { values: number[]; label: string };

type ArrayValues = { source: Source; target: number; options: Record<string, number> };

/**
 * A Dataset over the form's draft, so the data panel's tabs (generate, paste,
 * file, address) fill the draft rather than the link.
 */
function useDraftDataset(api: AlgorithmApi, form: UseFormReturn<ArrayValues>): Dataset {
  const source = form.watch('source');
  const generated = useMemo(
    () => ('kind' in source ? api.generate(source.kind, source.size, source.seed) : null),
    [api, source]
  );
  const set = (next: Source) => form.setValue('source', next, { shouldDirty: true });
  if ('kind' in source) {
    return {
      values: generated?.ok ? generated.value : null,
      label: `${source.size} ${source.kind} values, seed ${source.seed}`,
      error: generated && !generated.ok ? generated.failure.error : null,
      kind: source.kind,
      size: source.size,
      seed: source.seed,
      generate: (kind, size, seed) => set({ kind, size, seed }),
      use: (values, label) => set({ values, label })
    };
  }
  return {
    values: source.values,
    label: source.label,
    error: null,
    kind: 'list',
    size: source.values.length,
    seed: 1,
    generate: (kind, size, seed) => set({ kind, size, seed }),
    use: (values, label) => set({ values, label })
  };
}

/** The dialog's fields: the data panel over the draft, the target and the options. */
function ArrayFields({
  api,
  catalog,
  entry,
  search,
  form
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  entry: Entry;
  search: boolean;
  form: UseFormReturn<ArrayValues>;
}) {
  const id = useId();
  const draft = useDraftDataset(api, form);
  const target = form.watch('target');
  const options = form.watch('options');
  return (
    <div className="flex flex-col gap-6">
      <DataPanel
        bare
        api={api}
        kinds={catalog.datasets}
        maxValues={catalog.maxValues}
        dataset={draft}
      />
      {search || entry.options.length ? (
        <div className="grid gap-3 sm:grid-cols-3">
          {search ? (
            <Field
              label="Search for"
              htmlFor={id}
              hint={
                entry.input === 'sorted-ints'
                  ? 'The values are sorted first: this search needs sorted input.'
                  : undefined
              }
            >
              <NumberInput
                id={id}
                value={target}
                onValueChange={(n) => {
                  if (n !== null && Number.isSafeInteger(n))
                    form.setValue('target', n, { shouldDirty: true });
                }}
              />
            </Field>
          ) : null}
          <OptionFields
            entry={entry}
            values={options}
            onChange={(name, v) =>
              form.setValue(
                'options',
                { ...form.getValues('options'), [name]: v },
                { shouldDirty: true }
              )
            }
          />
        </div>
      ) : null}
    </div>
  );
}

/**
 * A list of values from the data panel (starting with the entry's example),
 * a target for a search, and the entry's options, set in the configure
 * dialog; the link keeps what the run uses.
 */
export function ArrayInput({
  api,
  catalog,
  entry,
  children
}: {
  api: AlgorithmApi;
  catalog: Catalog;
  entry: Entry;
  children: (input: RunInput | null) => ReactNode;
}) {
  const dataset = useDataset(api, 24, entry.example.values);
  const [options, , setOptions] = useEntryOptions(entry);
  const [t, setT] = useQueryState('t', parseAsInteger.withOptions({ history: 'replace' }));
  const search =
    entry.input === 'sorted-ints' || entry.input === 'any-ints' || entry.params.includes('target');
  const values = dataset.values;
  const target = t ?? entry.example.target ?? values?.[Math.floor((values?.length ?? 0) / 2)] ?? 0;
  const input: RunInput | null = values
    ? {
        values,
        ...(search ? { target } : {}),
        ...(entry.options.length ? { options } : {})
      }
    : null;
  const generator = catalog.datasets.includes(dataset.kind);
  const source: Source = generator
    ? { kind: dataset.kind, size: dataset.size, seed: dataset.seed }
    : { values: values ?? [], label: dataset.label };

  return (
    <div className="flex flex-col gap-6">
      <InputDialog<ArrayValues>
        entry={entry}
        values={{ source, target, options }}
        summary={[
          dataset.label || 'no data yet',
          ...(search ? [`target ${target}`] : []),
          ...Object.entries(entry.options.length ? options : {}).map(([k, v]) => `${k} ${v}`)
        ].join(' · ')}
        onApply={(v) => {
          // The data is written only when it changed, so applying a new target
          // leaves the example (or a long stored list) where it is.
          if (JSON.stringify(v.source) !== JSON.stringify(source)) {
            if ('kind' in v.source) dataset.generate(v.source.kind, v.source.size, v.source.seed);
            else dataset.use(v.source.values, v.source.label);
          }
          if (search && v.target !== target) void setT(v.target);
          if (entry.options.length) setOptions(v.options);
        }}
      >
        {(form) => (
          <ArrayFields
            api={api}
            catalog={catalog}
            entry={entry}
            search={search}
            form={form}
          />
        )}
      </InputDialog>
      {values && values.length > ANIMATION_LIMIT && entry.traced ? (
        <p className="bg-subtle rounded-md p-3 text-sm">
          {values.length} values are more than the animation can show clearly ({ANIMATION_LIMIT}).
          It still runs; for big data, the counts on the Compare page say more.
        </p>
      ) : null}
      {children(input)}
    </div>
  );
}
