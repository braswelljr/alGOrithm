'use client';

/**
 * Snake, played by hand or by a pathfinder from the catalog. The rules are in
 * lib/arena/snake.ts; this draws the board, keeps time and takes the keys. An
 * autopilot runs the chosen pathfinder (the real Go code, through the wasm)
 * on the board every step and takes the first move of its path.
 */
import { useEffect, useMemo, useRef, useState } from 'react';
import type React from 'react';
import { LuArrowDown, LuArrowLeft, LuArrowRight, LuArrowUp } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import {
  gridText,
  newGame,
  pilotMove,
  rng,
  step,
  type Dir,
  type SnakeState
} from '@/lib/arena/snake';
import { useArena } from '@/lib/arena/store';
import { cn } from '@/lib/utils';
import { cellCap, useArenaFull } from './arena-screen';
import type { GameProps } from './games';

const COLS = 22;
const ROWS = 14;
const TICK = { calm: 170, brisk: 115, fast: 75 } as const;
const MULTIPLIER = { calm: 1, brisk: 1.5, fast: 2 } as const;
const KEYS: Record<string, Dir> = {
  ArrowUp: 'up',
  ArrowDown: 'down',
  ArrowLeft: 'left',
  ArrowRight: 'right',
  w: 'up',
  s: 'down',
  a: 'left',
  d: 'right'
};

/** The pathfinders an autopilot may be: every entry that runs on a grid. */
export function pilotsOf(catalog: GameProps['catalog']) {
  return catalog.entries.filter((e) => e.input === 'grid');
}

export function SnakeGame({ api, catalog, paused, onScore, onOver }: GameProps) {
  const settings = useArena((s) => s.settings.snake);
  const pilot = useMemo(
    () => pilotsOf(catalog).find((e) => e.id === settings.pilot) ?? null,
    [catalog, settings.pilot]
  );
  // A generator seeded once per game, and the board it starts.
  const [{ rand, first }] = useState(() => {
    const r = rng(Date.now() & 0xffff);
    return { rand: r, first: newGame(COLS, ROWS, settings.walls, r) };
  });
  const [game, setGame] = useState<SnakeState>(first);
  const turn = useRef<Dir | null>(null);
  const over = useRef(false);
  const [visited, setVisited] = useState(0);

  const scoreOf = (s: SnakeState) =>
    Math.round(
      s.eaten * 10 * MULTIPLIER[settings.speed] * (settings.walls ? 1 : 0.8) * (pilot ? 0.25 : 1)
    );

  // Keys steer; held back from the site's shortcuts while playing.
  useEffect(() => {
    if (pilot) return;
    const onKey = (e: KeyboardEvent) => {
      const d = KEYS[e.key] ?? KEYS[e.key.toLowerCase()];
      if (!d || paused) return;
      e.preventDefault();
      e.stopPropagation();
      turn.current = d;
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  }, [paused, pilot]);

  // The clock: one step a tick, the autopilot planning before each.
  useEffect(() => {
    if (paused || !game.alive) return;
    const timer = window.setTimeout(() => {
      let dir = turn.current;
      turn.current = null;
      if (pilot) {
        const r = api.run(pilot.id, { text: gridText(game) }, 20_000);
        if (r.ok) {
          dir = pilotMove(game, r.value.output.path);
          setVisited(r.value.counts?.visit ?? 0);
        } else dir = pilotMove(game, []);
      }
      const next = step(game, dir, rand);
      setGame(next);
      onScore(scoreOf(next));
      if (!next.alive && !over.current) {
        over.current = true;
        onOver(
          scoreOf(next),
          `${next.eaten} eaten${pilot ? `, piloted by ${pilot.name}` : ''}, ${settings.speed}`
        );
      }
    }, TICK[settings.speed]);
    return () => window.clearTimeout(timer);
    // scoreOf reads the settings already listed.
  }, [game, paused, pilot, api, settings.speed, onScore, onOver]);

  return (
    <div className="flex min-h-0 w-full flex-1 flex-col items-center gap-4">
      <div className="text-muted flex w-full flex-wrap items-center justify-between gap-2 text-sm">
        <span>
          Length <strong className="text-fg tabular-nums">{game.body.length}</strong>
        </span>
        {pilot ? (
          <span>
            Autopilot <strong className="text-primary">{pilot.name}</strong>
            {visited ? ` · looked at ${visited} cells for this move` : ''}
          </span>
        ) : (
          <span>
            Speed <strong className="text-fg">{settings.speed}</strong>
            {settings.walls ? '' : ' · edges wrap'} · keys, or press and drag on the board
          </span>
        )}
      </div>
      <Board
        game={game}
        label={`Snake, ${game.body.length} long, ${game.eaten} eaten`}
        onSteer={pilot ? undefined : (d) => (turn.current = d)}
      />
      {/* A pad for touch screens; the arrow keys do the same. */}
      {pilot ? null : (
        <div
          className="grid grid-cols-3 gap-1.5 sm:hidden"
          role="group"
          aria-label="Steer"
        >
          <span />
          <Pad
            dir="up"
            onTurn={(d) => (turn.current = d)}
          />
          <span />
          <Pad
            dir="left"
            onTurn={(d) => (turn.current = d)}
          />
          <Pad
            dir="down"
            onTurn={(d) => (turn.current = d)}
          />
          <Pad
            dir="right"
            onTurn={(d) => (turn.current = d)}
          />
        </div>
      )}
    </div>
  );
}

/**
 * The board, drawn on a canvas at the screen's pixel density: the food
 * glowing, the body in a green that fades toward the tail, the head in the
 * primary with its eyes on the way it is going. Pressing or dragging on it
 * steers toward the pointer.
 */
function Board({
  game,
  label,
  onSteer
}: {
  game: SnakeState;
  label: string;
  onSteer?: (d: Dir) => void;
}) {
  const box = useRef<HTMLDivElement>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const full = useArenaFull();
  const [space, setSpace] = useState({ w: 0, h: 0 });

  useEffect(() => {
    const el = box.current;
    if (!el || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(([e]) =>
      setSpace({ w: e.contentRect.width, h: e.contentRect.height })
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  // The cell: as big as the space allows, up to the cap.
  const cell = space.w
    ? Math.max(8, Math.min(space.w / game.cols, (space.h || space.w) / game.rows, cellCap(full)))
    : 0;
  const width = cell * game.cols;

  useEffect(() => {
    const c = canvas.current;
    if (!c || !cell) return;
    const dpr = window.devicePixelRatio || 1;
    const height = cell * game.rows;
    c.width = Math.round(width * dpr);
    c.height = Math.round(height * dpr);
    c.style.width = `${width}px`;
    c.style.height = `${height}px`;
    const g = c.getContext('2d');
    if (!g) return;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    g.clearRect(0, 0, width, height);
    // The floor: a faint checker.
    for (let i = 0; i < game.cols * game.rows; i++) {
      const x = (i % game.cols) * cell;
      const y = Math.floor(i / game.cols) * cell;
      g.fillStyle = (i + Math.floor(i / game.cols)) % 2 ? '#1c1917' : '#171412';
      g.fillRect(x, y, cell, cell);
    }
    const at = (i: number) => [(i % game.cols) * cell, Math.floor(i / game.cols) * cell] as const;
    const r = cell * 0.22;
    const box = (x: number, y: number, inset: number) => {
      g.beginPath();
      g.roundRect(x + inset, y + inset, cell - inset * 2, cell - inset * 2, r);
      g.fill();
    };
    // The food, glowing.
    if (game.food >= 0) {
      const [x, y] = at(game.food);
      g.shadowColor = '#fbbf24';
      g.shadowBlur = cell * 0.8;
      g.fillStyle = '#fbbf24';
      g.beginPath();
      g.arc(x + cell / 2, y + cell / 2, cell * 0.32, 0, Math.PI * 2);
      g.fill();
      g.shadowBlur = 0;
    }
    // The body, fading toward the tail.
    game.body.slice(1).forEach((i, k) => {
      const [x, y] = at(i);
      const fade = 1 - (k / Math.max(game.body.length, 2)) * 0.55;
      g.fillStyle = `rgba(16, 185, 129, ${fade})`;
      box(x, y, cell * 0.08);
    });
    // The head, in the primary, with eyes on its way.
    const [hx, hy] = at(game.body[0]);
    g.shadowColor = '#00add8';
    g.shadowBlur = cell * 0.6;
    g.fillStyle = game.alive ? '#00add8' : '#ef4444';
    box(hx, hy, cell * 0.04);
    g.shadowBlur = 0;
    g.fillStyle = '#0c0a09';
    const eye = cell * 0.09;
    const [ex, ey] =
      game.dir === 'right'
        ? [0.68, 0.5]
        : game.dir === 'left'
          ? [0.32, 0.5]
          : game.dir === 'up'
            ? [0.5, 0.32]
            : [0.5, 0.68];
    const side = game.dir === 'up' || game.dir === 'down' ? [0.2, 0] : [0, 0.2];
    for (const s of [-1, 1]) {
      g.beginPath();
      g.arc(hx + (ex + side[0] * s) * cell, hy + (ey + side[1] * s) * cell, eye, 0, Math.PI * 2);
      g.fill();
    }
  }, [game, width, cell]);

  // Press or drag: turn toward the pointer, along the farther axis.
  const steer = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (!onSteer || (e.type === 'pointermove' && e.buttons === 0)) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const cell = rect.width / game.cols;
    const hx = ((game.body[0] % game.cols) + 0.5) * cell;
    const hy = (Math.floor(game.body[0] / game.cols) + 0.5) * cell;
    const dx = e.clientX - rect.left - hx;
    const dy = e.clientY - rect.top - hy;
    onSteer(Math.abs(dx) > Math.abs(dy) ? (dx > 0 ? 'right' : 'left') : dy > 0 ? 'down' : 'up');
  };

  return (
    <div
      ref={box}
      className="flex min-h-0 w-full flex-1 items-center justify-center"
    >
      <canvas
        ref={canvas}
        role="img"
        aria-label={label}
        onPointerDown={steer}
        onPointerMove={steer}
        className={cn(
          'ring-border block touch-none rounded-lg ring-2',
          onSteer && 'cursor-crosshair'
        )}
      />
    </div>
  );
}

const PAD_ICON = { up: LuArrowUp, down: LuArrowDown, left: LuArrowLeft, right: LuArrowRight };

function Pad({ dir, onTurn }: { dir: Dir; onTurn: (d: Dir) => void }) {
  const Icon = PAD_ICON[dir];
  return (
    <Button
      variant="secondary"
      size="icon"
      aria-label={`Turn ${dir}`}
      onPointerDown={() => onTurn(dir)}
    >
      <Icon aria-hidden />
    </Button>
  );
}
