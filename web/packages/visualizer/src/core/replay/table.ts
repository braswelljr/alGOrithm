/**
 * The state of a dynamic-programming table after some steps: the cells filled
 * so far, the cell written last, the cells it was computed from, and the
 * cells of the answer once traced back.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface TableState {
  cells: (number | null)[];
  current: number | null;
  reads: number[];
  path: number[];
  /** Cells marked by a note: "fault" or "hit" in page replacement. */
  marks: Map<number, string>;
  counts: Record<string, number>;
}

export const tableModel = (size: number): ReplayModel<TableState> => ({
  start: () => ({
    cells: Array.from({ length: size }, () => null),
    current: null,
    reads: [],
    path: [],
    marks: new Map(),
    counts: {}
  }),
  clone: (s) => ({
    cells: s.cells.slice(),
    current: s.current,
    reads: s.reads.slice(),
    path: s.path.slice(),
    marks: new Map(s.marks),
    counts: { ...s.counts }
  }),
  apply(s, e) {
    const name = kinds[e.kind] ?? 'unknown';
    countInto(s.counts, name);
    const fits = (k: number) => k >= 0 && k < size;
    if (name === 'set') {
      const i = e.indices[0];
      if (i !== undefined && fits(i) && e.values[0] !== undefined) {
        s.cells[i] = e.values[0];
        s.current = i;
      }
    } else if (name === 'compare') {
      s.reads = e.indices.filter(fits);
    } else if (name === 'mark' && e.note === 'path') {
      s.path = e.indices.filter(fits);
      s.current = null;
      s.reads = [];
    } else if (name === 'mark' && e.note) {
      for (const k of e.indices.filter(fits)) s.marks.set(k, e.note);
    }
  }
});

export class TableReplayer extends Replayer<TableState> {
  constructor(size: number, events: TraceEvent[]) {
    super(tableModel(size), events);
  }
}
