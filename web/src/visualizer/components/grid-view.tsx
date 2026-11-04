import type { CSSProperties } from 'react';
import type { GridState } from '../core/replay/grid';
import type { Grid } from '../core/schema';

/**
 * A pathfinding grid. Walls are hatched; open cells shade with their cost and
 * show it when above 1; the frontier is outlined; explored cells are dotted;
 * the cell explored last has a thick outline; the path is filled with a dot;
 * S and E mark the start and the end. Pass onToggle to let people draw walls.
 */
export function GridView({
  grid,
  state,
  onToggle,
  label = 'The grid'
}: {
  grid: Grid;
  state?: GridState;
  onToggle?: (cell: number) => void;
  label?: string;
}) {
  const path = new Set(state?.path ?? []);
  const cells = grid.cost.map((cost, i) => {
    const r = Math.floor(i / grid.cols) + 1;
    const c = (i % grid.cols) + 1;
    const start = i === grid.start;
    const end = i === grid.end;
    const attrs = {
      className: 'av-gcell',
      'data-wall': cost === 0 || undefined,
      'data-cost': cost > 1 || undefined,
      'data-frontier': (state?.frontier.has(i) && !path.has(i)) || undefined,
      'data-visited': (state?.visited.has(i) && !path.has(i)) || undefined,
      'data-current': state?.current === i || undefined,
      'data-path': path.has(i) || undefined,
      'data-start': start || undefined,
      'data-end': end || undefined,
      style: cost > 1 ? ({ '--av-cost': cost } as CSSProperties) : undefined
    };
    const text = grid.labels[i] || (start ? 'S' : end ? 'E' : cost > 1 ? String(cost) : '');
    const name = `Row ${r}, column ${c}: ${grid.labels[i] ? `${grid.labels[i]}, ` : ''}${
      start ? 'start' : end ? 'end' : cost === 0 ? 'wall' : `cost ${cost}`
    }${path.has(i) ? ', on the path' : state?.visited.has(i) ? ', explored' : state?.frontier.has(i) ? ', in the frontier' : ''}`;
    return onToggle && !start && !end ? (
      <button
        key={i}
        type="button"
        {...attrs}
        aria-label={`${name}. Toggle the wall.`}
        onClick={() => onToggle(i)}
      >
        {text}
      </button>
    ) : (
      <div
        key={i}
        {...attrs}
        role="gridcell"
        aria-label={name}
      >
        {text}
      </div>
    );
  });
  return (
    <div
      className="av-grid"
      role={onToggle ? 'group' : 'grid'}
      aria-label={`${label}: ${grid.rows} rows by ${grid.cols} columns`}
      style={
        {
          gridTemplateColumns: `repeat(${grid.cols}, minmax(0, 1fr))`,
          '--av-grid-ratio': `${grid.cols} / ${grid.rows}`
        } as CSSProperties
      }
    >
      {cells}
    </div>
  );
}
