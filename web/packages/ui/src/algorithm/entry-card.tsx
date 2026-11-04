'use client';

import * as React from 'react';
import { motion, useMotionValue, useSpring, useTransform } from 'motion/react';
import { PlayIcon } from 'lucide-react';
import { Card } from '@braswelljr/algorithm-ui/components/card';
import { ComplexityBadge } from '@braswelljr/algorithm-ui/algorithm/complexity-badge';
import { LevelBadge } from '@braswelljr/algorithm-ui/algorithm/level-badge';
import { TagChip } from '@braswelljr/algorithm-ui/algorithm/tag-chip';
import { useReducedMotion } from '@braswelljr/algorithm-ui/hooks/use-reduced-motion';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * One catalog entry as a card: its name (the link), summary, level, cost and
 * tags. The card tilts a few degrees toward the pointer (a tilt card, as in
 * beUI), never under reduced motion or on touch. The whole card is clickable
 * through its title link, so there is one tab stop, not one per card part.
 *
 *   <EntryCard name="Quick Sort" href="/docs/entry/?id=..." summary="..." level="intermediate"
 *     time="O(n log n)" why="..." tags={['divide-and-conquer']} animated />
 */
export function EntryCard({
  name,
  href,
  summary,
  level,
  time,
  space,
  why,
  tags = [],
  animated,
  className
}: {
  name: string;
  href: string;
  summary: string;
  level?: string;
  time?: string;
  space?: string;
  why?: string;
  tags?: string[];
  /** Shown with a play mark: the visualizer animates it. */
  animated?: boolean;
  className?: string;
}) {
  const reduce = useReducedMotion();
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const rx = useSpring(useTransform(y, [-0.5, 0.5], [4, -4]), { stiffness: 300, damping: 30 });
  const ry = useSpring(useTransform(x, [-0.5, 0.5], [-4, 4]), { stiffness: 300, damping: 30 });
  const onMove = (e: React.PointerEvent<HTMLElement>) => {
    if (reduce || e.pointerType !== 'mouse') return;
    const r = e.currentTarget.getBoundingClientRect();
    x.set((e.clientX - r.left) / r.width - 0.5);
    y.set((e.clientY - r.top) / r.height - 0.5);
  };
  const reset = () => {
    x.set(0);
    y.set(0);
  };
  return (
    <Card
      variant="interactive"
      size="md"
      className={cn('gap-3 px-(--card-spacing) hover:translate-y-0', className)}
      render={
        <motion.article
          onPointerMove={onMove}
          onPointerLeave={reset}
          style={reduce ? undefined : { rotateX: rx, rotateY: ry, transformPerspective: 800 }}
        />
      }
    >
      <header className="flex items-start justify-between gap-2">
        <h3 className="text-base leading-snug font-semibold text-fg">
          <a
            href={href}
            className="outline-none after:absolute after:inset-0 after:rounded-xl"
          >
            {name}
          </a>
        </h3>
        {animated ? (
          <span
            className="inline-flex shrink-0 items-center gap-1 text-xs text-muted"
            title="Animated in the visualizer"
          >
            <PlayIcon
              aria-hidden
              className="size-3.5"
            />
            <span className="sr-only">Animated</span>
          </span>
        ) : null}
      </header>
      <p className="line-clamp-3 text-sm text-pretty text-muted">{summary}</p>
      <footer className="relative z-10 mt-auto flex flex-wrap items-center gap-1.5">
        {level ? <LevelBadge level={level} /> : null}
        {time ? (
          <ComplexityBadge
            time={time}
            space={space}
            why={why}
          />
        ) : null}
        {tags.slice(0, 3).map((t) => (
          <TagChip
            key={t}
            tag={t}
          />
        ))}
      </footer>
    </Card>
  );
}
