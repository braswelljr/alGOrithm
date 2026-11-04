'use client';

import { useEffect, useRef, type ReactNode } from 'react';
import { Flame, Heart } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { Button } from '@/components/shared/button';
import { Card } from '@/components/ui/card';
import { Frame, FramePanel } from '@/components/ui/frame';
import { Stat, StatLabel, StatValue } from '@/components/ui/stat';
import type { GameItem } from '@/lib/quiz-game';
import { multiplier, questionCount, type GameEvent, type GameState } from '@/visualizer/core';
import { Burst } from './burst';
import { CountUp } from './count-up';
import { CountdownRing } from './countdown-ring';
import { QuizQuestion } from './quiz-question';

/** How long the visual plays the answered step before the explanation. */
const REVEAL_MS = 700;

/** The score, streak, lives and progress across the top of the game. */
function Hud({
  game,
  reduced,
  gameClock
}: {
  game: GameState;
  reduced: boolean;
  gameClock: ReactNode;
}) {
  const total = questionCount(game);
  const last = game.results.at(-1);
  return (
    <Card
      size="sm"
      className="flex-row flex-wrap items-center gap-x-6 gap-y-3 px-(--card-spacing)"
    >
      <Stat
        variant="plain"
        className="gap-0.5"
      >
        <StatLabel className="text-xs">Score</StatLabel>
        <StatValue className="relative text-lg sm:text-lg lg:text-lg">
          <CountUp
            value={game.score}
            reduced={reduced}
          />
          <AnimatePresence>
            {game.status !== 'question' && last?.right ? (
              <motion.span
                key={game.results.length}
                className="text-success absolute top-0 left-full ml-1 text-[0.8125rem] whitespace-nowrap"
                initial={reduced ? false : { opacity: 0, transform: 'translateY(6px)' }}
                animate={{ opacity: 1, transform: 'translateY(-14px)' }}
                exit={{ opacity: 0 }}
                transition={{ type: 'spring', duration: 0.5, bounce: 0.2 }}
              >
                +{last.points}
              </motion.span>
            ) : null}
          </AnimatePresence>
        </StatValue>
      </Stat>
      <Stat
        variant="plain"
        className="group gap-0.5"
        data-hot={game.streak >= 3 || undefined}
      >
        <StatLabel className="text-xs">Streak</StatLabel>
        <StatValue className="inline-flex items-center gap-1 text-lg sm:text-lg lg:text-lg">
          <Flame
            aria-hidden="true"
            className="text-muted group-data-hot:text-accent size-4 transition-transform duration-200 ease-[cubic-bezier(0.23,1,0.32,1)] group-data-hot:scale-125 motion-reduce:transition-none"
          />
          {game.streak}
          <span className="text-muted text-xs font-medium">
            x{multiplier(Math.max(1, game.streak))}
          </span>
        </StatValue>
      </Stat>
      {game.lives !== null && game.rules.lives !== null ? (
        <Stat
          variant="plain"
          className="gap-0.5"
        >
          <StatLabel className="text-xs">Lives</StatLabel>
          <StatValue
            className="inline-flex gap-0.5 text-lg sm:text-lg lg:text-lg"
            aria-label={`${game.lives} of ${game.rules.lives} lives left`}
            role="img"
          >
            {Array.from({ length: game.rules.lives }, (_, i) => (
              <Heart
                key={i}
                aria-hidden="true"
                className="text-danger size-4 fill-current data-lost:fill-none data-lost:opacity-40"
                data-lost={i >= game.lives! || undefined}
              />
            ))}
          </StatValue>
        </Stat>
      ) : null}
      <Stat
        variant="plain"
        className="gap-0.5"
      >
        <StatLabel className="text-xs">Question</StatLabel>
        <StatValue className="text-lg sm:text-lg lg:text-lg">
          {game.index + 1}
          {Number.isFinite(total) && game.rules.questions !== null ? ` / ${total}` : ''}
        </StatValue>
      </Stat>
      <div className="ml-auto">{gameClock}</div>
    </Card>
  );
}

/**
 * One question at a time: the real view of the run, paused at the question's
 * step, and the choices as cards. An answer plays the step on the view, then
 * shows the explanation.
 */
export function PlayScreen({
  game,
  items,
  send,
  reduced
}: {
  game: GameState;
  items: GameItem[];
  send(e: GameEvent): void;
  reduced: boolean;
}) {
  const item = items[game.index];
  const asking = game.status === 'question';
  const startedAt = useRef(0);

  // Time each question from when it is shown.
  useEffect(() => {
    if (asking) startedAt.current = performance.now();
  }, [asking, game.index]);

  // After an answer, let the view play the step, then reveal.
  useEffect(() => {
    if (game.status !== 'answered') return;
    if (reduced) {
      send({ type: 'REVEAL' });
      return;
    }
    const t = setTimeout(() => send({ type: 'REVEAL' }), REVEAL_MS);
    return () => clearTimeout(t);
  }, [game.status, game.index, reduced, send]);

  if (!item) return null;
  const q = item.question;
  const answered = !asking;
  const last = answered ? game.results[game.index] : undefined;
  const step = answered ? Math.min(item.frame.total, q.step + 1) : q.step;
  const { view, words } = item.frame.render(step, false);
  const lastQuestion = game.index + 1 >= Math.min(questionCount(game), items.length);

  const gameClock =
    game.rules.gameMs !== null ? (
      <CountdownRing
        durationMs={game.rules.gameMs}
        running={asking}
        resetKey="game"
        reduced={reduced}
        label="Time left in the game"
        onDone={() => send({ type: 'TIMEOUT', scope: 'game' })}
      />
    ) : game.rules.questionMs !== null ? (
      <CountdownRing
        durationMs={game.rules.questionMs}
        running={asking}
        resetKey={game.index}
        reduced={reduced}
        label="Time left for this question"
        onDone={() => send({ type: 'TIMEOUT', scope: 'question' })}
      />
    ) : null;

  const feedback =
    !last || reduced
      ? {}
      : last.right
        ? { transform: ['scale(1)', 'scale(1.015)', 'scale(1)'] }
        : {
            transform: [
              'translateX(0px)',
              'translateX(-8px)',
              'translateX(7px)',
              'translateX(-4px)',
              'translateX(0px)'
            ]
          };

  return (
    <div className="flex flex-col gap-4">
      <div className="relative">
        <Hud
          game={game}
          reduced={reduced}
          gameClock={gameClock}
        />
        <Burst
          fire={game.milestone ? game.index + 1 : 0}
          reduced={reduced}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
        <figure className="min-w-0">
          <Frame>
            <figcaption className="text-muted px-4 py-2 text-sm">
              {item.entry.name}
              <span className="sr-only">. {words}</span>
            </figcaption>
            <FramePanel className="p-4">
              {view ? (
                <div className="av-root">
                  <div className="av-stage">{view}</div>
                </div>
              ) : (
                <p className="text-muted text-sm">{q.context}</p>
              )}
            </FramePanel>
          </Frame>
        </figure>

        <motion.div
          key={game.index}
          animate={feedback}
          transition={{ duration: 0.35, ease: [0.23, 1, 0.32, 1] }}
        >
          <QuizQuestion
            question={q}
            index={game.index}
            picked={last ? last.pick : null}
            answered={answered}
            onAnswer={(choice) =>
              send({ type: 'ANSWER', choice, ms: performance.now() - startedAt.current })
            }
          >
            {game.status === 'revealing' ? (
              <motion.div
                className="flex flex-wrap items-center justify-between gap-3"
                initial={reduced ? false : { opacity: 0, transform: 'translateY(4px)' }}
                animate={{ opacity: 1, transform: 'translateY(0px)' }}
                transition={{ type: 'spring', duration: 0.3, bounce: 0 }}
              >
                {game.milestone ? (
                  <p className="font-medium">
                    <Flame
                      aria-hidden="true"
                      className="inline size-4"
                    />{' '}
                    {game.streak} in a row: x{multiplier(game.streak)} points.
                  </p>
                ) : null}
                <Button
                  variant="primary"
                  autoFocus
                  onClick={() => send({ type: 'NEXT' })}
                >
                  {lastQuestion || (game.lives !== null && game.lives <= 0)
                    ? 'See results'
                    : 'Next question'}
                </Button>
              </motion.div>
            ) : null}
          </QuizQuestion>
        </motion.div>
      </div>
    </div>
  );
}
