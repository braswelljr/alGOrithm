'use client';

import React, { useCallback, useId, useRef, useState } from 'react';
import { Tabs as TabsPrimitive } from '@base-ui/react/tabs';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { useIsomorphicLayoutEffect } from '@braswelljr/algorithm-ui/hooks/use-isomorphic-layout-effect';
import { useMediaQuery } from '@braswelljr/algorithm-ui/hooks/use-media-query';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { ChevronLeft, ChevronRight } from 'lucide-react';

export type TabsVariant = 'default' | 'underline';

export function Tabs({ className, ...props }: TabsPrimitive.Root.Props): React.ReactElement {
  return (
    <TabsPrimitive.Root
      className={cn('flex flex-col gap-2 data-[orientation=vertical]:flex-row', className)}
      data-slot="tabs"
      {...props}
    />
  );
}

/**
 * How far the arrows push the list, as a fraction of what is on screen.
 *
 * Not a whole viewport: a press that replaces every visible tab gives the reader
 * nothing to navigate by, and they lose their place. Four fifths leaves a column
 * of what they were just looking at on the other edge.
 */
const SCROLL_STEP = 0.8;

/** How much of each edge the arrows sit over, and so how much the mask fades. */
const EDGE = 40;

/**
 * A strip of tabs. A strip of one tab offers no choice, so it is not drawn:
 * the page shows its one panel, and a second tab (an ability, a role) brings
 * the strip back.
 */
export function TabsList(props: React.ComponentProps<typeof TabStrip>): React.ReactElement | null {
  if (React.Children.toArray(props.children).length < 2) return null;
  return <TabStrip {...props} />;
}

function TabStrip({
  variant = 'default',
  className,
  wrapperClassName,
  indicatorClassName,
  scrollable = true,
  children,
  ...props
}: TabsPrimitive.List.Props & {
  variant?: TabsVariant;
  indicatorClassName?: string;
  /** Classes for the scrolling frame around the list, where there is one. */
  wrapperClassName?: string;
  /**
   * Whether a list too wide for its space scrolls. On by default.
   *
   * Turn it off where the caller has already put the list in a scroller of its
   * own, so the two do not nest and produce a bar inside a bar.
   */
  scrollable?: boolean;
}): React.ReactElement {
  const rootRef = useRef<HTMLDivElement>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  const viewportId = useId();
  const reduce = useMediaQuery('(prefers-reduced-motion: reduce)');
  const [edges, setEdges] = useState({ overflow: false, left: false, right: false });

  /**
   * Whether the list is wider than its frame, and which ends are off screen.
   *
   * Measured against the frame rather than the viewport, because the arrows sit
   * over the viewport rather than beside it: they take no width, so the space
   * available does not change when they appear. Measuring the other way round
   * makes them flicker, each appearance shrinking the room that justified it.
   */
  const measure = useCallback(() => {
    const root = rootRef.current;
    const viewport = viewportRef.current;
    if (!root || !viewport) return;
    // The list's own width, not the viewport's scrollWidth: the indicator is
    // moved with a transform, and under the last tab it reaches a pixel or two
    // past the list, which scrollWidth counts and a reader never sees. Selected
    // on load, the last tab drew the arrows over a strip that fitted.
    const list = viewport.firstElementChild as HTMLElement | null;
    const overflow = (list?.offsetWidth ?? viewport.scrollWidth) > root.clientWidth + 1;
    const max = Math.max(0, viewport.scrollWidth - viewport.clientWidth);
    // Right-to-left reports a negative offset in modern browsers, and both
    // directions rubber-band past their own bounds on a touchpad. Clamp.
    const rtl = getComputedStyle(viewport).direction === 'rtl';
    const fromLeft = Math.max(
      0,
      Math.min(max, rtl ? max + viewport.scrollLeft : viewport.scrollLeft)
    );
    const next = { overflow, left: fromLeft > 1, right: fromLeft < max - 1 };
    setEdges((previous) =>
      previous.overflow === next.overflow &&
      previous.left === next.left &&
      previous.right === next.right
        ? previous
        : next
    );
  }, []);

  /**
   * Bring a tab into view, clear of the arrows.
   *
   * The selected tab can be off screen for reasons that have nothing to do with
   * scrolling: the page restored a selection from the URL, or the reader tabbed
   * to it with the keyboard. A tab that is selected and invisible is the state
   * this exists to prevent.
   *
   * `scrollBy` on the viewport rather than `scrollIntoView`, which also moves
   * every scrollable ancestor and can drag the whole page sideways.
   */
  const reveal = useCallback(
    (tab: HTMLElement | null) => {
      const viewport = viewportRef.current;
      if (!viewport || !tab) return;
      const frame = viewport.getBoundingClientRect();
      const item = tab.getBoundingClientRect();
      const max = Math.max(0, viewport.scrollWidth - viewport.clientWidth);
      const rtl = getComputedStyle(viewport).direction === 'rtl';
      const fromLeft = Math.max(
        0,
        Math.min(max, rtl ? max + viewport.scrollLeft : viewport.scrollLeft)
      );
      const left = frame.left + (fromLeft > 1 ? EDGE : 0);
      const right = frame.right - (fromLeft < max - 1 ? EDGE : 0);
      const delta =
        item.left < left ? item.left - left : item.right > right ? item.right - right : 0;
      if (delta) viewport.scrollBy({ left: delta, behavior: reduce ? 'instant' : 'smooth' });
    },
    [reduce]
  );

  useIsomorphicLayoutEffect(() => {
    if (!scrollable) return;
    const root = rootRef.current;
    const viewport = viewportRef.current;
    if (!root || !viewport) return;

    const update = () => {
      measure();
      reveal(viewport.querySelector<HTMLElement>('[role="tab"][data-selected]'));
    };
    // Three things change whether this overflows and none of them is a render:
    // the frame narrowing, the list growing, and a font arriving late. A resize
    // observer on both ends catches all three.
    const observer = new ResizeObserver(update);
    observer.observe(root);
    observer.observe(viewport);

    viewport.addEventListener('scroll', measure, { passive: true });
    // The indicator slides to the selected tab. While it moves it can reach
    // past the last tab and widen what scrolls, and its moving resizes no
    // frame, so measure again once it has arrived; otherwise the arrows a
    // passing frame drew stay over a strip that fits.
    viewport.addEventListener('transitionend', measure);
    update();
    return () => {
      observer.disconnect();
      viewport.removeEventListener('scroll', measure);
      viewport.removeEventListener('transitionend', measure);
    };
  }, [scrollable, measure, reveal, children]);

  const scroll = (direction: number) => {
    const viewport = viewportRef.current;
    if (!viewport) return;
    viewport.scrollBy({
      left: direction * viewport.clientWidth * SCROLL_STEP,
      behavior: reduce ? 'instant' : 'smooth'
    });
  };

  const list = (
    <TabsPrimitive.List
      className={cn(
        'relative z-0 flex w-fit items-center justify-center gap-x-0.5 text-fg',
        'data-[orientation=vertical]:flex-col',
        variant === 'default'
          ? 'rounded-lg bg-subtle p-0.5 text-neutral-950'
          : 'data-[orientation=horizontal]:py-1 data-[orientation=vertical]:px-1 *:data-[slot=tabs-tab]:hover:bg-subtle',
        className
      )}
      data-slot="tabs-list"
      {...props}
    >
      {children}
      <TabsPrimitive.Indicator
        className={cn(
          'absolute bottom-0 left-0 h-(--active-tab-height) w-(--active-tab-width) translate-x-(--active-tab-left) -translate-y-(--active-tab-bottom) transition-[width,height,translate] duration-200 ease-in-out',
          // The accent, not a neutral.
          //
          // Selection is one of the few things the brand colour is for on this
          // platform - active nav, the one primary CTA, selection and focus - and
          // the indicator is the selection. It was `zinc-800` over
          // `zinc-200`, which reads as another surface rather than as a mark,
          // and inverted between modes for no reason a reader can see: the
          // selected tab was near-black in light mode and near-white in dark.
          // `bg-primary` is the same gold in both, so a tab strip looks like
          // itself either way and like the active sidebar row above it.
          variant === 'underline'
            ? 'z-10 bg-primary data-[orientation=horizontal]:h-0.5 data-[orientation=horizontal]:translate-y-px data-[orientation=vertical]:w-0.5 data-[orientation=vertical]:-translate-x-px'
            : '-z-1 rounded-md bg-primary shadow-sm/5',
          indicatorClassName
        )}
        data-slot="tab-indicator"
      />
    </TabsPrimitive.List>
  );

  if (!scrollable) return list;

  /**
   * The scrolling frame.
   *
   * A vertical list needs none of this and gets it anyway, harmlessly: it stacks,
   * so it never overflows sideways, `overflow` stays false and no arrow is drawn.
   * Branching on the orientation would mean reading it out of the DOM before the
   * first paint, and measuring answers the same question later and more honestly.
   */
  return (
    <div
      ref={rootRef}
      data-slot="tabs-list-frame"
      className={cn('relative isolate flex max-w-full min-w-0 items-center', wrapperClassName)}
    >
      {edges.overflow ? (
        <TabsScrollButton
          direction="left"
          disabled={!edges.left}
          controls={viewportId}
          onClick={() => scroll(-1)}
        />
      ) : null}

      <div
        ref={viewportRef}
        id={viewportId}
        className="scrollbar-none max-w-full min-w-0 overflow-x-auto [&::-webkit-scrollbar]:hidden"
        // The labels fade under the arrows rather than running into them. Only
        // at an end there is more to see, so the mask is also the affordance:
        // a hard edge means that is the end of the list.
        style={
          edges.overflow
            ? {
                maskImage: `linear-gradient(to right, ${
                  edges.left ? `transparent, black ${EDGE}px` : 'black, black 0px'
                }, ${edges.right ? `black calc(100% - ${EDGE}px), transparent` : 'black 100%'})`
              }
            : undefined
        }
        // Reached with the keyboard rather than the pointer: arrow keys move the
        // selection through tabs that may be off screen, and focus that cannot
        // be seen is the accessibility failure this whole control risks.
        onFocusCapture={(event) => {
          if (event.target instanceof HTMLElement && event.target.getAttribute('role') === 'tab') {
            reveal(event.target);
          }
        }}
      >
        {list}
      </div>

      {edges.overflow ? (
        <TabsScrollButton
          direction="right"
          disabled={!edges.right}
          controls={viewportId}
          onClick={() => scroll(1)}
        />
      ) : null}
    </div>
  );
}

/**
 * One end of the strip.
 *
 * Overlaid on the viewport rather than placed beside it, so appearing costs the
 * tabs no width. Disabled at its own end rather than removed: a control that
 * vanishes moves the one beside it, and a reader who was aiming at it presses
 * the wrong thing.
 *
 * Not a `Button`: the house button carries a background, a height and a hover
 * surface, and this is a glyph floating over content that has to stay legible
 * through it.
 */
function TabsScrollButton({
  direction,
  disabled,
  controls,
  onClick
}: {
  direction: 'left' | 'right';
  disabled: boolean;
  controls: string;
  onClick: () => void;
}): React.ReactElement {
  const Icon = direction === 'left' ? ChevronLeft : ChevronRight;
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-sm"
      rounded="full"
      aria-label={direction === 'left' ? 'Scroll tabs left' : 'Scroll tabs right'}
      aria-controls={controls}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'absolute top-1/2 z-20 inline-flex -translate-1/2 items-center justify-center',
        // Faded out rather than hidden, so the strip does not resize at an end.
        'disabled:pointer-events-none disabled:opacity-0',
        direction === 'left' ? 'left-0' : 'right-0'
      )}
    >
      <Icon
        className="size-5"
        aria-hidden="true"
      />
    </Button>
  );
}

export function TabsTab({ className, ...props }: TabsPrimitive.Tab.Props): React.ReactElement {
  return (
    <TabsPrimitive.Tab
      className={cn(
        // Dark text on the selected tab, keyed on both `data-active` (what
        // base-ui sets on the tab under the indicator) and `data-selected`
        // (what this file's own `reveal` queries), both forced so a caller's
        // `text-*` cannot win by source order. The indicator behind a
        // selected tab is now `bg-primary`, a mid-gold: white on it measures
        // about 2.8:1 and fails AA, dark text about 7.4:1. Same rule as
        // `buttonVariants`' primary, and it holds in both modes because the gold
        // does.
        "relative flex h-9 shrink-0 grow cursor-pointer items-center justify-center gap-1.5 rounded-md px-[calc(--spacing(2.5)-1px)] text-base font-medium whitespace-nowrap transition-[color,background-color,box-shadow] outline-none focus-visible:ring-2 focus-visible:ring-fg focus-visible:ring-inset data-active:text-neutral-950! data-disabled:pointer-events-none data-selected:text-neutral-950! data-[orientation=vertical]:w-full data-[orientation=vertical]:justify-start sm:h-8 sm:text-sm [&_svg]:pointer-events-none [&_svg]:-mx-0.5 [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4.5 sm:[&_svg:not([class*='size-'])]:size-4",
        className
      )}
      data-slot="tabs-tab"
      {...props}
    />
  );
}

export function TabsPanel({
  className,
  keepMounted = true,
  ...props
}: TabsPrimitive.Panel.Props): React.ReactElement {
  return (
    <TabsPrimitive.Panel
      keepMounted={keepMounted}
      className={cn(
        'flex-1 rounded-md outline-none focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-control/40',
        'transition-[opacity,transform] duration-200 ease-(--ease-out-strong)',
        'data-starting-style:translate-y-1 data-starting-style:opacity-0',
        'data-ending-style:translate-y-1 data-ending-style:opacity-0',
        className
      )}
      data-slot="tabs-content"
      {...props}
    />
  );
}

export { TabsPrimitive, TabsTab as TabsTrigger, TabsPanel as TabsContent };
