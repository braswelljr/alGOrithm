'use client';

import { useHotkeys } from 'react-hotkeys-hook';
import { keysOf } from '@/config/shortcuts';
import { useShortcuts } from '@/context/use-shortcuts';
import { isTypingTarget } from '@/lib/shortcuts';
import type { Player } from '@/visualizer';
import { SPEEDS } from '@/visualizer/core';

const opts = {
  enableOnFormTags: false,
  preventDefault: true,
  // Never while typing, and leave keys alone inside a focused stage, which
  // handles its own (the package's player keys) so a key does not step twice.
  ignoreEventWhen: (e: KeyboardEvent) =>
    isTypingTarget(e.target) ||
    !!(e.target as Element | null)?.closest?.('.av-stage, [role=slider], [role=listbox]')
};

/**
 * The player's keys on the whole page, from the shortcuts registry: Space,
 * the arrows (Shift for 10), Home, End, R, [ and ]. Registers the "player"
 * scope so the help lists them first.
 */
export function usePlayerShortcuts(player: Player, enabled = true) {
  const { useScope } = useShortcuts();
  useScope('player');
  const send = player.send;
  const o = { ...opts, enabled };
  useHotkeys(keysOf('player.toggle'), () => send({ type: 'TOGGLE' }), o);
  useHotkeys(keysOf('player.forward'), () => send({ type: 'STEP_FORWARD', by: 1 }), o);
  useHotkeys(keysOf('player.back'), () => send({ type: 'STEP_BACK', by: 1 }), o);
  useHotkeys(keysOf('player.forward10'), () => send({ type: 'STEP_FORWARD', by: 10 }), o);
  useHotkeys(keysOf('player.back10'), () => send({ type: 'STEP_BACK', by: 10 }), o);
  useHotkeys(keysOf('player.first'), () => send({ type: 'SEEK', index: 0 }), o);
  useHotkeys(keysOf('player.last'), () => send({ type: 'SEEK', index: Infinity }), o);
  useHotkeys(keysOf('player.replay'), () => send({ type: 'REPLAY' }), o);
  const at = SPEEDS.indexOf(player.state.speed as (typeof SPEEDS)[number]);
  useHotkeys(
    keysOf('player.slower'),
    () => at > 0 && send({ type: 'SET_SPEED', speed: SPEEDS[at - 1] }),
    o,
    [at]
  );
  useHotkeys(
    keysOf('player.faster'),
    () => at >= 0 && at < SPEEDS.length - 1 && send({ type: 'SET_SPEED', speed: SPEEDS[at + 1] }),
    o,
    [at]
  );
}
