'use client';

import { useId, useState } from 'react';
import { CircleAlert } from 'lucide-react';
import { DataPanel } from '@/components/data/data-panel';
import { Button } from '@/components/shared/button';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Combobox } from '@/components/ui/combobox';
import { Heading } from '@/components/ui/heading';
import { useDataset } from '@/hooks/use-dataset';
import type { AlgorithmApi as Api, Catalog, RunResult } from '@/visualizer/core';
import { formatCount, formatNanos } from '@/visualizer/core';
import { CompareChart } from './compare-chart';
import { metricOf, metrics, type Metric } from './metrics';

/** Above this many values, quadratic sorts take too long to run in a tab. */
const QUADRATIC_LIMIT = 20_000;

function Compare({ api, catalog }: { api: Api; catalog: Catalog }) {
  const sorts = catalog.entries.filter(
    (e) => e.input === 'ints' && e.traced && /^sort/.test(e.category)
  );
  const names = new Map(sorts.map((e) => [e.id, e.name]));
  const dataset = useDataset(api, 2000);
  const [picked, setPicked] = useState<Set<string>>(() => new Set(sorts.map((e) => e.id)));
  const [metric, setMetric] = useState<Metric>('compare');
  const [results, setResults] = useState<RunResult[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const metricId = useId();

  const n = dataset.values?.length ?? 0;
  const slow = sorts.filter((e) => picked.has(e.id) && /^O\(n\^2\)/.test(e.time));
  const refused = n > QUADRATIC_LIMIT && slow.length > 0;

  const go = () => {
    if (!dataset.values || refused) return;
    setBusy(true);
    setError(null);
    // Let the page paint "Running" before the tab is busy.
    setTimeout(() => {
      const r = api.compare([...picked], { values: dataset.values ?? [] });
      if (r.ok) setResults([...r.value].sort((a, b) => metricOf(a, 'time') - metricOf(b, 'time')));
      else setError(r.failure.error);
      setBusy(false);
    }, 20);
  };

  return (
    <div className="flex flex-col gap-6">
      <DataPanel
        api={api}
        kinds={catalog.datasets}
        maxValues={catalog.maxValues}
        dataset={dataset}
      />
      <Card
        size="md"
        render={<section />}
        aria-labelledby="pick"
        className="px-(--card-spacing)"
      >
        <Heading
          level={2}
          id="pick"
          className="mb-3 font-semibold"
        >
          Sorts to compare
        </Heading>
        <fieldset className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
          <legend className="sr-only">Sorts</legend>
          {sorts.map((e) => (
            <label
              key={e.id}
              className="flex items-center gap-2 text-sm"
            >
              <Checkbox
                checked={picked.has(e.id)}
                onCheckedChange={(on) => {
                  const next = new Set(picked);
                  if (on) next.add(e.id);
                  else next.delete(e.id);
                  setPicked(next);
                }}
              />
              {e.name} <span className="text-muted font-mono text-xs">{e.time.split(',')[0]}</span>
            </label>
          ))}
        </fieldset>
        <div className="mt-4 flex flex-wrap items-center gap-3">
          <Button
            variant="primary"
            onClick={go}
            disabled={busy || picked.size === 0 || !dataset.values || refused}
          >
            {busy ? 'Running...' : `Run on ${formatCount(n)} values`}
          </Button>
          {refused ? (
            <Alert variant="warning">
              <CircleAlert aria-hidden />
              <AlertDescription>
                {slow.map((e) => e.name).join(', ')} would take minutes on {formatCount(n)} values.
                Uncheck {slow.length === 1 ? 'it' : 'them'}, or use at most{' '}
                {formatCount(QUADRATIC_LIMIT)}.
              </AlertDescription>
            </Alert>
          ) : null}
        </div>
      </Card>
      {error ? (
        <Alert variant="destructive">
          <CircleAlert aria-hidden />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {results ? (
        <Card
          size="md"
          render={<section />}
          aria-labelledby="res"
          className="flex flex-col gap-4 px-(--card-spacing)"
        >
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Heading
              level={2}
              id="res"
              className="font-semibold"
            >
              Results on {formatCount(n)} values
            </Heading>
            <div className="flex items-center gap-2">
              <label
                htmlFor={metricId}
                className="text-muted text-sm"
              >
                Chart
              </label>
              <Combobox
                id={metricId}
                className="w-40"
                value={metric}
                onChange={(value) =>
                  setMetric(metrics.find((m) => m.id === value)?.id ?? 'compare')
                }
                options={metrics.map((m) => ({ value: m.id, label: m.label }))}
              />
            </div>
          </div>
          <CompareChart
            results={results}
            names={names}
            metric={metric}
          />
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <caption className="sr-only">Every count, fastest first</caption>
              <thead>
                <tr className="border-border text-muted border-b text-left">
                  <th
                    scope="col"
                    className="py-2"
                  >
                    Sort
                  </th>
                  {metrics.map((m) => (
                    <th
                      key={m.id}
                      scope="col"
                      className="py-2 text-right"
                    >
                      {m.label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {results.map((r) => (
                  <tr
                    key={r.id}
                    className="border-border border-b last:border-0"
                  >
                    <th
                      scope="row"
                      className="py-2 text-left font-medium"
                    >
                      {names.get(r.id) ?? r.id}
                    </th>
                    {metrics.map((m) => (
                      <td
                        key={m.id}
                        className="py-2 text-right font-mono tabular-nums"
                      >
                        {m.id === 'time' ? formatNanos(r.nanos) : formatCount(metricOf(r, m.id))}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-muted text-xs">
            Times are one untraced run in this tab, so they vary with your device; the counts do
            not.
          </p>
        </Card>
      ) : null}
    </div>
  );
}

export function CompareApp() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Compare
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
