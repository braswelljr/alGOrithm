/** Where a learning path's step opens: the entry's page with the step's input in the link. */
import type { Entry, PathStep } from '@/visualizer/core';
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

/** A Mermaid node id from a path id: letters, digits and underscores only. */
const nodeId = (id: string) => `p_${id.replace(/[^A-Za-z0-9]/g, '_')}`;

/** A label safe inside Mermaid's quotes. */
const label = (text: string) => text.replace(/"/g, '#quot;');

/**
 * The learning paths as a Mermaid flowchart: one node per path, an arrow
 * from each prerequisite to the path that needs it, each node linking to its
 * page. Pure, so the chart can be tested without drawing it.
 */
export function pathsChart(
  paths: readonly { id: string; title: string; prerequisites: readonly string[] }[]
): string {
  const known = new Set(paths.map((p) => p.id));
  const lines = ['flowchart LR'];
  for (const p of paths) lines.push(`  ${nodeId(p.id)}["${label(p.title)}"]`);
  for (const p of paths)
    for (const pre of p.prerequisites)
      if (known.has(pre)) lines.push(`  ${nodeId(pre)} --> ${nodeId(p.id)}`);
  for (const p of paths) lines.push(`  click ${nodeId(p.id)} "${pathHref(p.id)}"`);
  return lines.join('\n');
}
