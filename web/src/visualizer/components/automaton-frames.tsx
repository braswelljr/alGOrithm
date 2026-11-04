/**
 * The frames of automata (one per move of the machine) and of cellular
 * automata (one per generation). RunView picks them through run-frames.
 */
import { automatonStep, describeAutomaton, generationEnds } from '../core/replay/automaton';
import { TableReplayer } from '../core/replay/table';
import type { Automaton, RunResult } from '../core/schema';
import { CellBoardView } from './cell-board-view';
import { InputStrip, StateMachineView } from './lazy-flow';
import { StackView } from './stack-view';
import { TableView } from './table-view';
import { TapeView } from './tape-view';

interface AutomatonFrame {
  total: number;
  unit: string;
  render(
    index: number,
    quiet: boolean
  ): { view: React.ReactNode; words: string; counts: Record<string, number> };
}

/** The verdict once the run is over, in words and as a badge. */
function Verdict({ machine, done }: { machine: Automaton; done: boolean }) {
  if (!done || machine.accepted === undefined) return null;
  return (
    <p
      className="data-[ok=false]:border-danger data-[ok=false]:text-danger data-[ok=true]:border-found data-[ok=true]:text-found m-0 inline-flex items-center gap-1.5 self-start rounded-full border px-3 py-1 text-sm font-semibold"
      data-ok={machine.accepted}
    >
      <span aria-hidden>{machine.accepted ? '✓' : '✕'}</span>
      {machine.accepted ? 'Accepted' : 'Rejected'}
    </p>
  );
}

/** An automaton's run: the state diagram, the input with its read head,
 * and the tape, the stack or the output when the machine has one. */
export function automatonFrame(r: RunResult): AutomatonFrame {
  const m = r.output.machine!;
  const total = Math.max(0, m.steps.length - 1);
  return {
    total,
    unit: 'Move',
    render(i, quiet) {
      const s = automatonStep(m, i);
      const done = i >= total;
      const state = s?.active.length === 1 ? m.states[s.active[0]!]?.name : undefined;
      return {
        view: (
          <div className="flex flex-col gap-4">
            <StateMachineView
              machine={m}
              step={s}
              stepIndex={i}
              quiet={quiet}
            />
            <div className="flex flex-wrap items-end gap-6">
              {m.input ? (
                <InputStrip
                  input={m.input}
                  head={s?.head ?? 0}
                  quiet={quiet}
                />
              ) : null}
              {s?.stack ? (
                <StackView
                  stack={s.stack}
                  quiet={quiet}
                />
              ) : null}
              {s?.output ? (
                <p className="m-0 flex flex-col gap-1 text-sm">
                  <span className="text-muted text-[11px] font-medium tracking-wide uppercase">
                    Output
                  </span>
                  <span className="font-mono">{s.output}</span>
                </p>
              ) : null}
              <Verdict
                machine={m}
                done={done}
              />
            </div>
            {s?.tape ? (
              <TapeView
                tape={s.tape}
                head={s.tapeHead}
                state={state}
                quiet={quiet}
              />
            ) : null}
            {r.output.table ? <TableView table={r.output.table} /> : null}
          </div>
        ),
        words: `Move ${Math.min(i, total)} of ${total}: ${describeAutomaton(m, i)}`,
        counts: { moves: Math.min(i, total) }
      };
    }
  };
}

/** Whether a run is a cellular automaton played by generation. */
export function isCellular(r: RunResult): boolean {
  return Boolean(r.output.table) && generationEnds(r.events).length > 0;
}

/** A cellular automaton, one generation a step: Game of Life's board, or an
 * elementary automaton's rows filling in from the top. */
export function cellularFrame(r: RunResult): AutomatonFrame {
  const t = r.output.table!;
  const rep = new TableReplayer(t.cells.length, r.events);
  const ends = generationEnds(r.events);
  const total = Math.max(0, ends.length - 1);
  return {
    total,
    unit: 'Generation',
    render(i, quiet) {
      const g = Math.max(0, Math.min(Math.trunc(i) || 0, total));
      const s = rep.stateAt(ends[g] ?? 0);
      const alive = s.cells.filter((c) => c === 1).length;
      return {
        view: (
          <CellBoardView
            rows={t.rows}
            cols={t.cols}
            cells={s.cells}
            generation={g}
            quiet={quiet}
          />
        ),
        words: `Generation ${g} of ${total}: ${alive} live cells.`,
        counts: { alive }
      };
    }
  };
}
