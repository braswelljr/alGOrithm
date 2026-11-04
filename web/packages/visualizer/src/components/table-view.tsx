import type { TableState } from '../core/replay/table';
import type { Table } from '../core/schema';

/**
 * A dynamic-programming table. Cells fill in as they are computed; the cell
 * written last is outlined and tinted; the cells it read have a dashed border;
 * the cells of the answer, once traced back, are filled and bold.
 */
export function TableView({ table, state }: { table: Table; state?: TableState }) {
  const values = state?.cells ?? table.cells;
  const reads = new Set(state?.reads ?? []);
  const path = new Set(state?.path ?? []);
  const rows = Array.from({ length: table.rows }, (_, r) => r);
  const cols = Array.from({ length: table.cols }, (_, c) => c);
  return (
    <div className="av-table-wrap">
      <table className="av-table">
        <caption className="av-sr-only">
          The table, {table.rows} by {table.cols}
        </caption>
        {table.colLabels.length > 0 ? (
          <thead>
            <tr>
              {table.rowLabels.length > 0 ? <td aria-hidden /> : null}
              {cols.map((c) => (
                <th
                  key={c}
                  scope="col"
                >
                  {table.colLabels[c] || (c === 0 && table.rowLabels.length ? '' : c)}
                </th>
              ))}
            </tr>
          </thead>
        ) : null}
        <tbody>
          {rows.map((r) => (
            <tr key={r}>
              {table.rowLabels.length > 0 ? <th scope="row">{table.rowLabels[r] || ''}</th> : null}
              {cols.map((c) => {
                const k = r * table.cols + c;
                const v = values[k];
                return (
                  <td
                    key={c}
                    data-empty={v === null || undefined}
                    data-read={reads.has(k) || undefined}
                    data-current={state?.current === k || undefined}
                    data-path={path.has(k) || undefined}
                    data-mark={state?.marks.get(k)}
                  >
                    {v === null ? (
                      <span className="av-sr-only">empty</span>
                    ) : table.blank !== undefined && v === table.blank ? (
                      <span aria-label="none">-</span>
                    ) : (
                      v
                    )}
                    {state?.marks.get(k) ? (
                      <span className="av-sr-only"> ({state.marks.get(k)})</span>
                    ) : null}
                    {path.has(k) ? <span className="av-sr-only"> (on the answer)</span> : null}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
