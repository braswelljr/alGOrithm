import { useEffect, useRef } from 'react';

const CELL = 40;
const EASE = 'cubic-bezier(0.23, 1, 0.32, 1)';

/**
 * A Turing machine's tape: cells slide under a fixed head, the state sits on
 * the head, and a cell just written flips in. Blanks show as an empty cell.
 * With reduced motion the tape jumps and nothing flips.
 */
export function TapeView({
  tape,
  head,
  state,
  quiet = false
}: {
  tape: string;
  head: number;
  state?: string;
  quiet?: boolean;
}) {
  const cells = [...(tape || '_')];
  const prev = useRef<string[]>(cells);
  const refs = useRef<(HTMLLIElement | null)[]>([]);
  useEffect(() => {
    const before = prev.current;
    prev.current = cells;
    if (quiet) return;
    cells.forEach((c, i) => {
      // A write changes the cell under the head; a tape that grew on the
      // left shifts every index, so compare against the old cell at the
      // same distance from the end.
      const old = before[before.length - (cells.length - i)];
      if (old !== undefined && old !== c && i === head) {
        refs.current[i]?.animate(
          [
            { transform: 'rotateX(90deg)', opacity: 0.4 },
            { transform: 'rotateX(0deg)', opacity: 1 }
          ],
          { duration: 220, easing: EASE }
        );
      }
    });
  });
  const offset = -(head + 0.5) * CELL;
  return (
    <figure
      className="m-0 flex flex-col items-center gap-2"
      aria-label={`Tape ${cells.join('')}, head on cell ${head + 1} reading ${cells[head] === '_' ? 'a blank' : cells[head]}${state ? `, state ${state}` : ''}`}
    >
      <div className="flex flex-col items-center">
        {state ? (
          <span className="bg-primary text-accent-fg mb-1 rounded-full px-2.5 py-0.5 font-mono text-xs font-semibold shadow-sm">
            {state}
          </span>
        ) : null}
        <span
          aria-hidden
          className="border-t-accent h-0 w-0 border-x-[7px] border-t-[8px] border-x-transparent"
        />
      </div>
      <div className="relative h-12 w-full overflow-hidden [mask-image:linear-gradient(to_right,transparent,black_12%,black_88%,transparent)]">
        <ol
          className="absolute top-0 left-1/2 flex [perspective:400px]"
          style={{
            transform: `translateX(${offset}px)`,
            transition: quiet ? undefined : `transform 220ms ${EASE}`
          }}
        >
          {cells.map((c, i) => (
            <li
              key={i}
              ref={(el) => {
                refs.current[i] = el;
              }}
              className="border-border bg-card data-[head=true]:border-accent flex h-12 items-center justify-center border-y border-r font-mono text-base first:border-l data-[head=true]:bg-[color-mix(in_oklch,var(--av-accent)_14%,var(--av-card))] data-[head=true]:font-bold"
              style={{ width: CELL }}
              data-head={i === head}
            >
              {c === '_' ? <span className="text-muted opacity-40">_</span> : c}
            </li>
          ))}
        </ol>
      </div>
    </figure>
  );
}
