import { cn } from '@/lib/utils';

/**
 * A block caret. It blinks while idle and holds still while typing, as a real
 * one does; under reduced motion it never blinks.
 */
export function Caret({ blinking = true, className }: { blinking?: boolean; className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        'bg-lit-face ml-px inline-block h-[1.15em] w-[0.55em] translate-y-[0.2em] rounded-[1px]',
        blinking && 'animate-caret',
        className
      )}
    />
  );
}
