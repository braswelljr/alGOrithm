import type { ReactNode } from 'react';

export type PaletteItem = {
  id: string;
  label: string;
  /** A second line: a summary, a category, a path. */
  hint?: string;
  /** More words to match: tags, aliases, the id. */
  keywords?: string[];
  /** A key binding to show beside it, from the shortcuts registry. */
  keys?: string;
  icon?: ReactNode;
  onSelect: () => void;
};

export type PaletteGroup = { label: string; items: PaletteItem[] };

/**
 * A fuzzy match: every character of the query, in order, somewhere in the
 * text. Scored so consecutive and word-start matches rank first.
 */
export function fuzzyScore(text: string, query: string): number {
  const t = text.toLowerCase();
  const q = query.toLowerCase().trim();
  if (!q) return 1;
  if (t.includes(q)) return 100 - t.indexOf(q);
  let score = 0;
  let at = -1;
  let run = 0;
  for (const ch of q) {
    if (ch === ' ') continue;
    const i = t.indexOf(ch, at + 1);
    if (i < 0) return 0;
    run = i === at + 1 ? run + 1 : 0;
    score += 1 + run * 2 + (i === 0 || /[\s\-_/]/.test(t[i - 1]) ? 3 : 0);
    at = i;
  }
  return score;
}

/** The groups with only the items that match, best first. */
export function filterPalette(groups: PaletteGroup[], query: string, limit = 8): PaletteGroup[] {
  return groups
    .map((g) => ({
      label: g.label,
      items: g.items
        .map((item) => ({
          item,
          score: Math.max(
            fuzzyScore(item.label, query) * 2,
            ...[item.hint ?? '', ...(item.keywords ?? [])].map((k) => fuzzyScore(k, query))
          )
        }))
        .filter((x) => x.score > 0)
        .sort((a, b) => b.score - a.score)
        .slice(0, query ? limit : limit * 2)
        .map((x) => x.item)
    }))
    .filter((g) => g.items.length > 0);
}

