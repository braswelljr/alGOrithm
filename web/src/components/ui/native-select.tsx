import type * as React from 'react';
import { ChevronDownIcon } from 'lucide-react';
import { controlSize, type ControlSize } from '@/lib/control';
import { cn } from '@/lib/utils';

/**
 * The platform's own `<select>`, styled to sit beside `Input` at the same
 * height. Native on purpose for plain forms: it registers with react-hook-form
 * like any input, works with the keyboard and screen readers as the platform
 * intends, and opens the device's own picker on a phone.
 */
function NativeSelect({
  className,
  children,
  size = 'md',
  ...props
}: Omit<React.ComponentProps<'select'>, 'size'> & { size?: ControlSize }) {
  return (
    <div
      className="relative w-full"
      data-slot="native-select-wrapper"
    >
      <select
        data-slot="native-select"
        data-size={size}
        className={cn(
          controlSize[size],
          'border-border bg-card text-fg focus-visible:border-control focus-visible:ring-fg/10 w-full min-w-0 appearance-none rounded-md border py-0 pe-9 shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-3 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-red-600 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
          className
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDownIcon
        aria-hidden
        className="text-muted pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2"
      />
    </div>
  );
}

export { NativeSelect };
