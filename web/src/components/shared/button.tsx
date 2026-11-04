import type * as React from 'react';
import { Button as UIButton } from '@/components/ui/button';

export type ButtonProps = React.ComponentProps<typeof UIButton>;

/**
 * The site's Button: the library's, with "ghost" as the site's default look.
 * One "primary" per view for its main action; ghost for everything else;
 * "outline" only where a bordered control must stand off a busy background
 * (docs/rules/use-the-component-library.md). An icon-only one takes size
 * "icon" or "icon-sm" and an aria-label.
 */
export function Button({ variant = 'ghost', size = 'default', ...props }: ButtonProps) {
  return (
    <UIButton
      variant={variant}
      size={size}
      {...props}
    />
  );
}
