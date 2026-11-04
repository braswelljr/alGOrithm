'use client';

import { Tooltip as TooltipPrimitive } from '@base-ui/react/tooltip';
import { cn } from '@/lib/utils';

/**
 * TooltipProvider: Wraps the application or a section to manage tooltip delays.
 *
 * A tooltip waits a moment (300ms) so it does not flash up under a pointer
 * passing over a toolbar; once one is open, its neighbours open at once.
 * A tooltip names or explains a control; it never holds the only copy of
 * something the reader needs, since a touch screen cannot hover.
 */
function TooltipProvider({ delay = 300, ...props }: TooltipPrimitive.Provider.Props) {
  return (
    <TooltipPrimitive.Provider
      data-slot="tooltip-provider"
      delay={delay}
      {...props}
    />
  );
}

/**
 * Tooltip: The root state manager for an individual tooltip.
 */
function Tooltip({ ...props }: TooltipPrimitive.Root.Props) {
  return (
    <TooltipPrimitive.Root
      data-slot="tooltip"
      {...props}
    />
  );
}

/**
 * TooltipTrigger: The element that triggers the tooltip.
 * Uses the 'render' pattern to inject event listeners into the sidebar buttons.
 */
function TooltipTrigger({ render, ...props }: TooltipPrimitive.Trigger.Props) {
  return (
    <TooltipPrimitive.Trigger
      data-slot="tooltip-trigger"
      render={render}
      {...props}
    />
  );
}

/**
 * TooltipContent: The actual popup balloon.
 * Updated to use the Zinc 950/50 high-contrast palette.
 */
function TooltipContent({
  className,
  side = 'top',
  sideOffset = 4,
  align = 'center',
  alignOffset = 0,
  children,
  ...props
}: TooltipPrimitive.Popup.Props &
  Pick<TooltipPrimitive.Positioner.Props, 'align' | 'alignOffset' | 'side' | 'sideOffset'>) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Positioner
        align={align}
        alignOffset={alignOffset}
        side={side}
        sideOffset={sideOffset}
        className="isolate z-50"
      >
        <TooltipPrimitive.Popup
          data-slot="tooltip-content"
          className={cn(
            // Layout & Typography
            'z-50 inline-flex w-fit max-w-xs origin-(--transform-origin) items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium shadow-md outline-none',
            // Zinc Colors: High contrast Zinc 950 in light mode, Zinc 50 in dark mode
            'bg-fg text-bg',
            // Handling for KBD components inside tooltips
            'has-data-[slot=kbd]:pr-1.5',
            // Animations
            'data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95',
            'data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95',
            'data-[side=bottom]:slide-in-from-top-2 data-[side=inline-end]:slide-in-from-left-2 data-[side=inline-start]:slide-in-from-right-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
            // Internal state animations
            'data-[state=delayed-open]:animate-in data-[state=delayed-open]:fade-in-0 data-[state=delayed-open]:zoom-in-95',
            className
          )}
          {...props}
        >
          {children}
          <TooltipPrimitive.Arrow
            className={cn(
              'z-50 size-2.5 translate-y-[calc(-50%-2px)] rotate-45 rounded-xs',
              'bg-fg',
              'data-[side=bottom]:top-1 data-[side=inline-end]:top-1/2! data-[side=inline-end]:-left-1 data-[side=inline-end]:-translate-y-1/2 data-[side=inline-start]:top-1/2! data-[side=inline-start]:-right-1 data-[side=inline-start]:-translate-y-1/2 data-[side=left]:top-1/2! data-[side=left]:-right-1 data-[side=left]:-translate-y-1/2 data-[side=right]:top-1/2! data-[side=right]:-left-1 data-[side=right]:-translate-y-1/2 data-[side=top]:-bottom-2.5'
            )}
          />
        </TooltipPrimitive.Popup>
      </TooltipPrimitive.Positioner>
    </TooltipPrimitive.Portal>
  );
}

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
