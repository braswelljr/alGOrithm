import type * as React from 'react';
import { Badge as UIBadge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

/** A small label: the library's Badge at its sm size, solid and set in mono
 *  for ids and tags, which read as written (never upper-cased). */
export function Badge({ className, ...props }: React.ComponentProps<typeof UIBadge>) {
  return (
    <UIBadge
      variant="secondary"
      size="sm"
      case="normal"
      className={cn('font-mono', className)}
      {...props}
    />
  );
}
