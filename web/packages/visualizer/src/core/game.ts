/**
 * The quiz game: a pure reducer over a list of questions, so every rule and
 * every transition is tested without a browser. It knows each question's
 * right choice and how many choices it has, never how anything looks; the
 * clock is the caller's, which reports time taken with ANSWER and the end of
 * a timer with TIMEOUT.
 *
 *          START              ANSWER or TIMEOUT (question)
 *   intro --------> question ---------------------------> answered
 *     ^                ^                                    |  | REVEAL
 *     |                |  NEXT (more to ask)                |  v
 *     |                +---------------------------- revealing
 *     |                                                     |
 *     |  RESTART (any)         NEXT (done, out of lives) or |
 *     +-------------------- finished <----------------------+
 *                              ^   TIMEOUT (game), time attack only
 *
 * An event that is not legal in the current state returns the same state.
 */

export type GameMode = 'classic' | 'time' | 'streak';
export type GameStatus = 'intro' | 'question' | 'answered' | 'revealing' | 'finished';

export interface GameRules {
  /** How many questions at most; null for as many as there are. */
  questions: number | null;
  /** Lives at the start; null for none (a miss costs only the streak). */
  lives: number | null;
  /** The whole game's time limit, in ms; null for none. */
  gameMs: number | null;
  /** Each question's time limit, in ms; null for none. */
  questionMs: number | null;
}

/** Classic: 10 questions, 3 lives. Time attack: as many as you can in 60 s.
 * Streak: until the first miss. */
export const GAME_RULES: Readonly<Record<GameMode, GameRules>> = {
  classic: { questions: 10, lives: 3, gameMs: null, questionMs: 20_000 },
  time: { questions: null, lives: null, gameMs: 60_000, questionMs: null },
  streak: { questions: null, lives: 1, gameMs: null, questionMs: 15_000 }
};

/** Points for a right answer before the multiplier. */
export const BASE_POINTS = 100;
/** The most a fast answer adds. */
export const TIME_BONUS = 50;
/** In time attack, an answer within this many ms earns some time bonus. */
export const SOFT_LIMIT_MS = 10_000;
/** A streak that is a multiple of this is a milestone worth celebrating. */
export const MILESTONE = 5;

/** The streak multiplier: x1 for streaks 1-3, x2 for 4-6, x3 for 7-9, x4 after. */
export function multiplier(streak: number): number {
  return Math.min(4, 1 + Math.floor(Math.max(0, streak - 1) / 3));
}

/** Points for a right answer: base x multiplier(streak) + a time bonus for
 * the share of limitMs left. streak counts this answer. */
export function points(streak: number, msTaken: number, limitMs: number): number {
  const left = limitMs > 0 ? Math.min(1, Math.max(0, (limitMs - msTaken) / limitMs)) : 0;
  return BASE_POINTS * multiplier(streak) + Math.round(TIME_BONUS * left);
}

export interface GameAnswer {
  /** The choice picked, or null when the timer ran out. */
  pick: number | null;
  right: boolean;
  points: number;
  /** Time taken, in ms. */
  ms: number;
}

export interface GameState {
  status: GameStatus;
  mode: GameMode;
  rules: GameRules;
  /** Each question's right choice and number of choices. */
  answers: number[];
  choices: number[];
  /** The question being asked or just answered. */
  index: number;
  score: number;
  streak: number;
  bestStreak: number;
  /** Lives left, or null in a mode without lives. */
  lives: number | null;
  results: GameAnswer[];
  /** Time spent answering, in ms. */
  elapsedMs: number;
  /** Whether the last answer reached a streak milestone. */
  milestone: boolean;
  /** Why the game ended. */
  ended: 'done' | 'lives' | 'time' | null;
}

export type GameEvent =
  | {
      type: 'START';
      mode: GameMode;
      answers: number[];
      choices: number[];
      /** Fewer questions than the mode's (a 5-question game for one entry). */
      questions?: number;
    }
  | { type: 'ANSWER'; choice: number; ms: number }
  | { type: 'TIMEOUT'; scope: 'question' | 'game' }
  | { type: 'REVEAL' }
  | { type: 'NEXT' }
  | { type: 'RESTART' };

export function initialGameState(mode: GameMode = 'classic'): GameState {
  const rules = GAME_RULES[mode];
  return {
    status: 'intro',
    mode,
    rules,
    answers: [],
    choices: [],
    index: 0,
    score: 0,
    streak: 0,
    bestStreak: 0,
    lives: rules.lives,
    results: [],
    elapsedMs: 0,
    milestone: false,
    ended: null
  };
}

/** How many questions this game asks. */
export function questionCount(s: GameState): number {
  return Math.min(s.rules.questions ?? Infinity, s.answers.length, s.choices.length);
}

/** Records an answer (or a timeout, with pick null). */
function record(s: GameState, pick: number | null, ms: number): GameState {
  const right = pick !== null && pick === s.answers[s.index];
  const streak = right ? s.streak + 1 : 0;
  const limit = s.rules.questionMs ?? SOFT_LIMIT_MS;
  const got = right ? points(streak, ms, limit) : 0;
  const taken = Math.max(0, Number.isFinite(ms) ? ms : 0);
  return {
    ...s,
    status: 'answered',
    score: s.score + got,
    streak,
    bestStreak: Math.max(s.bestStreak, streak),
    lives: s.lives !== null && !right ? s.lives - 1 : s.lives,
    results: [...s.results, { pick, right, points: got, ms: taken }],
    elapsedMs: s.elapsedMs + taken,
    milestone: right && streak % MILESTONE === 0
  };
}

export function gameReducer(s: GameState, e: GameEvent): GameState {
  switch (e.type) {
    case 'START': {
      const fresh = initialGameState(e.mode);
      const n = Math.min(e.answers.length, e.choices.length);
      if (n === 0) return s;
      const rules =
        e.questions !== undefined && e.questions > 0
          ? { ...fresh.rules, questions: Math.min(fresh.rules.questions ?? Infinity, e.questions) }
          : fresh.rules;
      return {
        ...fresh,
        rules,
        status: 'question',
        answers: e.answers.slice(0, n),
        choices: e.choices.slice(0, n)
      };
    }

    case 'ANSWER': {
      if (s.status !== 'question') return s;
      const c = Math.trunc(e.choice);
      if (!(c >= 0 && c < s.choices[s.index])) return s;
      return record(s, c, e.ms);
    }

    case 'TIMEOUT':
      if (e.scope === 'game') {
        return s.rules.gameMs !== null && s.status !== 'intro' && s.status !== 'finished'
          ? { ...s, status: 'finished', ended: 'time' }
          : s;
      }
      return s.status === 'question' && s.rules.questionMs !== null
        ? record(s, null, s.rules.questionMs)
        : s;

    case 'REVEAL':
      return s.status === 'answered' ? { ...s, status: 'revealing' } : s;

    case 'NEXT': {
      if (s.status !== 'answered' && s.status !== 'revealing') return s;
      if (s.lives !== null && s.lives <= 0) return { ...s, status: 'finished', ended: 'lives' };
      if (s.index + 1 >= questionCount(s)) return { ...s, status: 'finished', ended: 'done' };
      return { ...s, status: 'question', index: s.index + 1, milestone: false };
    }

    case 'RESTART':
      return s.status === 'intro' && s.answers.length === 0 ? s : initialGameState(s.mode);
  }
}

/** The numbers a results screen shows. */
export function gameSummary(s: GameState): {
  asked: number;
  right: number;
  accuracy: number;
  bestStreak: number;
  elapsedMs: number;
  score: number;
} {
  const asked = s.results.length;
  const right = s.results.filter((r) => r.right).length;
  return {
    asked,
    right,
    accuracy: asked ? right / asked : 0,
    bestStreak: s.bestStreak,
    elapsedMs: s.elapsedMs,
    score: s.score
  };
}
