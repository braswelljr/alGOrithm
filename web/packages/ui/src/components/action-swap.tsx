'use client';

import type * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { AnimatePresence, motion, useReducedMotion, type Variants } from 'motion/react';

const EASE_OUT = [0.23, 1, 0.32, 1] as const;
const SWAP_BLUR = 'blur(6px)';
const SWAP_TRANSITION = { duration: 0.2, ease: EASE_OUT } as const;

const ICON_VARIANTS: Variants = {
  initial: { opacity: 0, scale: 0.6, filter: SWAP_BLUR },
  animate: { opacity: 1, scale: 1, filter: 'blur(0px)', transition: SWAP_TRANSITION },
  exit: { opacity: 0, scale: 0.6, filter: SWAP_BLUR, transition: SWAP_TRANSITION }
};

export interface ActionSwapIconProps {
  /** Identity of the icon on show. A new value plays the swap. */
  value: string;
  children: React.ReactNode;
  className?: string;
}

/**
 * Trades one icon for another in the same slot. Both layers share a single
 * grid cell, so the outgoing and incoming icons overlap rather than push.
 */
export function ActionSwapIcon({ value, children, className }: ActionSwapIconProps) {
  const reduce = useReducedMotion();

  return (
    <span
      className={cn('relative inline-grid shrink-0 place-items-center overflow-hidden', className)}
    >
      <AnimatePresence
        mode="popLayout"
        initial={false}
      >
        <motion.span
          key={value}
          aria-hidden
          variants={ICON_VARIANTS}
          initial={reduce ? false : 'initial'}
          animate={reduce ? { opacity: 1, scale: 1, filter: 'blur(0px)' } : 'animate'}
          exit={reduce ? undefined : 'exit'}
          className="col-start-1 row-start-1 inline-flex items-center justify-center will-change-[opacity,filter,transform]"
        >
          {children}
        </motion.span>
      </AnimatePresence>
    </span>
  );
}
