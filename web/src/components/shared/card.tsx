import type * as React from 'react';
import { Card as UICard } from '@/components/ui/card';
import { cn } from '@/lib/utils';

/**
 * A padded panel: the library's Card with its body padding applied to the
 * whole card, for pages that put content straight inside. Compose the
 * library's CardHeader, CardTitle and CardContent for anything with parts.
 */
export function Card({ className, ...props }: React.ComponentProps<typeof UICard>) {
  return (
    <UICard
      size="md"
      className={cn('px-(--card-spacing)', className)}
      {...props}
    />
  );
}
