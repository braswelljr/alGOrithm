import type { Wave } from '../core/replay/circuit';

/**
 * A timing diagram: one square wave per signal, time to the right, drawn up
 * to tick `upTo` with a cursor there. Rising clock edges get a faint guide
 * line so the reader sees what changes on them; each segment also carries
 * its 1 or 0 as text when there is room.
 */
export function WaveformView({
  waves,
  ticks,
  upTo
}: {
  waves: Wave[];
  ticks: number;
  upTo: number;
}) {
  const left = 72;
  const step = Math.max(10, Math.min(36, 560 / Math.max(ticks, 1)));
  const lane = 34;
  const amp = 18;
  const W = left + ticks * step + 12;
  const H = waves.length * lane + 20;
  const clk = waves.find((w) => w.name === 'CLK');
  const edges: number[] = [];
  if (clk) for (let t = 1; t < ticks; t++) if (clk.levels[t] && !clk.levels[t - 1]) edges.push(t);
  const shown = Math.max(0, Math.min(upTo, ticks));
  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className="av:w-full"
      role="img"
      aria-label={`Timing diagram of ${waves.map((w) => w.name).join(', ')} over ${ticks} ticks, up to tick ${shown}`}
    >
      {edges.map((t) => (
        <line
          key={t}
          x1={left + t * step}
          x2={left + t * step}
          y1={4}
          y2={H - 16}
          className="av:stroke-border"
          strokeDasharray="2 4"
        />
      ))}
      {waves.map((w, i) => {
        const y0 = 10 + i * lane;
        const hi = y0 + 4;
        const lo = y0 + 4 + amp;
        let d = '';
        for (let t = 0; t < shown; t++) {
          const y = w.levels[t] ? hi : lo;
          const x = left + t * step;
          d += t === 0 ? `M${x} ${y}` : `L${x} ${y}`;
          d += `L${x + step} ${y}`;
        }
        return (
          <g key={w.name}>
            <text
              x={left - 10}
              y={y0 + 4 + amp / 2 + 4}
              textAnchor="end"
              className="av:fill-fg av:font-mono av:text-xs av:font-semibold"
            >
              {w.name}
            </text>
            <line
              x1={left}
              x2={left + ticks * step}
              y1={lo}
              y2={lo}
              className="av:stroke-border"
            />
            <path
              d={d}
              fill="none"
              strokeWidth={2}
              strokeLinejoin="round"
              className={w.name === 'CLK' ? 'av:stroke-muted' : 'av:stroke-high'}
            />
            {step >= 18
              ? w.levels.slice(0, shown).map((v, t) =>
                  t === 0 || v !== w.levels[t - 1] ? (
                    <text
                      key={t}
                      x={left + t * step + 3}
                      y={(v ? hi : lo) - 3}
                      className="av:fill-muted av:font-mono av:text-[9px]"
                    >
                      {v ? 1 : 0}
                    </text>
                  ) : null
                )
              : null}
          </g>
        );
      })}
      {shown > 0 && shown < ticks ? (
        <line
          x1={left + shown * step}
          x2={left + shown * step}
          y1={2}
          y2={H - 14}
          className="av:stroke-accent"
          strokeWidth={2}
        />
      ) : null}
      {Array.from({ length: Math.floor(ticks / 2) }, (_, c) => (
        <text
          key={c}
          x={left + (2 * c + 1) * step}
          y={H - 3}
          textAnchor="middle"
          className="av:fill-muted av:font-mono av:text-[9px]"
        >
          {step >= 14 || c % 4 === 0 ? c : ''}
        </text>
      ))}
    </svg>
  );
}
