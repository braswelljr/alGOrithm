'use client';

import { Radio as RadioPrimitive } from '@base-ui/react/radio';
import { RadioGroup as RadioGroupPrimitive } from '@base-ui/react/radio-group';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

function RadioGroup({ className, ...props }: RadioGroupPrimitive.Props) {
  return (
    <RadioGroupPrimitive
      data-slot="radio-group"
      className={cn('grid w-full gap-2', className)}
      {...props}
    />
  );
}

function RadioGroupItem({ className, ...props }: RadioPrimitive.Root.Props) {
  return (
    <RadioPrimitive.Root
      data-slot="radio-group-item"
      className={cn(
        'group/radio-group-item peer relative flex aspect-square size-4 shrink-0 rounded-full border transition-all outline-none',
        'after:absolute after:-inset-x-3 after:-inset-y-2',
        'border-control bg-card',
        'focus-visible:border-control focus-visible:ring-2 focus-visible:ring-fg/20',
        'data-checked:border-control data-checked:bg-fg',
        'disabled:cursor-not-allowed disabled:opacity-50 data-disabled:cursor-not-allowed data-disabled:opacity-50',
        'aria-invalid:border-red-600 aria-invalid:ring-2 aria-invalid:ring-red-600/20 dark:aria-invalid:border-red-400 dark:aria-invalid:ring-red-400/25',
        className
      )}
      {...props}
    >
      <RadioPrimitive.Indicator
        data-slot="radio-group-indicator"
        className="flex size-full items-center justify-center"
      >
        {/* The Dot: High contrast relative to the background */}
        <span className="size-1.5 rounded-full bg-card" />
      </RadioPrimitive.Indicator>
    </RadioPrimitive.Root>
  );
}

export { RadioGroup, RadioGroupItem };
