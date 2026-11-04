'use client';

import { type ReactNode } from 'react';
import { motion } from 'motion/react';
import { useReducedMotion } from '@/hooks/use-reduced-motion';

type RevealProps = {
  children: ReactNode;
  /** Seconds to wait, for staggering siblings. Keep the whole sequence short. */
  delay?: number;
  className?: string;
};

/**
 * Fades its content up into place the first time it scrolls into view.
 *
 * Only opacity and transform move, on a short ease-out, once. A reader who
 * asked for reduced motion gets the content in place with no movement at all.
 * Used on the landing page only: documentation and the terminal do not animate.
 *
 * The setting is read through `useReducedMotion` from our hooks, which renders
 * the server's answer during hydration: the server cannot know the setting,
 * and a first client render that disagreed with it would leave the server's
 * hidden starting style on the element.
 */
export function Reveal({ children, delay = 0, className }: RevealProps) {
  const reduced = useReducedMotion();

  if (reduced) return <div className={className}>{children}</div>;

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 12 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: '0px 0px -10% 0px' }}
      transition={{ duration: 0.4, delay, ease: [0.23, 1, 0.32, 1] }}
    >
      {children}
    </motion.div>
  );
}
