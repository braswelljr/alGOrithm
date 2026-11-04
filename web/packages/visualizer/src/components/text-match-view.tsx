import type { TextState } from '../core/replay/text';

/**
 * A string search: the text letter by letter, the pattern aligned under the
 * letter being compared, the window a rolling hash checked (underlined), and
 * every match found (double border). A mismatch is dashed.
 */
export function TextMatchView({
  text,
  pattern,
  state,
  matches
}: {
  text: string;
  pattern: string;
  state?: TextState;
  matches?: number[];
}) {
  const found = new Set<number>();
  const ranges = state?.found ?? (matches ?? []).map((m) => [m, m + pattern.length - 1] as const);
  for (const [a, b] of ranges) for (let k = a; k <= b; k++) found.add(k);
  const cmp = state?.compare ?? null;
  const win = state?.window ?? null;
  const offset = cmp ? cmp.text - cmp.pattern : (win?.[0] ?? null);
  const mismatch = cmp ? text[cmp.text] !== pattern[cmp.pattern] : false;
  return (
    <div>
      <ol
        className="av-text"
        aria-label="The text"
      >
        {[...text].map((ch, i) => (
          <li
            key={i}
            className="av-char"
            data-found={found.has(i) || undefined}
            data-window={(win && i >= win[0] && i <= win[1]) || undefined}
            data-compare={cmp?.text === i || undefined}
            data-mismatch={(cmp?.text === i && mismatch) || undefined}
          >
            <span>{ch === ' ' ? ' ' : ch}</span>
            <small>{i}</small>
          </li>
        ))}
      </ol>
      {offset !== null && offset >= 0 ? (
        <div
          className="av-pattern-row"
          aria-label={`The pattern, aligned at position ${offset}`}
        >
          {Array.from({ length: offset }, (_, i) => (
            <span
              key={`pad-${i}`}
              data-pad
              aria-hidden
            />
          ))}
          {[...pattern].map((ch, j) => (
            <span
              key={j}
              data-compare={cmp?.pattern === j || undefined}
            >
              {ch === ' ' ? ' ' : ch}
            </span>
          ))}
        </div>
      ) : null}
    </div>
  );
}
