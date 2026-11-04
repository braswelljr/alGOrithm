'use client';

import { useCallback, useMemo } from 'react';
import { createParser, useQueryState } from 'nuqs';
import { z } from 'zod';
import { inputSchema, type Entry, type Input } from '@/visualizer/core';

const recordSchema = z.record(z.string(), z.number().int());

/** A JSON value in the link, parsed with a schema; anything else reads as absent. */
function jsonParam<T>(schema: z.ZodType<T>) {
  return createParser({
    parse: (v: string) => {
      try {
        const r = schema.safeParse(JSON.parse(v));
        return r.success ? r.data : null;
      } catch {
        return null;
      }
    },
    serialize: (v: T) => JSON.stringify(v)
  });
}

const inputParam = jsonParam(inputSchema);
const optionsParam = jsonParam(recordSchema);

/**
 * An entry's input, kept in the link as `in` (JSON, validated) and starting
 * from the entry's example, so a first visit always shows something.
 */
export function useEntryInput(entry: Entry): [Input, (next: Input) => void] {
  const [q, setQ] = useQueryState('in', inputParam.withOptions({ history: 'replace' }));
  // A stable setter: callers memoize on it (RunPanel's onInputChange).
  const set = useCallback((next: Input) => void setQ(next), [setQ]);
  return [q ?? entry.example, set];
}

/**
 * The entry's numeric options, kept in the link as `opt`, each starting at
 * its default (or the example's value). Only names the entry declares count.
 * The third value sets them all at once (a form's Apply), which one-at-a-time
 * writes cannot: each would start from the same old values.
 */
export function useEntryOptions(
  entry: Entry
): [
  Record<string, number>,
  (name: string, v: number) => void,
  (all: Record<string, number>) => void
] {
  const [q, setQ] = useQueryState('opt', optionsParam.withOptions({ history: 'replace' }));
  const values = useMemo(() => {
    const out: Record<string, number> = {};
    for (const o of entry.options)
      out[o.name] = q?.[o.name] ?? entry.example.options?.[o.name] ?? o.default;
    return out;
  }, [entry, q]);
  return [
    values,
    (name, v) => void setQ({ ...values, [name]: v }),
    (all) => void setQ({ ...values, ...all })
  ];
}
