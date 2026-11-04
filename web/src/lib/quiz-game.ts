import { frameFor, runKind, type Frame } from '@/visualizer';
import {
  questionsFor,
  type AlgorithmApi,
  type Catalog,
  type Entry,
  type Input,
  type Question,
  type RunResult
} from '@/visualizer/core';
import { candidates, poolSize, seeded, shuffled, type GameConfig } from './quiz-seed';

/** One question of a game, with the run it is about, ready to draw. */
export interface GameItem {
  entry: Entry;
  input: Input;
  result: RunResult;
  frame: Frame;
  question: Question;
}

/** The most entries looked at to fill a game, so a narrow filter stays fast. */
const MAX_ENTRIES = 80;
/** Events kept per run: enough for any example. */
const EVENT_LIMIT = 200_000;

/**
 * The questions of a game: each one from the real run of an entry's example,
 * paused where the visualizer can show it. The same config (seed included)
 * gives the same questions in the same order.
 */
export function buildGame(api: AlgorithmApi, catalog: Catalog, c: GameConfig): GameItem[] {
  const need = poolSize(c);
  const perEntry = c.entry ? need : 2;
  const items: GameItem[] = [];
  const entries = candidates(catalog, c).slice(0, MAX_ENTRIES);
  for (const [i, entry] of entries.entries()) {
    if (items.length >= need) break;
    const input = entry.example;
    const run = api.run(entry.id, input, EVENT_LIMIT);
    if (!run.ok) continue;
    const frame = frameFor(runKind(run.value, entry.category), run.value, input);
    const asked = api.quiz(entry.id, input, perEntry, c.seed + i);
    if (!asked.ok) continue;
    for (const question of questionsFor(asked.value, frame.unit, frame.total)) {
      items.push({ entry, input, result: run.value, frame, question });
    }
  }
  // Mix entries together, except in a game about one entry, which follows its run.
  const ordered = c.entry ? items : shuffled(items, seeded(c.seed ^ 0x5eed));
  return ordered.slice(0, need);
}

/** Where "replay this step" opens the visualizer. */
export function replayHref(item: Pick<GameItem, 'entry' | 'input' | 'question'>): string {
  const q = new URLSearchParams({
    id: item.entry.id,
    in: JSON.stringify(item.input),
    step: String(item.question.step)
  });
  const page = item.entry.input === 'ops' ? '/structures/' : '/visualize/';
  return `${page}?${q.toString()}`;
}
