import { CheckIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

export type PathStep = {
  title: string;
  /** What to notice when you run it. */
  note?: string;
  href?: string;
  done?: boolean;
};

/**
 * A learning path's steps in order, as an ordered list: a numbered node per
 * step (a check once done, with "done" said in text), a line joining them,
 * and the current step marked with aria-current.
 *
 *   <PathStepper steps={steps} current={2} />
 */
export function PathStepper({
  steps,
  current,
  className
}: {
  steps: PathStep[];
  current?: number;
  className?: string;
}) {
  return (
    <ol className={cn('flex flex-col', className)}>
      {steps.map((s, i) => {
        const here = i === current;
        return (
          <li
            key={i}
            aria-current={here ? 'step' : undefined}
            className="relative flex gap-3 pb-6 last:pb-0"
          >
            {i < steps.length - 1 ? (
              <span
                aria-hidden
                className={cn(
                  'absolute top-8 bottom-0 left-4 w-px -translate-x-1/2',
                  s.done ? 'bg-primary' : 'bg-border'
                )}
              />
            ) : null}
            <span
              className={cn(
                'relative flex size-8 shrink-0 items-center justify-center rounded-full border text-sm font-semibold tabular-nums',
                s.done
                  ? 'border-accent bg-primary text-accent-fg'
                  : here
                    ? 'border-accent bg-card text-accent ring-accent/15 ring-4'
                    : 'border-control bg-card text-muted'
              )}
            >
              {s.done ? (
                <CheckIcon
                  aria-hidden
                  className="size-4"
                />
              ) : (
                i + 1
              )}
              <span className="sr-only">{s.done ? ', done' : here ? ', current step' : ''}</span>
            </span>
            <div className="flex min-w-0 flex-col gap-1 pt-1">
              {s.href ? (
                <a
                  href={s.href}
                  className="text-fg font-medium underline-offset-4 hover:underline"
                >
                  {s.title}
                </a>
              ) : (
                <span className="text-fg font-medium">{s.title}</span>
              )}
              {s.note ? <p className="text-muted text-sm text-pretty">{s.note}</p> : null}
            </div>
          </li>
        );
      })}
    </ol>
  );
}
