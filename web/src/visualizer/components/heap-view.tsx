import { cn } from '@/lib/utils';
import type { Exec, ExecFrame } from '../core/schema';

const W = 560;
const H = 300;
const R = 22;

const FILL: Record<string, string> = {
  white: 'fill-card stroke-fg',
  gray: 'fill-muted/60 stroke-fg',
  black: 'fill-fg stroke-fg',
  free: 'fill-transparent stroke-muted'
};
const TEXT: Record<string, string> = {
  white: 'fill-fg',
  gray: 'fill-fg',
  black: 'fill-bg',
  free: 'fill-muted'
};

function place(n: number): { x: number; y: number }[] {
  const cx = W / 2;
  const cy = H / 2 + 10;
  const rx = Math.min(W / 2 - 50, 60 + n * 18);
  const ry = Math.min(H / 2 - 40, 40 + n * 10);
  return Array.from({ length: n }, (_, i) => {
    if (n === 1) return { x: cx, y: cy };
    const a = -Math.PI / 2 + (2 * Math.PI * i) / n;
    return { x: cx + rx * Math.cos(a), y: cy + ry * Math.sin(a) };
  });
}

/**
 * A heap during garbage collection: each object a circle filled by its
 * color (white: not found, gray: found, black: done, free: reclaimed, drawn
 * dashed and faded), with its name and color written inside; references are
 * arrows, roots are marked, and reference counts or the copying collector's
 * to-space order are shown when the collector has them.
 */
export function HeapView({ exec, frame }: { exec: Exec; frame: ExecFrame }) {
  const pos = place(exec.objects.length);
  const color = (i: number) => frame.colors[i] ?? 'white';
  return (
    <figure className="m-0 flex flex-col gap-2">
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="w-full"
        role="img"
        aria-label={exec.objects
          .map((o, i) => `${o} ${color(i)}${frame.roots[i] ? ' (root)' : ''}`)
          .join(', ')}
      >
        <defs>
          <marker
            id="av-heap-head"
            viewBox="0 0 8 8"
            refX="7"
            refY="4"
            markerWidth="7"
            markerHeight="7"
            orient="auto"
          >
            <path
              d="M0,0 L8,4 L0,8 z"
              className="fill-muted"
            />
          </marker>
        </defs>
        {frame.refs.map((targets, i) =>
          targets.map((j, k) => {
            const a = pos[i];
            const b = pos[j];
            if (i === j) return null;
            const dx = b.x - a.x;
            const dy = b.y - a.y;
            const d = Math.hypot(dx, dy) || 1;
            return (
              <line
                key={`${i}-${j}-${k}`}
                x1={a.x + (dx / d) * R}
                y1={a.y + (dy / d) * R}
                x2={b.x - (dx / d) * (R + 4)}
                y2={b.y - (dy / d) * (R + 4)}
                strokeWidth={1.5}
                className="stroke-muted transition-opacity duration-300"
                opacity={color(i) === 'free' ? 0.25 : 1}
                markerEnd="url(#av-heap-head)"
              />
            );
          })
        )}
        {exec.objects.map((name, i) => {
          const c = color(i);
          const { x, y } = pos[i];
          return (
            <g
              key={name}
              className="transition-opacity duration-300 ease-out"
              opacity={c === 'free' ? 0.45 : 1}
            >
              {frame.focus === i ? (
                <circle
                  cx={x}
                  cy={y}
                  r={R + 5}
                  fill="none"
                  strokeWidth={2}
                  className="stroke-accent"
                />
              ) : null}
              <circle
                cx={x}
                cy={y}
                r={R}
                strokeWidth={1.5}
                strokeDasharray={c === 'free' ? '4 3' : undefined}
                className={cn('transition-[fill] duration-200', FILL[c])}
              />
              <text
                x={x}
                y={y - 2}
                textAnchor="middle"
                fontSize={13}
                fontWeight={600}
                className={TEXT[c]}
              >
                {name}
              </text>
              <text
                x={x}
                y={y + 11}
                textAnchor="middle"
                fontSize={8}
                className={TEXT[c]}
              >
                {c}
              </text>
              {frame.roots[i] ? (
                <text
                  x={x}
                  y={y - R - 6}
                  textAnchor="middle"
                  fontSize={9}
                  className="fill-accent"
                >
                  root
                </text>
              ) : null}
              {frame.counts.length ? (
                <text
                  x={x + R + 4}
                  y={y + R}
                  fontSize={10}
                  className="fill-fg"
                >
                  {frame.counts[i]} ref{frame.counts[i] === 1 ? '' : 's'}
                </text>
              ) : null}
            </g>
          );
        })}
      </svg>
      {frame.toSpace.length ? (
        <figcaption className="text-muted flex flex-wrap items-center gap-1 text-xs">
          to-space:
          {frame.toSpace.map((k, i) => (
            <span
              key={i}
              className="border-border bg-subtle text-fg rounded border px-1.5 py-0.5 font-mono"
            >
              {exec.objects[k]}
            </span>
          ))}
        </figcaption>
      ) : null}
    </figure>
  );
}
