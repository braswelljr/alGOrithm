import { cn } from '@/lib/utils';
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
    <div className="border-border max-h-[28rem] overflow-auto rounded-lg border">
      <table className="w-full border-collapse font-mono text-sm tabular-nums">
        <caption className="av-sr-only">
          Truth table{table.expression ? ` of ${table.expression}` : ''}: {n} inputs,{' '}
          {table.out.length} rows
        </caption>
        <thead className="bg-subtle sticky top-0">
          <tr>
            <th
              scope="col"
              className="text-muted px-2 py-1.5 text-left text-xs font-medium"
            >
              row
            </th>
            {table.vars.map((v) => (
              <th
                key={v}
                scope="col"
                className="text-fg px-3 py-1.5 text-center font-semibold"
              >
                {v}
              </th>
            ))}
            <th
              scope="col"
              className="border-border text-fg border-l px-3 py-1.5 text-center font-semibold"
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
                className={cn(
                  'border-border border-t transition-[background-color,opacity] duration-200',
                  isNow ? 'bg-accent/15' : o === 1 && shown ? 'bg-high/10' : '',
                  shown ? 'opacity-100' : 'opacity-35'
                )}
              >
                <td className="text-muted px-2 py-1 text-xs">{r}</td>
                {table.vars.map((v, k) => (
                  <td
                    key={v}
                    className="text-fg px-3 py-1 text-center"
                  >
                    {rowBit(r, k, n)}
                  </td>
                ))}
                <td
                  className={cn(
                    'border-border border-l px-3 py-1 text-center',
                    o === 1 ? 'text-high font-bold' : 'text-muted'
                  )}
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
    <div className="flex flex-wrap items-start gap-6">
      <svg
        viewBox={`0 0 ${W + 8} ${H + 8}`}
        className="w-full max-w-sm"
        role="img"
        aria-label={`Karnaugh map, rows ${kmap.rowVars} and columns ${kmap.colVars}, ${shown} of ${kmap.groups.length} groups shown`}
      >
        <text
          x={6}
          y={14}
          className="fill-muted font-mono text-[10px]"
        >
          {kmap.rowVars}\{kmap.colVars}
        </text>
        {kmap.colLabels.map((l, c) => (
          <text
            key={`c${c}`}
            x={head + c * cell + cell / 2}
            y={32}
            textAnchor="middle"
            className="fill-fg font-mono text-xs font-semibold"
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
            className="fill-fg font-mono text-xs font-semibold"
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
                className={cn('stroke-border', v === 1 ? 'fill-high/15' : 'fill-card')}
              />
              <text
                x={head + c * cell + cell / 2}
                y={head + r * cell + cell / 2 + 5}
                textAnchor="middle"
                className={cn('font-mono text-base', v === 1 ? 'fill-fg font-bold' : 'fill-muted')}
              >
                {outText(v)}
              </text>
              <text
                x={head + c * cell + 4}
                y={head + r * cell + 11}
                className="fill-muted font-mono text-[8px]"
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
              className="origin-center animate-[av-group-in_240ms_cubic-bezier(0.23,1,0.32,1)] motion-reduce:animate-none"
            />
          ));
        })}
      </svg>
      <ol className="m-0 flex list-none flex-col gap-1.5 p-0 text-sm">
        {kmap.groups.map((g, gi) => (
          <li
            key={gi}
            className={cn(
              'flex items-center gap-2 transition-opacity duration-200',
              gi < shown ? 'opacity-100' : 'opacity-30'
            )}
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
            <span className="text-muted">
              {gi + 1}. {g.length} cells
            </span>
            <code className="text-fg font-mono">{kmap.terms[gi]}</code>
          </li>
        ))}
        <li
          className={cn(
            'text-fg mt-1 font-mono font-semibold',
            shown >= kmap.groups.length ? '' : 'opacity-30'
          )}
        >
          F = {kmap.sop}
        </li>
      </ol>
    </div>
  );
}
