import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

/** A figure: a label, a value, and optional trend, indicator and description.
 *  `tile` (the default) frames it; `plain` sets it inline, as in a HUD row. */
function Stat({
  className,
  children,
  variant = 'tile',
  ...props
}: React.ComponentProps<'div'> & { variant?: 'tile' | 'plain' }) {
  return (
    <div
      data-slot="stat"
      data-variant={variant}
      className={cn(
        'relative overflow-hidden',
        'text-fg grid min-w-0 grid-cols-[minmax(0,1fr)_auto] gap-x-2 gap-y-1',
        variant === 'tile' &&
          'bg-card ring-border smooth-corners hocus:shadow-md rounded-lg p-3 shadow-xs ring-1 sm:gap-x-4 sm:p-4',
        '**:data-[slot=stat-label]:col-span-1 **:data-[slot=stat-value]:col-span-1',
        '**:data-[slot=stat-indicator]:col-start-2 **:data-[slot=stat-indicator]:row-span-2 **:data-[slot=stat-indicator]:row-start-1 **:data-[slot=stat-indicator]:self-start',
        '**:data-[slot=stat-description]:col-span-2 **:data-[slot=stat-separator]:col-span-2 **:data-[slot=stat-trend]:col-span-2',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
function StatLabel({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="stat-label"
      className={cn('text-muted text-sm font-medium', className)}
      {...props}
    />
  );
}
const statIndicatorVariants = cva(
  'flex shrink-0 items-center justify-center [&_svg]:pointer-events-none',
  {
    variants: {
      variant: {
        default: "text-muted [&_svg:not([class*='size-'])]:size-5",
        icon: "size-8 rounded-md border [&_svg:not([class*='size-'])]:size-3.5",
        badge:
          "h-6 min-w-6 rounded-sm border px-1.5 text-sm font-medium [&_svg:not([class*='size-'])]:size-3",
        action:
          "hover:bg-subtle size-8 cursor-pointer rounded-md transition-colors [&_svg:not([class*='size-'])]:size-4"
      },
      color: {
        default: 'bg-subtle text-muted',
        success: 'border-green-500/20 bg-green-500/10 text-green-600 dark:text-green-400',
        info: 'border-blue-500/20 bg-blue-500/10 text-blue-600 dark:text-blue-400',
        warning: 'border-orange-500/20 bg-orange-500/10 text-orange-600 dark:text-orange-400',
        error: 'border-red-600 bg-red-600/10 text-red-600 dark:border-red-400/20 dark:text-red-400'
      }
    },
    defaultVariants: {
      variant: 'default',
      color: 'default'
    }
  }
);
type StatIndicatorProps = Omit<React.ComponentProps<'div'>, 'color'> &
  VariantProps<typeof statIndicatorVariants>;
function StatIndicator({
  className,
  variant = 'default',
  color = 'default',
  ...props
}: StatIndicatorProps) {
  return (
    <div
      data-slot="stat-indicator"
      data-variant={variant}
      data-color={color}
      className={cn(statIndicatorVariants({ variant, color, className }))}
      {...props}
    />
  );
}
function StatValue({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="stat-value"
      className={cn(
        // A step smaller on a phone, where two sit side by side, and broken
        // rather than pushed past the tile when a figure runs long.
        'font-heading min-w-0 text-2xl font-bold tracking-tight break-words tabular-nums sm:text-3xl lg:text-4xl',
        className
      )}
      {...props}
    />
  );
}
function StatTrend({
  className,
  trend,
  ...props
}: React.ComponentProps<'div'> & {
  trend?: 'up' | 'down' | 'neutral';
}) {
  return (
    <div
      data-slot="stat-trend"
      data-trend={trend}
      className={cn(
        "inline-flex items-center gap-1 text-sm font-medium [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-3",
        {
          'text-green-600 dark:text-green-400': trend === 'up',
          'text-red-600 dark:text-red-400': trend === 'down',
          'text-muted': trend === 'neutral' || !trend
        },
        className
      )}
      {...props}
    />
  );
}
function StatSeparator({ ...props }: React.ComponentProps<typeof Separator>) {
  return (
    <Separator
      data-slot="stat-separator"
      className="my-2"
      {...props}
    />
  );
}
function StatDescription({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="stat-description"
      className={cn('text-muted text-sm', className)}
      {...props}
    />
  );
}

/** A link or two under the figure, divided from it by a rule: where to go to
 *  act on what the stat counts. */
function StatAction({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="stat-action"
      className={cn(
        'border-border mt-2 flex items-center gap-1 border-t pt-2 text-xs font-medium',
        '[&_a]:inline-flex [&_a]:items-center [&_a]:gap-1 [&_a]:underline-offset-4 hover:[&_a]:underline',
        className
      )}
      {...props}
    />
  );
}

export {
  Stat,
  StatAction,
  StatDescription,
  StatIndicator,
  StatLabel,
  StatSeparator,
  StatTrend,
  StatValue
};
