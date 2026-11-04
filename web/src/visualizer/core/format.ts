/** Formatting shared by the views and the site. */

/** A duration in nanoseconds, for people. */
export function formatNanos(ns: number): string {
  if (ns < 1_000) return `${ns} ns`;
  if (ns < 1_000_000) return `${(ns / 1_000).toFixed(1)} µs`;
  if (ns < 1_000_000_000) return `${(ns / 1_000_000).toFixed(2)} ms`;
  return `${(ns / 1_000_000_000).toFixed(2)} s`;
}

/** A count with thousands separators. */
export function formatCount(n: number): string {
  return n.toLocaleString('en-US');
}
