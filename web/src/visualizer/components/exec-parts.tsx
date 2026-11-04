/**
 * Small pieces the machine views share: a titled panel, a value that rolls
 * in when it changes, and the motion they use. Motion is transform and
 * opacity only, under 300ms, and none at all when quiet (reduced motion).
 */
import { useEffect, useRef, type ReactNode } from 'react';
import { CodeSample } from '@/components/ui/code-block';
import { cn } from '@/lib/utils';

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
      className={cn(
        'border-border bg-card flex min-w-0 flex-col gap-2 rounded-lg border p-3',
        className
      )}
      aria-label={title}
    >
      <h4 className="text-muted m-0 text-[11px] font-medium tracking-wide uppercase">{title}</h4>
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
      className="border-border bg-subtle data-[changed=true]:border-accent data-[changed=true]:bg-accent/10 flex min-w-0 flex-col items-center overflow-hidden rounded-md border px-2 py-1 transition-colors duration-200"
      data-changed={changed}
    >
      <span className="text-muted font-mono text-[11px]">{name}</span>
      <span
        ref={ref}
        className="font-mono text-sm font-semibold tabular-nums"
      >
        {value}
      </span>
      {changed ? <span className="sr-only">changed</span> : null}
    </div>
  );
}

/**
 * A code listing with the current line marked and kept in view, drawn as the
 * site's code block: the running line highlighted and a copy button (each line
 * already starts with its address, so the block adds no numbers of its own).
 */
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
  const host = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    // Scrolls the code block's own scroller, not the page.
    const line = host.current?.querySelector<HTMLElement>(`[data-line="${pc + 1}"]`);
    const box = host.current?.querySelector<HTMLElement>('[data-slot=code-block-content]');
    if (!line || !box) return;
    const top = line.offsetTop - box.clientHeight / 2 + line.clientHeight / 2;
    box.scrollTo({ top, behavior: quiet ? 'auto' : 'smooth' });
  }, [pc, quiet]);
  return (
    <div
      ref={host}
      className="min-w-0 [&_[data-slot=code-block-content]]:max-h-72"
    >
      <CodeSample
        title={label}
        language="text"
        code={code.join('\n')}
        highlight={[pc + 1]}
        size="sm"
      />
      {/* The running line, in words, for a screen reader. */}
      <p
        className="sr-only"
        aria-live="polite"
      >
        Line {pc + 1}: {code[pc] ?? ''}
      </p>
    </div>
  );
}
