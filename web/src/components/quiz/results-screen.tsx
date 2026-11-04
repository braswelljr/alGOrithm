'use client';

import { useEffect, useRef } from 'react';
import Link from 'next/link';
import { Check, Clock, Flame, Target, Trophy, X } from 'lucide-react';
import { motion } from 'motion/react';
import { CopyButton } from '@/components/run/copy-button';
import { Button } from '@/components/shared/button';
import { Card } from '@/components/ui/card';
import { Heading } from '@/components/ui/heading';
import { replayHref, type GameItem } from '@/lib/quiz-game';
import { gameSummary, type GameState } from '@/visualizer/core';
import { CountUp } from './count-up';

const endWords = {
  done: 'All questions answered.',
  lives: 'Out of lives.',
  time: 'Time is up.'
} as const;

/** The end of a game: the numbers, a review of every answer, and what next. */
export function ResultsScreen({
  game,
  items,
  newBest,
  shareUrl,
  reduced,
  onAgain,
  onNew,
  onMenu
}: {
  game: GameState;
  items: GameItem[];
  newBest: boolean;
  shareUrl: string;
  reduced: boolean;
  onAgain(): void;
  onNew(): void;
  onMenu(): void;
}) {
  const s = gameSummary(game);
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => heading.current?.focus(), []);
  const stats = [
    {
      icon: Target,
      label: 'Accuracy',
      value: `${Math.round(s.accuracy * 100)}%`,
      sub: `${s.right} of ${s.asked}`
    },
    { icon: Flame, label: 'Best streak', value: String(s.bestStreak), sub: 'in a row' },
    { icon: Clock, label: 'Time', value: `${(s.elapsedMs / 1000).toFixed(1)} s`, sub: 'answering' }
  ];
  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-col gap-1">
        <p className="text-muted text-sm">{game.ended ? endWords[game.ended] : ''}</p>
        <Heading
          level={2}
          ref={heading}
          tabIndex={-1}
          className="text-4xl font-semibold tracking-tight outline-none"
        >
          <CountUp
            value={s.score}
            reduced={reduced}
          />{' '}
          <span className="text-muted text-lg font-normal">points</span>
        </Heading>
        {newBest ? (
          <p className="border-border bg-subtle flex w-fit flex-wrap items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-sm">
            <Trophy
              aria-hidden="true"
              className="size-4"
            />
            A new best for this mode.
          </p>
        ) : null}
      </header>

      <dl className="grid gap-3 sm:grid-cols-3">
        {stats.map((st, i) => (
          <Card
            key={st.label}
            size="md"
            className="gap-1 px-(--card-spacing)"
            render={
              <motion.div
                initial={reduced ? false : { opacity: 0, transform: 'translateY(8px)' }}
                animate={{ opacity: 1, transform: 'translateY(0px)' }}
                transition={{ type: 'spring', duration: 0.4, bounce: 0, delay: i * 0.05 }}
              />
            }
          >
            <dt className="text-muted flex items-center gap-1.5 text-sm">
              <st.icon
                aria-hidden="true"
                className="size-4"
              />
              {st.label}
            </dt>
            <dd className="text-2xl font-semibold">{st.value}</dd>
            <dd className="text-muted text-xs">{st.sub}</dd>
          </Card>
        ))}
      </dl>

      <div className="flex flex-wrap gap-2">
        <Button
          variant="primary"
          onClick={onAgain}
        >
          Play the same questions again
        </Button>
        <Button onClick={onNew}>New game</Button>
        <Button
          variant="ghost"
          onClick={onMenu}
        >
          Change settings
        </Button>
        <CopyButton text={() => shareUrl}>Copy a link to this game</CopyButton>
      </div>

      <section aria-labelledby="review">
        <Heading
          level={3}
          id="review"
          className="mb-3 text-lg font-medium"
        >
          Review
        </Heading>
        <ol className="flex flex-col gap-2">
          {game.results.map((r, i) => {
            const item = items[i];
            if (!item) return null;
            const q = item.question;
            return (
              <li
                key={i}
                className="border-border flex items-start gap-3 rounded-xl border px-4 py-3"
                data-right={r.right || undefined}
              >
                <span className="pt-0.5">
                  {r.right ? (
                    <Check
                      aria-hidden="true"
                      className="text-success size-4"
                    />
                  ) : (
                    <X
                      aria-hidden="true"
                      className="text-danger size-4"
                    />
                  )}
                  <span className="sr-only">{r.right ? 'Right' : 'Wrong'}:</span>
                </span>
                <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                  <p className="text-muted text-sm">{item.entry.name}</p>
                  <p className="font-medium">{q.prompt}</p>
                  <p className="text-sm">
                    {r.right ? 'You said' : r.pick === null ? 'Out of time; it was' : 'It was'}{' '}
                    <strong>{q.choices[q.answer]}</strong>
                    {!r.right && r.pick !== null ? `, not ${q.choices[r.pick]}` : ''}.{' '}
                    {q.explanation}
                  </p>
                </div>
                <Link
                  className="shrink-0 text-sm underline"
                  href={replayHref(item)}
                >
                  Replay this step
                </Link>
              </li>
            );
          })}
        </ol>
      </section>
    </div>
  );
}
