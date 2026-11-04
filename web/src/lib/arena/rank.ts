/** The arena's ranks, by points earned. */

/** A rank for a point total, and how far the next one is. */
export const RANKS = [
  { name: 'Novice', from: 0 },
  { name: 'Apprentice', from: 100 },
  { name: 'Adept', from: 500 },
  { name: 'Expert', from: 1500 },
  { name: 'Master', from: 4000 },
  { name: 'Grandmaster', from: 10000 }
] as const;

export function rankOf(points: number) {
  let i = 0;
  while (i + 1 < RANKS.length && points >= RANKS[i + 1].from) i++;
  const next = RANKS[i + 1];
  return {
    name: RANKS[i].name,
    level: i + 1,
    next: next?.name ?? null,
    /** 0 to 1, through this rank toward the next. */
    progress: next ? (points - RANKS[i].from) / (next.from - RANKS[i].from) : 1,
    toNext: next ? next.from - points : 0
  };
}
