'use client';

import { useId, useState } from 'react';
import { Field, Input } from '@/components/shared/field';
import { Card } from '@/components/ui/card';
import { Table } from '@/components/ui/table';
import { Toggle } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';
import { growth, growthClasses } from '@/visualizer/core';

const W = 640;
const H = 320;
const PAD = { left: 56, right: 96, top: 12, bottom: 32 };
/** The y axis is logarithmic, from 1 to 10^MAX_DECADE operations. */
const MAX_DECADE = 12;

/**
 * A coordinate to two decimals. Math.log10 and ** may differ in the last bit
 * between the server's Node and the reader's browser, and the full-precision
 * points string then fails hydration.
 */
const round = (v: number) => Math.round(v * 100) / 100;
const x = (n: number, top: number) =>
  round(PAD.left + ((n - 1) / (top - 1)) * (W - PAD.left - PAD.right));
const y = (v: number) => {
  const d = Math.min(Math.max(Math.log10(Math.max(v, 1)), 0), MAX_DECADE);
  return round(H - PAD.bottom - (d / MAX_DECADE) * (H - PAD.top - PAD.bottom));
};

/** A count written for a person: 1,024, or 1.2e15 beyond a trillion. */
function count(v: number): string {
  if (!Number.isFinite(v)) return 'too many to write';
  if (v >= 1e12) return v.toExponential(1);
  return Math.round(v).toLocaleString('en-US');
}

/** How long v operations take at a billion a second, roughly. */
function duration(v: number): string {
  const s = v / 1e9;
  if (!Number.isFinite(s)) return 'forever';
  if (s < 1e-6) return 'under a microsecond';
  if (s < 1e-3) return `${Math.round(s * 1e6)} microseconds`;
  if (s < 1) return `${Math.round(s * 1e3)} milliseconds`;
  if (s < 60) return `${s.toFixed(1)} seconds`;
  if (s < 3600) return `${Math.round(s / 60)} minutes`;
  if (s < 86400 * 365) return `${Math.round(s / 3600)} hours`;
  const years = s / (86400 * 365);
  return years > 1e6 ? 'longer than people have existed' : `${Math.round(years)} years`;
}

/**
 * Every growth class drawn on one logarithmic chart, with a slider for n. The
 * table beside it gives the exact counts at the chosen n, so nothing depends
 * on reading a curve or telling colors apart.
 */
export function GrowthChart() {
  const [n, setN] = useState(32);
  const [focus, setFocus] = useState<string>('O(n log n)');
  const sliderId = useId();
  const top = 64;
  const ns = Array.from({ length: top }, (_, i) => i + 1);
  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_18rem]">
      <figure className="flex flex-col gap-3">
        <Card
          size="sm"
          className="px-(--card-spacing)"
        >
          <svg
            viewBox={`0 0 ${W} ${H}`}
            className="w-full"
            role="img"
            aria-labelledby={`${sliderId}-caption`}
          >
            {Array.from({ length: MAX_DECADE / 3 + 1 }, (_, i) => i * 3).map((d) => (
              <g key={d}>
                <line
                  x1={PAD.left}
                  x2={W - PAD.right}
                  y1={y(10 ** d)}
                  y2={y(10 ** d)}
                  stroke="var(--border)"
                />
                <text
                  x={PAD.left - 6}
                  y={y(10 ** d) + 4}
                  textAnchor="end"
                  fontSize="11"
                  fill="var(--muted)"
                >
                  {d === 0 ? '1' : `10^${d}`}
                </text>
              </g>
            ))}
            {growthClasses.map((c) => {
              const points = ns.map((v) => `${x(v, top)},${y(growth(c.class, v))}`).join(' ');
              const on = c.class === focus;
              const end = growth(c.class, top);
              return (
                <g key={c.class}>
                  <polyline
                    points={points}
                    fill="none"
                    stroke={on ? 'var(--accent)' : 'var(--muted)'}
                    strokeWidth={on ? 3 : 1.5}
                    strokeDasharray={on ? undefined : '4 3'}
                  />
                  <text
                    x={W - PAD.right + 6}
                    y={Math.max(y(end), PAD.top + 8)}
                    fontSize="11"
                    fill={on ? 'var(--fg)' : 'var(--muted)'}
                    fontWeight={on ? 600 : 400}
                  >
                    {c.class}
                  </text>
                </g>
              );
            })}
            <line
              x1={x(n, top)}
              x2={x(n, top)}
              y1={PAD.top}
              y2={H - PAD.bottom}
              stroke="var(--fg)"
            />
            <text
              x={x(n, top)}
              y={H - PAD.bottom + 20}
              textAnchor="middle"
              fontSize="11"
              fill="var(--fg)"
            >
              n = {n}
            </text>
          </svg>
        </Card>
        <figcaption
          id={`${sliderId}-caption`}
          className="text-muted text-sm"
        >
          How many steps each class takes as n grows from 1 to {top}, on a logarithmic scale where
          each grid line is a thousand times the one below. The selected class is solid; the others
          are dashed.
        </figcaption>
        <Field
          label={`n = ${n}`}
          htmlFor={sliderId}
          hint="Drag, or use the arrow keys."
        >
          <Input
            id={sliderId}
            type="range"
            min={1}
            max={top}
            value={n}
            onChange={(e) => setN(Number(e.target.value))}
            className="h-6 px-0"
          />
        </Field>
      </figure>
      <Table className="h-fit">
        <caption className="mb-2 text-left font-medium">
          At n = {n}, a billion steps a second
        </caption>
        <thead>
          <tr className="text-muted text-left">
            <th scope="col">Class</th>
            <th scope="col">Steps</th>
            <th scope="col">Time</th>
          </tr>
        </thead>
        <tbody>
          {growthClasses.map((c) => {
            const v = growth(c.class, n);
            return (
              <tr
                key={c.class}
                className={cn(c.class === focus && 'font-semibold')}
              >
                <th
                  scope="row"
                  className="py-1 pr-2 text-left font-mono"
                >
                  <Toggle
                    variant="chip"
                    size="sm"
                    pressed={c.class === focus}
                    onPressedChange={() => setFocus(c.class)}
                    className="font-mono"
                  >
                    {c.class}
                  </Toggle>
                </th>
                <td className="pr-2 font-mono">{count(v)}</td>
                <td className="text-muted">{duration(v)}</td>
              </tr>
            );
          })}
        </tbody>
      </Table>
    </div>
  );
}
