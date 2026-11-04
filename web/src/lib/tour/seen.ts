/**
 * Which tours a visitor has seen, kept in this browser. There are no accounts,
 * so one list serves the browser; another browser offers the welcome tour once
 * more, which is harmless. Every read and write is wrapped: a private window
 * or blocked site data leaves the tours working, unremembered.
 */

export const SEEN_KEY = 'algorithm:tours';

/** The ids of the tours this browser has seen. */
export function readSeen(): Set<string> {
  try {
    const raw = window.localStorage.getItem(SEEN_KEY);
    if (!raw) return new Set();
    const parsed: unknown = JSON.parse(raw);
    return new Set(
      Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === 'string') : []
    );
  } catch {
    return new Set();
  }
}

/** Saves the set; nothing happens when storage refuses. */
export function writeSeen(seen: Set<string>): void {
  try {
    if (seen.size < 1) window.localStorage.removeItem(SEEN_KEY);
    else window.localStorage.setItem(SEEN_KEY, JSON.stringify([...seen].sort()));
  } catch {
    // Not saved; the tour may be offered again next visit.
  }
}
