import type * as React from 'react';
import { MountainIcon, SproutIcon, TrendingUpIcon } from 'lucide-react';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';

const levels = {
  beginner: { label: 'Beginner', icon: SproutIcon },
  intermediate: { label: 'Intermediate', icon: TrendingUpIcon },
  advanced: { label: 'Advanced', icon: MountainIcon }
} as const;

export type Level = keyof typeof levels;

/**
 * How far along a learner should be for an entry: a color, an icon and the
 * word, so the level never depends on color alone.
 *
 *   <LevelBadge level="intermediate" />
 */
export function LevelBadge({
  level,
  className,
  ...props
}: { level: string } & React.ComponentProps<'span'>) {
  const l = levels[level as Level];
  if (!l) return null;
  const Icon = l.icon;
  return (
    <Badge
      data-slot="level-badge"
      data-level={level}
      variant={level as Level}
      className={className}
      {...props}
    >
      <Icon aria-hidden />
      {l.label}
    </Badge>
  );
}
