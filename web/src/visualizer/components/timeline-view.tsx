import { lanesOf } from '../core/replay/timeline';
import type { Span } from '../core/schema';

/** Fill classes by span state; each state also has its own pattern. */
const stateClass: Record<string, string> = {
  running: 'av-span-running',
  idle: 'av-span-idle',
  blocked: 'av-span-blocked',
  done: 'av-span-done'
};

/**
 * A timeline: one lane per process or thread, time to the right, a bar per
 * span labeled with its name. Running is solid, blocked is hatched, idle is
 * dotted, done is outlined. `now` draws a cursor at that time; `end` fixes
 * the axis so it does not grow while playing.
 */
export function TimelineView({
  spans,
  end,
  now,
  lanes: fixedLanes,
  focus
}: {
  spans: Span[];
  end: number;
  now?: number;
  /** The lanes, in order, when they must not change as spans appear. */
  lanes?: string[];
  /** A span to outline: the item being looked at. */
  focus?: Span;
}) {
  const lanes = fixedLanes ?? lanesOf(spans);
  const W = 640;
  const left = 96;
  const row = 26;
  const top = 8;
  const plot = W - left - 8;
  const x = (t: number) => left + (t / Math.max(end, 1)) * plot;
  const H = top + lanes.length * row + 24;
  const ticks = Math.min(end, 20);
  const tickStep = Math.max(1, Math.ceil(end / Math.max(ticks, 1)));
  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className="av-timeline"
      role="img"
      aria-label={`A timeline of ${lanes.length} lanes up to time ${now ?? end}`}
    >
      <defs>
        <pattern
          id="av-hatch-blocked"
          width="6"
          height="6"
          patternUnits="userSpaceOnUse"
          patternTransform="rotate(45)"
        >
          <rect
            width="6"
            height="6"
            className="av-hatch-a"
          />
          <rect
            width="2"
            height="6"
            className="av-fill-swap"
          />
        </pattern>
        <pattern
          id="av-dots-idle"
          width="5"
          height="5"
          patternUnits="userSpaceOnUse"
        >
          <circle
            cx="2.5"
            cy="2.5"
            r="1"
            className="av-fill-muted"
          />
        </pattern>
      </defs>
      {lanes.map((lane, i) => (
        <g key={lane}>
          <text
            x={left - 8}
            y={top + i * row + 17}
            textAnchor="end"
            fontSize={11}
            className="av-fill-fg"
          >
            {lane}
          </text>
          <line
            x1={left}
            x2={W - 8}
            y1={top + i * row + row - 2}
            y2={top + i * row + row - 2}
            className="av-stroke-border"
            strokeWidth={0.5}
          />
        </g>
      ))}
      {focus && lanes.includes(focus.lane) ? (
        <rect
          x={x(focus.start) - 2}
          y={top + lanes.indexOf(focus.lane) * row + 1}
          width={Math.max(x(focus.end) - x(focus.start), 0.5) + 4}
          height={row - 4}
          rx={4}
          fill="none"
          strokeWidth={2.5}
          strokeDasharray="5 3"
          className="av-stroke-fg"
        />
      ) : null}
      {spans.map((s, k) => {
        const i = lanes.indexOf(s.lane);
        const w = Math.max(x(s.end) - x(s.start), 0.5);
        const state = s.state ?? (s.lane === 'idle' ? 'idle' : 'running');
        const fill =
          state === 'blocked'
            ? 'url(#av-hatch-blocked)'
            : state === 'idle'
              ? 'url(#av-dots-idle)'
              : undefined;
        return (
          <g key={k}>
            <rect
              x={x(s.start)}
              y={top + i * row + 3}
              width={w}
              height={row - 8}
              rx={3}
              fill={fill}
              className={stateClass[state] ?? 'av-span-running'}
            />
            {w > 22 && s.label ? (
              <text
                x={x(s.start) + w / 2}
                y={top + i * row + 17}
                textAnchor="middle"
                fontSize={10}
                className={state === 'running' ? 'av-fill-bg' : 'av-fill-fg'}
              >
                {s.label}
              </text>
            ) : null}
          </g>
        );
      })}
      {Array.from({ length: Math.floor(end / tickStep) + 1 }, (_, k) => k * tickStep).map((t) => (
        <text
          key={t}
          x={x(t)}
          y={H - 6}
          textAnchor="middle"
          fontSize={9}
          className="av-fill-muted"
        >
          {t}
        </text>
      ))}
      {now !== undefined ? (
        <line
          x1={x(now)}
          x2={x(now)}
          y1={top}
          y2={H - 18}
          className="av-stroke-fg"
          strokeWidth={1.5}
          strokeDasharray="3 2"
        />
      ) : null}
    </svg>
  );
}
