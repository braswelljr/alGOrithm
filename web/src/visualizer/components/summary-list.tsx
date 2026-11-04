import { formatCount } from '../core/format';

/** "averageWaiting" -> "average waiting". */
function words(key: string): string {
  return key
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/[_-]+/g, ' ')
    .toLowerCase();
}

/** A run's statistics: average waiting time, page faults, lost updates... */
export function SummaryList({ summary }: { summary: Record<string, number> }) {
  const entries = Object.entries(summary);
  if (entries.length < 1) return null;
  return (
    <dl className="av-summary">
      {entries.map(([k, v]) => (
        <div key={k}>
          <dt>{words(k)}</dt>
          <dd>{Number.isInteger(v) ? formatCount(v) : v.toFixed(2)}</dd>
        </div>
      ))}
    </dl>
  );
}
