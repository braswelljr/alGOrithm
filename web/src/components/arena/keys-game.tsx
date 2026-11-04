'use client';

/**
 * Keys: the shortcut trainer as a scored round. Each right answer is worth 10
 * points, more as the streak grows (up to five times), and the round ends after
 * the number of right answers set in its menu.
 */
import { useCallback, useEffect, useRef, useState } from 'react';
import { ShortcutTrainer } from '@/components/shortcuts/shortcut-trainer';
import { useArena } from '@/lib/arena/store';
import type { GameProps } from './games';

export function KeysGame({ onScore, onOver }: GameProps) {
  const rounds = useArena((s) => s.settings.keys.rounds);
  const [right, setRight] = useState(0);
  const score = useRef(0);
  const best = useRef(0);
  const done = useRef(false);

  const onRight = useCallback(
    (streak: number) => {
      score.current += 10 * Math.min(streak, 5);
      best.current = Math.max(best.current, streak);
      onScore(score.current);
      setRight((n) => n + 1);
    },
    [onScore]
  );

  useEffect(() => {
    if (right >= rounds && !done.current) {
      done.current = true;
      onOver(score.current, `${rounds} right, best streak ${best.current}`);
    }
  }, [right, rounds, onOver]);

  return (
    <div className="flex w-full flex-col gap-3">
      <p className="text-muted text-sm">
        Right answers <strong className="text-fg tabular-nums">{right}</strong> of {rounds}
      </p>
      {/* The trainer draws its own surface; on the console it sits on a light card. */}
      <div className="bg-bg text-fg rounded-2xl p-4">
        <ShortcutTrainer onRight={onRight} />
      </div>
    </div>
  );
}
