import { Stat, StatLabel, StatValue } from '@/components/ui/stat';

const format = (v: number) =>
  Number.isInteger(v)
    ? v.toLocaleString('en-US')
    : v.toLocaleString('en-US', { maximumFractionDigits: 2 });

/** A run's named numbers as metric tiles, in the order the model gave them. */
export function Metrics({ summary, label }: { summary?: Record<string, number>; label: string }) {
  const rows = Object.entries(summary ?? {});
  if (rows.length < 1) return null;
  return (
    <ul
      aria-label={label}
      className="grid grid-cols-2 gap-3 lg:grid-cols-4"
    >
      {rows.map(([name, value]) => (
        <li key={name}>
          <Stat>
            <StatLabel>{name}</StatLabel>
            <StatValue className="text-xl sm:text-2xl lg:text-2xl">{format(value)}</StatValue>
          </Stat>
        </li>
      ))}
    </ul>
  );
}
