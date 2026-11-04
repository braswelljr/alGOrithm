'use client';

/**
 * Swap race: sort the bars in as few swaps as you can, then see how many every
 * sort in the catalog needed on the same values (the real Go code, through the
 * wasm, counting its own swaps).
 */
import { useEffect, useMemo, useState } from 'react';
import { motion, useReducedMotionConfig } from 'motion/react';
import { LuTrophy } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { useArena } from '@/lib/arena/store';
import { fewestSwaps } from '@/lib/arena/swaps';
import { cn } from '@/lib/utils';
import type { GameProps } from './games';

/** `n` distinct values in no order. */
function shuffled(n: number): number[] {
  const values = Array.from({ length: n }, (_, i) => (i + 1) * 3 + Math.floor(Math.random() * 3));
  for (let i = values.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [values[i], values[j]] = [values[j], values[i]];
  }
  // Never start sorted.
  return values.every((v, i) => i === 0 || values[i - 1] <= v) ? shuffled(n) : values;
}

export function SwapRace({ api, catalog, paused, onScore, onOver }: GameProps) {
  const size = useArena((s) => s.settings['swap-race'].size);
  const reduce = useReducedMotionConfig() ?? false;
  const [start] = useState(() => shuffled(size));
  const [values, setValues] = useState(start);
  const [picked, setPicked] = useState<number | null>(null);
  const [cursor, setCursor] = useState(0);
  const [swaps, setSwaps] = useState(0);
  const sorted = values.every((v, i) => i === 0 || values[i - 1] <= v);
  const best = useMemo(() => fewestSwaps(start), [start]);
  const top = Math.max(...start);

  // The sorts of the catalog on the same values, once the player is done.
  const rivals = useMemo(() => {
    if (!sorted) return [];
    return catalog.entries
      .filter((e) => e.category === 'sorting' && e.input === 'ints')
      .flatMap((e) => {
        const r = api.run(e.id, { values: start }, 100_000);
        if (!r.ok) return [];
        const n = r.value.counts?.swap ?? 0;
        return n > 0 ? [{ id: e.id, name: e.name, swaps: n }] : [];
      })
      .sort((a, b) => a.swaps - b.swaps);
  }, [sorted, catalog, api, start]);

  const score = sorted
    ? Math.round(
        50 +
          100 * (best / Math.max(swaps, best || 1)) +
          15 * rivals.filter((r) => swaps <= r.swaps).length
      )
    : 0;

  useEffect(() => {
    if (!sorted) return;
    onScore(score);
    const beaten = rivals.filter((r) => swaps <= r.swaps).length;
    onOver(
      score,
      `${size} bars in ${swaps} swaps (fewest ${best}), beat ${beaten} of ${rivals.length}`
    );
    // Once, when the bars come sorted.
  }, [sorted]);

  const pick = (i: number) => {
    if (sorted || paused) return;
    if (picked === null) setPicked(i);
    else if (picked === i) setPicked(null);
    else {
      setValues((v) => {
        const next = [...v];
        [next[picked], next[i]] = [next[i], next[picked]];
        return next;
      });
      setSwaps((n) => n + 1);
      setPicked(null);
    }
  };

  // Arrows move the cursor; Space or Enter picks.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (paused || sorted) return;
      if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
        e.preventDefault();
        e.stopPropagation();
        setCursor((c) => (c + (e.key === 'ArrowLeft' ? -1 : 1) + values.length) % values.length);
      } else if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault();
        e.stopPropagation();
        pick(cursor);
      }
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  });

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="text-muted flex flex-wrap items-center justify-between gap-2 text-sm">
        <span>
          Swaps <strong className="text-fg tabular-nums">{swaps}</strong>
        </span>
        <span>
          {sorted
            ? 'Sorted.'
            : picked === null
              ? 'Pick a bar, then the one to swap it with.'
              : 'Now pick the bar to swap it with.'}
        </span>
      </div>
      <ol
        aria-label="The bars"
        className="flex h-64 w-full items-end gap-2"
      >
        {values.map((v, i) => (
          <motion.li
            key={v}
            layout={!reduce}
            transition={{ type: 'spring', stiffness: 500, damping: 34 }}
            className="flex h-full min-w-0 flex-1 flex-col justify-end"
          >
            <Button
              variant="ghost"
              aria-label={`Bar ${i + 1}, ${v}${picked === i ? ', picked' : ''}`}
              aria-pressed={picked === i}
              onClick={() => pick(i)}
              className={cn(
                'h-auto w-full flex-col justify-end rounded-lg p-0 font-mono text-xs font-semibold',
                sorted
                  ? 'text-accent-fg bg-emerald-500'
                  : picked === i
                    ? 'bg-primary text-accent-fg'
                    : 'text-accent-fg bg-amber-400 hover:bg-amber-300',
                cursor === i && !sorted && 'ring-fg ring-offset-bg ring-2 ring-offset-2'
              )}
              style={{ height: `${Math.max(12, (v / top) * 100)}%` }}
            >
              <span className="pb-1.5">{v}</span>
            </Button>
          </motion.li>
        ))}
      </ol>
      {sorted ? (
        <section
          aria-label="The race"
          className="flex flex-col gap-2"
        >
          <Heading
            level={3}
            className="font-heading text-fg text-lg font-semibold"
          >
            The race
          </Heading>
          <p className="text-muted text-sm">
            You: <strong className="text-fg">{swaps}</strong> swaps. The fewest possible:{' '}
            <strong className="text-fg">{best}</strong>.
          </p>
          <ol className="grid gap-1.5 sm:grid-cols-2">
            {rivals.map((r) => {
              const won = swaps <= r.swaps;
              return (
                <li
                  key={r.id}
                  className={cn(
                    'flex items-center justify-between rounded-lg px-3 py-2 text-sm',
                    won ? 'bg-success/12 text-success' : 'bg-subtle text-muted'
                  )}
                >
                  <span className="flex items-center gap-2">
                    {won ? (
                      <LuTrophy
                        aria-hidden
                        className="size-4"
                      />
                    ) : null}
                    {r.name}
                  </span>
                  <span className="font-mono tabular-nums">{r.swaps}</span>
                </li>
              );
            })}
          </ol>
        </section>
      ) : null}
    </div>
  );
}
