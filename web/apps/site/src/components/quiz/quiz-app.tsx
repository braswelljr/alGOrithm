'use client';

import { useCallback, useEffect, useMemo, useReducer, useState } from 'react';
import dynamic from 'next/dynamic';
import { useReducedMotion } from '@braswelljr/algorithm-visualizer';
import {
  gameReducer,
  initialGameState,
  type AlgorithmApi,
  type Catalog,
  type GameEvent
} from '@braswelljr/algorithm-visualizer/core';
import { AnimatePresence, motion } from 'motion/react';
import { parseAsInteger, parseAsString, parseAsStringLiteral, useQueryStates } from 'nuqs';
import { WasmGate } from '@/components/shared/wasm-gate';
import { buildGame, type GameItem } from '@/lib/quiz-game';
import {
  browserStorage,
  domainsOf,
  levels,
  loadBests,
  newSeed,
  saveBest,
  type Bests,
  type GameConfig
} from '@/lib/quiz-seed';
import { StartScreen } from './start-screen';

// The game and results screens (gsap, the live view, the burst) load while the
// start screen is up, not before the page can be read.
const PlayScreen = dynamic(() => import('./play-screen').then((m) => m.PlayScreen), {
  ssr: false
});
const ResultsScreen = dynamic(() => import('./results-screen').then((m) => m.ResultsScreen), {
  ssr: false
});

/** The game's settings live in the link, so a game can be shared and played
 * again: the same seed gives the same questions
 * (docs/rules/state-a-reader-can-arrive-at-lives-in-the-url.md). */
const params = {
  mode: parseAsStringLiteral(['classic', 'time', 'streak'] as const).withDefault('classic'),
  domain: parseAsString.withDefault('mixed'),
  level: parseAsStringLiteral(levels).withDefault('any'),
  seed: parseAsInteger,
  entry: parseAsString,
  beat: parseAsInteger
};

function Game({ api, catalog }: { api: AlgorithmApi; catalog: Catalog }) {
  const [q, setQ] = useQueryStates(params, { history: 'replace' });
  const reduced = useReducedMotion();
  const [game, dispatch] = useReducer(gameReducer, q.mode, initialGameState);
  const [items, setItems] = useState<GameItem[]>([]);
  const [error, setError] = useState<string | null>(null);
  // The best score for this mode when the game started, to tell a new best.
  const [bestBefore, setBestBefore] = useState(0);
  const send = useCallback((e: GameEvent) => dispatch(e), []);

  // Read again whenever a game starts or ends, after the end saved its numbers.
  const bests: Bests = useMemo(() => loadBests(browserStorage()), [game.status]);

  const entry = q.entry ? catalog.entries.find((e) => e.id === q.entry) : undefined;
  const domains = useMemo(() => domainsOf(catalog), [catalog]);

  const start = (seed: number) => {
    const config: GameConfig = {
      mode: q.mode,
      domain: q.domain,
      level: q.level,
      seed,
      entry: entry?.id ?? null
    };
    const built = buildGame(api, catalog, config);
    if (built.length === 0) {
      setError('No questions fit these settings. Try another topic or difficulty.');
      return;
    }
    setError(null);
    setBestBefore(loadBests(browserStorage())[q.mode]?.score ?? 0);
    setItems(built);
    void setQ({ seed });
    send({
      type: 'START',
      mode: q.mode,
      answers: built.map((b) => b.question.answer),
      choices: built.map((b) => b.question.choices.length),
      questions: entry ? 5 : undefined
    });
  };

  // Keep the bests in this browser when a game ends.
  useEffect(() => {
    if (game.status === 'finished')
      saveBest(browserStorage(), game.mode, game.score, game.bestStreak);
  }, [game.status]);

  const shareUrl = () => {
    if (typeof window === 'undefined') return '';
    const u = new URL(window.location.href);
    u.searchParams.set('beat', String(game.score));
    return u.href;
  };

  const screen =
    game.status === 'intro' ? 'start' : game.status === 'finished' ? 'results' : 'play';
  const fade = reduced
    ? {
        initial: false as const,
        animate: { opacity: 1 },
        exit: { opacity: 1 },
        transition: { duration: 0 }
      }
    : {
        initial: { opacity: 0, transform: 'translateY(8px)' },
        animate: { opacity: 1, transform: 'translateY(0px)' },
        exit: { opacity: 0, transform: 'translateY(-4px)' },
        transition: { type: 'spring' as const, duration: 0.35, bounce: 0 }
      };

  return (
    <AnimatePresence
      mode="wait"
      initial={false}
    >
      <motion.div
        key={screen}
        {...fade}
      >
        {screen === 'start' ? (
          <StartScreen
            config={{
              mode: q.mode,
              domain: q.domain,
              level: q.level,
              seed: q.seed ?? 0,
              entry: q.entry
            }}
            domains={domains}
            entry={entry}
            bests={bests}
            beat={q.beat}
            error={error}
            onChange={(next) => {
              void setQ({ ...next, seed: null, beat: null });
            }}
            onStart={() => start(q.seed ?? newSeed())}
          />
        ) : screen === 'play' ? (
          <PlayScreen
            game={game}
            items={items}
            send={send}
            reduced={reduced}
          />
        ) : (
          <ResultsScreen
            game={game}
            items={items}
            newBest={game.score > bestBefore}
            shareUrl={shareUrl()}
            reduced={reduced}
            onAgain={() => start(q.seed ?? newSeed())}
            onNew={() => start(newSeed())}
            onMenu={() => {
              void setQ({ beat: null });
              send({ type: 'RESTART' });
            }}
          />
        )}
      </motion.div>
    </AnimatePresence>
  );
}

/** The quiz game, once the algorithms are loaded. */
export function QuizApp() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Game
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
