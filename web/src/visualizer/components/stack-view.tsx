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
      className="m-0 flex flex-col items-center gap-1"
      aria-label={`Stack, bottom first: ${items.join(' ') || 'empty'}`}
    >
      <span className="text-muted text-[11px] font-medium tracking-wide uppercase">Stack</span>
      <ol className="border-fg flex min-h-10 w-16 flex-col-reverse items-stretch rounded-b-md border-x-2 border-b-2 p-1">
        {items.map((c, i) => (
          <li
            key={`${i}-${c}`}
            ref={i === items.length - 1 ? topRef : undefined}
            className="border-border bg-card data-[bottom=true]:text-muted data-[top=true]:border-accent mt-1 flex h-7 items-center justify-center rounded border font-mono text-sm data-[top=true]:font-bold"
            data-top={i === items.length - 1}
            data-bottom={c === '$' && i === 0}
          >
            {c}
          </li>
        ))}
      </ol>
      {items.length ? (
        <span className="text-muted text-[11px]">top: {items[items.length - 1]}</span>
      ) : (
        <span className="text-muted text-[11px]">empty</span>
      )}
    </figure>
  );
}
