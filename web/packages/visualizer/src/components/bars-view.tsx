import type { ArrayStep } from '../core/replay/array';
import { Patterns } from './legend';

/**
 * An array as bars, for sorts. Each highlight has a shape as well as a color:
 * a triangle above a compared bar, a hatch on a swapped bar, a dot over a
 * written bar, a diamond over the pivot, a dashed outline round the smallest
 * so far, and a bar under the baseline once a value is in its final place.
 */
export function BarsView({ state, quiet = false }: { state: ArrayStep; quiet?: boolean }) {
  const n = state.values.length;
  const lo = Math.min(0, ...state.values);
  const hi = Math.max(1, ...state.values);
  const W = 10;
  const H = 200;
  const top = 16;
  const bottom = 10;
  const plot = H - top - bottom;
  const y = (v: number) => top + ((hi - v) / (hi - lo || 1)) * plot;
  const zero = y(0);
  const compared = new Set(quiet ? [] : state.compared);
  const swapped = new Set(quiet ? [] : state.swapped);
  const written = new Set(quiet ? [] : state.written.concat(state.visited));
  const pivot = new Set(state.pivot);
  const min = new Set(state.min);
  const flip = state.flip;
  const showLabels = n <= 32;

  return (
    <svg
      viewBox={`0 0 ${Math.max(n, 1) * W} ${H + (showLabels ? 12 : 0)}`}
      preserveAspectRatio="none"
      className="av-bars"
      role="img"
      aria-label={`${n} bars, one per value: ${state.values.slice(0, 40).join(', ')}${n > 40 ? ', ...' : ''}`}
    >
      <Patterns />
      {flip ? (
        <rect
          x={flip[0] * W}
          y={2}
          width={(flip[1] - flip[0] + 1) * W}
          height={4}
          className="av-fill-pivot"
        />
      ) : null}
      {state.range ? (
        <rect
          x={state.range[0] * W}
          y={top - 2}
          width={(state.range[1] - state.range[0] + 1) * W}
          height={H - top - bottom + 4}
          className="av-fill-range"
        />
      ) : null}
      <line
        x1={0}
        x2={n * W}
        y1={zero}
        y2={zero}
        className="av-stroke-border"
        strokeWidth={0.5}
      />
      {state.values.map((v, i) => {
        const x = i * W + 1;
        const y0 = Math.min(y(v), zero);
        const h = Math.max(Math.abs(zero - y(v)), 0.75);
        const cls = swapped.has(i)
          ? 'av-stroke-swap'
          : written.has(i)
            ? 'av-fill-write'
            : pivot.has(i)
              ? 'av-fill-pivot'
              : compared.has(i)
                ? 'av-fill-compare'
                : state.sorted[i]
                  ? 'av-fill-sorted'
                  : 'av-fill-bar';
        return (
          <g key={i}>
            <rect
              x={x}
              y={y0}
              width={W - 2}
              height={h}
              fill={swapped.has(i) ? 'url(#av-hatch-swap)' : undefined}
              className={cls}
              strokeWidth={swapped.has(i) ? 0.6 : 0}
            />
            {min.has(i) ? (
              <rect
                x={x - 0.5}
                y={y0 - 0.5}
                width={W - 1}
                height={h + 1}
                fill="none"
                className="av-stroke-pivot"
                strokeWidth={0.8}
                strokeDasharray="2 1"
              />
            ) : null}
            {state.sorted[i] ? (
              <rect
                x={x}
                y={H - bottom + 3}
                width={W - 2}
                height={3}
                className="av-fill-fg"
              />
            ) : null}
            {compared.has(i) ? (
              <path
                d={`M${x} ${top - 12} L${x + W - 2} ${top - 12} L${x + (W - 2) / 2} ${top - 4} Z`}
                className="av-fill-compare"
              />
            ) : null}
            {written.has(i) ? (
              <circle
                cx={x + (W - 2) / 2}
                cy={top - 8}
                r={3}
                className="av-fill-write"
              />
            ) : null}
            {state.found === i ? (
              <path
                d={`M${x} ${top - 8} L${x + 3} ${top - 4} L${x + W - 2} ${top - 13}`}
                fill="none"
                strokeWidth={1.6}
                className="av-stroke-found"
              />
            ) : null}
            {pivot.has(i) ? (
              <path
                d={`M${x + 4} ${top - 13} L${x + 8} ${top - 8} L${x + 4} ${top - 3} L${x} ${top - 8} Z`}
                className="av-fill-pivot"
              />
            ) : null}
            {showLabels ? (
              <text
                x={x + (W - 2) / 2}
                y={H + 10}
                textAnchor="middle"
                className="av-fill-muted"
                fontSize={4}
              >
                {v}
              </text>
            ) : null}
          </g>
        );
      })}
    </svg>
  );
}
