import { pegName, type HanoiState } from '../core/replay/hanoi';

/** Three pegs and their disks. The disk just moved is outlined and tinted. */
export function HanoiView({ state, disks }: { state: HanoiState; disks: number }) {
  const W = 300;
  const H = 30 + disks * 14;
  const unit = 80 / Math.max(disks, 1);
  return (
    <svg
      viewBox={`0 0 ${W} ${H + 16}`}
      className="av-hanoi"
      role="img"
      aria-label={
        state.move
          ? `Moved the top disk from peg ${pegName(state.move.from)} to peg ${pegName(state.move.to)}. Pegs: ${state.pegs.map((p, i) => `${pegName(i)} has ${p.join(', ') || 'nothing'}`).join('; ')}`
          : 'All disks on peg A'
      }
    >
      {state.pegs.map((peg, p) => {
        const cx = 50 + p * 100;
        return (
          <g key={p}>
            <rect
              x={cx - 1.5}
              y={10}
              width={3}
              height={H - 10}
              className="av-fill-border"
            />
            <rect
              x={cx - 45}
              y={H}
              width={90}
              height={3}
              className="av-fill-muted"
            />
            <text
              x={cx}
              y={H + 14}
              textAnchor="middle"
              fontSize={10}
              className="av-fill-fg"
            >
              {pegName(p)}
            </text>
            {peg.map((d, k) => {
              const w = 10 + d * unit;
              const moved = state.move?.to === p && k === peg.length - 1;
              return (
                <g key={d}>
                  <rect
                    x={cx - w / 2}
                    y={H - (k + 1) * 14}
                    width={w}
                    height={12}
                    rx={3}
                    className={moved ? 'av-fill-compare av-stroke-fg' : 'av-fill-pivot'}
                    strokeWidth={moved ? 1.5 : 0}
                  />
                  <text
                    x={cx}
                    y={H - (k + 1) * 14 + 9}
                    textAnchor="middle"
                    fontSize={8}
                    className="av-fill-bg"
                  >
                    {d}
                  </text>
                </g>
              );
            })}
          </g>
        );
      })}
    </svg>
  );
}
