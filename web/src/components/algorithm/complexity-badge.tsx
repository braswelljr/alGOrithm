'use client';

import { ClockIcon } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

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
      // A cost reads as written: O(n log n), never O(N LOG N).
      case="normal"
      className={className}
    >
      <ClockIcon
        aria-hidden
        className="text-muted"
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
          <Button
            variant="ghost"
            size="xs"
            aria-label={`Time ${time}${space ? `, space ${space}` : ''}. ${why}`}
            className="h-auto rounded-md p-0 hover:bg-transparent"
          />
        }
      >
        {badge}
      </TooltipTrigger>
      <TooltipContent className="max-w-72 text-pretty">{why}</TooltipContent>
    </Tooltip>
  );
}
