'use client';

/**
 * The path race's map, chosen by its look: each kind of map the Go side can
 * generate, drawn small on a canvas from a fixed seed (walls dark, dearer
 * cells warmer, S and the flag marked), as a radio group of keycaps.
 */
import { useEffect, useRef } from 'react';
import { parseGridText } from '@/lib/grid-text';
import { cn } from '@/lib/utils';
import type { AlgorithmApi } from '@/visualizer/core';

const ROWS = 11;
const COLS = 21;

function Thumb({ api, kind }: { api: AlgorithmApi; kind: string }) {
  const canvas = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const c = canvas.current;
    const g = c?.getContext('2d');
    const r = api.generateGrid(kind, ROWS, COLS, 7);
    const grid = r.ok ? parseGridText(r.value) : null;
    if (!c || !g || !grid) return;
    const cell = 6;
    const dpr = window.devicePixelRatio || 1;
    c.width = COLS * cell * dpr;
    c.height = ROWS * cell * dpr;
    g.setTransform(dpr, 0, 0, dpr, 0, 0);
    grid.cost.forEach((k, i) => {
      g.fillStyle =
        i === grid.start
          ? '#00add8'
          : i === grid.end
            ? '#fbbf24'
            : k === 0
              ? '#292524'
              : k === 1
                ? '#e7e5e4'
                : `rgba(180, 83, 9, ${0.25 + k * 0.08})`;
      g.fillRect((i % COLS) * cell, Math.floor(i / COLS) * cell, cell, cell);
    });
  }, [api, kind]);
  return (
    <canvas
      ref={canvas}
      aria-hidden
      className="block h-auto w-full rounded-md [image-rendering:pixelated]"
    />
  );
}

export function MapPicker({
  api,
  kinds,
  value,
  onChange
}: {
  api: AlgorithmApi;
  kinds: string[];
  value: string;
  onChange: (kind: string) => void;
}) {
  return (
    <div
      role="radiogroup"
      aria-label="Map"
      className="grid grid-cols-2 gap-3 sm:grid-cols-4"
      onKeyDown={(e) => {
        const i = kinds.indexOf(value);
        const step =
          e.key === 'ArrowRight' || e.key === 'ArrowDown'
            ? 1
            : e.key === 'ArrowLeft' || e.key === 'ArrowUp'
              ? -1
              : 0;
        if (!step) return;
        e.preventDefault();
        e.stopPropagation();
        const next = kinds[(i + step + kinds.length) % kinds.length];
        onChange(next);
        e.currentTarget.querySelector<HTMLElement>(`[data-kind="${next}"]`)?.focus();
      }}
    >
      {kinds.map((k) => {
        const on = k === value;
        return (
          // ui-ok: a map is a radio drawn as a keycap tile with its picture
          <button
            key={k}
            type="button"
            role="radio"
            aria-checked={on}
            data-kind={k}
            tabIndex={on ? 0 : -1}
            onClick={() => onChange(k)}
            className={cn(
              'focus-visible:ring-accent flex flex-col gap-2 rounded-xl p-2 text-start text-sm font-medium capitalize outline-none focus-visible:ring-2',
              on ? 'keycap keycap-primary' : 'bg-subtle hover:bg-fg/10'
            )}
          >
            <Thumb
              api={api}
              kind={k}
            />
            {k}
          </button>
        );
      })}
    </div>
  );
}
