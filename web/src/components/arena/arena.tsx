'use client';

/**
 * The game arena: a select screen of games, a game's own menu (Play, How to
 * play, Settings, Scores, Back) and the game itself, with a pause menu and a
 * game-over screen, all on one console screen. Arrows choose, Enter takes,
 * Escape goes back, as a game does; pointer and touch work too. Points, bests
 * and settings stay in this browser (lib/arena/store).
 *
 * Which screen is up and which game lives in the link (?game=snake&screen=menu),
 * so the browser's Back goes back a screen.
 *
 * ui-ok: primaries: each screen (select, menu, play, pause) leads with its own.
 */
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { AnimatePresence, motion, useReducedMotionConfig } from 'motion/react';
import { parseAsString, parseAsStringLiteral, useQueryStates } from 'nuqs';
import {
  LuArrowLeft,
  LuBookOpen,
  LuChartBar,
  LuPause,
  LuPlay,
  LuRotateCcw,
  LuSettings,
  LuTrophy
} from 'react-icons/lu';
import { NumberInput } from '@/components/shared/number-input';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Button } from '@/components/ui/button';
import { FacetFilter } from '@/components/ui/facet-filter';
import { Heading } from '@/components/ui/heading';
import { Item, ItemContent, ItemDescription, ItemMedia, ItemTitle } from '@/components/ui/item';
import { Switch } from '@/components/ui/switch';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { rankOf } from '@/lib/arena/rank';
import { useArena, type GameId } from '@/lib/arena/store';
import { cn } from '@/lib/utils';
import type { AlgorithmApi, Catalog } from '@/visualizer/core';
import { ArenaScreen, KeyHints } from './arena-screen';
import { GAME_INFO, gameInfo, type GameInfo } from './games';
import { MapPicker } from './map-picker';
import { pilotsOf } from './snake-game';

const SCREENS = ['select', 'menu', 'play'] as const;
const MENU = ['play', 'how', 'settings', 'scores', 'back'] as const;
type MenuItem = (typeof MENU)[number];
const MENU_LABEL: Record<MenuItem, string> = {
  play: 'Play',
  how: 'How to play',
  settings: 'Settings',
  scores: 'Scores',
  back: 'Back to the games'
};
const MENU_ICON = {
  play: LuPlay,
  how: LuBookOpen,
  settings: LuSettings,
  scores: LuChartBar,
  back: LuArrowLeft
};

/** Arrow keys over a list of choices, Enter to take one, Escape to go back. */
function useMenuKeys({
  count,
  index,
  setIndex,
  onEnter,
  onBack,
  axis = 'y',
  enabled = true
}: {
  count: number;
  index: number;
  setIndex: (i: number) => void;
  onEnter: (i: number) => void;
  onBack?: () => void;
  axis?: 'x' | 'y';
  enabled?: boolean;
}) {
  useEffect(() => {
    if (!enabled) return;
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement | null;
      if (t?.closest('input, textarea, [role=listbox], [role=dialog]')) return;
      const prev = axis === 'y' ? 'ArrowUp' : 'ArrowLeft';
      const next = axis === 'y' ? 'ArrowDown' : 'ArrowRight';
      if (e.key === prev || e.key === next) {
        e.preventDefault();
        setIndex((index + (e.key === next ? 1 : -1) + count) % count);
      } else if (e.key === 'Enter') {
        if (t?.closest('button, a')) return;
        e.preventDefault();
        onEnter(index);
      } else if (e.key === 'Escape' && onBack) {
        e.preventDefault();
        onBack();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [count, index, setIndex, onEnter, onBack, axis, enabled]);
}

/** The player's record: points, rank and the way to the next. */
function PlayerCard() {
  const points = useArena((s) => s.points);
  const rank = rankOf(points);
  return (
    <div className="bg-subtle ring-border flex min-w-56 flex-col gap-1.5 rounded-2xl p-3 ring-1">
      <div className="flex items-center justify-between gap-3">
        <span className="text-muted text-xs font-semibold tracking-widest uppercase">
          Level {rank.level} · {rank.name}
        </span>
        <span className="text-warning flex items-center gap-1 font-mono text-sm font-bold tabular-nums">
          <LuTrophy aria-hidden />
          {points.toLocaleString('en-US')}
        </span>
      </div>
      <div
        role="progressbar"
        aria-label={rank.next ? `Toward ${rank.next}` : 'The top rank'}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(rank.progress * 100)}
        className="bg-border h-1.5 overflow-hidden rounded-full"
      >
        <div
          className="bg-primary h-full rounded-full"
          style={{ width: `${rank.progress * 100}%` }}
        />
      </div>
      <span className="text-muted text-[11px]">
        {rank.next
          ? `${rank.toNext.toLocaleString('en-US')} points to ${rank.next}`
          : 'The top rank'}
      </span>
    </div>
  );
}

/**
 * The record: each game's best, how often it was played and its last score,
 * and the latest games across them all, newest first.
 */
function Records() {
  const best = useArena((s) => s.best);
  const played = useArena((s) => s.played);
  const history = useArena((s) => s.history);
  const last = useMemo(() => {
    const out: Partial<Record<GameId, number>> = {};
    for (const h of history) if (out[h.game] === undefined) out[h.game] = h.score;
    return out;
  }, [history]);
  const nameOf = (id: GameId) => gameInfo(id)?.name ?? id;
  return (
    <section
      aria-labelledby="arena-records"
      data-tour="game-records"
      className="grid gap-4 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]"
    >
      <div className="flex flex-col gap-2">
        <Heading
          level={3}
          id="arena-records"
          className="flex items-center gap-2 text-base font-semibold"
        >
          <LuTrophy
            aria-hidden
            className="text-warning"
          />
          High scores
        </Heading>
        <Table>
          <caption className="sr-only">Each game's best score, games played and last score</caption>
          <TableHeader>
            <TableRow>
              <TableHead scope="col">Game</TableHead>
              <TableHead
                scope="col"
                className="text-end"
              >
                Best
              </TableHead>
              <TableHead
                scope="col"
                className="text-end"
              >
                Played
              </TableHead>
              <TableHead
                scope="col"
                className="text-end"
              >
                Last
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[...GAME_INFO]
              .sort((a, b) => (best[b.id] ?? 0) - (best[a.id] ?? 0))
              .map((g) => (
                <TableRow key={g.id}>
                  <TableCell className="font-medium">{g.name}</TableCell>
                  <TableCell className="text-warning text-end font-mono font-semibold tabular-nums">
                    {best[g.id] ?? 0}
                  </TableCell>
                  <TableCell className="text-end font-mono tabular-nums">
                    {played[g.id] ?? 0}
                  </TableCell>
                  <TableCell className="text-muted text-end font-mono tabular-nums">
                    {last[g.id] ?? '-'}
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </div>
      <div className="flex flex-col gap-2">
        <Heading
          level={3}
          className="text-base font-semibold"
        >
          The latest games
        </Heading>
        {history.length ? (
          <ol className="flex flex-col gap-1.5 text-sm">
            {history.slice(0, 8).map((h) => (
              <li
                key={h.at}
                className="bg-subtle/80 flex items-center gap-3 rounded-lg px-3 py-2"
              >
                <span className="w-24 shrink-0 font-medium">{nameOf(h.game)}</span>
                <span className="text-muted min-w-0 flex-1 truncate">{h.note ?? ''}</span>
                <span className="font-mono font-semibold tabular-nums">{h.score}</span>
              </li>
            ))}
          </ol>
        ) : (
          <p className="text-muted text-sm">
            No games yet. Every finished game is kept here, in this browser.
          </p>
        )}
      </div>
    </section>
  );
}

/** The first screen: every game as a cartridge. */
function SelectScreen({ onPick }: { onPick: (id: GameId) => void }) {
  const [index, setIndex] = useState(0);
  const best = useArena((s) => s.best);
  const played = useArena((s) => s.played);
  useMenuKeys({
    count: GAME_INFO.length,
    index,
    setIndex,
    onEnter: (i) => onPick(GAME_INFO[i].id),
    axis: 'x'
  });
  return (
    <div
      data-tour="game-select"
      className="flex flex-1 flex-col gap-8 p-6 sm:p-10"
    >
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-primary text-xs font-semibold tracking-[0.3em] uppercase">alGOrithm</p>
          <Heading
            level={2}
            className="font-heading text-3xl font-bold tracking-tight"
          >
            <span className="shimmer text-fg [animation-duration:3.5s]">The arena</span>
          </Heading>
          <p className="text-muted mt-2 max-w-xl text-sm">
            Games played against the real algorithms. Choose one to open its menu.
          </p>
        </div>
        <PlayerCard />
      </header>
      <ul
        aria-label="Games"
        className="grid gap-2 sm:grid-cols-2 xl:grid-cols-3"
      >
        {GAME_INFO.map((g, i) => {
          const Icon = g.icon;
          const on = i === index;
          return (
            <li key={g.id}>
              <Item
                size="sm"
                render={
                  // ui-ok: a game is the select screen's item itself, drawn by Item
                  <button
                    type="button"
                    onClick={() => onPick(g.id)}
                    onMouseEnter={() => setIndex(i)}
                    onFocus={() => setIndex(i)}
                    aria-current={on ? 'true' : undefined}
                  />
                }
                className={cn(
                  'bg-subtle/80 hover:bg-subtle w-full text-start',
                  // The chosen game is a keycap, the primary cap, as a list's chosen row is.
                  on && 'keycap keycap-primary'
                )}
              >
                <ItemMedia
                  variant="icon"
                  className={cn(on ? 'text-accent-fg' : g.tone)}
                >
                  <Icon aria-hidden />
                </ItemMedia>
                <ItemContent>
                  <ItemTitle>{g.name}</ItemTitle>
                  <ItemDescription className={cn('line-clamp-2', on && 'text-accent-fg/80')}>
                    {g.tagline}
                  </ItemDescription>
                </ItemContent>
                <span
                  className={cn(
                    'shrink-0 font-mono text-xs',
                    on ? 'text-accent-fg/80' : 'text-muted'
                  )}
                >
                  {best[g.id] ?? 0}
                  <span className="sr-only"> best, played {played[g.id] ?? 0}</span>
                </span>
              </Item>
            </li>
          );
        })}
      </ul>
      <Records />
      <div className="mt-auto">
        <KeyHints
          hints={[
            ['← →', 'choose'],
            ['Enter', 'open'],
            ['Shift T', 'tour']
          ]}
        />
      </div>
    </div>
  );
}

/** A game's settings, written straight to the arena's store. */
function SettingsPanel({
  game,
  api,
  catalog
}: {
  game: GameId;
  api: AlgorithmApi;
  catalog: Catalog;
}) {
  const settings = useArena((s) => s.settings);
  const set = useArena((s) => s.setSettings);
  if (game === 'snake') {
    const pilots = pilotsOf(catalog);
    return (
      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">Speed</span>
          <div
            role="radiogroup"
            aria-label="Speed"
            className="flex gap-2"
          >
            {(['calm', 'brisk', 'fast'] as const).map((sp) => (
              <Button
                key={sp}
                role="radio"
                aria-checked={settings.snake.speed === sp}
                variant={settings.snake.speed === sp ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => set('snake', { speed: sp })}
                className="capitalize"
              >
                {sp}
              </Button>
            ))}
          </div>
        </div>
        <label className="flex items-center justify-between gap-4 text-sm">
          <span>
            <span className="font-medium">Walls</span>
            <span className="text-muted block">Off, the snake wraps round the edges.</span>
          </span>
          <Switch
            checked={settings.snake.walls}
            onCheckedChange={(v) => set('snake', { walls: v })}
          />
        </label>
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">Autopilot</span>
          <FacetFilter
            single
            title="Autopilot"
            size="default"
            className="w-full justify-start"
            options={[
              { value: 'off', label: 'Off: I steer' },
              ...pilots.map((e) => ({ value: e.id, label: e.name }))
            ]}
            value={[settings.snake.pilot ?? 'off']}
            onValueChange={(v) => set('snake', { pilot: v[0] === 'off' ? null : (v[0] ?? null) })}
          />
          <span className="text-muted text-xs">
            A pathfinder from the catalog plans every move, as the real Go code.
          </span>
        </div>
      </div>
    );
  }
  if (game === 'race') {
    const kinds = catalog.gridKinds.length ? catalog.gridKinds : ['weighted'];
    return (
      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">Map</span>
          <MapPicker
            api={api}
            kinds={kinds}
            value={settings.race.map}
            onChange={(map) => set('race', { map })}
          />
          <span className="text-muted text-xs">
            A weighted map rewards the algorithms that weigh costs.
          </span>
        </div>
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium">Size</span>
          <div
            role="radiogroup"
            aria-label="Size"
            className="flex gap-2"
          >
            {(['small', 'medium', 'large'] as const).map((sz) => (
              <Button
                key={sz}
                role="radio"
                aria-checked={settings.race.size === sz}
                variant={settings.race.size === sz ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => set('race', { size: sz })}
                className="capitalize"
              >
                {sz}
              </Button>
            ))}
          </div>
        </div>
      </div>
    );
  }
  if (game === 'swap-race') {
    return (
      <label className="flex flex-col gap-2 text-sm">
        <span className="font-medium">Bars</span>
        <NumberInput
          min={5}
          max={14}
          value={settings['swap-race'].size}
          onValueChange={(n) => n !== null && set('swap-race', { size: n })}
          aria-label="Bars"
          className="sm:max-w-xs"
        />
        <span className="text-muted text-xs">More bars, more swaps, more points.</span>
      </label>
    );
  }
  if (game === 'bricks') {
    return <p className="text-muted text-sm">Bricks has no settings: it speeds up by itself.</p>;
  }
  return (
    <label className="flex flex-col gap-2 text-sm">
      <span className="font-medium">Right answers in a round</span>
      <NumberInput
        min={5}
        max={40}
        value={settings.keys.rounds}
        onValueChange={(n) => n !== null && set('keys', { rounds: n })}
        aria-label="Right answers in a round"
        className="sm:max-w-xs"
      />
    </label>
  );
}

function ScoresPanel({ game }: { game: GameId }) {
  const best = useArena((s) => s.best[game] ?? 0);
  const played = useArena((s) => s.played[game] ?? 0);
  // The store's own array, filtered here: a selector that built a new array
  // each call would make the store look changed on every render.
  const all = useArena((s) => s.history);
  const mine = useMemo(() => all.filter((h) => h.game === game), [all, game]);
  const history = mine.slice(0, 8);
  // The game's own high scores, best first.
  const top = useMemo(() => [...mine].sort((a, b) => b.score - a.score).slice(0, 5), [mine]);
  return (
    <div className="flex flex-col gap-4">
      <dl className="grid grid-cols-2 gap-3">
        <div className="bg-subtle rounded-xl p-3">
          <dt className="text-muted text-xs uppercase">Best</dt>
          <dd className="text-warning font-mono text-2xl font-bold">{best}</dd>
        </div>
        <div className="bg-subtle rounded-xl p-3">
          <dt className="text-muted text-xs uppercase">Played</dt>
          <dd className="font-mono text-2xl font-bold">{played}</dd>
        </div>
      </dl>
      {top.length ? (
        <div className="flex flex-col gap-1.5">
          <Heading
            level={4}
            className="text-sm font-semibold"
          >
            High scores
          </Heading>
          <ol className="flex flex-col gap-1 text-sm">
            {top.map((h, i) => (
              <li
                key={h.at}
                className="flex items-center gap-3 rounded-lg px-3 py-1.5"
              >
                <span className="text-muted w-5 font-mono">{i + 1}</span>
                <span className="min-w-0 flex-1 truncate">{h.note ?? ''}</span>
                <span className="text-warning font-mono font-semibold tabular-nums">{h.score}</span>
              </li>
            ))}
          </ol>
        </div>
      ) : null}
      {history.length ? (
        <Heading
          level={4}
          className="text-sm font-semibold"
        >
          The latest games
        </Heading>
      ) : null}
      {history.length ? (
        <ol className="flex flex-col gap-1.5 text-sm">
          {history.map((h) => (
            <li
              key={h.at}
              className="bg-subtle flex items-center justify-between gap-3 rounded-lg px-3 py-2"
            >
              <span className="text-muted min-w-0 truncate">{h.note ?? ''}</span>
              <span className="font-mono font-semibold tabular-nums">{h.score}</span>
            </li>
          ))}
        </ol>
      ) : (
        <p className="text-muted text-sm">No games yet. Your scores stay in this browser.</p>
      )}
    </div>
  );
}

/** A game's own menu, as a game has one. */
function MenuScreen({
  info,
  api,
  catalog,
  onPlay,
  onBack
}: {
  info: GameInfo;
  api: AlgorithmApi;
  catalog: Catalog;
  onPlay: () => void;
  onBack: () => void;
}) {
  const [index, setIndex] = useState(0);
  const item = MENU[index];
  const take = useCallback(
    (i: number) => {
      const m = MENU[i];
      if (m === 'play') onPlay();
      else if (m === 'back') onBack();
      else setIndex(i);
    },
    [onPlay, onBack]
  );
  useMenuKeys({ count: MENU.length, index, setIndex, onEnter: take, onBack });
  const Icon = info.icon;
  return (
    <div className="grid flex-1 gap-8 p-6 sm:p-10 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-3">
          <span
            className={cn(
              'bg-subtle ring-border grid size-12 place-items-center rounded-xl ring-1 [&_svg]:size-7!',
              info.tone
            )}
          >
            <Icon aria-hidden />
          </span>
          <Heading
            level={2}
            className="font-heading text-3xl font-bold tracking-tight"
          >
            {info.name}
          </Heading>
        </div>
        <nav
          aria-label={`${info.name} menu`}
          data-tour="game-menu"
        >
          <ul className="flex flex-col gap-1.5">
            {MENU.map((m, i) => {
              const ItemIcon = MENU_ICON[m];
              const on = i === index;
              return (
                <li key={m}>
                  <Button
                    onClick={() => take(i)}
                    onMouseEnter={() => setIndex(i)}
                    onFocus={() => setIndex(i)}
                    aria-current={on ? 'true' : undefined}
                    variant={on ? 'primary' : 'plain'}
                    size="sm"
                    className="w-full justify-start gap-2.5 rounded-lg px-3"
                  >
                    <ItemIcon aria-hidden />
                    {MENU_LABEL[m]}
                  </Button>
                </li>
              );
            })}
          </ul>
        </nav>
        <div className="mt-auto">
          <KeyHints
            hints={[
              ['↑ ↓', 'choose'],
              ['Enter', 'take'],
              ['Esc', 'back']
            ]}
          />
        </div>
      </div>
      <AnimatePresence
        mode="wait"
        initial={false}
      >
        <motion.section
          key={item}
          aria-label={MENU_LABEL[item]}
          initial={{ opacity: 0, x: 12 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: -8 }}
          transition={{ duration: 0.16 }}
          className="bg-subtle ring-border flex flex-col gap-4 rounded-2xl p-6 ring-1"
        >
          <Heading
            level={3}
            className="font-heading text-xl font-semibold"
          >
            {MENU_LABEL[item]}
          </Heading>
          {item === 'play' ? (
            <>
              <p className="text-muted">{info.tagline}</p>
              <Button
                variant="primary"
                size="lg"
                className="w-fit px-6"
                onClick={onPlay}
              >
                <LuPlay aria-hidden />
                Start
              </Button>
            </>
          ) : item === 'how' ? (
            <div className="text-muted [&_strong]:text-fg flex flex-col gap-3">{info.howTo}</div>
          ) : item === 'settings' ? (
            <SettingsPanel
              game={info.id}
              api={api}
              catalog={catalog}
            />
          ) : item === 'scores' ? (
            <ScoresPanel game={info.id} />
          ) : (
            <p className="text-muted">Back to every game.</p>
          )}
        </motion.section>
      </AnimatePresence>
    </div>
  );
}

/** The game, its bar, its pause menu and its end. */
function PlayScreen({
  info,
  api,
  catalog,
  onMenu
}: {
  info: GameInfo;
  api: AlgorithmApi;
  catalog: Catalog;
  onMenu: () => void;
}) {
  const [round, setRound] = useState(0);
  const [score, setScore] = useState(0);
  const [paused, setPaused] = useState(false);
  const [over, setOver] = useState<{ score: number; best: boolean } | null>(null);
  const best = useArena((s) => s.best[info.id] ?? 0);
  const record = useArena((s) => s.record);
  const recorded = useRef(false);

  const onOver = useCallback(
    (final: number, note?: string) => {
      if (recorded.current) return;
      recorded.current = true;
      const r = record(info.id, final, note);
      setOver({ score: final, best: r.best });
    },
    [info.id, record]
  );
  const restart = () => {
    recorded.current = false;
    setOver(null);
    setPaused(false);
    setScore(0);
    setRound((n) => n + 1);
  };

  // Escape pauses and resumes; the game holds still while paused.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape' || over) return;
      e.preventDefault();
      e.stopPropagation();
      setPaused((p) => !p);
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  }, [over]);

  const Game = info.Game;
  return (
    <div className="relative flex min-h-0 flex-1 flex-col gap-5 p-4 sm:p-6">
      {/* Room on the right for the full-screen key. */}
      <header className="flex flex-wrap items-center justify-between gap-3 pe-12">
        <Heading
          level={2}
          className="font-heading text-2xl font-bold"
        >
          {info.name}
        </Heading>
        <div className="flex items-center gap-4 font-mono text-sm">
          <span>
            Score <strong className="text-primary text-lg tabular-nums">{score}</strong>
          </span>
          <span className="text-muted">
            Best <strong className="text-warning tabular-nums">{best}</strong>
          </span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPaused(true)}
            disabled={!!over}
          >
            <LuPause aria-hidden />
            Pause
          </Button>
        </div>
      </header>
      <div className="flex min-h-0 flex-1 flex-col items-stretch">
        <Game
          key={round}
          api={api}
          catalog={catalog}
          paused={paused || !!over}
          onScore={setScore}
          onOver={onOver}
        />
      </div>
      {paused || over ? (
        <div
          role="dialog"
          aria-modal="true"
          aria-label={over ? 'Game over' : 'Paused'}
          className="bg-bg/85 absolute inset-0 z-20 grid place-items-center p-6 backdrop-blur-sm"
        >
          <div className="bg-subtle ring-border flex w-full max-w-sm flex-col items-center gap-5 rounded-3xl p-8 text-center ring-1">
            <p className="text-primary text-xs font-semibold tracking-[0.3em] uppercase">
              {over ? 'Game over' : 'Paused'}
            </p>
            {over ? (
              <>
                <p className="font-heading text-5xl font-bold tabular-nums">{over.score}</p>
                {over.best ? (
                  <p className="text-warning flex items-center gap-2 font-semibold">
                    <LuTrophy aria-hidden />A new best
                  </p>
                ) : (
                  <p className="text-muted text-sm">Best {best}</p>
                )}
                <p className="text-muted text-sm">+{over.score} points to your record</p>
              </>
            ) : null}
            <div className="flex w-full flex-col gap-2">
              {over ? null : (
                <Button
                  variant="primary"
                  className="h-11"
                  onClick={() => setPaused(false)}
                  autoFocus
                >
                  <LuPlay aria-hidden />
                  Resume
                </Button>
              )}
              <Button
                variant={over ? 'primary' : 'secondary'}
                className="h-11"
                onClick={restart}
                autoFocus={!!over}
              >
                <LuRotateCcw aria-hidden />
                {over ? 'Play again' : 'Restart'}
              </Button>
              <Button
                variant="ghost"
                className="h-11"
                onClick={onMenu}
              >
                <LuArrowLeft aria-hidden />
                Back to the menu
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function Arena({ api, catalog }: { api: AlgorithmApi; catalog: Catalog }) {
  const [q, setQ] = useQueryStates({
    game: parseAsString,
    screen: parseAsStringLiteral(SCREENS).withDefault('select')
  });
  const info = q.game ? gameInfo(q.game) : undefined;
  const screen = info ? q.screen : 'select';
  const reduce = useReducedMotionConfig() ?? false;

  return (
    <ArenaScreen
      label={
        screen === 'select'
          ? 'The arena: choose a game'
          : `${info?.name}: ${screen === 'menu' ? 'menu' : 'playing'}`
      }
    >
      <AnimatePresence
        mode="wait"
        initial={false}
      >
        <motion.div
          key={`${screen}-${info?.id ?? ''}`}
          className="flex min-h-0 flex-1 flex-col"
          initial={reduce ? false : { opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={reduce ? { opacity: 0, transition: { duration: 0 } } : { opacity: 0, scale: 1.02 }}
          transition={{ duration: 0.18 }}
        >
          {screen === 'select' || !info ? (
            <SelectScreen onPick={(id) => void setQ({ game: id, screen: 'menu' })} />
          ) : screen === 'menu' ? (
            <MenuScreen
              info={info}
              api={api}
              catalog={catalog}
              onPlay={() => void setQ({ screen: 'play' })}
              onBack={() => void setQ({ game: null, screen: null })}
            />
          ) : (
            <PlayScreen
              info={info}
              api={api}
              catalog={catalog}
              onMenu={() => void setQ({ screen: 'menu' })}
            />
          )}
        </motion.div>
      </AnimatePresence>
    </ArenaScreen>
  );
}

export function GameArena() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Arena
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
