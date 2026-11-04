/**
 * Positions for a graph's nodes. A directed acyclic graph is laid out in
 * layers (each node one layer below its deepest predecessor), so edges flow
 * downward; any other graph goes on a circle, which keeps every edge visible.
 */
import type { Point } from './layout';
import type { Graph } from './schema';

const LAYER_GAP = 110;
const NODE_GAP = 120;

/** The layer of each node, or null when the graph has a cycle or is undirected. */
export function layers(g: Graph): number[] | null {
  if (!g.directed) return null;
  const n = g.nodes.length;
  const indeg = new Array<number>(n).fill(0);
  const out: number[][] = Array.from({ length: n }, () => []);
  for (const e of g.edges) {
    out[e.from].push(e.to);
    indeg[e.to]++;
  }
  const layer = new Array<number>(n).fill(0);
  const queue = indeg.flatMap((d, i) => (d === 0 ? [i] : []));
  let seen = 0;
  while (queue.length) {
    const v = queue.shift()!;
    seen++;
    for (const w of out[v]) {
      layer[w] = Math.max(layer[w], layer[v] + 1);
      if (--indeg[w] === 0) queue.push(w);
    }
  }
  return seen === n ? layer : null;
}

export function layoutGraph(g: Graph): Point[] {
  const n = g.nodes.length;
  const layer = layers(g);
  if (layer) {
    const counts = new Map<number, number>();
    const widest = Math.max(
      1,
      ...[...new Set(layer)].map((l) => layer.filter((x) => x === l).length)
    );
    return layer.map((l) => {
      const k = counts.get(l) ?? 0;
      counts.set(l, k + 1);
      const inLayer = layer.filter((x) => x === l).length;
      return { x: (k + (widest - inLayer) / 2) * NODE_GAP, y: l * LAYER_GAP };
    });
  }
  const radius = Math.max(120, (n * NODE_GAP) / (2 * Math.PI));
  return g.nodes.map((_, i) => {
    const a = (2 * Math.PI * i) / Math.max(1, n) - Math.PI / 2;
    return { x: radius + radius * Math.cos(a), y: radius + radius * Math.sin(a) };
  });
}
