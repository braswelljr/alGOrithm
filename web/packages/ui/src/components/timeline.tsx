import * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

function Timeline({
  defaultValue,
  className,
  ...props
}: React.ComponentProps<'ol'> & { defaultValue?: number }) {
  return (
    <ol
      data-slot="timeline"
      data-orientation="vertical"
      data-default-value={defaultValue}
      className={cn('group/timeline flex flex-col gap-6', className)}
      {...props}
    />
  );
}

function TimelineItem({
  step,
  className,
  ...props
}: React.ComponentProps<'li'> & { step: number }) {
  return (
    <li
      data-slot="timeline-item"
      data-step={step}
      className={cn('group/timeline-item relative flex flex-col', className)}
      {...props}
    />
  );
}

function TimelineHeader({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="timeline-header"
      className={cn('flex items-center gap-3', className)}
      {...props}
    />
  );
}

function TimelineSeparator({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="timeline-separator"
      className={cn(
        'bg-subtle',
        // From under this marker to the next one, across the gap between.
        'absolute top-10 -bottom-5 left-4 w-0.5 -translate-x-1/2 rounded-full',
        'group-last/timeline-item:hidden',
        className
      )}
      {...props}
    />
  );
}

/** The colours a step may wear: what kind of thing happened, at a glance. */
const TIMELINE_TONES = {
  neutral: 'border-border bg-card text-muted',
  primary:
    'border-primary-600 bg-primary-600 text-white dark:border-primary-400 dark:bg-primary-400 dark:text-bg',
  success:
    'border-emerald-600 bg-emerald-600 text-white dark:border-emerald-400 dark:bg-emerald-400 dark:text-bg',
  info: 'border-sky-600 bg-sky-600 text-white dark:border-sky-400 dark:bg-sky-400 dark:text-bg',
  warning:
    'border-amber-500 bg-amber-500 text-white dark:border-amber-500 dark:bg-amber-500 dark:text-bg',
  danger: 'border-red-600 bg-red-600 text-white dark:border-red-400 dark:bg-red-400 dark:text-bg',
  accent:
    'border-violet-600 bg-violet-600 text-white dark:border-violet-400 dark:bg-violet-400 dark:text-bg'
} as const;

type TimelineTone = keyof typeof TIMELINE_TONES;

/**
 * A step's marker. Current marks where things stand: outlined in its tone
 * with a ring, where the steps before it are filled.
 */
function TimelineIndicator({
  className,
  tone = 'neutral',
  current = false,
  ...props
}: React.ComponentProps<'div'> & { tone?: TimelineTone; current?: boolean }) {
  return (
    <div
      data-slot="timeline-indicator"
      data-current={current || undefined}
      className={cn(
        'relative flex size-8 shrink-0 items-center justify-center rounded-full border-2',
        TIMELINE_TONES[tone],
        current && 'bg-card ring-4 ring-border',
        current &&
          tone === 'primary' &&
          'text-primary-700 ring-primary-100 dark:text-primary-300 dark:ring-primary-900',
        current &&
          tone === 'success' &&
          'text-emerald-700 ring-emerald-100 dark:text-emerald-300 dark:ring-emerald-900',
        current &&
          tone === 'info' &&
          'text-sky-700 ring-sky-100 dark:text-sky-300 dark:ring-sky-900',
        current &&
          tone === 'warning' &&
          'text-amber-700 ring-amber-100 dark:text-amber-300 dark:ring-amber-900',
        current &&
          tone === 'danger' &&
          'text-red-700 ring-red-100 dark:text-red-300 dark:ring-red-900',
        current &&
          tone === 'accent' &&
          'text-violet-700 ring-violet-100 dark:text-violet-300 dark:ring-violet-900',
        className
      )}
      {...props}
    />
  );
}

function TimelineTitle({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="timeline-title"
      className={cn('text-base font-semibold text-fg', className)}
      {...props}
    />
  );
}

function TimelineContent({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="timeline-content"
      className={cn('mt-1 pl-11', className)}
      {...props}
    />
  );
}

function TimelineDate({ className, ...props }: React.ComponentProps<'span'>) {
  return (
    <span
      data-slot="timeline-date"
      className={cn('mt-1 block text-sm text-muted', className)}
      {...props}
    />
  );
}

export type { TimelineTone };
export {
  Timeline,
  TimelineItem,
  TimelineHeader,
  TimelineSeparator,
  TimelineIndicator,
  TimelineTitle,
  TimelineContent,
  TimelineDate
};
