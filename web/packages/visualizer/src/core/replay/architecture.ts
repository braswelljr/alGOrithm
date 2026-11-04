/**
 * The state of a system design at a step: which requests are on which links,
 * each component's latest load, and the running totals of requests by status.
 * Step 0 is the design before any request; step i shows frame i - 1.
 */
import type { Architecture, ArchitectureFrame, RequestStatus } from '../schema';

export interface ArchitectureState {
  step: number;
  frame?: ArchitectureFrame;
  /** The latest load each component reported up to this step. */
  loads: Map<string, { load: number; queue: number; hitRate: number }>;
  /** Requests seen so far, by status. */
  totals: Record<RequestStatus, number>;
}

const empty = (): Record<RequestStatus, number> => ({
  ok: 0,
  hit: 0,
  miss: 0,
  error: 0,
  queued: 0,
  dropped: 0
});

/** The state after step steps (clamped to the frames there are). */
export function architectureAt(a: Architecture, step: number): ArchitectureState {
  const n = Math.max(0, Math.min(step, a.frames.length));
  const loads: ArchitectureState['loads'] = new Map();
  const totals = empty();
  for (let k = 0; k < n; k++) {
    const f = a.frames[k]!;
    for (const q of f.requests) totals[q.status]++;
    for (const l of f.loads)
      loads.set(l.component, { load: l.load, queue: l.queue, hitRate: l.hitRate });
  }
  return { step: n, frame: n > 0 ? a.frames[n - 1] : undefined, loads, totals };
}

/** Columns for the components, left to right in the direction requests
 * flow: a component's column is one more than the furthest component that
 * links to it, counting each component once so cycles stop. Rows stack the
 * components of a column. */
export function layoutArchitecture(a: Architecture): Map<string, { x: number; y: number }> {
  const incoming = new Map<string, string[]>();
  for (const c of a.components) incoming.set(c.id, []);
  for (const l of a.links) if (l.from !== l.to) incoming.get(l.to)?.push(l.from);
  const col = new Map<string, number>();
  const visiting = new Set<string>();
  const depth = (id: string): number => {
    const known = col.get(id);
    if (known !== undefined) return known;
    if (visiting.has(id)) return 0;
    visiting.add(id);
    let d = 0;
    for (const from of incoming.get(id) ?? []) d = Math.max(d, depth(from) + 1);
    visiting.delete(id);
    col.set(id, d);
    return d;
  };
  for (const c of a.components) depth(c.id);
  const columns = new Map<number, string[]>();
  for (const c of a.components) {
    const k = col.get(c.id) ?? 0;
    columns.set(k, [...(columns.get(k) ?? []), c.id]);
  }
  const tallest = Math.max(1, ...[...columns.values()].map((v) => v.length));
  const pos = new Map<string, { x: number; y: number }>();
  for (const [k, ids] of columns) {
    const offset = ((tallest - ids.length) * 120) / 2;
    ids.forEach((id, row) => pos.set(id, { x: k * 230, y: offset + row * 120 }));
  }
  return pos;
}

const STATUS_WORDS: Record<RequestStatus, string> = {
  ok: 'ok',
  hit: 'hit',
  miss: 'miss',
  error: 'failed',
  queued: 'queued',
  dropped: 'dropped'
};

/** The words for a step: its note, then each request on its link. */
export function describeArchitecture(a: Architecture, s: ArchitectureState): string {
  if (!s.frame) return 'The design before any request.';
  const label = (id: string) => a.components.find((c) => c.id === id)?.label ?? id;
  const moves = s.frame.requests.map((q) => {
    const l = a.links[q.link]!;
    return `${label(l.from)} to ${label(l.to)} (${STATUS_WORDS[q.status]})`;
  });
  return moves.length ? `${s.frame.note}. ${moves.join('; ')}.` : `${s.frame.note}.`;
}

export { STATUS_WORDS };
