/**
 * The player: a finite state machine over a run's steps, written as a pure
 * reducer so every transition can be tested without a browser.
 *
 *            LOAD             LOADED
 *   idle ----------> loading --------> ready
 *     ^                 |                | PLAY / TOGGLE / REPLAY
 *     | RESET (any)     | FAIL           v
 *     |                 v             playing <---------+
 *     |               error              |  ^           |
 *     |                         PAUSE    |  | PLAY      | REPLAY
 *     |                                  v  |           |
 *     |                               paused ---------> finished
 *     |                      (TICK, STEP_FORWARD or SEEK reaching the end)
 *
 * An event that is not legal in the current state returns the same state
 * object, unchanged. Nothing here starts playing on its own: only PLAY,
 * TOGGLE and REPLAY do, and those come from the visitor
 * (docs/rules/motion-can-be-paused-stepped-and-reduced.md).
 */

export type PlayerStatus =
  'idle' | 'loading' | 'ready' | 'playing' | 'paused' | 'finished' | 'error';

export interface PlayerState {
  status: PlayerStatus;
  /** How many steps have been applied: 0 is the input, total is the end. */
  index: number;
  /** The number of steps. */
  total: number;
  /** Steps per second while playing. */
  speed: number;
  /** Fractional steps carried between ticks. */
  carry: number;
  /** Why loading failed, in the error state. */
  error: string | null;
}

export type PlayerEvent =
  | { type: 'LOAD' }
  | { type: 'LOADED'; total: number; index?: number }
  | { type: 'FAIL'; error: string }
  | { type: 'PLAY' }
  | { type: 'PAUSE' }
  | { type: 'TOGGLE' }
  | { type: 'STEP_FORWARD'; by?: number }
  | { type: 'STEP_BACK'; by?: number }
  | { type: 'SEEK'; index: number }
  | { type: 'SET_SPEED'; speed: number }
  | { type: 'TICK'; dt: number }
  | { type: 'REPLAY' }
  | { type: 'RESET' };

/** The speeds a player offers, in steps per second. */
export const SPEEDS = [1, 2, 5, 10, 25, 50, 100, 250, 500, 1000] as const;
export const MIN_SPEED = 1;
export const MAX_SPEED = 1000;

/** Above this speed, views drop per-step highlights, which would otherwise
 * flash more than three times a second for long stretches. */
export const FAST_SPEED = 50;

export function initialPlayerState(speed = 10): PlayerState {
  return {
    status: 'idle',
    index: 0,
    total: 0,
    speed: clampSpeed(speed),
    carry: 0,
    error: null
  };
}

function clampSpeed(s: number): number {
  return Number.isFinite(s) ? Math.max(MIN_SPEED, Math.min(MAX_SPEED, s)) : 10;
}

function clampIndex(i: number, total: number): number {
  if (Number.isNaN(i)) return 0;
  if (i === Infinity) return total;
  if (i === -Infinity) return 0;
  return Math.max(0, Math.min(Math.trunc(i), total));
}

/** States in which there is a run to move through. */
const loaded: ReadonlySet<PlayerStatus> = new Set(['ready', 'playing', 'paused', 'finished']);

/** Where a stop at index leaves the player. */
function rest(index: number, total: number): PlayerStatus {
  return total > 0 && index >= total ? 'finished' : 'paused';
}

export function playerReducer(s: PlayerState, e: PlayerEvent): PlayerState {
  switch (e.type) {
    case 'RESET':
      return s.status === 'idle' && s.index === 0 && s.total === 0
        ? s
        : { ...initialPlayerState(s.speed) };

    case 'LOAD':
      return s.status === 'loading'
        ? s
        : { ...s, status: 'loading', index: 0, total: 0, carry: 0, error: null };

    case 'LOADED': {
      if (s.status !== 'loading') return s;
      const total = Math.max(0, Math.trunc(e.total));
      const index = clampIndex(e.index ?? 0, total);
      return {
        ...s,
        status: index === 0 ? 'ready' : rest(index, total),
        total,
        index,
        carry: 0
      };
    }

    case 'FAIL':
      return s.status === 'loading' ? { ...s, status: 'error', error: e.error } : s;

    case 'PLAY':
      if (!loaded.has(s.status) || s.status === 'playing' || s.total === 0) return s;
      return {
        ...s,
        status: 'playing',
        index: s.status === 'finished' ? 0 : s.index,
        carry: 0
      };

    case 'PAUSE':
      return s.status === 'playing' ? { ...s, status: 'paused', carry: 0 } : s;

    case 'TOGGLE':
      return s.status === 'playing'
        ? playerReducer(s, { type: 'PAUSE' })
        : playerReducer(s, { type: 'PLAY' });

    case 'STEP_FORWARD':
    case 'STEP_BACK': {
      if (!loaded.has(s.status)) return s;
      const by = Math.abs(e.by ?? 1);
      const index = clampIndex(e.type === 'STEP_FORWARD' ? s.index + by : s.index - by, s.total);
      // A step that cannot move (back at the start, forward at the end) does nothing.
      if (index === s.index) return s;
      return { ...s, status: rest(index, s.total), index, carry: 0 };
    }

    case 'SEEK': {
      if (!loaded.has(s.status)) return s;
      const index = clampIndex(e.index, s.total);
      const status = s.status === 'playing' && index < s.total ? 'playing' : rest(index, s.total);
      if (index === s.index && status === s.status) return s;
      return { ...s, status, index, carry: 0 };
    }

    case 'SET_SPEED': {
      const speed = clampSpeed(e.speed);
      return speed === s.speed ? s : { ...s, speed };
    }

    case 'TICK': {
      if (s.status !== 'playing' || !(e.dt > 0)) return s;
      const carry = s.carry + (e.dt * s.speed) / 1000;
      const advance = Math.floor(carry);
      if (advance === 0) return { ...s, carry };
      const index = Math.min(s.total, s.index + advance);
      return index >= s.total
        ? { ...s, status: 'finished', index: s.total, carry: 0 }
        : { ...s, index, carry: carry - advance };
    }

    case 'REPLAY':
      return loaded.has(s.status) && s.total > 0
        ? { ...s, status: 'playing', index: 0, carry: 0 }
        : s;
  }
}

/** A machine: its starting state and its transition function. */
export interface PlayerMachine {
  initial: PlayerState;
  transition: (s: PlayerState, e: PlayerEvent) => PlayerState;
}

export function createPlayerMachine(options: { speed?: number } = {}): PlayerMachine {
  return { initial: initialPlayerState(options.speed), transition: playerReducer };
}

/** Which events change something in state s: for enabling controls. */
export function can(s: PlayerState, e: PlayerEvent): boolean {
  return playerReducer(s, e) !== s;
}
