import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import type { Exec } from '../core/schema';
import { useEnter } from './exec-parts';

const KIND_CLASS: Record<string, string> = {
  keyword: 'decoration-accent',
  ident: 'decoration-pivot',
  number: 'decoration-compare',
  operator: 'decoration-swap',
  punct: 'decoration-muted'
};

/**
 * A lexer's progress: the source with every token read so far underlined in
 * its kind's color, the current one highlighted, and a strip of the tokens
 * as chips naming their kind, so kind never rests on color alone.
 */
export function TokensView({
  exec,
  focus,
  quiet = false
}: {
  exec: Exec;
  focus: number;
  quiet?: boolean;
}) {
  const src = exec.source ?? '';
  const parts: ReactNode[] = [];
  let at = 0;
  exec.tokens.forEach((t, i) => {
    if (t.start > at) parts.push(src.slice(at, t.start));
    const text = src.slice(t.start, t.end);
    if (i > focus || t.kind === 'eof') {
      parts.push(text);
    } else {
      parts.push(
        <mark
          key={i}
          data-current={i === focus}
          className={cn(
            'text-fg data-[current=true]:bg-accent/20 data-[current=true]:ring-accent rounded-sm bg-transparent underline decoration-2 underline-offset-4 transition-colors duration-150 data-[current=true]:ring-2',
            KIND_CLASS[t.kind] ?? ''
          )}
        >
          {text}
        </mark>
      );
    }
    at = t.end;
  });
  if (at < src.length) parts.push(src.slice(at));
  const read = exec.tokens.slice(0, focus + 1);
  const chip = useEnter<HTMLLIElement>(focus, quiet, [
    { transform: 'translateY(6px)', opacity: 0 },
    { transform: 'translateY(0)', opacity: 1 }
  ]);
  return (
    <div className="flex flex-col gap-3">
      <pre
        className="border-border bg-card m-0 max-h-80 overflow-auto rounded-lg border p-3 font-mono text-[13px] leading-7"
        aria-label="Source"
      >
        {parts}
      </pre>
      <ol
        className="m-0 flex max-h-32 list-none flex-wrap gap-1 overflow-auto p-0"
        aria-label={`${read.length} tokens read`}
      >
        {read.map((t, i) => (
          <li
            key={i}
            ref={i === focus ? chip : undefined}
            data-current={i === focus}
            className="border-border bg-subtle data-[current=true]:border-accent flex items-baseline gap-1 rounded-md border px-1.5 py-0.5 font-mono text-xs"
          >
            <span className="text-muted text-[10px]">{t.kind}</span>
            <span>{t.kind === 'eof' ? 'end' : t.text}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
