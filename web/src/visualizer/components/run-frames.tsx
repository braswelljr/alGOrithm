/**
 * One frame builder per kind of run: how many steps it has, and how to draw
 * step i with its words and counts. RunView picks the builder; each builder
 * hands its view only the small typed props it needs.
 */
import type { ReactNode } from 'react';
import { ArrayReplayer } from '../core/replay/array';
import {
  describeArrayEvent,
  describeGraphEvent,
  describeGridEvent,
  describePointsEvent,
  describeTableEvent,
  describeTextEvent
} from '../core/replay/describe';
import { GraphReplayer } from '../core/replay/graph';
import { GridReplayer } from '../core/replay/grid';
import { disksFor, hanoiAt, parseMoves, pegName } from '../core/replay/hanoi';
import { PointsReplayer } from '../core/replay/points';
import { TableReplayer } from '../core/replay/table';
import { TextReplayer } from '../core/replay/text';
import { chosenAt, describeTime, lanesOf, spansUntil, timelineEnd } from '../core/replay/timeline';
import type { Input, RunResult } from '../core/schema';
import { architectureFrame } from './architecture-frames';
import { automatonFrame, cellularFrame, isCellular } from './automaton-frames';
import { BarsView } from './bars-view';
import { BoardView } from './board-view';
import { CellsView } from './cells-view';
import { execFrame } from './exec-frames';
import { GridView } from './grid-view';
import { HanoiView } from './hanoi-view';
import { GraphFlow, StructureFlow } from './lazy-flow';
import { Legend } from './legend';
import { LinesView } from './lines-view';
import { bitsFrame, circuitFrame, logicFrame, logicKind, waveformFrame } from './logic-frames';
import { PointsView } from './points-view';
import { StructureLog } from './structure-log';
import { TableView } from './table-view';
import { TextMatchView } from './text-match-view';
import { TimelineView } from './timeline-view';
import { TrackView } from './track-view';

/** How a run is drawn. */
export type RunKind =
  | 'automaton'
  | 'circuit'
  | 'waveform'
  | 'logic'
  | 'bits'
  | 'cellular'
  | 'timeline'
  | 'track'
  | 'grid'
  | 'graph'
  | 'points'
  | 'board'
  | 'table'
  | 'text'
  | 'structure'
  | 'hanoi'
  | 'bars'
  | 'cells'
  | 'lines'
  | 'exec'
  | 'architecture'
  | 'none';

/** Picks the view for a run from what its output holds. A square table from
 * a backtracking entry is a board (queens, sudoku). */
export function runKind(result: RunResult, category?: string): RunKind {
  const o = result.output;
  if (o.architecture) return 'architecture';
  if (o.exec) return 'exec';
  if (o.machine) return 'automaton';
  if (isCellular(result)) return 'cellular';
  const logic = logicKind(result);
  if (logic) return logic;
  if (o.timeline.length) return 'timeline';
  if (o.track) return 'track';
  if (o.grid) return 'grid';
  // A nodes-by-nodes table (all-pairs distances) is what the events fill;
  // the graph is then drawn beside it, still.
  const matrix = o.graph && o.table && o.table.cells.length === o.graph.nodes.length ** 2;
  if (o.graph && !matrix) return 'graph';
  if (o.points.length) return 'points';
  if (o.table && category === 'backtracking' && o.table.rows === o.table.cols) return 'board';
  if (o.table && result.events.length) return 'table';
  if (o.matches) return 'text';
  if (o.steps.length) return 'structure';
  if (o.lines.length && parseMoves(o.lines).length === o.lines.length) return 'hanoi';
  if (o.initial.length && result.events.length) return o.index !== undefined ? 'cells' : 'bars';
  if (o.table) return 'table';
  if (o.lines.length) return 'lines';
  return 'none';
}

export interface Frame {
  total: number;
  /** Labels the scrubber: "Step", "Time", "Operation", "Move". */
  unit: string;
  render(
    index: number,
    quiet: boolean
  ): { view: ReactNode; words: string; counts: Record<string, number> };
  legend?: ReactNode;
}

const of = (i: number, total: number) => `${Math.min(i, total)} of ${total}`;

function arrayFrame(kind: 'bars' | 'cells', r: RunResult): Frame {
  const rep = new ArrayReplayer(r.output.initial, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i, quiet) {
      const s = rep.stateAt(i);
      return {
        view:
          kind === 'bars' ? (
            <BarsView
              state={s}
              quiet={quiet}
            />
          ) : (
            <CellsView
              state={s}
              quiet={quiet}
            />
          ),
        words: `Step ${of(s.step, rep.length)}: ${describeArrayEvent(rep.lastEvent(s.step), s.values)}`,
        counts: s.counts
      };
    },
    legend:
      kind === 'bars' ? (
        <Legend marks={['compare', 'swap', 'write', 'pivot', 'min', 'sorted', 'found']} />
      ) : (
        <Legend marks={['compare', 'range', 'found']} />
      )
  };
}

function gridFrame(r: RunResult): Frame {
  const g = r.output.grid!;
  const rep = new GridReplayer(g.cost.length, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <GridView
            grid={g}
            state={s}
          />
        ),
        words: `Step ${of(i, rep.length)}: ${describeGridEvent(rep.lastEvent(i), g.cols)}`,
        counts: s.counts
      };
    }
  };
}

function graphFrame(r: RunResult): Frame {
  const g = r.output.graph!;
  const rep = new GraphReplayer(g.nodes.length, g.edges.length, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <GraphFlow
            graph={g}
            state={s}
          />
        ),
        words: `Step ${of(i, rep.length)}: ${describeGraphEvent(rep.lastEvent(i), g.nodes, g.edges)}`,
        counts: s.counts
      };
    }
  };
}

function pointsFrame(r: RunResult): Frame {
  const pts = r.output.points;
  const rep = new PointsReplayer(pts.length, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <PointsView
            points={pts}
            state={s}
            segments={r.output.segments}
          />
        ),
        words: `Step ${of(i, rep.length)}: ${describePointsEvent(rep.lastEvent(i), pts)}`,
        counts: s.counts
      };
    }
  };
}

function boardFrame(r: RunResult): Frame {
  const t = r.output.table!;
  const rep = new TableReplayer(t.cells.length, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <BoardView
            table={t}
            state={rep.length ? s : undefined}
          />
        ),
        words: rep.length
          ? `Step ${of(i, rep.length)}: ${describeTableEvent(rep.lastEvent(i), t.cols, t.rowLabels, t.colLabels)}`
          : 'The board.',
        counts: s.counts
      };
    }
  };
}

function tableFrame(r: RunResult): Frame {
  const t = r.output.table!;
  const rep = new TableReplayer(t.cells.length, r.events);
  if (rep.length < 1) {
    return {
      total: 0,
      unit: 'Step',
      render: () => ({ view: <TableView table={t} />, words: '', counts: {} })
    };
  }
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <TableView
            table={t}
            state={s}
          />
        ),
        words: `Step ${of(i, rep.length)}: ${describeTableEvent(rep.lastEvent(i), t.cols, t.rowLabels, t.colLabels)}`,
        counts: s.counts
      };
    }
  };
}

function textFrame(r: RunResult, input: Input): Frame {
  const text = r.output.text ?? input.text ?? '';
  const pattern = input.other ?? '';
  const rep = new TextReplayer(text.length, r.events);
  return {
    total: rep.length,
    unit: 'Step',
    render(i) {
      const s = rep.stateAt(i);
      return {
        view: (
          <TextMatchView
            text={text}
            pattern={pattern}
            state={s}
          />
        ),
        words: `Step ${of(i, rep.length)}: ${describeTextEvent(rep.lastEvent(i), text, pattern)}`,
        counts: s.counts
      };
    }
  };
}

function structureFrame(r: RunResult): Frame {
  const steps = r.output.steps;
  const last = Math.max(0, steps.length - 1);
  return {
    total: last,
    unit: 'Operation',
    render(i) {
      const k = Math.max(0, Math.min(i, last));
      const st = steps[k];
      const op = `${st.op.name}${st.op.value ? ` ${st.op.value}` : ''}`;
      return {
        view: (
          <div className="av-stack">
            <StructureFlow
              snapshot={st.snapshot}
              previous={k > 0 ? steps[k - 1].snapshot : undefined}
            />
            <StructureLog
              steps={steps}
              current={k}
            />
          </div>
        ),
        words:
          k === 0
            ? `The start: ${st.result ?? ''}`
            : `Operation ${of(k, last)}, ${op}: ${st.error ? `error, ${st.error}` : (st.result ?? 'done')}.`,
        counts: {}
      };
    }
  };
}

function hanoiFrame(r: RunResult): Frame {
  const moves = parseMoves(r.output.lines);
  const disks = disksFor(moves.length);
  return {
    total: moves.length,
    unit: 'Move',
    render(i) {
      const s = hanoiAt(disks, moves, i);
      return {
        view: (
          <HanoiView
            state={s}
            disks={disks}
          />
        ),
        words: s.move
          ? `Move ${of(i, moves.length)}: the top disk from ${pegName(s.move.from)} to ${pegName(s.move.to)}.`
          : `${disks} disks on peg A.`,
        counts: {}
      };
    }
  };
}

function timelineFrame(r: RunResult): Frame {
  const spans = r.output.timeline;
  const end = timelineEnd(spans);
  if (r.events.length) {
    // One span per item, revealed as the item is chosen (greedy schedules).
    const lanes = lanesOf(spans);
    const total = r.events.length;
    return {
      total,
      unit: 'Step',
      render(i) {
        const { chosen, compared } = chosenAt(r.events, i);
        const focus = compared !== null ? spans[compared] : undefined;
        return {
          view: (
            <TimelineView
              spans={spans.filter((_, k) => chosen.has(k))}
              end={end}
              lanes={lanes}
              focus={focus}
            />
          ),
          words:
            i === 0
              ? 'Nothing chosen yet.'
              : `Step ${of(i, total)}: ${chosen.size} chosen${focus ? `; looking at ${focus.label || focus.lane}` : ''}.`,
          counts: {}
        };
      }
    };
  }
  return {
    total: end,
    unit: 'Time',
    render(t) {
      return {
        view: (
          <TimelineView
            spans={spansUntil(spans, t)}
            end={end}
            now={t}
          />
        ),
        words: describeTime(spans, t),
        counts: {}
      };
    }
  };
}

function trackFrame(r: RunResult): Frame {
  const track = r.output.track!;
  const last = Math.max(0, track.path.length - 1);
  return {
    total: last,
    unit: 'Seek',
    render(i) {
      const k = Math.min(i, last);
      return {
        view: (
          <TrackView
            track={track}
            upTo={k}
          />
        ),
        words:
          k === 0
            ? `The head starts at track ${track.path[0]}.`
            : `Seek ${of(k, last)}: from track ${track.path[k - 1]} to ${track.path[k]}, ${Math.abs(track.path[k] - track.path[k - 1])} tracks.`,
        counts: {}
      };
    }
  };
}

function linesFrame(r: RunResult): Frame {
  const lines = r.output.lines;
  return {
    total: lines.length,
    unit: 'Line',
    render(i) {
      return {
        view: (
          <LinesView
            lines={lines}
            upTo={i}
          />
        ),
        words: lines[i - 1] ?? 'The start.',
        counts: {}
      };
    }
  };
}

export function frameFor(
  kind: RunKind,
  result: RunResult,
  input: Input,
  onInputChange?: (input: Input) => void
): Frame {
  switch (kind) {
    case 'circuit':
      return circuitFrame(result, input, onInputChange);
    case 'waveform':
      return waveformFrame(result);
    case 'logic':
      return logicFrame(result);
    case 'bits':
      return bitsFrame(result);
    case 'automaton':
      return automatonFrame(result);
    case 'cellular':
      return cellularFrame(result);
    case 'bars':
    case 'cells':
      return arrayFrame(kind, result);
    case 'grid':
      return gridFrame(result);
    case 'graph':
      return graphFrame(result);
    case 'points':
      return pointsFrame(result);
    case 'board':
      return boardFrame(result);
    case 'table':
      return tableFrame(result);
    case 'text':
      return textFrame(result, input);
    case 'structure':
      return structureFrame(result);
    case 'hanoi':
      return hanoiFrame(result);
    case 'timeline':
      return timelineFrame(result);
    case 'track':
      return trackFrame(result);
    case 'lines':
      return linesFrame(result);
    case 'exec':
      return execFrame(result);
    case 'architecture':
      return architectureFrame(result);
    case 'none':
      return { total: 0, unit: 'Step', render: () => ({ view: null, words: '', counts: {} }) };
  }
}
