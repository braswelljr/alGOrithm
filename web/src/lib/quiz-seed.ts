/**
 * The quiz game's plain parts: its settings and how they read from the link,
 * the seeded choice of entries (same seed, same entries, same questions), and
 * the best scores kept in this browser. Nothing here draws or runs anything.
 */
import type { Catalog, Entry, GameMode } from '@/visualizer/core';

export const levels = ['any', 'beginner', 'intermediate', 'advanced'] as const;
export type QuizLevel = (typeof levels)[number];

export const modes: { id: GameMode; name: string; blurb: string }[] = [
  { id: 'classic', name: 'Classic', blurb: '10 questions, 3 lives.' },
  { id: 'time', name: 'Time attack', blurb: 'As many as you can in 60 seconds.' },
  { id: 'streak', name: 'Streak', blurb: 'Endless, until the first miss.' }
];

export interface GameConfig {
  mode: GameMode;
  /** "mixed" or a catalog topic ("Algorithms", "Operating systems", ...). */
  domain: string;
  level: QuizLevel;
  seed: number;
  /** One entry's id for a short game about it alone, or null. */
  entry: string | null;
}

/** Questions to gather for a game: enough to never run out. */
export function poolSize(c: Pick<GameConfig, 'mode' | 'entry'>): number {
  if (c.entry) return 5;
  return c.mode === 'classic' ? 10 : 40;
}

/** A small, fast, seeded random number generator (mulberry32): [0, 1). */
export function seeded(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** A shuffled copy, the same for the same random source. */
export function shuffled<T>(xs: readonly T[], rand: () => number): T[] {
  const out = xs.slice();
  for (let i = out.length - 1; i > 0; i--) {
    const j = Math.floor(rand() * (i + 1));
    [out[i], out[j]] = [out[j], out[i]];
  }
  return out;
}

/** A fresh seed for a new game. */
export function newSeed(): number {
  return Math.floor(Math.random() * 1_000_000) + 1;
}

/** The topics a game can be about, in catalog order. */
export function domainsOf(catalog: Pick<Catalog, 'categories'>): string[] {
  const out: string[] = [];
  for (const c of catalog.categories) if (!out.includes(c.topic)) out.push(c.topic);
  return out;
}

/** The entries a game may ask about, in a seeded order. */
export function candidates(
  catalog: Pick<Catalog, 'categories' | 'entries'>,
  c: GameConfig
): Entry[] {
  if (c.entry) return catalog.entries.filter((e) => e.id === c.entry);
  const topic = new Map(catalog.categories.map((k) => [k.id, k.topic]));
  const fit = catalog.entries.filter(
    (e) =>
      (c.domain === 'mixed' || topic.get(e.category) === c.domain) &&
      (c.level === 'any' || e.level === c.level)
  );
  return shuffled(fit, seeded(c.seed));
}

/** The best score and streak for each mode, kept in this browser only. */
export type Bests = Partial<Record<GameMode, { score: number; streak: number }>>;

const BEST_KEY = 'algorithm.quiz.best';

export function loadBests(storage: Pick<Storage, 'getItem'> | undefined): Bests {
  try {
    const raw = storage?.getItem(BEST_KEY);
    if (!raw) return {};
    const v: unknown = JSON.parse(raw);
    if (typeof v !== 'object' || v === null) return {};
    const out: Bests = {};
    for (const m of modes) {
      const b: unknown = Reflect.get(v, m.id);
      if (typeof b !== 'object' || b === null) continue;
      const score: unknown = Reflect.get(b, 'score');
      const streak: unknown = Reflect.get(b, 'streak');
      if (typeof score === 'number' && typeof streak === 'number') out[m.id] = { score, streak };
    }
    return out;
  } catch {
    return {};
  }
}

/** Saves a finished game's numbers if they beat the bests; returns the new
 * bests and whether the score was a new best. */
export function saveBest(
  storage: Pick<Storage, 'getItem' | 'setItem'> | undefined,
  mode: GameMode,
  score: number,
  streak: number
): { bests: Bests; newBest: boolean } {
  const bests = loadBests(storage);
  const old = bests[mode] ?? { score: 0, streak: 0 };
  const newBest = score > old.score;
  bests[mode] = { score: Math.max(old.score, score), streak: Math.max(old.streak, streak) };
  try {
    storage?.setItem(BEST_KEY, JSON.stringify(bests));
  } catch {
    // Storage may be full or blocked; the game still works without it.
  }
  return { bests, newBest };
}

/** The browser's localStorage, or undefined where it cannot be reached. */
export function browserStorage(): Storage | undefined {
  try {
    return typeof window === 'undefined' ? undefined : window.localStorage;
  } catch {
    return undefined;
  }
}
