import { useEffect, useMemo } from 'react';
import type { AlgorithmApi } from '../core/api';
import type { Input, RunResult } from '../core/schema';
import { useAlgorithm } from '../hooks/use-algorithm';
import type { Player } from '../hooks/use-player';
import { RunView } from './run-view';

/**
 * Runs one entry on an input and shows it step by step. Give it the API, or
 * put it inside an <AlgorithmProvider> (or neither: it then loads
 * /algorithm.wasm and /wasm_exec.js).
 *
 *   <Visualizer entryId="quick-sort" input={{ values: [5, 3, 9, 1] }} />
 */
export function Visualizer({
  entryId,
  input,
  api: given,
  limit = 0,
  player,
  initialStep,
  onResult
}: {
  entryId: string;
  input: Input;
  api?: AlgorithmApi;
  /** Most events kept: 0 is the Go side's default. */
  limit?: number;
  player?: Player;
  initialStep?: number;
  onResult?: (r: RunResult) => void;
}) {
  const loaded = useAlgorithm();
  const api = given ?? (loaded.status === 'ready' ? loaded.api : null);
  const key = JSON.stringify(input);
  const outcome = useMemo(() => {
    if (!api) return null;
    return api.run(entryId, input, limit);
    // input is compared by its JSON so a new object with the same values does not re-run
  }, [api, entryId, key, limit]);
  const entry = useMemo(() => api?.catalog().entries.find((e) => e.id === entryId), [api, entryId]);
  useEffect(() => {
    if (outcome?.ok) onResult?.(outcome.value);
  }, [outcome]);

  if (!api) {
    return loaded.status === 'error' ? (
      <p
        className="av-root av-error"
        role="alert"
      >
        The algorithms could not be loaded: {loaded.error}
      </p>
    ) : (
      <p
        className="av-root av-muted"
        role="status"
      >
        Loading the algorithms...
      </p>
    );
  }
  if (!outcome) return null;
  if (!outcome.ok) {
    return (
      <div
        className="av-root av-error"
        role="alert"
      >
        <p style={{ margin: 0 }}>{outcome.failure.error}</p>
        {outcome.failure.problems.length ? (
          <ul>
            {outcome.failure.problems.map((p, i) => (
              <li key={i}>
                {p.line ? `line ${p.line}` : `item ${p.column ?? '?'}`}: "{p.value}" is {p.reason}
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    );
  }
  return (
    <RunView
      result={outcome.value}
      input={input}
      entry={entry}
      player={player}
      initialStep={initialStep}
    />
  );
}
