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
      className="av:m-0 av:flex av:flex-col av:items-center av:gap-2"
      aria-label={`Tape ${cells.join('')}, head on cell ${head + 1} reading ${cells[head] === '_' ? 'a blank' : cells[head]}${state ? `, state ${state}` : ''}`}
    >
      <div className="av:flex av:flex-col av:items-center">
        {state ? (
          <span className="av:mb-1 av:rounded-full av:bg-accent av:px-2.5 av:py-0.5 av:font-mono av:text-xs av:font-semibold av:text-accent-fg av:shadow-sm">
            {state}
          </span>
        ) : null}
        <span
          aria-hidden
          className="av:h-0 av:w-0 av:border-x-[7px] av:border-t-[8px] av:border-x-transparent av:border-t-accent"
        />
      </div>
      <div className="av:relative av:h-12 av:w-full av:overflow-hidden av:[mask-image:linear-gradient(to_right,transparent,black_12%,black_88%,transparent)]">
        <ol
          className="av:absolute av:top-0 av:left-1/2 av:flex av:[perspective:400px]"
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
              className="av:flex av:h-12 av:items-center av:justify-center av:border-y av:border-r av:border-border av:bg-card av:font-mono av:text-base av:first:border-l av:data-[head=true]:border-accent av:data-[head=true]:bg-[color-mix(in_oklch,var(--av-accent)_14%,var(--av-card))] av:data-[head=true]:font-bold"
              style={{ width: CELL }}
              data-head={i === head}
            >
              {c === '_' ? <span className="av:text-muted av:opacity-40">_</span> : c}
            </li>
          ))}
        </ol>
      </div>
    </figure>
  );
}
