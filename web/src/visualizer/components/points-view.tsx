import type { PointsState } from '../core/replay/points';
import type { PlanePoint } from '../core/schema';

const SIZE = 360;
const PAD = 24;

/**
 * Points in the plane as an SVG scatter with axes. Points compared last are
 * ringed, visited points are hollow, the hull grows as a dashed polygon, the
 * closest pair is joined by a thick line, points inside are squares, and a
 * clustered point is tinted by its group and labeled with its number (so no
 * state relies on color alone). Pass onAdd to let people click to add a
 * point.
 */
export function PointsView({
  points,
  state,
  onAdd,
  segments = [],
  label = 'The points'
}: {
  points: PlanePoint[];
  state?: PointsState;
  /** Pairs of point indexes to draw as line segments. */
  segments?: [number, number][];
  onAdd?: (p: PlanePoint) => void;
  label?: string;
}) {
  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);
  const minX = Math.min(0, ...xs);
  const maxX = Math.max(10, ...xs);
  const minY = Math.min(0, ...ys);
  const maxY = Math.max(10, ...ys);
  const span = Math.max(maxX - minX, maxY - minY) || 1;
  const sx = (x: number) => PAD + ((x - minX) / span) * (SIZE - 2 * PAD);
  const sy = (y: number) => SIZE - PAD - ((y - minY) / span) * (SIZE - 2 * PAD);
  const compared = new Set(state?.compared ?? []);
  const hull = state?.hull ?? [];
  const pair = state?.pair;
  return (
    <figure
      style={{ margin: 0 }}
      aria-label={label}
    >
      <svg
        className="av-points"
        viewBox={`0 0 ${SIZE} ${SIZE}`}
        role="img"
        aria-label={`${points.length} points${hull.length ? `, a hull of ${hull.length}` : ''}${pair ? ', a closest pair' : ''}`}
        onClick={
          onAdd
            ? (ev) => {
                const box = ev.currentTarget.getBoundingClientRect();
                const px = ((ev.clientX - box.left) / box.width) * SIZE;
                const py = ((ev.clientY - box.top) / box.height) * SIZE;
                onAdd({
                  x: Math.round(minX + ((px - PAD) / (SIZE - 2 * PAD)) * span),
                  y: Math.round(minY + ((SIZE - PAD - py) / (SIZE - 2 * PAD)) * span)
                });
              }
            : undefined
        }
      >
        <line
          className="av-axis"
          x1={PAD}
          y1={SIZE - PAD}
          x2={SIZE - PAD}
          y2={SIZE - PAD}
        />
        <line
          className="av-axis"
          x1={PAD}
          y1={PAD}
          x2={PAD}
          y2={SIZE - PAD}
        />
        {segments.map(([a, b], k) =>
          points[a] && points[b] ? (
            <line
              key={`s${k}`}
              className="av-segment"
              x1={sx(points[a].x)}
              y1={sy(points[a].y)}
              x2={sx(points[b].x)}
              y2={sy(points[b].y)}
            />
          ) : null
        )}
        {hull.length > 1 ? (
          <polygon
            className="av-hull"
            points={hull.map((k) => `${sx(points[k].x)},${sy(points[k].y)}`).join(' ')}
          />
        ) : null}
        {pair ? (
          <line
            className="av-pair"
            x1={sx(points[pair[0]].x)}
            y1={sy(points[pair[0]].y)}
            x2={sx(points[pair[1]].x)}
            y2={sy(points[pair[1]].y)}
          />
        ) : null}
        {points.map((p, i) => {
          const inside = state?.inside.has(i);
          const group = state?.groups.get(i);
          const attrs = {
            className: 'av-pt',
            'data-compared': compared.has(i) || undefined,
            'data-visited': state?.visited.has(i) || undefined,
            'data-hull': hull.includes(i) || undefined,
            'data-group': group === undefined ? undefined : group < 0 ? 'noise' : group % 6
          };
          const tag =
            group === undefined ? null : (
              <text
                key={`g${i}`}
                className="av-pt-group"
                x={sx(p.x) + 6}
                y={sy(p.y) - 6}
              >
                {group < 0 ? '×' : group}
              </text>
            );
          const mark = inside ? (
            <rect
              key={i}
              {...attrs}
              x={sx(p.x) - 4}
              y={sy(p.y) - 4}
              width={8}
              height={8}
            />
          ) : (
            <circle
              key={i}
              {...attrs}
              cx={sx(p.x)}
              cy={sy(p.y)}
              r={compared.has(i) ? 6 : 4}
            />
          );
          return (
            <g key={i}>
              {mark}
              {tag}
            </g>
          );
        })}
      </svg>
    </figure>
  );
}
