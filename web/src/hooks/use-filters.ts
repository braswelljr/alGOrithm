'use client';

import { parseAsArrayOf, parseAsString, useQueryStates } from 'nuqs';
import type { Filters } from '@/lib/filters';

const list = parseAsArrayOf(parseAsString).withDefault([]);

/** The catalog filters, in the link as ?kind=...&level=...&tag=a,b. */
export function useFilters(): [Filters, (next: Partial<Filters>) => void] {
  const [f, set] = useQueryStates({ kind: list, level: list, tag: list }, { history: 'replace' });
  return [f, (next) => void set(next)];
}
