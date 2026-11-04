/**
 * Grid text in the format the Go side reads ("#" wall, "." open, "1"-"9"
 * cost, "S" start, "E" end), parsed here only to draw and edit a grid before
 * it runs; the Go side parses it again and is the judge.
 */
import type { Grid } from '@braswelljr/algorithm-visualizer/core';

export const MAX_GRID_CELLS = 100 * 100;

export function parseGridText(text: string): Grid | null {
  const lines = text
    .replace(/\r/g, '')
    .split('\n')
    .map((l) => l.replace(/\s+$/, ''))
    .filter((l) => l.trim() !== '');
  if (!lines.length) return null;
  const cols = Math.max(...lines.map((l) => l.length));
  if (lines.length * cols > MAX_GRID_CELLS) return null;
  const cost: number[] = [];
  let start = -1;
  let end = -1;
  for (const [r, line] of lines.entries()) {
    for (let c = 0; c < cols; c++) {
      const ch = line[c] ?? '.';
      const i = r * cols + c;
      if (ch === '#') cost.push(0);
      else if (ch === '.' || ch === ' ') cost.push(1);
      else if (ch >= '1' && ch <= '9') cost.push(Number(ch));
      else if (ch === 'S' || ch === 's') {
        cost.push(1);
        start = i;
      } else if (ch === 'E' || ch === 'e') {
        cost.push(1);
        end = i;
      } else return null;
    }
  }
  if (start < 0) start = cost.findIndex((c) => c > 0);
  if (end < 0)
    for (let i = cost.length - 1; i >= 0 && end < 0; i--) if (cost[i] > 0 && i !== start) end = i;
  if (start < 0 || end < 0) return null;
  return { rows: lines.length, cols, cost, start, end, labels: [] };
}

export function gridToText(g: Grid): string {
  let out = '';
  g.cost.forEach((c, i) => {
    out += i === g.start ? 'S' : i === g.end ? 'E' : c === 0 ? '#' : c === 1 ? '.' : String(c);
    if (i % g.cols === g.cols - 1) out += '\n';
  });
  return out;
}

/** Toggles a wall: open becomes wall, wall becomes open. */
export function toggleWall(g: Grid, cell: number): Grid {
  if (cell === g.start || cell === g.end) return g;
  const cost = g.cost.slice();
  cost[cell] = cost[cell] === 0 ? 1 : 0;
  return { ...g, cost };
}

/** A grid kept in the link: rows joined by "~". */
export const encodeGrid = (text: string) => text.trim().split('\n').join('~');
export const decodeGrid = (s: string) => s.split('~').join('\n');

/** A small grid for when the wasm build cannot generate one. */
export const SAMPLE_GRID = [
  'S....#......',
  '.###.#.####.',
  '...#...#....',
  '##.#####.##.',
  '...#.....#..',
  '.#...###.#.E'
].join('\n');
