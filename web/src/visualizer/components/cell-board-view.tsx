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
      className="m-0 flex flex-col items-center gap-2"
      aria-label={`Generation ${generation}: ${alive} live cells on a ${rows} by ${cols} board`}
    >
      <div
        className="border-border bg-border grid gap-px overflow-auto rounded-[var(--av-radius)] border p-px"
        style={style}
        aria-hidden
      >
        {cells.map((c, i) => (
          <span
            key={i}
            className="bg-card data-[alive=true]:bg-primary data-[pending=true]:bg-subtle data-[pending=true]:opacity-50"
            data-alive={c === 1}
            data-pending={c === null}
            style={quiet ? undefined : { transition: 'background-color 180ms ease-out' }}
          />
        ))}
      </div>
      <figcaption className="text-muted text-xs">
        Generation {generation} · {alive} alive
      </figcaption>
    </figure>
  );
}
