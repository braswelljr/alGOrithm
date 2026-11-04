import type { RunResult } from '@/visualizer/core';
import { formatCount, formatNanos } from '@/visualizer/core';
import { metricOf, metrics, type Metric } from './metrics';

/** One bar per algorithm, with its number written beside it. */
export function CompareChart({
  results,
  names,
  metric
}: {
  results: RunResult[];
  names: Map<string, string>;
  metric: Metric;
}) {
  const max = Math.max(1, ...results.map((r) => metricOf(r, metric)));
  const label = metrics.find((m) => m.id === metric)?.label ?? metric;
  const fmt = (v: number) => (metric === 'time' ? formatNanos(v) : formatCount(v));
  const row = 28;
  return (
    <figure>
      <svg
        viewBox={`0 0 600 ${results.length * row}`}
        className="w-full"
        role="img"
        aria-label={`${label} per algorithm: ${results.map((r) => `${names.get(r.id) ?? r.id} ${fmt(metricOf(r, metric))}`).join(', ')}`}
      >
        {results.map((r, i) => {
          const v = metricOf(r, metric);
          const w = (v / max) * 380;
          return (
            <g
              key={r.id}
              transform={`translate(0 ${i * row})`}
            >
              <text
                x={130}
                y={18}
                textAnchor="end"
                fontSize={12}
                className="fill-fg"
              >
                {names.get(r.id) ?? r.id}
              </text>
              <rect
                x={140}
                y={5}
                width={Math.max(w, 1)}
                height={18}
                rx={3}
                className="fill-accent"
              />
              <text
                x={146 + w}
                y={18}
                fontSize={11}
                className="fill-muted font-mono"
              >
                {fmt(v)}
              </text>
            </g>
          );
        })}
      </svg>
      <figcaption className="text-muted mt-1 text-xs">
        {label}, fewer is better. The table below has every number.
      </figcaption>
    </figure>
  );
}
