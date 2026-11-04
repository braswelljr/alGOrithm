'use client';

import { Separator as SeparatorPrimitive } from '@base-ui/react/separator';
import { cn } from '@/lib/utils';

function Separator({ className, orientation = 'horizontal', ...props }: SeparatorPrimitive.Props) {
  return (
    <SeparatorPrimitive
      data-slot="separator"
      orientation={orientation}
      className={({ orientation: o }) =>
        cn(
          'shrink-0',
          'bg-subtle',
          o === 'horizontal' ? 'h-px w-full' : 'h-full w-px self-stretch',
          className
        )
      }
      {...props}
    />
  );
}

export { Separator };
