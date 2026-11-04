'use client';

import type React from 'react';
import { Popover as PopoverPrimitive } from '@base-ui/react/popover';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

export const PopoverCreateHandle: typeof PopoverPrimitive.createHandle =
  PopoverPrimitive.createHandle;

export const Popover: typeof PopoverPrimitive.Root = PopoverPrimitive.Root;

export function PopoverTrigger({
  className,
  children,
  ...props
}: PopoverPrimitive.Trigger.Props): React.ReactElement {
  return (
    <PopoverPrimitive.Trigger
      className={className}
      data-slot="popover-trigger"
      {...props}
    >
      {children}
    </PopoverPrimitive.Trigger>
  );
}

export function PopoverPopup({
  children,
  className,
  side = 'bottom',
  align = 'center',
  sideOffset = 4,
  alignOffset = 0,
  tooltipStyle = false,
  arrow = false,
  anchor,
  portalProps,
  ...props
}: PopoverPrimitive.Popup.Props & {
  portalProps?: PopoverPrimitive.Portal.Props;
  side?: PopoverPrimitive.Positioner.Props['side'];
  align?: PopoverPrimitive.Positioner.Props['align'];
  sideOffset?: PopoverPrimitive.Positioner.Props['sideOffset'];
  alignOffset?: PopoverPrimitive.Positioner.Props['alignOffset'];
  tooltipStyle?: boolean;
  /**
   * Points the popup at what opened it.
   *
   * Off by default, and that is the decision rather than an oversight. An
   * arrow earns its place where the anchor is small or ambiguous - the "+N
   * more" in a month cell, an icon button in a toolbar - because it answers
   * "which of these did I open". Under a full-width input, a date picker say,
   * the relationship is already plain and the arrow is a notch on an edge.
   */
  arrow?: boolean;
  anchor?: PopoverPrimitive.Positioner.Props['anchor'];
}): React.ReactElement {
  return (
    <PopoverPrimitive.Portal {...portalProps}>
      <PopoverPrimitive.Positioner
        align={align}
        alignOffset={alignOffset}
        anchor={anchor}
        className="z-50 h-(--positioner-height) w-(--positioner-width) max-w-(--available-width) transition-[top,left,right,bottom,transform] data-instant:transition-none"
        data-slot="popover-positioner"
        side={side}
        sideOffset={sideOffset}
      >
        <PopoverPrimitive.Popup
          className={cn(
            'relative flex h-(--popup-height,auto) w-(--popup-width,auto) origin-(--transform-origin) rounded-lg border border-border bg-card text-fg shadow-lg/5 transition-[width,height,scale,opacity] outline-none not-dark:bg-clip-padding',
            'before:pointer-events-none before:absolute before:inset-0 before:rounded-[calc(var(--radius-lg)-1px)] before:shadow-[0_1px_--theme(--color-zinc-950/4%)] dark:before:shadow-[0_-1px_--theme(--color-white/6%)]',
            'has-data-[slot=calendar]:rounded-xl has-data-[slot=calendar]:before:rounded-[calc(var(--radius-xl)-1px)] data-ending-style:scale-98 data-ending-style:opacity-0 data-starting-style:scale-98 data-starting-style:opacity-0',
            tooltipStyle &&
              'w-fit rounded-md text-sm text-balance shadow-md/5 before:rounded-[calc(var(--radius-md)-1px)]',
            className
          )}
          data-slot="popover-popup"
          {...props}
        >
          {arrow && <PopoverArrow />}
          <PopoverPrimitive.Viewport
            className={cn(
              'relative size-full max-h-(--available-height) overflow-clip px-(--viewport-inline-padding) py-2 [--viewport-inline-padding:--spacing(2)]',
              'has-data-[slot=calendar]:p-2',
              '**:data-current:w-[calc(var(--popup-width)-2*var(--viewport-inline-padding)-2px)] **:data-current:opacity-100 **:data-current:transition-opacity **:data-current:data-ending-style:opacity-0 data-instant:transition-none',
              '**:data-previous:w-[calc(var(--popup-width)-2*var(--viewport-inline-padding)-2px)] **:data-previous:opacity-100 **:data-previous:transition-opacity **:data-previous:data-ending-style:opacity-0 **:data-current:data-starting-style:opacity-0 **:data-previous:data-starting-style:opacity-0',
              tooltipStyle
                ? 'py-1 [--viewport-inline-padding:--spacing(2)]'
                : 'not-data-transitioning:overflow-y-auto'
            )}
            data-slot="popover-viewport"
          >
            {children}
          </PopoverPrimitive.Viewport>
        </PopoverPrimitive.Popup>
      </PopoverPrimitive.Positioner>
    </PopoverPrimitive.Portal>
  );
}

export function PopoverClose({ ...props }: PopoverPrimitive.Close.Props): React.ReactElement {
  return (
    <PopoverPrimitive.Close
      data-slot="popover-close"
      {...props}
    />
  );
}

export function PopoverTitle({
  className,
  ...props
}: PopoverPrimitive.Title.Props): React.ReactElement {
  return (
    <PopoverPrimitive.Title
      className={cn('text-lg leading-none font-semibold', className)}
      data-slot="popover-title"
      {...props}
    />
  );
}

export function PopoverDescription({
  className,
  ...props
}: PopoverPrimitive.Description.Props): React.ReactElement {
  return (
    <PopoverPrimitive.Description
      className={cn('text-sm text-fg', className)}
      data-slot="popover-description"
      {...props}
    />
  );
}

/**
 * The arrow, as a rotated square that continues the popup's own outline.
 *
 * # Why this is not the tooltip's arrow
 *
 * A tooltip is a solid dark chip, so its arrow is a solid block and needs no
 * border: any two edges of it are the same colour as the middle. A popover is
 * a bordered light surface, so an arrow drawn the same way is a diamond
 * stuck to the edge with the popup's border line running straight past behind
 * it. The two outward-facing edges have to carry the border and the two inward
 * ones must not, which is why the border classes below are keyed on the side
 * the popup opened to.
 *
 * That is almost certainly why this component was exported and mounted
 * nowhere: as a bare primitive it drew an unstyled square, so nothing used it
 * and popovers simply went without.
 */
export function PopoverArrow({
  className,
  ...props
}: PopoverPrimitive.Arrow.Props): React.ReactElement {
  return (
    <PopoverPrimitive.Arrow
      data-slot="popover-arrow"
      className={cn(
        'z-50 size-2.5 rotate-45 rounded-xs border-border bg-card',
        // Which two edges face out depends on which way the popup opened, and
        // only those two carry the border. Rotated 45 degrees, the top and
        // left edges form the point that aims upward.
        'data-[side=bottom]:-top-1 data-[side=bottom]:border-t data-[side=bottom]:border-l',
        'data-[side=top]:-bottom-1 data-[side=top]:border-r data-[side=top]:border-b',
        'data-[side=left]:-right-1 data-[side=left]:border-t data-[side=left]:border-r',
        'data-[side=right]:-left-1 data-[side=right]:border-b data-[side=right]:border-l',
        'data-[side=inline-end]:-left-1 data-[side=inline-end]:border-b data-[side=inline-end]:border-l',
        'data-[side=inline-start]:-right-1 data-[side=inline-start]:border-t data-[side=inline-start]:border-r',
        className
      )}
      {...props}
    />
  );
}

export { PopoverPrimitive, PopoverPopup as PopoverContent };
