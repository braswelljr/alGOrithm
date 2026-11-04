'use client';

import { useMemo, useSyncExternalStore } from 'react';
import { parseAsInteger, parseAsString, useQueryStates } from 'nuqs';
import type { AlgorithmApi as Api } from '@/visualizer/core';

/** The most values written into a link; more are kept in this browser. */
export const URL_VALUES = 200;

const params = {
  data: parseAsString,
  n: parseAsInteger,
  seed: parseAsInteger,
  v: parseAsString
};

// Data too large for a link lives here (and in sessionStorage, so a reload
// keeps it). It never leaves the browser
// (docs/rules/the-users-data-never-leaves-the-browser.md).
const STORE_KEY = 'algorithm:dataset';
let local: { values: number[]; label: string } | null = null;
const listeners = new Set<() => void>();

function readStored(): { values: number[]; label: string } | null {
  if (local) return local;
  try {
    const raw = sessionStorage.getItem(STORE_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed === 'object' &&
      parsed !== null &&
      'values' in parsed &&
      'label' in parsed &&
      Array.isArray(parsed.values) &&
      parsed.values.every((x) => Number.isInteger(x)) &&
      typeof parsed.label === 'string'
    ) {
      local = { values: parsed.values, label: parsed.label };
    }
  } catch {
    // Storage blocked or corrupt: the data is simply gone.
  }
  return local;
}

function store(values: number[], label: string) {
  local = { values, label };
  try {
    sessionStorage.setItem(STORE_KEY, JSON.stringify(local));
  } catch {
    // Too large or blocked: kept in memory for this page only.
  }
  listeners.forEach((l) => l());
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

/** Parses the v parameter: integers separated by commas. */
export function parseList(v: string): number[] | null {
  const parts = v.split(',').filter((p) => p.trim() !== '');
  const out = parts.map((p) => Number(p.trim()));
  return out.length > 0 && out.every((x) => Number.isSafeInteger(x)) ? out : null;
}

export interface Dataset {
  values: number[] | null;
  label: string;
  error: string | null;
  kind: string;
  size: number;
  seed: number;
  /** Use a generated dataset. */
  generate(kind: string, size: number, seed: number): void;
  /** Use these values; a short list goes into the link, a long one stays here. */
  use(values: number[], label: string): void;
}

/** The dataset a page works on, read from and written to the URL. */
export function useDataset(api: Api | null, defaultSize: number, fallback?: number[]): Dataset {
  const [q, setQ] = useQueryStates(params, { history: 'replace' });
  const stored = useSyncExternalStore(subscribe, readStored, () => null);

  // With nothing in the link, the example (when given) comes first.
  const kind = q.data ?? (fallback?.length ? 'example' : 'random');
  const size = q.n ?? defaultSize;
  const seed = q.seed ?? 1;

  const { values, label, error } = useMemo((): {
    values: number[] | null;
    label: string;
    error: string | null;
  } => {
    if (kind === 'example' && fallback?.length) {
      return { values: fallback, label: `the example's ${fallback.length} values`, error: null };
    }
    if (kind === 'list') {
      const list = q.v ? parseList(q.v) : null;
      return list
        ? { values: list, label: `${list.length} values from the link`, error: null }
        : {
            values: null,
            label: '',
            error: 'The values in the link are not whole numbers separated by commas.'
          };
    }
    if (kind === 'local') {
      return stored
        ? { values: stored.values, label: stored.label, error: null }
        : {
            values: null,
            label: '',
            error:
              'The data for this link was kept in another browser tab or session. Load it again.'
          };
    }
    if (!api) return { values: null, label: '', error: null };
    const r = api.generate(kind, size, seed);
    return r.ok
      ? { values: r.value, label: `${size} ${kind} values, seed ${seed}`, error: null }
      : { values: null, label: '', error: r.failure.error };
  }, [api, kind, size, seed, q.v, stored, fallback]);

  return {
    values,
    label,
    error,
    kind,
    size,
    seed,
    generate: (k, n, s) => void setQ({ data: k, n, seed: s, v: null }),
    use(vals, lbl) {
      if (vals.length <= URL_VALUES) {
        void setQ({ data: 'list', v: vals.join(','), n: null, seed: null });
      } else {
        store(vals, lbl);
        void setQ({ data: 'local', v: null, n: null, seed: null });
      }
    }
  };
}
