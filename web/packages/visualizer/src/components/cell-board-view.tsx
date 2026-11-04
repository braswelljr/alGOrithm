import type { CSSProperties } from 'react';

/**
 * A board of cells alive or dead, for cellular automata: a live cell is a
 * filled square, a dead one an outline, a cell not computed yet empty.
 * Births and deaths fade, unless motion is reduced.
 */
export function CellBoardView({
  rows,
  cols,
  cells,
  generation,
  quiet = false
}: {
  rows: number;
  cols: number;
  cells: (number | null)[];
  generation: number;
  quiet?: boolean;
}) {
  const alive = cells.filter((c) => c === 1).length;
  const size = cols > 60 ? 8 : cols > 30 ? 12 : 18;
  const style: CSSProperties = {
    gridTemplateColumns: `repeat(${cols}, ${size}px)`,
    gridAutoRows: `${size}px`
  };
  return (
    <figure
      className="av:m-0 av:flex av:flex-col av:items-center av:gap-2"
      aria-label={`Generation ${generation}: ${alive} live cells on a ${rows} by ${cols} board`}
    >
      <div
        className="av:grid av:gap-px av:overflow-auto av:rounded-[var(--av-radius)] av:border av:border-border av:bg-border av:p-px"
        style={style}
        aria-hidden
      >
        {cells.map((c, i) => (
          <span
            key={i}
            className="av:bg-card av:data-[alive=true]:bg-accent av:data-[pending=true]:bg-subtle av:data-[pending=true]:opacity-50"
            data-alive={c === 1}
            data-pending={c === null}
            style={quiet ? undefined : { transition: 'background-color 180ms ease-out' }}
          />
        ))}
      </div>
      <figcaption className="av:text-xs av:text-muted">
        Generation {generation} · {alive} alive
      </figcaption>
    </figure>
  );
}
