import { useCallback, useEffect, useReducer, type KeyboardEvent } from 'react';
import {
  can,
  FAST_SPEED,
  initialPlayerState,
  playerReducer,
  type PlayerEvent,
  type PlayerState
} from '../core/machine';
import { useReducedMotion } from './use-reduced-motion';

export interface Player {
  state: PlayerState;
  send(e: PlayerEvent): void;
  /** Whether an event would change anything: for disabling controls. */
  can(e: PlayerEvent): boolean;
  /** Starts a new run of `total` steps, paused at `index`. */
  load(total: number, index?: number): void;
  /** Views drop per-step highlights when this is set. */
  quiet: boolean;
  reducedMotion: boolean;
  /** Space or K toggles; arrows step (Shift: 10); Home and End jump. */
  onKeyDown(e: KeyboardEvent): void;
}

/**
 * Drives the player machine (core/machine.ts) in React: TICK events come from
 * requestAnimationFrame while, and only while, the state is "playing".
 */
export function usePlayer(options: { speed?: number } = {}): Player {
  const [state, send] = useReducer(playerReducer, options.speed, initialPlayerState);
  const reducedMotion = useReducedMotion();
  const playing = state.status === 'playing';

  useEffect(() => {
    if (!playing) return;
    let frame = 0;
    let last = performance.now();
    const tick = (now: number) => {
      send({ type: 'TICK', dt: now - last });
      last = now;
      frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [playing]);

  const load = useCallback((total: number, index = 0) => {
    send({ type: 'LOAD' });
    send({ type: 'LOADED', total, index });
  }, []);

  const onKeyDown = useCallback((e: KeyboardEvent) => {
    const t = e.target;
    if (
      t instanceof HTMLElement &&
      t !== e.currentTarget &&
      t.closest('input, textarea, select, button, a, [contenteditable]')
    ) {
      return;
    }
    const by = e.shiftKey ? 10 : 1;
    const map: Record<string, PlayerEvent | undefined> = {
      ' ': { type: 'TOGGLE' },
      k: { type: 'TOGGLE' },
      ArrowRight: { type: 'STEP_FORWARD', by },
      ArrowLeft: { type: 'STEP_BACK', by },
      Home: { type: 'SEEK', index: 0 },
      End: { type: 'SEEK', index: Infinity },
      r: { type: 'REPLAY' }
    };
    const ev = map[e.key];
    if (ev) {
      e.preventDefault();
      send(ev);
    }
  }, []);

  return {
    state,
    send,
    can: (e) => can(state, e),
    load,
    quiet: playing && state.speed > FAST_SPEED,
    reducedMotion,
    onKeyDown
  };
}
