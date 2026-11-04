/**
 * The frames of digital logic runs: a circuit (one step per gate delay), a
 * truth table or Karnaugh map (one per row, then one per group), a timing
 * diagram (one per clock tick) and rows of bits (one per row). RunView picks
 * them through run-frames.
 */
import type { ReactNode } from 'react';
import {
  circuitAt,
  describeCircuit,
  describeLogic,
  describeWave,
  groupsShown,
  isWaveform,
  logicSteps,
  toggledInput,
  wavesOf
} from '../core/replay/circuit';
import type { Input, RunResult } from '../core/schema';
import { BitsView } from './bits-view';
import { CircuitView } from './lazy-flow';
import { KMapView, TruthTableView } from './logic-view';
import { WaveformView } from './waveform-view';

interface LogicFrame {
  total: number;
  unit: string;
  render(
    index: number,
    quiet: boolean
  ): { view: ReactNode; words: string; counts: Record<string, number> };
}

/** Whether the run is drawn by one of these frames, and which. */
export function logicKind(
  result: RunResult
): 'circuit' | 'waveform' | 'logic' | 'bits' | undefined {
  const o = result.output;
  if (o.circuit) return 'circuit';
  if (isWaveform(o.timeline)) return 'waveform';
  if (o.logic) return 'logic';
  if (o.bits.length) return 'bits';
  return undefined;
}

/** A circuit, one step per gate delay. With onInputChange, the reader can
 * flip an input and the run is redone on the new input. */
export function circuitFrame(
  r: RunResult,
  input: Input,
  onInputChange?: (input: Input) => void
): LogicFrame {
  const c = r.output.circuit!;
  return {
    total: Math.max(0, c.frames.length - 1),
    unit: 'Gate delay',
    render(i, quiet) {
      const s = circuitAt(c, i);
      const onToggle = onInputChange
        ? (gate: number) => {
            const next = toggledInput(c, input, gate, s.levels);
            if (next) onInputChange(next);
          }
        : undefined;
      return {
        view: (
          <CircuitView
            circuit={c}
            state={s}
            onToggle={c.phases.length > 1 ? undefined : onToggle}
            quiet={quiet}
          />
        ),
        words: describeCircuit(c, s),
        counts: {}
      };
    }
  };
}

export function logicFrame(r: RunResult): LogicFrame {
  const t = r.output.logic!;
  const k = t.kmap;
  const total = logicSteps(t);
  return {
    total,
    unit: k ? 'Step' : 'Row',
    render(i) {
      const step = Math.max(0, Math.min(i, total));
      const shown = groupsShown(k, t.out.length, step);
      return {
        view: (
          <div className="flex flex-col gap-5">
            {k ? (
              <KMapView
                kmap={k}
                shown={shown}
                current={shown > 0 ? shown - 1 : undefined}
              />
            ) : null}
            <TruthTableView
              table={t}
              upTo={Math.min(step, t.out.length)}
            />
          </div>
        ),
        words: describeLogic(t, step),
        counts: {}
      };
    }
  };
}

export function waveformFrame(r: RunResult): LogicFrame {
  const { waves, ticks } = wavesOf(r.output.timeline);
  return {
    total: ticks,
    unit: 'Tick',
    render(i) {
      const t = Math.max(0, Math.min(i, ticks));
      return {
        view: (
          <WaveformView
            waves={waves}
            ticks={ticks}
            upTo={t}
          />
        ),
        words: describeWave(waves, t),
        counts: {}
      };
    }
  };
}

export function bitsFrame(r: RunResult): LogicFrame {
  const rows = r.output.bits;
  return {
    total: rows.length,
    unit: 'Row',
    render(i) {
      const n = Math.max(0, Math.min(i, rows.length));
      const row = rows[n - 1];
      return {
        view: (
          <BitsView
            rows={rows}
            upTo={n}
          />
        ),
        words: row
          ? `${row.label}: ${row.fields.map((f) => `${f.name} ${f.bits}`).join(', ')}${row.note ? ` (${row.note})` : ''}.`
          : `${rows.length} rows of bits.`,
        counts: {}
      };
    }
  };
}
