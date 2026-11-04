/**
 * A timeline (a Gantt chart of CPU scheduling, or one lane per thread of a
 * concurrency simulation) played as time passes: the player's index is the
 * time, and a view draws each span cut off at it.
 */
import { kinds, type Span, type TraceEvent } from '../schema';

/** The lanes in order of first appearance, "idle" last. */
export function lanesOf(spans: Span[]): string[] {
  const seen: string[] = [];
  for (const s of spans) if (!seen.includes(s.lane)) seen.push(s.lane);
  return [...seen.filter((l) => l !== 'idle'), ...seen.filter((l) => l === 'idle')];
}

/** When the last span ends. */
export function timelineEnd(spans: Span[]): number {
  return Math.ceil(Math.max(0, ...spans.map((s) => s.end)));
}

/** The spans as far as time t: later ones dropped, the current one cut at t. */
export function spansUntil(spans: Span[], t: number): Span[] {
  return spans.filter((s) => s.start < t).map((s) => (s.end > t ? { ...s, end: t } : s));
}

/** What is happening at time t, in words. */
export function describeTime(spans: Span[], t: number): string {
  const now = spans.filter((s) => s.start <= t && t < s.end);
  if (now.length === 0) return `Time ${t}.`;
  return `Time ${t}: ${now.map((s) => `${s.lane} ${s.state ?? 'running'}${s.label && s.label !== s.lane ? ` (${s.label})` : ''}`).join(', ')}.`;
}

/** Which items of an event-driven timeline (one span per item, in item
 * order) are chosen after the first n events, and which item was compared
 * last: for greedy schedules (activity selection, job sequencing). */
export function chosenAt(
  events: TraceEvent[],
  n: number
): { chosen: Set<number>; compared: number | null } {
  const chosen = new Set<number>();
  let compared: number | null = null;
  for (const e of events.slice(0, n)) {
    if (kinds[e.kind] === 'compare') compared = e.indices[0] ?? null;
    else if (kinds[e.kind] === 'mark' && e.note === 'chosen')
      for (const i of e.indices) chosen.add(i);
  }
  return { chosen, compared };
}
