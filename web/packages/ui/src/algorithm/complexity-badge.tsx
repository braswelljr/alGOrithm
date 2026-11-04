'use client';

import { ClockIcon } from 'lucide-react';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';
import { Tooltip, TooltipContent, TooltipTrigger } from '@braswelljr/algorithm-ui/components/tooltip';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * An entry's cost in Big O, with the reason in a tooltip. The reason is also
 * the button's accessible description, since a touch screen cannot hover.
 *
 *   <ComplexityBadge time="O(n log n)" space="O(n)" why="log n levels of n merges." />
 */
export function ComplexityBadge({
  time,
  space,
  why,
  className
}: {
  time: string;
  space?: string;
  why?: string;
  className?: string;
}) {
  const badge = (
    <Badge
      variant="cost"
      className={cn('gap-1.5', className)}
    >
      <ClockIcon
        aria-hidden
        className="size-3.5 text-muted"
      />
      <span>
        <span className="sr-only">Time </span>
        {time}
      </span>
      {space ? (
        <span className="text-muted">
          <span aria-hidden>·</span>
          <span className="sr-only">, space</span> {space}
        </span>
      ) : null}
    </Badge>
  );
  if (!why) return badge;
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <button
            type="button"
            aria-label={`Time ${time}${space ? `, space ${space}` : ''}. ${why}`}
            className="rounded-md"
          />
        }
      >
        {badge}
      </TooltipTrigger>
      <TooltipContent className="max-w-72 text-pretty">{why}</TooltipContent>
    </Tooltip>
  );
}
