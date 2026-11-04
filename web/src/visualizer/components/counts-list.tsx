import { formatCount, formatNanos } from '../core/format';

const names: Record<string, string> = {
  compare: 'comparisons',
  swap: 'swaps',
  set: 'writes',
  visit: 'reads',
  push: 'added',
  pop: 'removed'
};

/** Operation counts so far, and optionally the run's time. */
export function CountsList({ counts, nanos }: { counts: Record<string, number>; nanos?: number }) {
  const items = Object.entries(names).filter(([k]) => (counts[k] ?? 0) > 0);
  return (
    <ul
      className="av-counts"
      aria-label="Counts so far"
    >
      {items.map(([k, label]) => (
        <li key={k}>
          {formatCount(counts[k] ?? 0)} {label}
        </li>
      ))}
      {nanos !== undefined ? <li>{formatNanos(nanos)} untraced</li> : null}
    </ul>
  );
}
