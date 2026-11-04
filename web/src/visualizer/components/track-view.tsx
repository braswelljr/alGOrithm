import type { Track } from '../core/schema';

/**
 * A disk head's path: tracks across, one stop per row down. The stops so far
 * are joined by a line, each marked with a dot and its track number; the
 * stop just reached is a larger ring. The total distance is written below.
 */
export function TrackView({ track, upTo }: { track: Track; upTo?: number }) {
  const stops = track.path.slice(0, (upTo ?? track.path.length - 1) + 1);
  const W = 640;
  const left = 16;
  const top = 18;
  const step = 22;
  const H = top + Math.max(track.path.length - 1, 1) * step + 28;
  const x = (t: number) => left + (t / Math.max(track.size - 1, 1)) * (W - 2 * left);
  let moved = 0;
  for (let i = 1; i < stops.length; i++) moved += Math.abs(stops[i] - stops[i - 1]);
  return (
    <figure style={{ margin: 0 }}>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="av-track"
        role="img"
        aria-label={`The head visits ${stops.join(', ')}: ${moved} tracks moved so far`}
      >
        <line
          x1={left}
          x2={W - left}
          y1={top - 8}
          y2={top - 8}
          className="av-stroke-border"
        />
        {[0, Math.floor((track.size - 1) / 2), track.size - 1].map((t) => (
          <text
            key={t}
            x={x(t)}
            y={top - 11}
            textAnchor="middle"
            fontSize={9}
            className="av-fill-muted"
          >
            {t}
          </text>
        ))}
        <polyline
          points={stops.map((t, i) => `${x(t)},${top + i * step}`).join(' ')}
          fill="none"
          className="av-stroke-pivot"
          strokeWidth={2}
        />
        {stops.map((t, i) => (
          <g key={i}>
            <circle
              cx={x(t)}
              cy={top + i * step}
              r={i === stops.length - 1 ? 6 : 3.5}
              className={i === stops.length - 1 ? 'av-ring-current' : 'av-fill-pivot'}
            />
            <text
              x={x(t) + 9}
              y={top + i * step + 4}
              fontSize={10}
              className="av-fill-fg"
            >
              {t}
            </text>
          </g>
        ))}
      </svg>
      <figcaption className="av-muted">
        {moved} tracks moved over {Math.max(stops.length - 1, 0)} seeks
      </figcaption>
    </figure>
  );
}
