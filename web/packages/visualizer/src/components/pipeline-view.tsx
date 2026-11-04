import { useEffect, useRef } from 'react';
import type { Exec } from '../core/schema';
import { EASE_OUT } from './exec-parts';

const CELL = 44; // px per cycle column
const ROW = 32; // px per instruction row
const LABEL = 200; // px for the instruction column

const STAGE_CLASS: Record<string, string> = {
  IF: 'av:bg-pivot/20 av:border-pivot',
  ID: 'av:bg-compare/20 av:border-compare',
  EX: 'av:bg-accent/20 av:border-accent',
  MEM: 'av:bg-write/20 av:border-write',
  WB: 'av:bg-sorted/20 av:border-sorted',
  stall: 'av:border-dashed av:border-muted av:text-muted',
  flush: 'av:border-danger av:text-danger'
};

/**
 * A pipeline as instructions by cycles, filled up to the current cycle: each
 * stage a chip that slides in as its cycle comes, a stall a dashed bubble, a
 * flushed instruction struck through with an ×. Forwarded values are arrows
 * from the producer's result to the consumer's execute stage.
 */
export function PipelineView({
  exec,
  cycle,
  quiet = false
}: {
  exec: Exec;
  cycle: number;
  quiet?: boolean;
}) {
  const cycles = exec.cells[0]?.length ?? 0;
  const column = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (quiet) return;
    column.current?.parentElement
      ?.querySelectorAll<HTMLElement>(`[data-cycle="${cycle}"][data-stage]`)
      .forEach((el) =>
        el.animate(
          [
            { transform: 'translateX(-10px)', opacity: 0 },
            { transform: 'translateX(0)', opacity: 1 }
          ],
          { duration: 200, easing: EASE_OUT }
        )
      );
  }, [cycle, quiet]);
  const width = LABEL + cycles * CELL;
  const height = ROW * (exec.rows.length + 1);
  const shown = exec.forwards.filter((f) => f.cycle <= cycle);
  return (
    <figure className="av:m-0 av:overflow-x-auto">
      <div
        className="av:relative av:font-mono av:text-xs"
        style={{ width, height }}
        role="table"
        aria-label={`Pipeline at cycle ${cycle + 1} of ${cycles}`}
      >
        <div
          role="row"
          className="av:flex"
          style={{ height: ROW }}
        >
          <span
            role="columnheader"
            className="av:flex av:items-center av:px-2 av:text-muted"
            style={{ width: LABEL }}
          >
            instruction
          </span>
          {Array.from({ length: cycles }, (_, c) => (
            <span
              key={c}
              role="columnheader"
              ref={c === cycle ? column : undefined}
              data-current={c === cycle}
              className="av:flex av:items-center av:justify-center av:text-muted av:data-[current=true]:font-bold av:data-[current=true]:text-accent"
              style={{ width: CELL }}
            >
              {c + 1}
            </span>
          ))}
        </div>
        <div
          aria-hidden
          className="av:pointer-events-none av:absolute av:top-0 av:rounded-md av:bg-accent/5 av:transition-transform av:duration-200 av:ease-out"
          style={{ width: CELL, height, transform: `translateX(${LABEL + cycle * CELL}px)` }}
        />
        {exec.rows.map((label, r) => (
          <div
            key={r}
            role="row"
            className="av:flex av:items-center"
            style={{ height: ROW }}
          >
            <span
              role="rowheader"
              data-flushed={exec.flushed[r]}
              className="av:truncate av:px-2 av:data-[flushed=true]:text-muted av:data-[flushed=true]:line-through"
              style={{ width: LABEL }}
              title={label}
            >
              {label}
            </span>
            {exec.cells[r].map((stage, c) => {
              const visible = stage !== '' && c <= cycle;
              return (
                <span
                  key={c}
                  role="cell"
                  className="av:flex av:items-center av:justify-center"
                  style={{ width: CELL, height: ROW }}
                >
                  {visible ? (
                    <span
                      data-stage={stage}
                      data-cycle={c}
                      className={`av:flex av:h-6 av:w-10 av:items-center av:justify-center av:rounded av:border ${STAGE_CLASS[stage] ?? ''}`}
                    >
                      {stage === 'flush' ? '×' : stage === 'stall' ? 'stall' : stage}
                      {stage === 'flush' ? <span className="av:sr-only">flushed</span> : null}
                    </span>
                  ) : null}
                </span>
              );
            })}
          </div>
        ))}
        <svg
          aria-hidden
          className="av:pointer-events-none av:absolute av:top-0 av:left-0 av:text-accent"
          width={width}
          height={height}
        >
          <defs>
            <marker
              id="av-fwd-head"
              viewBox="0 0 8 8"
              refX="7"
              refY="4"
              markerWidth="6"
              markerHeight="6"
              orient="auto"
            >
              <path
                d="M0,0 L8,4 L0,8 z"
                fill="currentColor"
              />
            </marker>
          </defs>
          {shown.map((f, i) => {
            const x1 = LABEL + (f.cycle - 1) * CELL + CELL / 2;
            const y1 = ROW * (f.from + 1) + ROW / 2 + 8;
            const x2 = LABEL + f.cycle * CELL + CELL / 2 - 12;
            const y2 = ROW * (f.to + 1) + ROW / 2 - 8;
            return (
              <path
                key={i}
                d={`M${x1},${y1} C${x1 + 16},${y1} ${x2 - 16},${y2} ${x2},${y2}`}
                fill="none"
                stroke="currentColor"
                strokeWidth={1.5}
                markerEnd="url(#av-fwd-head)"
              />
            );
          })}
        </svg>
      </div>
      {shown.length ? (
        <figcaption className="av:mt-2 av:text-xs av:text-muted">
          Forwarded so far:{' '}
          {shown
            .map(
              (f) => `r${f.reg} from row ${f.from + 1} to row ${f.to + 1} at cycle ${f.cycle + 1}`
            )
            .join('; ')}
        </figcaption>
      ) : null}
    </figure>
  );
}
