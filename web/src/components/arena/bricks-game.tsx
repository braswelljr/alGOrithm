'use client';

/**
 * Bricks, a falling-blocks game, drawn on a canvas. The rules are in
 * lib/arena/bricks.ts; this keeps time and takes the keys and the mouse.
 * Keys: left and right move, up or X turns, down drops a row, Space drops to
 * the bottom. Mouse: moving over the well slides the piece under the pointer,
 * a click turns it, a right click drops it. A pad of keycaps does the same on
 * a touch screen.
 */
import { useEffect, useRef, useState } from 'react';
import type React from 'react';
import { LuArrowDown, LuArrowLeft, LuArrowRight, LuChevronsDown, LuRotateCw } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import {
  cells,
  drop,
  fall,
  fallMs,
  ghost,
  newBricks,
  rotate,
  shift,
  type BricksState,
  type PieceKind
} from '@/lib/arena/bricks';
import { rng } from '@/lib/arena/snake';
import { cellCap, useArenaFull } from './arena-screen';
import type { GameProps } from './games';

const ROWS = 20;
const COLS = 10;
const COLOR: Record<PieceKind, string> = {
  I: '#00add8',
  O: '#fbbf24',
  T: '#a855f7',
  S: '#10b981',
  Z: '#ef4444',
  J: '#6366f1',
  L: '#f97316'
};

export function BricksGame({ paused, onScore, onOver }: GameProps) {
  const [{ rand, first }] = useState(() => {
    const r = rng(Date.now() & 0xffff);
    return { rand: r, first: newBricks(ROWS, COLS, r) };
  });
  const [game, setGame] = useState<BricksState>(first);
  const over = useRef(false);

  const act = (f: (s: BricksState) => BricksState) => {
    if (paused) return;
    setGame((s) => f(s));
  };

  // Gravity, faster each level.
  useEffect(() => {
    if (paused || game.over) return;
    const timer = window.setTimeout(() => setGame((s) => fall(s, rand)), fallMs(game.lines));
    return () => window.clearTimeout(timer);
  }, [game, paused, rand]);

  useEffect(() => {
    onScore(game.score);
    if (game.over && !over.current) {
      over.current = true;
      onOver(game.score, `${game.lines} lines, level ${1 + Math.floor(game.lines / 10)}`);
    }
  }, [game, onScore, onOver]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const map: Record<string, (s: BricksState) => BricksState> = {
        ArrowLeft: (s) => shift(s, 0, -1),
        ArrowRight: (s) => shift(s, 0, 1),
        ArrowDown: (s) => fall(s, rand),
        ArrowUp: rotate,
        x: rotate,
        ' ': (s) => drop(s, rand)
      };
      const f = map[e.key] ?? map[e.key.toLowerCase()];
      if (!f || paused) return;
      e.preventDefault();
      e.stopPropagation();
      setGame((s) => f(s));
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  }, [paused, rand]);

  return (
    <div className="flex min-h-0 w-full flex-1 flex-col items-center gap-4 lg:flex-row lg:items-start lg:justify-center">
      <Well
        game={game}
        onAct={act}
        rand={rand}
      />
      <aside
        aria-label="Bricks: next and score"
        className="flex flex-row flex-wrap gap-3 lg:w-48 lg:flex-col"
      >
        <div className="bg-subtle ring-border rounded-xl p-3 ring-1">
          <p className="text-muted mb-2 text-xs tracking-widest uppercase">Next</p>
          <Preview kind={game.next} />
        </div>
        <dl className="bg-subtle ring-border grid grid-cols-2 gap-2 rounded-xl p-3 font-mono text-sm ring-1 lg:grid-cols-1">
          <div>
            <dt className="text-muted text-xs uppercase">Lines</dt>
            <dd className="text-lg font-bold tabular-nums">{game.lines}</dd>
          </div>
          <div>
            <dt className="text-muted text-xs uppercase">Level</dt>
            <dd className="text-lg font-bold tabular-nums">{1 + Math.floor(game.lines / 10)}</dd>
          </div>
        </dl>
        {/* A pad of keycaps for touch screens; the keys do the same. */}
        <div
          role="group"
          aria-label="Controls"
          className="grid grid-cols-4 gap-1.5 lg:grid-cols-2"
        >
          <Pad
            label="Left"
            icon={LuArrowLeft}
            onPress={() => act((s) => shift(s, 0, -1))}
          />
          <Pad
            label="Turn"
            icon={LuRotateCw}
            onPress={() => act(rotate)}
          />
          <Pad
            label="Right"
            icon={LuArrowRight}
            onPress={() => act((s) => shift(s, 0, 1))}
          />
          <Pad
            label="Down"
            icon={LuArrowDown}
            onPress={() => act((s) => fall(s, rand))}
          />
          <Button
            variant="primary"
            className="col-span-4 lg:col-span-2"
            onClick={() => act((s) => drop(s, rand))}
          >
            <LuChevronsDown aria-hidden />
            Drop
          </Button>
        </div>
      </aside>
    </div>
  );
}

function Pad({
  label,
  icon: Icon,
  onPress
}: {
  label: string;
  icon: React.ComponentType<{ 'aria-hidden'?: boolean }>;
  onPress: () => void;
}) {
  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label={label}
      onClick={onPress}
    >
      <Icon aria-hidden />
    </Button>
  );
}

/** A cell as a small bevelled brick. */
function brick(g: CanvasRenderingContext2D, x: number, y: number, size: number, color: string) {
  g.fillStyle = color;
  g.fillRect(x + 1, y + 1, size - 2, size - 2);
  g.fillStyle = 'rgba(255,255,255,0.28)';
  g.fillRect(x + 1, y + 1, size - 2, Math.max(2, size * 0.18));
  g.fillStyle = 'rgba(0,0,0,0.25)';
  g.fillRect(x + 1, y + size - 1 - Math.max(2, size * 0.14), size - 2, Math.max(2, size * 0.14));
}

function Well({
  game,
  onAct,
  rand
}: {
  game: BricksState;
  onAct: (f: (s: BricksState) => BricksState) => void;
  rand: () => number;
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

  const cell = space.w
    ? Math.max(12, Math.min(space.w / COLS, (space.h || 600) / ROWS, cellCap(full) + 4))
    : 0;

  useEffect(() => {
    const c = canvas.current;
    if (!c || !cell) return;
    const w = cell * COLS;
    const h = cell * ROWS;
    const dpr = window.devicePixelRatio || 1;
    c.width = Math.round(w * dpr);
    c.height = Math.round(h * dpr);
    c.style.width = `${w}px`;
    c.style.height = `${h}px`;
    const g = c.getContext('2d');
    if (!g) return;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    g.fillStyle = '#0c0a09';
    g.fillRect(0, 0, w, h);
    g.strokeStyle = 'rgba(255,255,255,0.04)';
    for (let x = 0; x <= COLS; x++) {
      g.beginPath();
      g.moveTo(x * cell, 0);
      g.lineTo(x * cell, h);
      g.stroke();
    }
    for (let y = 0; y <= ROWS; y++) {
      g.beginPath();
      g.moveTo(0, y * cell);
      g.lineTo(w, y * cell);
      g.stroke();
    }
    game.well.forEach((k, i) => {
      if (k) brick(g, (i % COLS) * cell, Math.floor(i / COLS) * cell, cell, COLOR[k]);
    });
    if (!game.over) {
      // The ghost: where the piece will land.
      g.strokeStyle = COLOR[game.piece.kind];
      g.globalAlpha = 0.5;
      for (const [r, col] of cells(ghost(game))) {
        if (r >= 0) g.strokeRect(col * cell + 2, r * cell + 2, cell - 4, cell - 4);
      }
      g.globalAlpha = 1;
      for (const [r, col] of cells(game.piece)) {
        if (r >= 0) brick(g, col * cell, r * cell, cell, COLOR[game.piece.kind]);
      }
    }
  }, [game, cell]);

  // The mouse: move to slide the piece under the pointer, click to turn, right click to drop.
  const onMove = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (e.pointerType !== 'mouse' || !cell) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const want = Math.floor((e.clientX - rect.left) / cell);
    const centre = Math.round(
      cells(game.piece).reduce((a, [, c]) => a + c, 0) / cells(game.piece).length - 0.5
    );
    if (want !== centre) onAct((s) => shift(s, 0, Math.sign(want - centre)));
  };

  return (
    <div
      ref={box}
      className="flex min-h-[20rem] w-full max-w-md flex-1 items-start justify-center"
    >
      <canvas
        ref={canvas}
        role="img"
        aria-label={`The well: ${game.lines} lines cleared`}
        onPointerMove={onMove}
        onClick={() => onAct(rotate)}
        onContextMenu={(e) => {
          e.preventDefault();
          onAct((s) => drop(s, rand));
        }}
        className="ring-border block cursor-pointer touch-none rounded-lg ring-2"
      />
    </div>
  );
}

/** The next piece, small. */
function Preview({ kind }: { kind: PieceKind }) {
  const canvas = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const c = canvas.current;
    const g = c?.getContext('2d');
    if (!c || !g) return;
    const size = 18;
    const dpr = window.devicePixelRatio || 1;
    c.width = 4 * size * dpr;
    c.height = 2 * size * dpr;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    g.clearRect(0, 0, 4 * size, 2 * size);
    for (const [r, col] of cells({ kind, turn: 0, row: 0, col: 0 })) {
      brick(g, col * size, (r - (kind === 'I' ? 1 : 0)) * size, size, COLOR[kind]);
    }
  }, [kind]);
  return (
    <canvas
      ref={canvas}
      aria-label={`Next: ${kind}`}
      role="img"
      className="h-9 w-[72px]"
    />
  );
}
