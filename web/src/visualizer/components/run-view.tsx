import { useEffect, useMemo, useRef } from 'react';
import type { Entry, Input, RunResult } from '../core/schema';
import { usePlayer, type Player } from '../hooks/use-player';
import { CountsList } from './counts-list';
import { GraphFlow } from './lazy-flow';
import { LinesView } from './lines-view';
import { PlayerControls } from './player-controls';
import { ResultSummary } from './result-summary';
import { frameFor, runKind } from './run-frames';
import { StepDescription } from './step-description';
import { SummaryList } from './summary-list';
import { TableView } from './table-view';

/** What the output holds besides what the main view draws. */
function Extras({ result, kind }: { result: RunResult; kind: ReturnType<typeof runKind> }) {
  const o = result.output;
  return (
    <>
      {o.summary ? <SummaryList summary={o.summary} /> : null}
      {o.graph && kind !== 'graph' ? <GraphFlow graph={o.graph} /> : null}
      {o.table && kind !== 'table' && kind !== 'board' ? <TableView table={o.table} /> : null}
      {o.lines.length && kind !== 'lines' && kind !== 'hanoi' ? (
        <LinesView lines={o.lines} />
      ) : null}
    </>
  );
}

/**
 * Draws a run with the view that fits it and a player over its steps. It
 * never re-runs anything, so it replays a saved or imported result as well as
 * a fresh one. Pass `player` to control it from outside (the site keeps the
 * step in the URL); otherwise it makes its own. Pass `onInputChange` to let
 * the reader edit the input from the view itself (flipping a circuit's input
 * switches); the caller runs the new input and passes the new result. */
export function RunView({
  result,
  input,
  entry,
  player: external,
  initialStep = 0,
  label,
  onInputChange
}: {
  result: RunResult;
  input: Input;
  entry?: Pick<Entry, 'category' | 'id' | 'name'>;
  player?: Player;
  initialStep?: number;
  label?: string;
  onInputChange?: (input: Input) => void;
}) {
  const own = usePlayer();
  const player = external ?? own;
  const kind = runKind(result, entry?.category);
  const frame = useMemo(
    () => frameFor(kind, result, input, onInputChange),
    [kind, result, input, onInputChange]
  );
  // Load once per run, not per frame: the frame is rebuilt whenever the input
  // or the callback is a new value, and reloading then would reset the step
  // (or, with a parent that re-renders on every load, never settle).
  const { load } = player;
  const { total } = frame;
  const loadedFor = useRef<unknown>(null);
  useEffect(() => {
    if (loadedFor.current === result) return;
    loadedFor.current = result;
    load(total, initialStep);
  }, [result, total, load]);

  const { view, words, counts } = frame.render(player.state.index, player.quiet);
  const animated = frame.total > 0;
  return (
    <section
      className="av-root av-stack"
      aria-label={label ?? (entry ? `${entry.name}, step by step` : 'The run, step by step')}
    >
      {view ? (
        <div
          className="av-stage"
          tabIndex={animated ? 0 : undefined}
          onKeyDown={animated ? player.onKeyDown : undefined}
          data-tour="run-view"
        >
          {view}
        </div>
      ) : null}
      {animated ? (
        <>
          <StepDescription
            text={words}
            busy={player.state.status === 'playing'}
          />
          <div data-tour="player-controls">
            <PlayerControls
              player={player}
              label={frame.unit}
            />
          </div>
          {frame.legend}
          <CountsList
            counts={counts}
            nanos={result.nanos}
          />
        </>
      ) : null}
      {!animated || player.state.status === 'finished' ? (
        <ResultSummary
          entry={entry}
          input={input}
          output={result.output}
        />
      ) : null}
      <Extras
        result={result}
        kind={kind}
      />
      {result.truncated ? (
        <p className="av-muted">
          Only the first {result.events.length} steps were kept; the counts cover the whole run.
        </p>
      ) : null}
    </section>
  );
}
