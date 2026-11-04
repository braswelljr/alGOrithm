/**
 * The state of a string search after some steps: the letter of the text
 * being compared with a letter of the pattern, the window a rolling hash
 * checked, and every match found so far.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface TextState {
  /** Text index and pattern index of the comparison just made. */
  compare: { text: number; pattern: number } | null;
  /** The window checked, [start, end] inclusive. */
  window: [number, number] | null;
  found: [number, number][];
  counts: Record<string, number>;
}

export const textModel = (length: number): ReplayModel<TextState> => ({
  start: () => ({ compare: null, window: null, found: [], counts: {} }),
  clone: (s) => ({
    compare: s.compare ? { ...s.compare } : null,
    window: s.window ? [s.window[0], s.window[1]] : null,
    found: s.found.map(([a, b]) => [a, b]),
    counts: { ...s.counts }
  }),
  apply(s, e) {
    const name = kinds[e.kind] ?? 'unknown';
    countInto(s.counts, name);
    const [a, b] = e.indices;
    const fits = (k: number | undefined): k is number => k !== undefined && k >= 0 && k < length;
    if (name === 'compare' && fits(a)) {
      s.compare = { text: a, pattern: e.values[0] ?? 0 };
    } else if (name === 'visit' && fits(a) && fits(b)) {
      s.window = [a, b];
      s.compare = null;
    } else if (name === 'mark' && e.note === 'found' && fits(a) && fits(b)) {
      s.found.push([a, b]);
    }
  }
});

export class TextReplayer extends Replayer<TextState> {
  constructor(length: number, events: TraceEvent[]) {
    super(textModel(length), events);
  }
}
