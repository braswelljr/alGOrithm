'use client';

import type React from 'react';
import { ScrollArea as ScrollAreaPrimitive } from '@base-ui/react/scroll-area';
import { cn } from '@/lib/utils';

/**
 * Which axes a scroller may scroll on.
 *
 * `both` is the default because it is what every existing caller already got,
 * and it is right for a document pane or a wide table. It is wrong for a nav
 * rail or a board column, which have a fixed width and should never travel
 * sideways: a rail that scrolls horizontally hides the start of its own labels,
 * and the reader has no way to tell that it did.
 *
 * # Why this cannot be a class
 *
 * base-ui sets `style={{ overflow: "scroll" }}` inline on the viewport, and an
 * inline style beats any utility, so `overflow-x-hidden` in a `className` is
 * silently inert - which is why three rails and the workflow board all scrolled
 * sideways while looking as though they had been told not to. The override has
 * to be a style too, and it has to be a longhand: the shorthand is already
 * there, and `overflowX` after it in the same object is what wins.
 */
export type ScrollAxis = 'both' | 'vertical' | 'horizontal';

export function ScrollArea({
  className,
  children,
  scrollFade = false,
  scrollbarGutter = false,
  orientation = 'both',
  ...props
}: ScrollAreaPrimitive.Root.Props & {
  scrollFade?: boolean;
  scrollbarGutter?: boolean;
  /** Which axes may scroll. Defaults to `both`; see `ScrollAxis`. */
  orientation?: ScrollAxis;
}): React.ReactElement {
  return (
    <ScrollAreaPrimitive.Root
      className={cn('size-full min-h-0', className)}
      {...props}
    >
      <ScrollAreaPrimitive.Viewport
        // Merged over base-ui's own `overflow: scroll`, which is inline and so
        // cannot be reached from `className`. Undefined for `both`, so the
        // default path sets nothing at all.
        style={
          orientation === 'vertical'
            ? { overflowX: 'hidden' }
            : orientation === 'horizontal'
              ? { overflowY: 'hidden' }
              : undefined
        }
        className={cn(
          'focus-visible:ring-fg focus-visible:ring-offset-bg h-full rounded-[inherit] transition-shadow outline-none focus-visible:ring-2 focus-visible:ring-offset-1 data-has-overflow-x:overscroll-x-contain data-has-overflow-y:overscroll-y-contain',
          scrollFade &&
            'mask-t-from-[calc(100%-min(var(--fade-size),var(--scroll-area-overflow-y-start)))] mask-r-from-[calc(100%-min(var(--fade-size),var(--scroll-area-overflow-x-end)))] mask-b-from-[calc(100%-min(var(--fade-size),var(--scroll-area-overflow-y-end)))] mask-l-from-[calc(100%-min(var(--fade-size),var(--scroll-area-overflow-x-start)))] [--fade-size:1.5rem]',
          scrollbarGutter && 'data-has-overflow-x:pb-2.5 data-has-overflow-y:pe-2.5'
        )}
        data-slot="scroll-area-viewport"
      >
        {children}
      </ScrollAreaPrimitive.Viewport>
      {orientation !== 'horizontal' ? <ScrollBar orientation="vertical" /> : null}
      {orientation !== 'vertical' ? <ScrollBar orientation="horizontal" /> : null}
      {/* The corner is the square where the two bars meet, so it only exists
          when both of them do. */}
      {orientation === 'both' ? (
        <ScrollAreaPrimitive.Corner data-slot="scroll-area-corner" />
      ) : null}
    </ScrollAreaPrimitive.Root>
  );
}

export function ScrollBar({
  className,
  orientation = 'vertical',
  ...props
}: ScrollAreaPrimitive.Scrollbar.Props): React.ReactElement {
  return (
    <ScrollAreaPrimitive.Scrollbar
      className={cn(
        'm-1 flex opacity-0 transition-opacity delay-300 data-hovering:opacity-100 data-hovering:delay-0 data-hovering:duration-100 data-scrolling:opacity-100 data-scrolling:delay-0 data-scrolling:duration-100 data-[orientation=horizontal]:h-1.5 data-[orientation=horizontal]:flex-col data-[orientation=vertical]:w-1.5',
        className
      )}
      data-slot="scroll-area-scrollbar"
      orientation={orientation}
      {...props}
    >
      <ScrollAreaPrimitive.Thumb
        className="bg-fg/20 relative flex-1 rounded-full"
        data-slot="scroll-area-thumb"
      />
    </ScrollAreaPrimitive.Scrollbar>
  );
}

export { ScrollAreaPrimitive };
