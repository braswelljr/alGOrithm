/**
 * Automata and cellular automata, played step by step. An automaton's run
 * carries its own steps (the active states, the edges taken, the tape or the
 * stack), so playing it is reading step i. A cellular automaton is a table
 * filled by Set events, played one generation at a time: the steps are the
 * positions of its "generation" marks.
 */
import { kinds, type Automaton, type AutomatonStep, type TraceEvent } from '../schema';

export interface StatePoint {
  x: number;
  y: number;
}

/** Spacing between the columns and rows of the automaton layout. */
export const STATE_GAP_X = 170;
export const STATE_GAP_Y = 120;

/**
 * Places the states left to right by their distance from the start (breadth
 * first over the transitions), each column centered; states the start never
 * reaches go in a last column. Deterministic, so a replay never moves a state.
 */
export function layoutAutomaton(m: Automaton): StatePoint[] {
  const n = m.states.length;
  const start = Math.max(
    0,
    m.states.findIndex((s) => s.start)
  );
  const depth = Array.from({ length: n }, () => -1);
  const out = Array.from({ length: n }, () => [] as number[]);
  for (const e of m.edges) if (e.from !== e.to) out[e.from]?.push(e.to);
  if (n > 0) {
    depth[start] = 0;
    const queue = [start];
    for (let i = 0; i < queue.length; i++) {
      const s = queue[i]!;
      for (const t of out[s] ?? []) {
        if (depth[t] === -1) {
          depth[t] = depth[s]! + 1;
          queue.push(t);
        }
      }
    }
  }
  const deepest = Math.max(0, ...depth);
  const columns = new Map<number, number[]>();
  depth.forEach((d, i) => {
    const col = d === -1 ? deepest + 1 : d;
    if (!columns.has(col)) columns.set(col, []);
    columns.get(col)!.push(i);
  });
  const tallest = Math.max(1, ...[...columns.values()].map((c) => c.length));
  const pos: StatePoint[] = Array.from({ length: n }, () => ({ x: 0, y: 0 }));
  for (const [col, members] of columns) {
    const offset = ((tallest - members.length) * STATE_GAP_Y) / 2;
    members.forEach((s, row) => {
      pos[s] = { x: col * STATE_GAP_X, y: offset + row * STATE_GAP_Y };
    });
  }
  return pos;
}

/** The step to show at player index i (the start at 0, clamped at the end). */
export function automatonStep(m: Automaton, i: number): AutomatonStep | undefined {
  if (m.steps.length === 0) return undefined;
  return m.steps[Math.max(0, Math.min(Math.trunc(i) || 0, m.steps.length - 1))];
}

/** What the machine is doing at step i, in words. */
export function describeAutomaton(m: Automaton, i: number): string {
  const s = automatonStep(m, i);
  if (!s) return `${m.states.length} states and ${m.edges.length} transitions.`;
  const names = s.active.map((a) => m.states[a]?.name ?? '?');
  const where = names.length ? `In ${names.join(', ')}.` : 'In no state.';
  return `${s.note ? `${s.note.charAt(0).toUpperCase()}${s.note.slice(1)}.` : ''} ${where}`.trim();
}

/** The indexes of the events after which each generation is complete:
 * a cellular automaton's steps. Empty when the events have no generations. */
export function generationEnds(events: TraceEvent[]): number[] {
  const ends: number[] = [];
  events.forEach((e, k) => {
    if (kinds[e.kind] === 'mark' && e.note === 'generation') ends.push(k + 1);
  });
  return ends;
}
