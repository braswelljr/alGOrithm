import type { ReactNode } from 'react';
import type { Exec } from '../core/schema';
import { useEnter } from './exec-parts';

const KIND_CLASS: Record<string, string> = {
  keyword: 'av:decoration-accent',
  ident: 'av:decoration-pivot',
  number: 'av:decoration-compare',
  operator: 'av:decoration-swap',
  punct: 'av:decoration-muted'
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
          className={`av:rounded-sm av:bg-transparent av:text-fg av:underline av:decoration-2 av:underline-offset-4 av:transition-colors av:duration-150 av:data-[current=true]:bg-accent/20 av:data-[current=true]:ring-2 av:data-[current=true]:ring-accent ${KIND_CLASS[t.kind] ?? ''}`}
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
    <div className="av:flex av:flex-col av:gap-3">
      <pre
        className="av:m-0 av:max-h-80 av:overflow-auto av:rounded-lg av:border av:border-border av:bg-card av:p-3 av:font-mono av:text-[13px] av:leading-7"
        aria-label="Source"
      >
        {parts}
      </pre>
      <ol
        className="av:m-0 av:flex av:max-h-32 av:list-none av:flex-wrap av:gap-1 av:overflow-auto av:p-0"
        aria-label={`${read.length} tokens read`}
      >
        {read.map((t, i) => (
          <li
            key={i}
            ref={i === focus ? chip : undefined}
            data-current={i === focus}
            className="av:flex av:items-baseline av:gap-1 av:rounded-md av:border av:border-border av:bg-subtle av:px-1.5 av:py-0.5 av:font-mono av:text-xs av:data-[current=true]:border-accent"
          >
            <span className="av:text-[10px] av:text-muted">{t.kind}</span>
            <span>{t.kind === 'eof' ? 'end' : t.text}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
