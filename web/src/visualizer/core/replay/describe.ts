/**
 * Describes a step in words, for the live region under each view and for the
 * step log (docs/rules/the-markup-says-what-the-thing-is.md).
 */
import { kinds, type TraceEvent } from '../schema';

const at = (values: number[] | undefined, i: number | undefined) =>
  values && i !== undefined && i >= 0 && i < values.length ? String(values[i]) : '?';

/** An array algorithm's step; values is the array before or after it. */
export function describeArrayEvent(e: TraceEvent | undefined, values: number[]): string {
  if (!e) return 'The input, before the first step.';
  const name = kinds[e.kind] ?? 'unknown';
  const [i, j] = e.indices;
  switch (name) {
    case 'compare':
      if (e.indices.length === 1) {
        return e.values.length === 1
          ? `Compare ${at(values, i)} at position ${i} with ${e.values[0]}.`
          : `Look at ${at(values, i)} at position ${i}.`;
      }
      return `Compare ${at(values, i)} at position ${i} with ${at(values, j)} at position ${j}.`;
    case 'swap':
      return `Swap positions ${i} and ${j}: now ${at(values, i)} and ${at(values, j)}.`;
    case 'set':
      return `Write ${e.values[0]} at position ${i}.`;
    case 'visit':
      return `Read ${at(values, i)} at position ${i}.`;
    case 'mark':
      switch (e.note) {
        case 'sorted':
          return `Position ${e.indices.join(', ')} is in its final place.`;
        case 'pivot':
          return e.indices.length
            ? `The pivot is ${at(values, i)} at position ${i}.`
            : 'The pivot is in place.';
        case 'min':
          return e.indices.length
            ? `The smallest so far is ${at(values, i)} at position ${i}.`
            : 'The pass is done.';
        case 'range':
          return `Search positions ${i} to ${j}.`;
        case 'flip':
          return e.indices.length ? `Flip positions ${i} to ${j}.` : 'The flip is done.';
        case 'found':
          return `Found the target at position ${i}.`;
        case 'digit':
          return `Sort on the ${e.values[0]}s digit.`;
        case 'grow':
          return `Full: grow the array to capacity ${i}.`;
        case 'pop':
          return `Remove position ${i}.`;
      }
      return `Mark ${e.note ?? ''}.`;
  }
  return `${name}.`;
}

/** A grid search's step. */
export function describeGridEvent(e: TraceEvent | undefined, cols: number): string {
  if (!e) return 'The grid, before the search starts.';
  const name = kinds[e.kind] ?? 'unknown';
  const cell = (c: number | undefined) =>
    c === undefined ? '?' : `row ${Math.floor(c / cols) + 1}, column ${(c % cols) + 1}`;
  const cost = e.values[0] !== undefined ? ` (cost so far ${e.values[0]})` : '';
  switch (name) {
    case 'push':
      return `Add ${cell(e.indices[0])} to the frontier${cost}.`;
    case 'visit':
      return `Explore ${cell(e.indices[0])}${cost}.`;
    case 'mark':
      return e.note === 'path'
        ? `The path: ${e.indices.length} cells, start to end.`
        : `Mark ${e.note ?? ''}.`;
  }
  return `${name}.`;
}

/** A dynamic-programming step; cols is the table's width. */
export function describeTableEvent(
  e: TraceEvent | undefined,
  cols: number,
  rowLabels: string[],
  colLabels: string[]
): string {
  if (!e) return 'An empty table.';
  const name = kinds[e.kind] ?? 'unknown';
  const cell = (k: number) => {
    const r = Math.floor(k / cols);
    const c = k % cols;
    const rl = rowLabels[r] ? ` (${rowLabels[r]})` : '';
    const cl = colLabels[c] ? ` (${colLabels[c]})` : '';
    return rowLabels.length ? `row ${r}${rl}, column ${c}${cl}` : `cell ${c}${cl}`;
  };
  switch (name) {
    case 'set':
      return `Fill ${cell(e.indices[0] ?? 0)} with ${e.values[0]}.`;
    case 'compare':
      return `Read ${e.indices.map(cell).join('; ')}.`;
    case 'mark':
      if (e.note === 'path') return `Trace back the answer through ${e.indices.length} cells.`;
      if (e.note === 'fault') return `A page fault at ${cell(e.indices[0] ?? 0)}.`;
      if (e.note === 'hit')
        return `A hit at ${cell(e.indices[0] ?? 0)}: the page is already in memory.`;
      return e.note ?? '';
  }
  return `${name}.`;
}

/** A string search's step. */
export function describeTextEvent(e: TraceEvent | undefined, text: string, pattern: string) {
  if (!e) return 'The text and the pattern, before the search.';
  const name = kinds[e.kind] ?? 'unknown';
  const [a, b] = e.indices;
  switch (name) {
    case 'compare': {
      const j = e.values[0] ?? 0;
      const same = text[a ?? -1] === pattern[j];
      return `Compare text letter ${a} "${text[a ?? -1] ?? '?'}" with pattern letter ${j} "${pattern[j] ?? '?'}": ${same ? 'a match' : 'different'}.`;
    }
    case 'visit':
      return `Check the hash of the window from ${a} to ${b}.`;
    case 'mark':
      return e.note === 'found' ? `A match from ${a} to ${b}.` : '';
  }
  return `${name}.`;
}

/** A graph algorithm's step, with the node names. */
export function describeGraphEvent(
  e: TraceEvent | undefined,
  names: string[],
  edges: { from: number; to: number }[]
): string {
  if (!e) return 'The graph, before the first step.';
  const name = kinds[e.kind] ?? 'unknown';
  const node = (k: number | undefined) =>
    k !== undefined && names[k] !== undefined ? names[k] : '?';
  const edge = (k: number) => (edges[k] ? `${node(edges[k].from)}-${node(edges[k].to)}` : '?');
  const [n] = e.indices;
  switch (name) {
    case 'push':
      return `Add ${node(n)} to the frontier.`;
    case 'visit':
      return `Visit ${node(n)}.`;
    case 'set':
      return `The distance to ${node(n)} is now ${e.values[0]}.`;
    case 'compare':
      return `Look at ${e.indices.map(node).join(' and ')}.`;
    case 'mark':
      switch (e.note) {
        case 'path':
          return `The path: ${e.indices.map(node).join(' to ')}.`;
        case 'order':
          return `Place ${e.indices.map(node).join(', ')} next in the order.`;
        case 'group':
          return `${e.indices.map(node).join(', ')} ${e.indices.length > 1 ? 'are' : 'is'} in group ${e.values[0]}.`;
        case 'edge':
          return `Consider edge ${e.indices.map(edge).join(', ')}.`;
        case 'tree':
          return `Take edge ${e.indices.map(edge).join(', ')}.`;
        case 'flow':
          return `Flow ${e.indices.map((k, i) => `${e.values[i]} on ${edge(k)}`).join(', ')}.`;
        case 'cycle':
          return `A cycle: ${e.indices.map(node).join(', ')}.`;
      }
      return `Mark ${e.note ?? ''}.`;
  }
  return `${name}.`;
}

/** A geometry algorithm's step. */
export function describePointsEvent(
  e: TraceEvent | undefined,
  points: { x: number; y: number }[]
): string {
  if (!e) return 'The points, before the first step.';
  const name = kinds[e.kind] ?? 'unknown';
  const p = (k: number) => (points[k] ? `(${points[k].x}, ${points[k].y})` : '?');
  switch (name) {
    case 'compare':
      return `Compare ${e.indices.map(p).join(' and ')}.`;
    case 'visit':
      return `Look at ${e.indices.map(p).join(', ')}.`;
    case 'mark':
      switch (e.note) {
        case 'hull':
          return `The hull so far has ${e.indices.length} points.`;
        case 'pair':
          return `The closest pair so far: ${p(e.indices[0])} and ${p(e.indices[1])}.`;
        case 'inside':
          return `${e.indices.map(p).join(', ')} ${e.indices.length > 1 ? 'are' : 'is'} inside.`;
        case 'group':
          return (e.values[0] ?? 0) < 0
            ? `${e.indices.length} ${e.indices.length === 1 ? 'point is' : 'points are'} noise.`
            : `Group ${e.values[0] ?? 0} now has ${e.indices.length} ${e.indices.length === 1 ? 'point' : 'points'}.`;
      }
      return `Mark ${e.note ?? ''}.`;
  }
  return `${name}.`;
}
