'use client';

import type React from 'react';
import { Checkbox as CheckboxPrimitive } from '@base-ui/react/checkbox';
import { cn } from '@/lib/utils';

export function Checkbox({
  className,
  ...props
}: CheckboxPrimitive.Root.Props): React.ReactElement {
  return (
    <CheckboxPrimitive.Root
      className={cn(
        // The edge is stone-500, 3:1 against the page, since it is all that shows an
        // unticked box is there; `after` widens what a finger can hit to 24.
        'group border-control bg-card ring-fg/10 focus-visible:ring-fg relative inline-flex size-4.5 shrink-0 items-center justify-center rounded-sm border shadow-sm transition-all outline-none after:absolute after:-inset-1 focus-visible:ring-2 focus-visible:ring-offset-1 data-disabled:cursor-not-allowed data-disabled:opacity-50 sm:size-4',
        '',
        'aria-invalid:border-red-600 focus-visible:aria-invalid:border-red-700 focus-visible:aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400',
        className
      )}
      data-slot="checkbox"
      {...props}
    >
      <CheckboxPrimitive.Indicator
        className={cn(
          'absolute -inset-px flex items-center justify-center rounded-sm transition-colors data-unchecked:hidden',
          'data-checked:bg-fg data-checked:text-bg',
          'group-aria-invalid:data-checked:bg-red-600 group-aria-invalid:data-checked:text-white'
        )}
        data-slot="checkbox-indicator"
        render={(props: React.ComponentProps<'span'>, state: CheckboxPrimitive.Indicator.State) => (
          <span {...props}>
            {state.indeterminate ? (
              <svg
                aria-hidden="true"
                className="size-3.5 sm:size-3"
                fill="none"
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="3.5"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M5.252 12h13.496" />
              </svg>
            ) : (
              <svg
                aria-hidden="true"
                className="size-3.5 sm:size-3"
                fill="none"
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="3.5"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path d="M5.252 12.7 10.2 18.63 18.748 5.37" />
              </svg>
            )}
          </span>
        )}
      />
    </CheckboxPrimitive.Root>
  );
}

export { CheckboxPrimitive };
