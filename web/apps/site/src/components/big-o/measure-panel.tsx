'use client';

import { useMemo, useState } from 'react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { Table } from '@braswelljr/algorithm-ui/components/table';
import {
  predict,
  type AlgorithmApi,
  type Catalog,
  type Measurement
} from '@braswelljr/algorithm-visualizer/core';
import { parseAsString, useQueryStates } from 'nuqs';
import { WasmGate } from '@/components/shared/wasm-gate';
import { Button } from '@/components/ui/button';
import { Field } from '@/components/ui/field';
import { Select } from '@/components/ui/select';

/** The input kinds measure() can grow with one size (internal/core Measurable). */
const growable = new Set([
  'ints',
  'sorted-ints',
  'any-ints',
  'text',
  'two-texts',
  'pattern',
  'number',
  'grid',
  'graph',
  'points',
  'amount',
  'pages',
  'tracks'
]);

const sizeSets: Record<string, number[]> = {
  small: [10, 14, 18, 22],
  medium: [100, 200, 400, 800],
  large: [500, 1000, 2000, 4000]
};

const W = 560;
const H = 260;
const PAD = { left: 64, right: 16, top: 12, bottom: 32 };

function Plot({ m }: { m: Measurement }) {
  const top = m.samples[m.samples.length - 1].n;
  const value = (s: Measurement['samples'][number]) => (m.byTime ? s.nanos : s.ops);
  const fit = m.fits[0];
  const steps = 48;
  const curve = Array.from({ length: steps + 1 }, (_, i) => Math.max(1, (top * i) / steps));
  const peak = Math.max(...m.samples.map(value), ...curve.map((n) => predict(fit, n)), 1);
  const x = (n: number) => PAD.left + (n / top) * (W - PAD.left - PAD.right);
  const y = (v: number) => H - PAD.bottom - (v / peak) * (H - PAD.top - PAD.bottom);
  return (
    <Card
      size="sm"
      className="px-(--card-spacing)"
    >
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="w-full"
        role="img"
        aria-label={`${m.name}: measured ${m.byTime ? 'time' : 'operations'} at each size, as dots, and the ${fit.class} fit as a line`}
      >
        <line
          x1={PAD.left}
          x2={W - PAD.right}
          y1={H - PAD.bottom}
          y2={H - PAD.bottom}
          stroke="var(--border)"
        />
        <line
          x1={PAD.left}
          x2={PAD.left}
          y1={PAD.top}
          y2={H - PAD.bottom}
          stroke="var(--border)"
        />
        <text
          x={PAD.left - 6}
          y={PAD.top + 10}
          textAnchor="end"
          fontSize="11"
          fill="var(--muted)"
        >
          {Math.round(peak).toLocaleString('en-US')}
        </text>
        <polyline
          points={curve.map((n) => `${x(n)},${y(predict(fit, n))}`).join(' ')}
          fill="none"
          stroke="var(--accent)"
          strokeWidth={2}
        />
        {m.samples.map((s) => (
          <g key={s.n}>
            <circle
              cx={x(s.n)}
              cy={y(value(s))}
              r={5}
              fill="var(--fg)"
            />
            <text
              x={x(s.n)}
              y={H - PAD.bottom + 18}
              textAnchor="middle"
              fontSize="11"
              fill="var(--muted)"
            >
              {s.n}
            </text>
          </g>
        ))}
      </svg>
    </Card>
  );
}

function Result({ m }: { m: Measurement }) {
  const verdict =
    m.agrees === undefined
      ? 'The stated cost uses a size this measurement holds fixed, so the two cannot be compared.'
      : m.agrees
        ? 'It grows no faster than stated.'
        : 'It grows faster than stated on this input.';
  return (
    <div className="flex flex-col gap-4">
      <p
        role="status"
        className="text-sm"
      >
        <strong>
          {m.best}, off by {(100 * m.error).toFixed(1)}% ({m.quality} fit).
        </strong>{' '}
        Stated: <code>{m.stated}</code>. {verdict}
        {m.byTime
          ? ' This entry reports no steps, so the fit is on time, which varies from run to run.'
          : ''}
      </p>
      <Plot m={m} />
      <Table>
        <caption className="sr-only">Measured samples</caption>
        <thead>
          <tr className="text-left text-muted">
            <th scope="col">n</th>
            <th scope="col">Steps</th>
            <th scope="col">Growth from the last size</th>
            <th scope="col">{m.best} predicts</th>
          </tr>
        </thead>
        <tbody className="font-mono">
          {m.samples.map((s, i) => (
            <tr key={s.n}>
              <th
                scope="row"
                className="text-left font-normal"
              >
                {s.n}
              </th>
              <td>{s.ops.toLocaleString('en-US')}</td>
              <td>
                {i > 0 && m.samples[i - 1].ops > 0
                  ? `x${(s.ops / m.samples[i - 1].ops).toFixed(2)}`
                  : '-'}
              </td>
              <td>{Math.round(predict(m.fits[0], s.n)).toLocaleString('en-US')}</td>
            </tr>
          ))}
        </tbody>
      </Table>
      <p className="text-xs text-muted">
        Next best: {m.fits[1]?.class}, off by{' '}
        {m.fits[1] && m.fits[1].error >= 0
          ? `${(100 * m.fits[1].error).toFixed(1)}%`
          : 'too much to count'}
        .{m.note ? ` ${m.note}.` : ''} The same in a terminal: <code>algorithm bench {m.id}</code>.
      </p>
    </div>
  );
}

function Panel({ api, catalog }: { api: AlgorithmApi; catalog: Catalog }) {
  const [q, setQ] = useQueryStates({
    entry: parseAsString.withDefault(''),
    order: parseAsString.withDefault('random'),
    sizes: parseAsString.withDefault('medium')
  });
  const entries = useMemo(
    () =>
      catalog.entries
        .filter((e) => e.traced && growable.has(e.input))
        .sort((a, b) => a.name.localeCompare(b.name)),
    [catalog]
  );
  const [result, setResult] = useState<ReturnType<AlgorithmApi['measure']> | null>(null);
  // With no entry in the link, start from the first one that can be measured.
  const entry = entries.find((e) => e.id === q.entry) ?? entries[0];
  const id = entry?.id ?? '';
  const measure = () => setResult(api.measure(id, sizeSets[q.sizes] ?? [], q.order));
  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-3 sm:grid-cols-[2fr_1fr_1fr_auto] sm:items-end">
        <Field
          label="Entry"
          htmlFor="measure-entry"
        >
          <Select
            id="measure-entry"
            value={id}
            onChange={(e) => {
              setResult(null);
              void setQ({ entry: e.target.value });
            }}
          >
            {entries.map((e) => (
              <option
                key={e.id}
                value={e.id}
              >
                {e.name}
              </option>
            ))}
          </Select>
        </Field>
        <Field
          label="Values"
          htmlFor="measure-order"
        >
          <Select
            id="measure-order"
            value={q.order}
            onChange={(e) => {
              setResult(null);
              void setQ({ order: e.target.value });
            }}
          >
            {catalog.datasets.map((d) => (
              <option
                key={d}
                value={d}
              >
                {d}
              </option>
            ))}
          </Select>
        </Field>
        <Field
          label="Sizes"
          htmlFor="measure-sizes"
        >
          <Select
            id="measure-sizes"
            value={q.sizes}
            onChange={(e) => {
              setResult(null);
              void setQ({ sizes: e.target.value });
            }}
          >
            {Object.entries(sizeSets).map(([k, v]) => (
              <option
                key={k}
                value={k}
              >
                {v.join(', ')}
              </option>
            ))}
          </Select>
        </Field>
        <Button onClick={measure}>Measure</Button>
      </div>
      {entry ? (
        <p className="text-sm text-muted">
          Stated: <code>{entry.time}</code>. {entry.why}
        </p>
      ) : null}
      {result === null ? null : result.ok ? (
        <Result m={result.value} />
      ) : (
        <p
          role="alert"
          className="text-danger"
        >
          {result.failure.error}
        </p>
      )}
    </div>
  );
}

/** Measure any entry in the browser and see its counts against the fitted curve. */
export function MeasurePanel() {
  return (
    <WasmGate>
      {(api, catalog) => (
        <Panel
          api={api}
          catalog={catalog}
        />
      )}
    </WasmGate>
  );
}
