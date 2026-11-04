/**
 * The state of a pathfinding run on a grid after some steps: the frontier
 * (cells pushed, not yet explored), the explored cells, the cell explored last,
 * the cost so far of each reached cell, and the path once found.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface GridState {
  frontier: Set<number>;
  visited: Set<number>;
  current: number | null;
  cost: Map<number, number>;
  path: number[];
  counts: Record<string, number>;
}

export const gridModel = (cells: number): ReplayModel<GridState> => ({
  start: () => ({
    frontier: new Set(),
    visited: new Set(),
    current: null,
    cost: new Map(),
    path: [],
    counts: {}
  }),
  clone: (s) => ({
    frontier: new Set(s.frontier),
    visited: new Set(s.visited),
    current: s.current,
    cost: new Map(s.cost),
    path: s.path.slice(),
    counts: { ...s.counts }
  }),
  apply(s, e) {
    const name = kinds[e.kind] ?? 'unknown';
    countInto(s.counts, name);
    const c = e.indices[0];
    const ok = c !== undefined && c >= 0 && c < cells;
    if (name === 'push' && ok) {
      s.frontier.add(c);
      if (e.values[0] !== undefined) s.cost.set(c, e.values[0]);
    } else if (name === 'visit' && ok) {
      s.frontier.delete(c);
      s.visited.add(c);
      s.current = c;
      if (e.values[0] !== undefined) s.cost.set(c, e.values[0]);
    } else if (name === 'mark' && e.note === 'path') {
      s.path = e.indices.filter((k) => k >= 0 && k < cells);
      s.current = null;
    }
  }
});

export class GridReplayer extends Replayer<GridState> {
  constructor(cells: number, events: TraceEvent[]) {
    super(gridModel(cells), events);
  }
}
