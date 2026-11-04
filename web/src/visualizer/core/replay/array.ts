/**
 * The state of an array algorithm (a sort or a search) after some steps.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface ArrayState {
  values: number[];
  /** Lasting marks. */
  sorted: boolean[];
  pivot: number[];
  min: number[];
  /** The range still searched, or a flipped prefix. */
  range: [number, number] | null;
  flip: [number, number] | null;
  found: number | null;
  /** The place value a radix pass is sorting on (1, 10, 100, ...). */
  digit: number | null;
  counts: Record<string, number>;
}

/** Highlights of the step's own event. */
export interface ArrayStep extends ArrayState {
  step: number;
  compared: number[];
  swapped: number[];
  written: number[];
  visited: number[];
}

function inBounds(s: ArrayState, i: number | undefined): i is number {
  return i !== undefined && Number.isInteger(i) && i >= 0 && i < s.values.length;
}

function pair(s: ArrayState, e: TraceEvent): [number, number] | null {
  const [i, j] = e.indices;
  return inBounds(s, i) && inBounds(s, j) ? [i, j] : null;
}

export function arrayModel(initial: number[]): ReplayModel<ArrayState> {
  return {
    start: () => ({
      values: initial.slice(),
      sorted: initial.map(() => false),
      pivot: [],
      min: [],
      range: null,
      flip: null,
      found: null,
      digit: null,
      counts: {}
    }),
    clone: (s) => ({
      values: s.values.slice(),
      sorted: s.sorted.slice(),
      pivot: s.pivot.slice(),
      min: s.min.slice(),
      range: s.range ? [s.range[0], s.range[1]] : null,
      flip: s.flip ? [s.flip[0], s.flip[1]] : null,
      found: s.found,
      digit: s.digit,
      counts: { ...s.counts }
    }),
    apply(s, e) {
      const name = kinds[e.kind] ?? 'unknown';
      countInto(s.counts, name);
      const [i, j] = e.indices;
      if (name === 'swap' && inBounds(s, i) && inBounds(s, j)) {
        [s.values[i], s.values[j]] = [s.values[j], s.values[i]];
      } else if (name === 'set' && inBounds(s, i) && e.values[0] !== undefined) {
        s.values[i] = e.values[0];
      } else if (name === 'mark') {
        switch (e.note) {
          case 'sorted':
            for (const k of e.indices) if (inBounds(s, k)) s.sorted[k] = true;
            break;
          case 'pivot':
            s.pivot = e.indices.filter((k) => inBounds(s, k));
            break;
          case 'min':
            s.min = e.indices.filter((k) => inBounds(s, k));
            break;
          case 'range':
            s.range = pair(s, e);
            break;
          case 'flip':
            s.flip = pair(s, e);
            break;
          case 'found':
            s.found = inBounds(s, i) ? i : null;
            break;
          case 'digit':
            s.digit = e.values[0] ?? null;
            break;
        }
      }
    }
  };
}

/** Replays an array algorithm's events over the array they index into. */
export class ArrayReplayer {
  private readonly r: Replayer<ArrayState>;

  constructor(initial: number[], events: TraceEvent[]) {
    this.r = new Replayer(arrayModel(initial), events);
  }

  get length(): number {
    return this.r.length;
  }

  get events(): TraceEvent[] {
    return this.r.events;
  }

  lastEvent(n: number) {
    return this.r.lastEvent(n);
  }

  stateAt(n: number): ArrayStep {
    const s = this.r.stateAt(n);
    const step = Math.max(0, Math.min(Math.trunc(n) || 0, this.r.length));
    const last = this.r.lastEvent(step);
    const name = last ? kinds[last.kind] : undefined;
    const idx = last ? last.indices.filter((k) => inBounds(s, k)) : [];
    return {
      ...s,
      step,
      compared: name === 'compare' ? idx : [],
      swapped: name === 'swap' ? idx : [],
      written: name === 'set' ? idx : [],
      visited: name === 'visit' ? idx : []
    };
  }

  /** The values after every event. */
  final(): number[] {
    return this.r.stateAt(this.r.length).values;
  }
}
