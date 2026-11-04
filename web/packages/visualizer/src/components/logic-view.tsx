import { rowBit } from '../core/replay/circuit';
import type { KMap, LogicTable } from '../core/schema';

const outText = (v: number) => (v === 2 ? 'X' : String(v));

/**
 * A truth table, one row per combination of the inputs, revealing rows up to
 * `upTo` and marking the current one. True rows are tinted and bold, so the
 * output reads without color too.
 */
export function TruthTableView({ table, upTo }: { table: LogicTable; upTo: number }) {
  const n = table.vars.length;
  const current = upTo - 1;
  return (
    <div className="av:max-h-[28rem] av:overflow-auto av:rounded-lg av:border av:border-border">
      <table className="av:w-full av:border-collapse av:font-mono av:text-sm av:tabular-nums">
        <caption className="av-sr-only">
          Truth table{table.expression ? ` of ${table.expression}` : ''}: {n} inputs,{' '}
          {table.out.length} rows
        </caption>
        <thead className="av:sticky av:top-0 av:bg-subtle">
          <tr>
            <th
              scope="col"
              className="av:px-2 av:py-1.5 av:text-left av:text-xs av:font-medium av:text-muted"
            >
              row
            </th>
            {table.vars.map((v) => (
              <th
                key={v}
                scope="col"
                className="av:px-3 av:py-1.5 av:text-center av:font-semibold av:text-fg"
              >
                {v}
              </th>
            ))}
            <th
              scope="col"
              className="av:border-l av:border-border av:px-3 av:py-1.5 av:text-center av:font-semibold av:text-fg"
            >
              {table.expression && table.expression.length <= 24 ? table.expression : 'out'}
            </th>
          </tr>
        </thead>
        <tbody>
          {table.out.map((o, r) => {
            const shown = r < upTo;
            const isNow = r === current;
            return (
              <tr
                key={r}
                aria-current={isNow ? 'step' : undefined}
                className={`av:border-t av:border-border av:transition-[background-color,opacity] av:duration-200 ${
                  isNow ? 'av:bg-accent/15' : o === 1 && shown ? 'av:bg-high/10' : ''
                } ${shown ? 'av:opacity-100' : 'av:opacity-35'}`}
              >
                <td className="av:px-2 av:py-1 av:text-xs av:text-muted">{r}</td>
                {table.vars.map((v, k) => (
                  <td
                    key={v}
                    className="av:px-3 av:py-1 av:text-center av:text-fg"
                  >
                    {rowBit(r, k, n)}
                  </td>
                ))}
                <td
                  className={`av:border-l av:border-border av:px-3 av:py-1 av:text-center ${
                    o === 1 ? 'av:font-bold av:text-high' : 'av:text-muted'
                  }`}
                >
                  {shown ? outText(o) : '?'}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

/** Each group's rectangles: a group that wraps around an edge is split into
 * the runs it covers on each side. */
function rects(k: KMap, cells: number[]) {
  const rows = [...new Set(cells.map((c) => Math.floor(c / k.cols)))].sort((a, b) => a - b);
  const cols = [...new Set(cells.map((c) => c % k.cols))].sort((a, b) => a - b);
  const runs = (xs: number[]) => {
    const out: [number, number][] = [];
    for (const x of xs) {
      const last = out[out.length - 1];
      if (last && x === last[1] + 1) last[1] = x;
      else out.push([x, x]);
    }
    return out;
  };
  const out: { r0: number; r1: number; c0: number; c1: number }[] = [];
  for (const [r0, r1] of runs(rows))
    for (const [c0, c1] of runs(cols)) out.push({ r0, r1, c0, c1 });
  return out;
}

/** Outline styles per group: each has its own dash pattern and number. */
const groupStyle = [
  { color: 'var(--av-accent)', dash: undefined },
  { color: 'var(--av-swap)', dash: '6 3' },
  { color: 'var(--av-sorted)', dash: '2 3' },
  { color: 'var(--av-write)', dash: '8 3 2 3' },
  { color: 'var(--av-compare)', dash: undefined },
  { color: 'var(--av-pivot)', dash: '6 3' }
];

/**
 * A Karnaugh map: rows and columns in Gray code, so neighbors (wrapping
 * around) differ in one variable. Groups of the minimal cover appear one at a
 * time as rounded outlines, each numbered and listed with its term.
 */
export function KMapView({
  kmap,
  shown,
  current
}: {
  kmap: KMap;
  shown: number;
  current?: number;
}) {
  const cell = 56;
  const head = 44;
  const W = head + kmap.cols * cell;
  const H = head + kmap.rows * cell;
  return (
    <div className="av:flex av:flex-wrap av:items-start av:gap-6">
      <svg
        viewBox={`0 0 ${W + 8} ${H + 8}`}
        className="av:w-full av:max-w-sm"
        role="img"
        aria-label={`Karnaugh map, rows ${kmap.rowVars} and columns ${kmap.colVars}, ${shown} of ${kmap.groups.length} groups shown`}
      >
        <text
          x={6}
          y={14}
          className="av:fill-muted av:font-mono av:text-[10px]"
        >
          {kmap.rowVars}\{kmap.colVars}
        </text>
        {kmap.colLabels.map((l, c) => (
          <text
            key={`c${c}`}
            x={head + c * cell + cell / 2}
            y={32}
            textAnchor="middle"
            className="av:fill-fg av:font-mono av:text-xs av:font-semibold"
          >
            {l}
          </text>
        ))}
        {kmap.rowLabels.map((l, r) => (
          <text
            key={`r${r}`}
            x={head - 10}
            y={head + r * cell + cell / 2 + 4}
            textAnchor="end"
            className="av:fill-fg av:font-mono av:text-xs av:font-semibold"
          >
            {l}
          </text>
        ))}
        {kmap.cells.map((v, i) => {
          const r = Math.floor(i / kmap.cols);
          const c = i % kmap.cols;
          return (
            <g key={i}>
              <rect
                x={head + c * cell}
                y={head + r * cell}
                width={cell}
                height={cell}
                className={`av:stroke-border ${v === 1 ? 'av:fill-high/15' : 'av:fill-card'}`}
              />
              <text
                x={head + c * cell + cell / 2}
                y={head + r * cell + cell / 2 + 5}
                textAnchor="middle"
                className={`av:font-mono av:text-base ${v === 1 ? 'av:fill-fg av:font-bold' : 'av:fill-muted'}`}
              >
                {outText(v)}
              </text>
              <text
                x={head + c * cell + 4}
                y={head + r * cell + 11}
                className="av:fill-muted av:font-mono av:text-[8px]"
              >
                m{kmap.minterm[i]}
              </text>
            </g>
          );
        })}
        {kmap.groups.slice(0, shown).map((g, gi) => {
          const st = groupStyle[gi % groupStyle.length];
          const inset = 5 + (gi % 3) * 3;
          return rects(kmap, g).map((rc, k) => (
            <rect
              key={`${gi}.${k}`}
              x={head + rc.c0 * cell + inset}
              y={head + rc.r0 * cell + inset}
              width={(rc.c1 - rc.c0 + 1) * cell - 2 * inset}
              height={(rc.r1 - rc.r0 + 1) * cell - 2 * inset}
              rx={14}
              fill="none"
              stroke={st.color}
              strokeWidth={gi === current ? 3.5 : 2.5}
              strokeDasharray={st.dash}
              className="av:origin-center av:animate-[av-group-in_240ms_cubic-bezier(0.23,1,0.32,1)] av:motion-reduce:animate-none"
            />
          ));
        })}
      </svg>
      <ol className="av:m-0 av:flex av:list-none av:flex-col av:gap-1.5 av:p-0 av:text-sm">
        {kmap.groups.map((g, gi) => (
          <li
            key={gi}
            className={`av:flex av:items-center av:gap-2 av:transition-opacity av:duration-200 ${gi < shown ? 'av:opacity-100' : 'av:opacity-30'}`}
          >
            <svg
              width={22}
              height={12}
              aria-hidden="true"
            >
              <rect
                x={1}
                y={1}
                width={20}
                height={10}
                rx={5}
                fill="none"
                stroke={groupStyle[gi % groupStyle.length].color}
                strokeWidth={2}
                strokeDasharray={groupStyle[gi % groupStyle.length].dash}
              />
            </svg>
            <span className="av:text-muted">
              {gi + 1}. {g.length} cells
            </span>
            <code className="av:font-mono av:text-fg">{kmap.terms[gi]}</code>
          </li>
        ))}
        <li
          className={`av:mt-1 av:font-mono av:font-semibold av:text-fg ${shown >= kmap.groups.length ? '' : 'av:opacity-30'}`}
        >
          F = {kmap.sop}
        </li>
      </ol>
    </div>
  );
}
