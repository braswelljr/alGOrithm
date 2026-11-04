import type { RunResult } from '@braswelljr/algorithm-visualizer/core';

export type Metric = 'compare' | 'swap' | 'set' | 'time';

export const metrics: { id: Metric; label: string }[] = [
  { id: 'compare', label: 'Comparisons' },
  { id: 'swap', label: 'Swaps' },
  { id: 'set', label: 'Writes' },
  { id: 'time', label: 'Time' }
];

export function metricOf(r: RunResult, m: Metric): number {
  return m === 'time' ? r.nanos : (r.counts?.[m] ?? 0);
}
