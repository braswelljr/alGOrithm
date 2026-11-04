/**
 * Small pieces the machine views share: a titled panel, a value that rolls
 * in when it changes, and the motion they use. Motion is transform and
 * opacity only, under 300ms, and none at all when quiet (reduced motion).
 */
import { useEffect, useRef, type ReactNode } from 'react';

export const EASE_OUT = 'cubic-bezier(0.23, 1, 0.32, 1)';
/** A spring-like settle with a little overshoot, for things pushed in. */
export const SPRING = 'cubic-bezier(0.34, 1.36, 0.64, 1)';

/** Plays keyframes on the element whenever key changes (not on the first
 * render), unless quiet. */
export function useEnter<T extends HTMLElement | SVGElement>(
  key: unknown,
  quiet: boolean,
  keyframes: Keyframe[],
  options: KeyframeAnimationOptions = { duration: 220, easing: EASE_OUT }
) {
  const ref = useRef<T | null>(null);
  const first = useRef(true);
  useEffect(() => {
    if (first.current) {
      first.current = false;
      return;
    }
    if (!quiet) ref.current?.animate(keyframes, options);
  }, [key, quiet]);
  return ref;
}

export function Panel({
  title,
  children,
  className = ''
}: {
  title: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section
      className={`av:flex av:min-w-0 av:flex-col av:gap-2 av:rounded-lg av:border av:border-border av:bg-card av:p-3 ${className}`}
      aria-label={title}
    >
      <h4 className="av:m-0 av:text-[11px] av:font-medium av:tracking-wide av:text-muted av:uppercase">
        {title}
      </h4>
      {children}
    </section>
  );
}

/** A value in a cell; when it changes it rolls in from above and the cell is
 * marked changed (an outline and the word "changed" for screen readers). */
export function Value({
  name,
  value,
  changed,
  quiet
}: {
  name: string;
  value: string;
  changed: boolean;
  quiet: boolean;
}) {
  const ref = useEnter<HTMLSpanElement>(changed ? value : null, quiet || !changed, [
    { transform: 'translateY(-60%)', opacity: 0 },
    { transform: 'translateY(0)', opacity: 1 }
  ]);
  return (
    <div
      className="av:flex av:min-w-0 av:flex-col av:items-center av:overflow-hidden av:rounded-md av:border av:border-border av:bg-subtle av:px-2 av:py-1 av:transition-colors av:duration-200 av:data-[changed=true]:border-accent av:data-[changed=true]:bg-accent/10"
      data-changed={changed}
    >
      <span className="av:font-mono av:text-[11px] av:text-muted">{name}</span>
      <span
        ref={ref}
        className="av:font-mono av:text-sm av:font-semibold av:tabular-nums"
      >
        {value}
      </span>
      {changed ? <span className="av:sr-only">changed</span> : null}
    </div>
  );
}

/** A code listing with the current line marked and kept in view. */
export function Listing({
  code,
  pc,
  quiet,
  label
}: {
  code: string[];
  pc: number;
  quiet: boolean;
  label: string;
}) {
  const current = useRef<HTMLLIElement | null>(null);
  useEffect(() => {
    current.current?.scrollIntoView({ block: 'nearest', behavior: quiet ? 'auto' : 'smooth' });
  }, [pc, quiet]);
  return (
    <Panel title={label}>
      <ol className="av:m-0 av:max-h-72 av:list-none av:overflow-auto av:p-0 av:font-mono av:text-[13px] av:leading-6">
        {code.map((line, i) => (
          <li
            key={i}
            ref={i === pc ? current : undefined}
            aria-current={i === pc ? 'step' : undefined}
            className="av:flex av:gap-2 av:rounded av:border-l-2 av:border-transparent av:px-2 av:whitespace-pre av:text-muted av:transition-colors av:duration-150 av:aria-[current=step]:border-accent av:aria-[current=step]:bg-accent/10 av:aria-[current=step]:text-fg"
          >
            <span
              aria-hidden
              className="av:w-3 av:text-accent"
            >
              {i === pc ? '▸' : ''}
            </span>
            {line || ' '}
          </li>
        ))}
      </ol>
    </Panel>
  );
}
