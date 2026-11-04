import type * as React from 'react';
import { Badge as UIBadge } from '@braswelljr/algorithm-ui/components/badge';
import { cn } from '@/utils/utils';

/** A small label: the library's Badge, soft and set in mono for costs and ids. */
export function Badge({ className, ...props }: React.ComponentProps<typeof UIBadge>) {
  return (
    <UIBadge
      variant="secondary"
      className={cn('font-mono', className)}
      {...props}
    />
  );
}
