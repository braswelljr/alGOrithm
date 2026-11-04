/**
 * What changed between two snapshots of a data structure, so a view can
 * highlight the nodes and links the last operation touched. Nodes are matched
 * by label and position among equal labels, since ids are renumbered on every
 * snapshot.
 */
import type { Snapshot } from '../schema';

export interface SnapshotDiff {
  /** Ids (in the new snapshot) of nodes that are new or moved. */
  added: Set<number>;
  /** Labels of nodes that are gone. */
  removed: string[];
  /** Keys "from-to" (new ids) of edges that are new. */
  newEdges: Set<string>;
}

/** A node's identity across snapshots: its label and its parent's label. */
function keysOf(s: Snapshot): string[] {
  const label = new Map(s.nodes.map((n) => [n.id, n.label]));
  const seen = new Map<string, number>();
  return s.nodes.map((n) => {
    const base = `${n.label}\u0000${n.parent >= 0 ? (label.get(n.parent) ?? '') : ''}`;
    const k = (seen.get(base) ?? 0) + 1;
    seen.set(base, k);
    return `${base}\u0000${k}`;
  });
}

export function diffSnapshots(before: Snapshot | undefined, after: Snapshot): SnapshotDiff {
  if (!before) return { added: new Set(), removed: [], newEdges: new Set() };
  const was = new Set(keysOf(before));
  const now = keysOf(after);
  const added = new Set<number>();
  after.nodes.forEach((n, i) => {
    if (!was.has(now[i])) added.add(n.id);
  });
  const nowSet = new Set(now);
  const beforeKeys = keysOf(before);
  const removed = before.nodes.filter((_, i) => !nowSet.has(beforeKeys[i])).map((n) => n.label);

  const labelBefore = new Map(before.nodes.map((n) => [n.id, n.label]));
  const labelAfter = new Map(after.nodes.map((n) => [n.id, n.label]));
  const oldEdges = new Set(
    before.edges.map((e) => `${labelBefore.get(e.from)}\u0000${labelBefore.get(e.to)}`)
  );
  const newEdges = new Set<string>();
  for (const e of after.edges) {
    if (!oldEdges.has(`${labelAfter.get(e.from)}\u0000${labelAfter.get(e.to)}`)) {
      newEdges.add(`${e.from}-${e.to}`);
    }
  }
  return { added, removed, newEdges };
}
