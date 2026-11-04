/** Where a learning path's step opens: the entry's page with the step's input in the link. */
import type { Entry, PathStep } from '@braswelljr/algorithm-visualizer/core';
import { entryHref } from './entries';

/** The page that runs a step: its entry with the step's input and options preloaded. */
export function stepHref(entry: Entry, step: PathStep): string {
  const base = entryHref(entry);
  if (!step.input) return base;
  const { options, ...input } = step.input;
  const params = new URLSearchParams();
  if (Object.keys(input).length > 0) params.set('in', JSON.stringify(input));
  if (options && Object.keys(options).length > 0) params.set('opt', JSON.stringify(options));
  const query = params.toString();
  return query ? `${base}&${query}` : base;
}

/** The page of one learning path. */
export function pathHref(id: string): string {
  return `/learn/${id}/`;
}
