'use client';

/**
 * Path race: you against every pathfinder in the catalog, across a generated
 * map from S to E. Each rival drives the path its algorithm found (the real Go
 * code, through the wasm) after waiting the time it spent thinking, the cells
 * it looked at; entering a cell costs its weight, so a dear cell slows every
 * car. You drive with the arrow keys, W A S D, or by pressing and dragging
 * toward where you want to go. The clock is in lib/arena/race.ts.
 */
import { useEffect, useMemo, useRef, useState } from 'react';
import type React from 'react';
import { LuFlag, LuTrophy } from 'react-icons/lu';
import { Heading } from '@/components/ui/heading';
import { arrivals, places, positionAt, racePoints } from '@/lib/arena/race';
import { useArena } from '@/lib/arena/store';
import { parseGridText } from '@/lib/grid-text';
import { cn } from '@/lib/utils';
import type { Grid } from '@/visualizer/core';
import { cellCap, useArenaFull } from './arena-screen';
import type { GameProps } from './games';

const SIZES = { small: [11, 21], medium: [15, 31], large: [21, 41] } as const;
/** A step into a cell of weight 1, in milliseconds. */
const STEP_MS = 140;
/** The thinking time for each cell a rival's algorithm looked at. */
const THINK_MS_PER_CELL = 6;
const COUNTDOWN_MS = 3000;
const RIVAL_COLORS = ['#f59e0b', '#a855f7', '#ef4444', '#10b981', '#ec4899', '#eab308', '#6366f1'];
const KEYS: Record<string, [number, number]> = {
  ArrowUp: [-1, 0],
  ArrowDown: [1, 0],
  ArrowLeft: [0, -1],
  ArrowRight: [0, 1],
  w: [-1, 0],
  s: [1, 0],
  a: [0, -1],
  d: [0, 1]
};

type Rival = {
  id: string;
  name: string;
  color: string;
  path: number[];
  times: number[];
  finish: number;
  looked: number;
};

/** Where the player's car is: a cell, or on its way into the next. */
type Car = { from: number; to: number; leftAt: number; arriveAt: number; finish: number | null };

export function PathRace({ api, catalog, paused, onScore, onOver }: GameProps) {
  const settings = useArena((s) => s.settings.race);
  const [rows, cols] = SIZES[settings.size];

  // The map and the rivals' routes, once per race.
  const [seed] = useState(() => Math.floor(Math.random() * 1e6));
  const grid = useMemo<Grid | null>(() => {
    const kinds = catalog.gridKinds.length ? catalog.gridKinds : ['weighted'];
    const kind = kinds.includes(settings.map) ? settings.map : kinds[0];
    const r = api.generateGrid(kind, rows, cols, seed);
    return r.ok ? parseGridText(r.value) : null;
  }, [api, catalog.gridKinds, settings.map, rows, cols, seed]);

  const rivals = useMemo<Rival[]>(() => {
    if (!grid) return [];
    const text = gridLines(grid);
    return catalog.entries
      .filter((e) => e.input === 'grid')
      .flatMap((e, i) => {
        const r = api.run(e.id, { text }, 50_000);
        if (!r.ok || r.value.output.path.length < 2) return [];
        const path = r.value.output.path;
        const looked = r.value.counts?.visit ?? path.length;
        const times = arrivals(path, grid.cost, STEP_MS, looked * THINK_MS_PER_CELL);
        return [
          {
            id: e.id,
            name: e.name,
            color: RIVAL_COLORS[i % RIVAL_COLORS.length],
            path,
            times,
            finish: times[times.length - 1],
            looked
          }
        ];
      });
  }, [api, catalog.entries, grid]);

  // The race clock: milliseconds since the green light, held while paused.
  const clock = useRef({ elapsed: -COUNTDOWN_MS, last: 0 });
  const [now, setNow] = useState(-COUNTDOWN_MS);
  const car = useRef<Car | null>(null);
  const held = useRef<[number, number] | null>(null);
  const [result, setResult] = useState<{ time: number; place: number } | null>(null);
  // Where the player's car is drawn, set each frame from the ref.
  const [mine, setMine] = useState<Spot>(null);

  useEffect(() => {
    if (grid && !car.current)
      car.current = { from: grid.start, to: grid.start, leftAt: 0, arriveAt: 0, finish: null };
  }, [grid]);

  // Keys drive while held.
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      const d = KEYS[e.key] ?? KEYS[e.key.toLowerCase()];
      if (!d) return;
      e.preventDefault();
      e.stopPropagation();
      held.current = d;
    };
    const up = (e: KeyboardEvent) => {
      const d = KEYS[e.key] ?? KEYS[e.key.toLowerCase()];
      if (d && held.current === d) held.current = null;
    };
    window.addEventListener('keydown', down, { capture: true });
    window.addEventListener('keyup', up, { capture: true });
    return () => {
      window.removeEventListener('keydown', down, { capture: true });
      window.removeEventListener('keyup', up, { capture: true });
    };
  }, []);

  // One frame: the clock, then the player's next cell while a way is held.
  useEffect(() => {
    if (!grid || result) return;
    let frame = 0;
    const tick = (ts: number) => {
      frame = requestAnimationFrame(tick);
      const c = clock.current;
      if (!c.last) c.last = ts;
      const dt = Math.min(ts - c.last, 100);
      c.last = ts;
      if (paused) return;
      c.elapsed += dt;
      setNow(c.elapsed);
      const me = car.current;
      if (me) {
        setMine(
          me.to === me.from || c.elapsed <= me.leftAt
            ? { from: me.from, to: me.from, along: 0 }
            : {
                from: me.from,
                to: me.to,
                along: Math.min(1, (c.elapsed - me.leftAt) / (me.arriveAt - me.leftAt))
              }
        );
      }
      if (!me || c.elapsed < 0 || me.finish !== null) return;
      if (c.elapsed >= me.arriveAt) {
        me.from = me.to;
        if (me.from === grid.end) {
          me.finish = me.arriveAt;
          const all = [me.finish, ...rivals.map((r) => r.finish)];
          const place = places(all)[0];
          setResult({ time: me.finish, place });
          const score = racePoints(place, all.length);
          onScore(score);
          onOver(score, `place ${place} of ${all.length} in ${(me.finish / 1000).toFixed(1)}s`);
          return;
        }
        const d = held.current;
        if (d) {
          const r = Math.floor(me.from / grid.cols) + d[0];
          const col = (me.from % grid.cols) + d[1];
          const next = r * grid.cols + col;
          if (r >= 0 && r < grid.rows && col >= 0 && col < grid.cols && grid.cost[next] > 0) {
            me.to = next;
            me.leftAt = c.elapsed;
            me.arriveAt = c.elapsed + grid.cost[next] * STEP_MS;
          }
        }
      }
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [grid, rivals, paused, result, onScore, onOver]);

  if (!grid) return <p className="text-muted">The map could not be generated.</p>;

  // Press or drag: drive toward the pointer, along the farther axis.
  const steer = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (e.type === 'pointerup' || e.type === 'pointerleave') {
      held.current = null;
      return;
    }
    if (e.type === 'pointermove' && e.buttons === 0) return;
    const me = car.current;
    if (!me) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const cell = rect.width / grid.cols;
    const cx = ((me.from % grid.cols) + 0.5) * cell;
    const cy = (Math.floor(me.from / grid.cols) + 0.5) * cell;
    const dx = e.clientX - rect.left - cx;
    const dy = e.clientY - rect.top - cy;
    held.current = Math.abs(dx) > Math.abs(dy) ? [0, dx > 0 ? 1 : -1] : [dy > 0 ? 1 : -1, 0];
  };

  const standings = [
    ...(result
      ? [{ id: 'you', name: 'You', color: '#00add8', finish: result.time, looked: 0 }]
      : []),
    ...rivals
  ].sort((a, b) => a.finish - b.finish);

  return (
    <div className="flex min-h-0 w-full flex-1 flex-col gap-4 lg:flex-row">
      <div className="flex min-h-0 flex-1 flex-col gap-3">
        <div className="text-muted flex flex-wrap items-center justify-between gap-2 font-mono text-sm">
          <span>
            {now < 0 ? (
              <strong className="text-primary text-lg">Ready... {Math.ceil(-now / 1000)}</strong>
            ) : (
              <>
                Time{' '}
                <strong className="text-fg tabular-nums">
                  {(Math.max(0, now) / 1000).toFixed(1)}s
                </strong>
              </>
            )}
          </span>
          <span className="text-muted">Arrow keys, W A S D, or press and drag toward the way</span>
        </div>
        <RaceTrack
          grid={grid}
          rivals={rivals}
          now={now}
          me={mine}
          onPointer={steer}
        />
      </div>
      <aside
        aria-label="The field"
        className="flex shrink-0 flex-col gap-2 lg:w-72"
      >
        <Heading
          level={3}
          className="font-heading text-fg flex items-center gap-2 text-lg font-semibold"
        >
          <LuFlag aria-hidden />
          {result ? 'Results' : 'The field'}
        </Heading>
        <ol className="flex flex-col gap-1.5 text-sm">
          {(result ? standings : rivals).map((r, i) => {
            const there = now >= r.finish;
            return (
              <li
                key={r.id}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-3 py-2',
                  r.id === 'you' ? 'bg-primary/20 text-fg' : 'bg-subtle text-muted'
                )}
              >
                {result ? <span className="text-muted w-5 font-mono">{i + 1}</span> : null}
                <span
                  aria-hidden
                  className="size-3 shrink-0 rounded-full"
                  style={{ background: r.color }}
                />
                <span className="min-w-0 flex-1 truncate">{r.name}</span>
                <span className="font-mono text-xs tabular-nums">
                  {result || there
                    ? `${(r.finish / 1000).toFixed(1)}s`
                    : r.looked
                      ? `thinks ${r.looked}`
                      : ''}
                </span>
                {result && i === 0 ? (
                  <LuTrophy
                    aria-hidden
                    className="text-warning"
                  />
                ) : null}
              </li>
            );
          })}
        </ol>
        <p className="text-muted text-xs">
          Each rival waits for the cells its algorithm looked at, then drives the path it found. A
          dear cell, a 9, is nine times as slow to enter.
        </p>
      </aside>
    </div>
  );
}

/** The grid as text again, for the pathfinders. */
function gridLines(g: Grid): string {
  const out: string[] = [];
  for (let r = 0; r < g.rows; r++) {
    let line = '';
    for (let c = 0; c < g.cols; c++) {
      const i = r * g.cols + c;
      line +=
        i === g.start
          ? 'S'
          : i === g.end
            ? 'E'
            : g.cost[i] === 0
              ? '#'
              : g.cost[i] === 1
                ? '.'
                : String(g.cost[i]);
    }
    out.push(line);
  }
  return out.join('\n');
}

type Spot = { from: number; to: number; along: number } | null;

/**
 * The track on a canvas, fitted to the space it is given: walls, the cost of
 * each cell as a warmer floor, the start and the flag, every rival's car and
 * its trail, and the player's car in the primary.
 */
function RaceTrack({
  grid,
  rivals,
  now,
  me,
  onPointer
}: {
  grid: Grid;
  rivals: Rival[];
  now: number;
  me: Spot;
  onPointer: (e: React.PointerEvent<HTMLCanvasElement>) => void;
}) {
  const box = useRef<HTMLDivElement>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const [space, setSpace] = useState({ w: 0, h: 0 });
  const full = useArenaFull();

  useEffect(() => {
    const el = box.current;
    if (!el || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(([e]) =>
      setSpace({ w: e.contentRect.width, h: e.contentRect.height })
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const c = canvas.current;
    if (!c || !space.w) return;
    const cell = Math.max(
      4,
      Math.min(space.w / grid.cols, (space.h || space.w) / grid.rows, cellCap(full))
    );
    const w = cell * grid.cols;
    const h = cell * grid.rows;
    const dpr = window.devicePixelRatio || 1;
    c.width = Math.round(w * dpr);
    c.height = Math.round(h * dpr);
    c.style.width = `${w}px`;
    c.style.height = `${h}px`;
    const g = c.getContext('2d');
    if (!g) return;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    const xy = (i: number) => [(i % grid.cols) * cell, Math.floor(i / grid.cols) * cell] as const;
    for (let i = 0; i < grid.cost.length; i++) {
      const [x, y] = xy(i);
      const k = grid.cost[i];
      g.fillStyle =
        k === 0 ? '#0c0a09' : k === 1 ? '#292524' : `rgba(180, 83, 9, ${0.18 + k * 0.07})`;
      g.fillRect(x, y, cell, cell);
      if (k === 0) {
        g.fillStyle = '#44403c';
        g.fillRect(x + cell * 0.1, y + cell * 0.1, cell * 0.8, cell * 0.8);
      }
    }
    const flag = (i: number, color: string) => {
      const [x, y] = xy(i);
      g.fillStyle = color;
      g.fillRect(x + cell * 0.15, y + cell * 0.15, cell * 0.7, cell * 0.7);
    };
    flag(grid.start, '#57534e');
    // The finish: a checkered square.
    {
      const [x, y] = xy(grid.end);
      const q = cell / 4;
      for (let a = 0; a < 4; a++)
        for (let b = 0; b < 4; b++) {
          g.fillStyle = (a + b) % 2 ? '#fafaf9' : '#0c0a09';
          g.fillRect(x + a * q, y + b * q, q, q);
        }
    }
    const centre = (spot: { from: number; to: number; along: number }) => {
      const [ax, ay] = xy(spot.from);
      const [bx, by] = xy(spot.to);
      return [
        ax + (bx - ax) * spot.along + cell / 2,
        ay + (by - ay) * spot.along + cell / 2
      ] as const;
    };
    const drawCar = (x: number, y: number, color: string, size: number) => {
      g.shadowColor = color;
      g.shadowBlur = cell * 0.6;
      g.fillStyle = color;
      g.beginPath();
      g.arc(x, y, size, 0, Math.PI * 2);
      g.fill();
      g.shadowBlur = 0;
    };
    for (const r of rivals) {
      const spot = positionAt(r.path, r.times, Math.max(0, now));
      if (!spot) continue;
      // The trail so far.
      const reached = r.times.findIndex((t) => t > now);
      const upto = reached < 0 ? r.path.length : reached;
      g.strokeStyle = r.color;
      g.globalAlpha = 0.35;
      g.lineWidth = Math.max(1.5, cell * 0.14);
      g.beginPath();
      r.path.slice(0, upto).forEach((p, k) => {
        const [x, y] = xy(p);
        if (k === 0) g.moveTo(x + cell / 2, y + cell / 2);
        else g.lineTo(x + cell / 2, y + cell / 2);
      });
      g.stroke();
      g.globalAlpha = 1;
      const [x, y] = centre(spot);
      drawCar(x, y, r.color, cell * 0.26);
    }
    if (me) {
      const [x, y] = centre(me);
      drawCar(x, y, '#00add8', cell * 0.34);
      g.fillStyle = '#0c0a09';
      g.beginPath();
      g.arc(x, y, cell * 0.12, 0, Math.PI * 2);
      g.fill();
    }
  }, [grid, rivals, now, me, space, full]);

  return (
    <div
      ref={box}
      className="flex min-h-[18rem] w-full flex-1 items-center justify-center"
    >
      <canvas
        ref={canvas}
        role="img"
        aria-label="The race track"
        onPointerDown={onPointer}
        onPointerMove={onPointer}
        onPointerUp={onPointer}
        onPointerLeave={onPointer}
        className="ring-border block cursor-crosshair touch-none rounded-lg ring-1"
      />
    </div>
  );
}
