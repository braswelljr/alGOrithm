/**
 * The state of a geometry algorithm after some steps: the points compared
 * last, the points visited, the hull so far (in order), the closest pair, and
 * the points found inside a polygon, and each point's group (a cluster, or
 * -1 for noise) from Mark "group" events, whose values[0] is the group.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface PointsState {
  compared: number[];
  visited: Set<number>;
  current: number | null;
  hull: number[];
  pair: [number, number] | null;
  inside: Set<number>;
  groups: Map<number, number>;
  counts: Record<string, number>;
}

export const pointsModel = (count: number): ReplayModel<PointsState> => {
  const fits = (k: number) => k >= 0 && k < count;
  return {
    start: () => ({
      compared: [],
      visited: new Set(),
      current: null,
      hull: [],
      pair: null,
      inside: new Set(),
      groups: new Map(),
      counts: {}
    }),
    clone: (s) => ({
      compared: s.compared.slice(),
      visited: new Set(s.visited),
      current: s.current,
      hull: s.hull.slice(),
      pair: s.pair ? [s.pair[0], s.pair[1]] : null,
      inside: new Set(s.inside),
      groups: new Map(s.groups),
      counts: { ...s.counts }
    }),
    apply(s, e) {
      const name = kinds[e.kind] ?? 'unknown';
      countInto(s.counts, name);
      const pts = e.indices.filter(fits);
      if (name === 'compare') s.compared = pts;
      else if (name === 'visit' && pts.length) {
        for (const k of pts) s.visited.add(k);
        s.current = pts[pts.length - 1];
      } else if (name === 'mark') {
        if (e.note === 'hull') s.hull = pts;
        else if (e.note === 'pair' && pts.length >= 2) s.pair = [pts[0], pts[1]];
        else if (e.note === 'inside') for (const k of pts) s.inside.add(k);
        else if (e.note === 'group') for (const k of pts) s.groups.set(k, e.values[0] ?? 0);
      }
    }
  };
};

export class PointsReplayer extends Replayer<PointsState> {
  constructor(count: number, events: TraceEvent[]) {
    super(pointsModel(count), events);
  }
}
