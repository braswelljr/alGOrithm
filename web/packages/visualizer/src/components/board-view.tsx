import type { TableState } from '../core/replay/table';
import type { Table } from '../core/schema';

/**
 * A square board for backtracking: N-queens (cells 1 are queens, drawn as Q)
 * or a puzzle of digits (sudoku; 0 is empty). The cell written last is
 * outlined; cells being checked are dashed; 3x3 boxes are ruled when the board
 * is 9 by 9.
 */
export function BoardView({ table, state }: { table: Table; state?: TableState }) {
  const values = state?.cells ?? table.cells;
  const queens = table.cells.every((v) => v === 0 || v === 1);
  const reads = new Set(state?.reads ?? []);
  const n = table.cols;
  return (
    <div
      className="av-board"
      role="grid"
      aria-label={`A ${n} by ${n} board`}
      style={{ gridTemplateColumns: `repeat(${n}, minmax(0, 2.25rem))` }}
      data-boxes={n === 9 || undefined}
    >
      {values.map((v, k) => {
        const r = Math.floor(k / n);
        const c = k % n;
        const piece = queens ? (v === 1 ? 'Q' : '') : v ? String(v) : '';
        return (
          <div
            key={k}
            role="gridcell"
            className="av-bcell"
            data-dark={(r + c) % 2 === 1 || undefined}
            data-current={state?.current === k || undefined}
            data-read={reads.has(k) || undefined}
            data-box-right={n === 9 && c % 3 === 2 && c < 8 ? true : undefined}
            data-box-bottom={n === 9 && r % 3 === 2 && r < 8 ? true : undefined}
            aria-label={`Row ${r + 1}, column ${c + 1}: ${piece ? (queens ? 'queen' : piece) : 'empty'}`}
          >
            {piece}
          </div>
        );
      })}
    </div>
  );
}
