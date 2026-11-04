'use client';

import * as React from 'react';
import { useReducedMotion } from '@/hooks/use-reduced-motion';
import { cn } from '@/lib/utils';

const glyphs = '01<>/{}[]=+*#%&ABCDEFGHIJKLMNOPQRSTUVWXYZ';

/**
 * A heading that resolves out of random characters, left to right, once, when
 * it first appears (like a search settling on its answer). The real text is
 * always the accessible name, and under reduced motion it is shown at once.
 * Inspired by the text-scramble items in beUI, Sora UI and Atelier UI;
 * written for this site.
 *
 *   <TextScramble as="h1" text="Sorting from zero" />
 */
export function TextScramble({
  text,
  as: Tag = 'span',
  duration = 700,
  className
}: {
  text: string;
  as?: 'span' | 'h1' | 'h2' | 'h3' | 'p';
  duration?: number;
  className?: string;
}) {
  const reduce = useReducedMotion();
  const [shown, setShown] = React.useState(text);
  React.useEffect(() => {
    if (reduce) {
      setShown(text);
      return;
    }
    let frame = 0;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration);
      const fixed = Math.floor(t * text.length);
      let out = '';
      for (let i = 0; i < text.length; i++) {
        const ch = text[i];
        out += i < fixed || ch === ' ' ? ch : glyphs[Math.floor(Math.random() * glyphs.length)];
      }
      setShown(out);
      if (t < 1) frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [text, reduce, duration]);
  return (
    <Tag
      className={cn(className)}
      aria-label={text}
    >
      <span aria-hidden>{shown}</span>
    </Tag>
  );
}
