'use client';

import { useCallback, useSyncExternalStore } from 'react';
import { z } from 'zod';
import { savedRunSchema, type SavedRun } from '@/visualizer/core';

/** How many runs the history keeps. */
export const HISTORY_SIZE = 20;
const KEY = 'algorithm:history';

// Kept in this browser only (docs/rules/the-users-data-never-leaves-the-browser.md).
// Results are not stored: a run is replayed by running it again, which is fast
// and keeps the history small.
const historySchema = z.array(savedRunSchema.omit({ result: true }));
type HistoryItem = z.infer<typeof historySchema>[number];

let cache: HistoryItem[] | null = null;
const listeners = new Set<() => void>();
const EMPTY: HistoryItem[] = [];

function read(): HistoryItem[] {
  if (cache) return cache;
  try {
    const raw = localStorage.getItem(KEY);
    const parsed = raw ? historySchema.safeParse(JSON.parse(raw)) : null;
    cache = parsed?.success ? parsed.data : [];
  } catch {
    cache = [];
  }
  return cache;
}

function write(items: HistoryItem[]) {
  cache = items;
  try {
    localStorage.setItem(KEY, JSON.stringify(items));
  } catch {
    // Storage blocked or full: the history lasts for this page only.
  }
  listeners.forEach((l) => l());
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

/** The last runs in this browser, newest first. */
export function useRunHistory() {
  const items = useSyncExternalStore(subscribe, read, () => EMPTY);
  const add = useCallback((run: SavedRun) => {
    const item: HistoryItem = {
      format: run.format,
      version: run.version,
      entry: run.entry,
      input: run.input,
      savedAt: run.savedAt
    };
    const key = JSON.stringify([item.entry, item.input]);
    const rest = read().filter((r) => JSON.stringify([r.entry, r.input]) !== key);
    write([item, ...rest].slice(0, HISTORY_SIZE));
  }, []);
  const clear = useCallback(() => write([]), []);
  return { items, add, clear };
}
