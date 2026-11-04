/**
 * The state of a graph algorithm after some steps: the nodes in the frontier
 * (pushed, not yet visited), visited, the node visited last, each node's
 * distance (Set), the path, the order nodes were placed in (a topological
 * order), each node's group (components, colors, sides), the edges chosen
 * (a spanning tree), the flow on each edge, and a cycle when one is found.
 */
import { kinds, type TraceEvent } from '../schema';
import { countInto, Replayer, type ReplayModel } from './replayer';

export interface GraphState {
  frontier: Set<number>;
  visited: Set<number>;
  current: number | null;
  /** The edge being considered at this step (Kruskal), cleared by the next event. */
  considering: number | null;
  distance: Map<number, number>;
  path: number[];
  order: Map<number, number>;
  group: Map<number, number>;
  tree: Set<number>;
  flow: Map<number, number>;
  cycle: Set<number>;
  counts: Record<string, number>;
}

export const graphModel = (nodes: number, edges: number): ReplayModel<GraphState> => {
  const node = (k: number) => k >= 0 && k < nodes;
  const edge = (k: number) => k >= 0 && k < edges;
  return {
    start: () => ({
      frontier: new Set(),
      visited: new Set(),
      current: null,
      considering: null,
      distance: new Map(),
      path: [],
      order: new Map(),
      group: new Map(),
      tree: new Set(),
      flow: new Map(),
      cycle: new Set(),
      counts: {}
    }),
    clone: (s) => ({
      frontier: new Set(s.frontier),
      visited: new Set(s.visited),
      current: s.current,
      considering: s.considering,
      distance: new Map(s.distance),
      path: s.path.slice(),
      order: new Map(s.order),
      group: new Map(s.group),
      tree: new Set(s.tree),
      flow: new Map(s.flow),
      cycle: new Set(s.cycle),
      counts: { ...s.counts }
    }),
    apply(s, e) {
      const name = kinds[e.kind] ?? 'unknown';
      countInto(s.counts, name);
      s.considering = null;
      const n = e.indices[0];
      switch (name) {
        case 'push':
          if (n !== undefined && node(n)) s.frontier.add(n);
          break;
        case 'visit':
          if (n !== undefined && node(n)) {
            s.frontier.delete(n);
            s.visited.add(n);
            s.current = n;
          }
          break;
        case 'set':
          if (n !== undefined && node(n) && e.values[0] !== undefined) {
            s.distance.set(n, e.values[0]);
            s.current = n;
          }
          break;
        case 'mark':
          switch (e.note) {
            case 'path':
              s.path = e.indices.filter(node);
              s.current = null;
              break;
            case 'order':
              for (const k of e.indices.filter(node)) s.order.set(k, s.order.size + 1);
              break;
            case 'group':
              if (e.values[0] !== undefined)
                for (const k of e.indices.filter(node)) s.group.set(k, e.values[0]);
              break;
            case 'edge':
              if (n !== undefined && edge(n)) s.considering = n;
              break;
            case 'tree':
              for (const k of e.indices.filter(edge)) s.tree.add(k);
              break;
            case 'flow':
              e.indices.forEach((k, i) => {
                if (edge(k) && e.values[i] !== undefined) s.flow.set(k, e.values[i]);
              });
              break;
            case 'cycle':
              s.cycle = new Set(e.indices.filter(node));
              break;
          }
          break;
      }
    }
  };
};

export class GraphReplayer extends Replayer<GraphState> {
  constructor(nodes: number, edges: number, events: TraceEvent[]) {
    super(graphModel(nodes, edges), events);
  }
}
