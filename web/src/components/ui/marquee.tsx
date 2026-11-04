import { type ComponentPropsWithoutRef, type ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface MarqueeProps extends ComponentPropsWithoutRef<'div'> {
  /**
   * Optional CSS class name to apply custom styles
   */
  className?: string;
  /**
   * Whether to reverse the animation direction
   * @default false
   */
  reverse?: boolean;
  /**
   * Whether to pause the animation on hover (and while focus is inside it)
   * @default false
   */
  pauseOnHover?: boolean;
  /**
   * Content to be displayed in the marquee
   */
  children: ReactNode;
  /**
   * Whether to animate vertically instead of horizontally
   * @default false
   */
  vertical?: boolean;
  /**
   * Number of times to repeat the content
   * @default 4
   */
  repeat?: number;
}

/**
 * Items sliding past in a loop: the children repeated `repeat` times, each copy
 * moving by its own width plus the gap. Set `--duration` and `--gap` in `style`
 * to tune it. Only the first copy is exposed: the repeats are hidden from
 * assistive technology and unreachable by keyboard, so every item is read once.
 * Under reduced motion it stands still, shows one copy and scrolls by hand.
 * Stop it from outside with `data-paused="true"` on an ancestor.
 */
export function Marquee({
  className,
  reverse = false,
  pauseOnHover = false,
  children,
  vertical = false,
  repeat = 4,
  ...props
}: MarqueeProps) {
  return (
    <div
      {...props}
      className={cn(
        'group flex gap-(--gap) overflow-hidden p-2 [--duration:40s] [--gap:1rem] motion-reduce:overflow-auto',
        {
          'flex-row': !vertical,
          'flex-col': vertical
        },
        className
      )}
    >
      {Array(Math.max(2, repeat))
        .fill(0)
        .map((_, i) => (
          <div
            key={i}
            aria-hidden={i > 0 || undefined}
            inert={i > 0}
            className={cn('flex shrink-0 justify-around gap-(--gap)', {
              'animate-marquee flex-row': !vertical,
              'animate-marquee-vertical flex-col': vertical,
              'group-focus-within:[animation-play-state:paused] group-hover:[animation-play-state:paused]':
                pauseOnHover,
              '[animation-direction:reverse]': reverse,
              // Standing still, the repeats would only show the items again.
              'motion-reduce:hidden': i > 0
            })}
          >
            {children}
          </div>
        ))}
    </div>
  );
}
