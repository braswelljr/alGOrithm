/** Points as text, one "x y" a line, the way the geometry entries read them. */
export interface XY {
  x: number;
  y: number;
}

/** The points in text; lines that are not two whole numbers are skipped. */
export function parsePoints(text: string): XY[] {
  const out: XY[] = [];
  for (const line of text.split('\n')) {
    const f = line.trim().split(/[\s,]+/);
    if (f.length !== 2) continue;
    const [x, y] = f.map(Number);
    if (Number.isSafeInteger(x) && Number.isSafeInteger(y)) out.push({ x, y });
  }
  return out;
}

export const formatPoints = (ps: XY[]) => ps.map((p) => `${p.x} ${p.y}`).join('\n');

/** n points in 0..size from seed (mulberry32), so a link can carry the seed. */
export function randomPoints(n: number, seed: number, size = 100): XY[] {
  let a = seed >>> 0;
  const next = () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
  return Array.from({ length: n }, () => ({
    x: Math.round(next() * size),
    y: Math.round(next() * size)
  }));
}
