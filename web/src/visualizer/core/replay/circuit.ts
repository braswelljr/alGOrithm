/**
 * Circuits, truth tables, Karnaugh maps and timing diagrams step from their
 * output, not from events: a circuit's frames are its gate delays, a truth
 * table's steps are its rows, and a timing diagram's steps are clock ticks.
 * These helpers turn a step number into what to draw and what to say.
 */
import type { Circuit, Input, KMap, Level, LogicTable, Span } from '../schema';

export interface CircuitState {
  /** The frame shown: the number of gate delays since the inputs changed. */
  frame: number;
  levels: Level[];
  /** Gates whose output changed in this frame. */
  changed: Set<number>;
  /** The label of the set of inputs being applied, for sequences. */
  phase?: string;
  /** The gate delays since this phase's inputs were applied. */
  delay: number;
}

/** The state after step i (clamped), and which gates just changed. */
export function circuitAt(c: Circuit, i: number): CircuitState {
  const last = Math.max(0, c.frames.length - 1);
  const frame = Math.max(0, Math.min(Math.trunc(i) || 0, last));
  const levels = c.frames[frame] ?? [];
  const prev = frame > 0 ? c.frames[frame - 1] : undefined;
  const changed = new Set<number>();
  levels.forEach((l, k) => {
    if (!prev || prev[k] !== l) changed.add(k);
  });
  let phase: string | undefined;
  let start = 0;
  for (const p of c.phases) {
    if (p.frame <= frame) {
      phase = p.label;
      start = p.frame;
    }
  }
  return { frame, levels, changed, phase, delay: frame - start };
}

const levelWord = (l: Level) => (l === 1 ? '1' : l === 0 ? '0' : 'unknown');

/** Says what happened at the step, in words. */
export function describeCircuit(c: Circuit, s: CircuitState): string {
  const inputs = c.gates
    .map((g, k) => (g.kind === 'input' ? `${g.name}=${levelWord(s.levels[k] ?? -1)}` : ''))
    .filter(Boolean)
    .join(' ');
  if (s.delay === 0) {
    return `${s.phase ? `${s.phase}: ` : ''}the inputs are applied (${inputs}); nothing has passed through a gate yet.`;
  }
  const moved = [...s.changed]
    .filter((k) => c.gates[k]?.kind !== 'input')
    .map((k) => `${c.gates[k].name} (${c.gates[k].kind}) became ${levelWord(s.levels[k])}`);
  const outputs = c.gates
    .map((g, k) => (g.kind === 'output' ? `${g.name}=${levelWord(s.levels[k] ?? -1)}` : ''))
    .filter(Boolean)
    .join(' ');
  const head = `After ${s.delay} gate delay${s.delay === 1 ? '' : 's'}`;
  if (!moved.length)
    return `${head}: nothing changes, the circuit has settled. Outputs: ${outputs}.`;
  return `${head}: ${moved.join('; ')}. Outputs: ${outputs}.`;
}

/** The input the run would take with input gate `gate` flipped, or undefined
 * when the run cannot be edited that way (a sequence of inputs). */
export function toggledInput(
  c: Circuit,
  input: Input,
  gate: number,
  levels: Level[]
): Input | undefined {
  const g = c.gates[gate];
  if (!g || g.kind !== 'input' || c.phases.length > 1) return undefined;
  const inputs = c.gates.flatMap((x, k) => (x.kind === 'input' ? [k] : []));
  const values = input.values ?? [];
  const bus = /^([A-Z])(\d+)$/.exec(g.name);
  // Two numbers A and B, bit by bit (adders, comparators).
  if (bus && values.length === 2 && (bus[1] === 'A' || bus[1] === 'B')) {
    const which = bus[1] === 'A' ? 0 : 1;
    const next = [...values];
    next[which] = values[which] ^ (1 << Number(bus[2]));
    return { ...input, values: next };
  }
  // One bit per input, in order.
  if (values.length === inputs.length) {
    const at = inputs.indexOf(gate);
    const next = [...values];
    next[at] = values[at] ? 0 : 1;
    return { ...input, values: next };
  }
  // An expression: the inputs set in the key.
  if (input.text !== undefined) {
    const key = inputs
      .map((k) => {
        const on = k === gate ? levels[k] !== 1 : levels[k] === 1;
        return `${c.gates[k].name}=${on ? 1 : 0}`;
      })
      .join(' ');
    return { ...input, key };
  }
  return undefined;
}

/** The row's bits for variable k of n, the first variable most significant. */
export function rowBit(row: number, k: number, n: number): number {
  return (row >> (n - 1 - k)) & 1;
}

/** A truth table is stepped row by row, then a Karnaugh map group by group. */
export function logicSteps(t: LogicTable): number {
  return t.kmap ? t.out.length + t.kmap.groups.length : t.out.length;
}

export function describeLogic(t: LogicTable, step: number): string {
  const n = t.vars.length;
  if (step === 0) {
    return t.kmap
      ? `A Karnaugh map of ${n} variables: rows ${t.kmap.rowVars}, columns ${t.kmap.colVars}, in Gray code order.`
      : `${t.out.length} rows: one for every combination of ${t.vars.join(', ')}.`;
  }
  if (step <= t.out.length) {
    const row = step - 1;
    const vals = t.vars.map((v, k) => `${v}=${rowBit(row, k, n)}`).join(' ');
    const out = t.out[row] === 2 ? 'does not matter' : t.out[row];
    return `Row ${row}: ${vals} gives ${out}.`;
  }
  const g = step - t.out.length - 1;
  const k = t.kmap!;
  return `Group ${g + 1} of ${k.groups.length}: ${k.groups[g].length} cells, the term ${k.terms[g]}.${
    g === k.groups.length - 1 ? ` Together: ${k.sop}.` : ''
  }`;
}

/** The groups shown by step: none until the rows are done, then one more each step. */
export function groupsShown(k: KMap | undefined, rows: number, step: number): number {
  if (!k) return 0;
  return Math.max(0, Math.min(k.groups.length, step - rows));
}

export interface Wave {
  name: string;
  /** The level at each tick. */
  levels: boolean[];
}

/** Whether a timeline is a timing diagram: every span is high or low. */
export function isWaveform(spans: Span[]): boolean {
  return spans.length > 0 && spans.every((s) => s.state === 'high' || s.state === 'low');
}

/** Rebuilds each signal's level per tick from the spans. */
export function wavesOf(spans: Span[]): { waves: Wave[]; ticks: number } {
  const order: string[] = [];
  const byLane = new Map<string, boolean[]>();
  let ticks = 0;
  for (const s of spans) {
    if (!byLane.has(s.lane)) {
      byLane.set(s.lane, []);
      order.push(s.lane);
    }
    const l = byLane.get(s.lane)!;
    for (let t = s.start; t < s.end; t++) l[t] = s.state === 'high';
    ticks = Math.max(ticks, s.end);
  }
  return { waves: order.map((name) => ({ name, levels: byLane.get(name)! })), ticks };
}

export function describeWave(waves: Wave[], t: number): string {
  if (t === 0) return 'The timing diagram: time runs to the right, one clock cycle is two ticks.';
  const tick = t - 1;
  const clk = waves.find((w) => w.name === 'CLK');
  const edge = clk && clk.levels[tick] && !clk.levels[tick - 1];
  const vals = waves
    .filter((w) => w.name !== 'CLK')
    .map((w) => `${w.name}=${w.levels[tick] ? 1 : 0}`)
    .join(' ');
  return `Tick ${tick}${edge ? ' (rising clock edge)' : ''}: ${vals}.`;
}
