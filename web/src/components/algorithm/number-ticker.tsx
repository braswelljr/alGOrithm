'use client';

import * as React from 'react';
import { animate, motion, useMotionValue, useTransform } from 'motion/react';
import { useReducedMotion } from '@/hooks/use-reduced-motion';
import { cn } from '@/lib/utils';

/**
 * A count that rolls to its new value: comparisons, swaps, a score. The
 * number is tabular so the width does not jump, the final value is what a
 * screen reader hears (the rolling digits are hidden from it), and under
 * reduced motion it changes at once.
 *
 *   <NumberTicker value={counts.compare} label="comparisons" />
 */
export function NumberTicker({
  value,
  label,
  duration = 0.45,
  format = (n) => Math.round(n).toLocaleString('en-US'),
  className
}: {
  value: number;
  /** What the number counts, read after it: "12 comparisons". */
  label?: string;
  duration?: number;
  format?: (n: number) => string;
  className?: string;
}) {
  const reduce = useReducedMotion();
  const mv = useMotionValue(value);
  const text = useTransform(mv, format);
  React.useEffect(() => {
    if (reduce) {
      mv.set(value);
      return;
    }
    const controls = animate(mv, value, { duration, ease: [0.23, 1, 0.32, 1] });
    return () => controls.stop();
  }, [value, reduce, duration, mv]);
  return (
    <span className={cn('tabular-nums', className)}>
      <motion.span aria-hidden>{text}</motion.span>
      <span className="sr-only">
        {format(value)}
        {label ? ` ${label}` : ''}
      </span>
    </span>
  );
}
