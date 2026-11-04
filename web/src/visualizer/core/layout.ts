/**
 * Positions for a snapshot's nodes, so a node-and-edge view (React Flow, or
 * anything else) can draw it. Rows and sets line up; trees get a tidy layout:
 * a binary tree (edges labeled "left" and "right") by in-order position, so a
 * lone left child sits to the left; any other tree, forest included, by
 * placing leaves in slots and each parent over the middle of its children.
 */
import type { Snapshot } from './schema';

export interface Point {
  x: number;
  y: number;
}

export const NODE_WIDTH = 64;
export const NODE_HEIGHT = 40;
const GAP_X = 28;
const GAP_Y = 64;
const SET_COLUMNS = 8;

export function layoutSnapshot(s: Snapshot): Map<number, Point> {
  const pos = new Map<number, Point>();
  if (s.layout === 'row') {
    s.nodes.forEach((n, i) => pos.set(n.id, { x: i * (NODE_WIDTH + GAP_X), y: 0 }));
    return pos;
  }
  if (s.layout === 'set') {
    s.nodes.forEach((n, i) =>
      pos.set(n.id, {
        x: (i % SET_COLUMNS) * (NODE_WIDTH + GAP_X),
        y: Math.floor(i / SET_COLUMNS) * (NODE_HEIGHT + GAP_X)
      })
    );
    return pos;
  }

  // Children in the order the edges list them, with left/right sides known.
  const children = new Map<number, { id: number; side?: string }[]>();
  const hasParent = new Set<number>();
  for (const e of s.edges) {
    if (!children.has(e.from)) children.set(e.from, []);
    children.get(e.from)?.push({ id: e.to, side: e.label });
    hasParent.add(e.to);
  }
  // A node whose parent is given but has no edge still hangs under it.
  for (const n of s.nodes) {
    if (n.parent >= 0 && !hasParent.has(n.id)) {
      if (!children.has(n.parent)) children.set(n.parent, []);
      children.get(n.parent)?.push({ id: n.id });
      hasParent.add(n.id);
    }
  }
  const roots = s.nodes.filter((n) => !hasParent.has(n.id)).map((n) => n.id);
  const binary =
    s.edges.length > 0 && s.edges.every((e) => e.label === 'left' || e.label === 'right');
  const step = NODE_WIDTH + GAP_X;
  let slot = 0;
  const seen = new Set<number>();

  const place = (id: number, depth: number): number => {
    if (seen.has(id)) return pos.get(id)?.x ?? 0; // a cycle: draw it once
    seen.add(id);
    const kids = children.get(id) ?? [];
    let x: number;
    if (binary) {
      const left = kids.find((k) => k.side === 'left');
      const right = kids.find((k) => k.side === 'right');
      if (left) place(left.id, depth + 1);
      x = slot++ * step;
      pos.set(id, { x, y: depth * GAP_Y });
      if (right) place(right.id, depth + 1);
      return x;
    }
    if (kids.length < 1) {
      x = slot++ * step;
    } else {
      const xs = kids.map((k) => place(k.id, depth + 1));
      x = (Math.min(...xs) + Math.max(...xs)) / 2;
    }
    pos.set(id, { x, y: depth * GAP_Y });
    return x;
  };

  for (const r of roots) {
    place(r, 0);
    slot += 1; // a gap between the trees of a forest
  }
  // Anything unreachable (should not happen) goes on a line below.
  const maxY = Math.max(0, ...[...pos.values()].map((p) => p.y));
  let extra = 0;
  for (const n of s.nodes) {
    if (!pos.has(n.id)) pos.set(n.id, { x: extra++ * step, y: maxY + GAP_Y });
  }
  return pos;
}
