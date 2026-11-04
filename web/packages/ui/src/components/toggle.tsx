'use client';

import { Toggle as TogglePrimitive } from '@base-ui/react/toggle';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, type VariantProps } from 'class-variance-authority';

const toggleVariants = cva(
  "relative inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 rounded-lg border border-border text-base font-medium whitespace-nowrap transition-shadow outline-none select-none before:pointer-events-none before:absolute before:inset-0 before:rounded-[calc(var(--radius-lg)-1px)] hover:bg-bg focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-1 focus-visible:ring-offset-bg disabled:pointer-events-none disabled:opacity-64 data-pressed:bg-accent data-pressed:text-accent-fg data-pressed:transition-none sm:text-sm pointer-coarse:after:absolute pointer-coarse:after:size-full pointer-coarse:after:min-h-11 pointer-coarse:after:min-w-11 [&_svg]:pointer-events-none [&_svg]:-mx-0.5 [&_svg]:shrink-0 [&_svg:not([class*='opacity-'])]:opacity-80 [&_svg:not([class*='size-'])]:size-4.5 sm:[&_svg:not([class*='size-'])]:size-4",
  {
    defaultVariants: {
      size: 'default',
      variant: 'default'
    },
    variants: {
      size: {
        default: 'h-9 min-w-9 px-[calc(--spacing(2)-1px)] sm:h-8 sm:min-w-8',
        lg: 'h-10 min-w-10 px-[calc(--spacing(2.5)-1px)] sm:h-9 sm:min-w-9',
        sm: 'h-8 min-w-8 px-[calc(--spacing(1.5)-1px)] sm:h-7 sm:min-w-7'
      },
      variant: {
        default: 'border-transparent',
        outline:
          'border-border bg-clip-padding shadow-xs not-disabled:not-active:not-data-pressed:before:shadow-[0_1px_--theme(--color-black/4%)] dark:not-disabled:not-data-pressed:before:shadow-[0_-1px_--theme(--color-white/4%)] dark:not-disabled:not-active:not-data-pressed:before:shadow-[0_-1px_--theme(--color-white/8%)] [:disabled,:active,[data-pressed]]:shadow-none',
        // Semantic tone variants: neutral until pressed, then filled with the
        // tone. `primary` is the base's own pressed state - gold with dark text
        // on it, which is what every active control in the app uses - and is
        // kept as a name so a caller can say which it means. The `!` matters
        // where a call site passes its own background.
        primary: 'border-transparent data-pressed:bg-accent! data-pressed:text-accent-fg!',
        /** A filter or option chip: a pill, outlined, filled with the accent when on. */
        chip: 'rounded-full border-border px-3 text-xs before:rounded-full hover:border-control hover:bg-subtle data-pressed:border-accent',
        info: 'border-sky-200 data-pressed:bg-sky-500! data-pressed:text-white! dark:border-sky-800/60',
        success:
          'border-emerald-200 data-pressed:bg-emerald-500! data-pressed:text-white! dark:border-emerald-800/60',
        warning:
          'border-amber-200 data-pressed:bg-amber-500! data-pressed:text-zinc-950! dark:border-amber-800/60',
        danger:
          'border-red-200 data-pressed:bg-red-500! data-pressed:text-white! dark:border-red-800/60'
      }
    }
  }
);

function Toggle({
  className,
  variant,
  size,
  ...props
}: TogglePrimitive.Props & VariantProps<typeof toggleVariants>) {
  return (
    <TogglePrimitive
      className={cn(toggleVariants({ className, size, variant }))}
      data-slot="toggle"
      {...props}
    />
  );
}

export { Toggle, toggleVariants };
