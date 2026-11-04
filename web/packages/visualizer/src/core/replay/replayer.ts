/**
 * Replays an event log into states, keeping a checkpoint every CHECKPOINT
 * events so that the state after any step is a short replay away. Each kind
 * of view (arrays, grids, tables, text) supplies how one event changes its
 * state; this class does the bookkeeping.
 */
import type { TraceEvent } from '../schema';

/** How many events lie between checkpoints. */
export const CHECKPOINT = 256;

export interface ReplayModel<S> {
  /** The state before any event. */
  start(): S;
  /** A deep copy of s. */
  clone(s: S): S;
  /** Applies e to s in place; events that do not fit are ignored. */
  apply(s: S, e: TraceEvent): void;
}

export class Replayer<S> {
  readonly events: TraceEvent[];
  private readonly model: ReplayModel<S>;
  private readonly checkpoints: S[] = [];

  constructor(model: ReplayModel<S>, events: TraceEvent[]) {
    this.model = model;
    this.events = events;
    const s = model.start();
    for (let k = 0; k <= events.length; k++) {
      if (k % CHECKPOINT === 0) this.checkpoints.push(model.clone(s));
      if (k < events.length) model.apply(s, events[k]);
    }
  }

  /** The number of events: the last step. */
  get length(): number {
    return this.events.length;
  }

  /** The event applied last at step n (undefined at step 0). */
  lastEvent(n: number): TraceEvent | undefined {
    return n > 0 ? this.events[Math.min(n, this.events.length) - 1] : undefined;
  }

  /** The state after the first n events. */
  stateAt(n: number): S {
    const step = Math.max(0, Math.min(Math.trunc(n) || 0, this.events.length));
    const base = Math.floor(step / CHECKPOINT);
    const s = this.model.clone(this.checkpoints[base]);
    for (let k = base * CHECKPOINT; k < step; k++) this.model.apply(s, this.events[k]);
    return s;
  }
}

/** Counts events per kind name. */
export function countInto(counts: Record<string, number>, name: string) {
  counts[name] = (counts[name] ?? 0) + 1;
}
