import type * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { ChevronDownIcon } from 'lucide-react';

/**
 * The platform's own `<select>`, styled to sit beside `Input` at the same
 * height. Native on purpose for plain forms: it registers with react-hook-form
 * like any input, works with the keyboard and screen readers as the platform
 * intends, and opens the device's own picker on a phone.
 */
function NativeSelect({ className, children, ...props }: React.ComponentProps<'select'>) {
  return (
    <div
      className="relative w-full"
      data-slot="native-select-wrapper"
    >
      <select
        data-slot="native-select"
        className={cn(
          'h-10 w-full min-w-0 appearance-none rounded-md border border-border bg-card py-2 pr-9 pl-3 text-sm text-fg shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-control focus-visible:ring-3 focus-visible:ring-fg/10 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
          className
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDownIcon
        aria-hidden
        className="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted"
      />
    </div>
  );
}

export { NativeSelect };
