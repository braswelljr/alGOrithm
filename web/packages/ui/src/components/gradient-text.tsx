import { type ComponentProps } from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

/**
 * Text in the color of a lit key. Use it on a few words of a heading, never a
 * paragraph: it is emphasis, and emphasis spent everywhere is none.
 */
export function GradientText({ className, ...props }: ComponentProps<'span'>) {
  return (
    <span
      className={cn('text-lit', className)}
      {...props}
    />
  );
}
