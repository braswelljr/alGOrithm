import { useEffect, useRef } from 'react';

const EASE = 'cubic-bezier(0.23, 1, 0.32, 1)';

/**
 * A pushdown automaton's stack, top at the top: a new top drops in, and the
 * bottom marker "$" is shown apart. With reduced motion it just changes.
 */
export function StackView({ stack, quiet = false }: { stack: string; quiet?: boolean }) {
  const items = [...stack];
  const topRef = useRef<HTMLLIElement | null>(null);
  const before = useRef(items.length);
  useEffect(() => {
    const grew = items.length > before.current;
    before.current = items.length;
    if (grew && !quiet) {
      topRef.current?.animate(
        [
          { transform: 'translateY(-10px)', opacity: 0 },
          { transform: 'translateY(0)', opacity: 1 }
        ],
        { duration: 200, easing: EASE }
      );
    }
  });
  return (
    <figure
      className="av:m-0 av:flex av:flex-col av:items-center av:gap-1"
      aria-label={`Stack, bottom first: ${items.join(' ') || 'empty'}`}
    >
      <span className="av:text-[11px] av:font-medium av:tracking-wide av:text-muted av:uppercase">
        Stack
      </span>
      <ol className="av:flex av:min-h-10 av:w-16 av:flex-col-reverse av:items-stretch av:rounded-b-md av:border-x-2 av:border-b-2 av:border-fg av:p-1">
        {items.map((c, i) => (
          <li
            key={`${i}-${c}`}
            ref={i === items.length - 1 ? topRef : undefined}
            className="av:mt-1 av:flex av:h-7 av:items-center av:justify-center av:rounded av:border av:border-border av:bg-card av:font-mono av:text-sm av:data-[bottom=true]:text-muted av:data-[top=true]:border-accent av:data-[top=true]:font-bold"
            data-top={i === items.length - 1}
            data-bottom={c === '$' && i === 0}
          >
            {c}
          </li>
        ))}
      </ol>
      {items.length ? (
        <span className="av:text-[11px] av:text-muted">top: {items[items.length - 1]}</span>
      ) : (
        <span className="av:text-[11px] av:text-muted">empty</span>
      )}
    </figure>
  );
}
