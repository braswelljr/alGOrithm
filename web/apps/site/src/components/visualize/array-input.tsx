'use client';

import { useId, type ReactNode } from 'react';
import type {
  AlgorithmApi,
  Catalog,
  Entry,
  Input as RunInput
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsInteger, useQueryState } from 'nuqs';
import { DataPanel } from '@/components/data/data-panel';
import { Field, Input } from '@/components/ui/field';
import { useDataset } from '@/hooks/use-dataset';
import { useEntryOptions } from '@/hooks/use-entry-input';
import { OptionFields } from './option-fields';

/** Above this many values the bars are too thin to follow. */
export const ANIMATION_LIMIT = 300;

/**
 * A list of values from the data panel (starting with the entry's example),
 * a target for a search, and the entry's options.
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
  const id = useId();
  const dataset = useDataset(api, 24, entry.example.values);
  const [options, setOption] = useEntryOptions(entry);
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
  return (
    <div className="flex flex-col gap-6">
      <DataPanel
        api={api}
        kinds={catalog.datasets}
        maxValues={catalog.maxValues}
        dataset={dataset}
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
              <Input
                id={id}
                type="number"
                value={target}
                onChange={(e) => {
                  const n = Number(e.target.value);
                  if (Number.isSafeInteger(n)) void setT(n);
                }}
              />
            </Field>
          ) : null}
          <OptionFields
            entry={entry}
            values={options}
            onChange={setOption}
          />
        </div>
      ) : null}
      {values && values.length > ANIMATION_LIMIT && entry.traced ? (
        <p className="rounded-md border border-border bg-subtle p-3 text-sm">
          {values.length} values are more than the animation can show clearly ({ANIMATION_LIMIT}).
          It still runs; for big data, the counts on the Compare page say more.
        </p>
      ) : null}
      {children(input)}
    </div>
  );
}
