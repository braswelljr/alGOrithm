import type { ArrayStep } from '../core/replay/array';

/**
 * A search's array as cells. Cells outside the range still searched are faded
 * and struck through; the probe is outlined with a "probe" tag; the found cell
 * gets a double border and the word "found".
 */
export function CellsView({ state, quiet = false }: { state: ArrayStep; quiet?: boolean }) {
  const probe = new Set(quiet ? [] : state.compared.concat(state.visited));
  const [lo, hi] = state.range ?? [0, state.values.length - 1];
  return (
    <ol
      className="av-cells"
      aria-label="The array searched"
    >
      {state.values.map((v, i) => {
        const outside = state.range !== null && (i < lo || i > hi);
        const found = state.found === i;
        const probing = probe.has(i);
        return (
          <li
            key={i}
            className="av-cell"
            data-outside={outside || undefined}
            data-probe={probing || undefined}
            data-found={found || undefined}
          >
            {probing ? <span className="av-tag">probe</span> : null}
            {found ? <span className="av-tag">found</span> : null}
            <span>{v}</span>
            <small>{i}</small>
            {outside ? <span className="av-sr-only">(ruled out)</span> : null}
          </li>
        );
      })}
    </ol>
  );
}
