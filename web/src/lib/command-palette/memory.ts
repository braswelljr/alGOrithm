/**
 * What the command palette remembers in this browser: the places and entries
 * opened lately, and the ones pinned (after gsl-special-courses).
 *
 * A convenience, so it lives in local storage and nowhere else. It can come
 * back empty (a private window, cleared site data, storage switched off), and
 * every read and write is guarded so the palette works the same without it.
 * The site has no accounts, so the key names this browser ('local'); the
 * functions keep the id so the shape matches the source.
 *
 * Only what is needed to draw a row and go back to it is kept: a title, a
 * second line and an address. Nothing is re-fetched to show it, so a recent
 * row says what the record was called when it was opened.
 */

import { useSyncExternalStore } from 'react';
import { z } from 'zod';

/** One remembered row. */
export type RememberedEntry = {
  id: string;
  title: string;
  subtitle?: string;
  href: string;
  /** What kind of thing it is, for its icon: a record kind, or 'page'. */
  kind: string;
};

type Memory = { recent: RememberedEntry[]; pinned: RememberedEntry[] };

/** How many of each are kept. Past this the oldest falls off. */
export const RECENT_LIMIT = 6;
export const PINNED_LIMIT = 10;

const EMPTY: Memory = { recent: [], pinned: [] };

const keyFor = (userId: string) => `algorithm.command-palette.${userId}`;

/** The one key the site remembers under: there are no accounts. */
export const LOCAL = 'local';

const cache = new Map<string, Memory>();
const listeners = new Set<() => void>();

/** A remembered row as storage holds it; anything else is dropped. */
const entrySchema = z.object({
  id: z.string(),
  title: z.string(),
  subtitle: z.string().optional(),
  href: z.string(),
  kind: z.string()
});

/** The stored value, checked: a list that is not one, or a row that does not
 *  parse, reads as empty rather than breaking the palette. */
const memorySchema = z.object({
  recent: z.array(z.unknown()).catch([]),
  pinned: z.array(z.unknown()).catch([])
});

const rows = (list: unknown[]): RememberedEntry[] =>
  list.flatMap((row) => {
    const r = entrySchema.safeParse(row);
    return r.success ? [r.data] : [];
  });

function read(userId: string): Memory {
  const hit = cache.get(userId);
  if (hit) return hit;
  let memory = EMPTY;
  try {
    const raw = window.localStorage.getItem(keyFor(userId));
    if (raw) {
      const parsed = memorySchema.safeParse(JSON.parse(raw));
      if (parsed.success)
        memory = { recent: rows(parsed.data.recent), pinned: rows(parsed.data.pinned) };
    }
  } catch {
    // Storage unavailable or the value unreadable: start from nothing.
  }
  cache.set(userId, memory);
  return memory;
}

function write(userId: string, memory: Memory) {
  cache.set(userId, memory);
  try {
    window.localStorage.setItem(keyFor(userId), JSON.stringify(memory));
  } catch {
    // Kept for this session only.
  }
  listeners.forEach((listener) => listener());
}

const plain = ({ id, title, subtitle, href, kind }: RememberedEntry): RememberedEntry => ({
  id,
  title,
  subtitle,
  href,
  kind
});

/** Put an entry at the top of the recent list, once. */
export function rememberVisit(userId: string, entry: RememberedEntry) {
  const memory = read(userId);
  const recent = [plain(entry), ...memory.recent.filter((e) => e.id !== entry.id)].slice(
    0,
    RECENT_LIMIT
  );
  write(userId, { ...memory, recent });
}

/** Pin an entry, or unpin it when it is pinned already. */
export function togglePin(userId: string, entry: RememberedEntry) {
  const memory = read(userId);
  const pinned = memory.pinned.some((e) => e.id === entry.id)
    ? memory.pinned.filter((e) => e.id !== entry.id)
    : [plain(entry), ...memory.pinned].slice(0, PINNED_LIMIT);
  write(userId, { ...memory, pinned });
}

/** Forget the recent list. Pins stay. */
export function clearRecent(userId: string) {
  write(userId, { ...read(userId), recent: [] });
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

/** The remembered rows for an account, kept current as they change. */
export function usePaletteMemory(userId: string | undefined): Memory {
  return useSyncExternalStore(
    subscribe,
    () => (userId ? read(userId) : EMPTY),
    () => EMPTY
  );
}

/** For tests: forget everything held in memory. */
export function resetPaletteMemory() {
  cache.clear();
}
