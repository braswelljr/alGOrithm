'use client';

import React, { useCallback, useId, useRef, useState } from 'react';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { useIsomorphicLayoutEffect } from '@braswelljr/algorithm-ui/hooks/use-isomorphic-layout-effect';
import { useMediaQuery } from '@braswelljr/algorithm-ui/hooks/use-media-query';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { ChevronLeft, ChevronRight } from 'lucide-react';

/** How far an arrow pushes the row: four fifths of what is on screen, so a
 *  column of what was just seen stays at the other edge, as the tabs do. */
const SCROLL_STEP = 0.8;

/** How much of each edge the arrows sit over, and so how much the mask fades. */
const EDGE = 36;

/**
 * A row of actions that never runs past its space: what does not fit scrolls
 * sideways, with an arrow at each end that has more to see, the way a tab
 * strip does. The arrows sit over the row rather than beside it, so appearing
 * costs the row no width, and the actions fade under them. `label` names the
 * row for the arrows: "actions".
 */
export function OverflowRow({
  className,
  rowClassName,
  label = 'actions',
  children
}: {
  /** Classes for the frame: its width, say. */
  className?: string;
  /** Classes for the row of actions inside it. */
  rowClassName?: string;
  label?: string;
  children: React.ReactNode;
}): React.ReactElement {
  const rootRef = useRef<HTMLDivElement>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  const viewportId = useId();
  const reduce = useMediaQuery('(prefers-reduced-motion: reduce)');
  const [edges, setEdges] = useState({ overflow: false, left: false, right: false });

  const measure = useCallback(() => {
    const viewport = viewportRef.current;
    if (!viewport) return;
    const max = Math.max(0, viewport.scrollWidth - viewport.clientWidth);
    const overflow = max > 1;
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

  useIsomorphicLayoutEffect(() => {
    const root = rootRef.current;
    const viewport = viewportRef.current;
    if (!root || !viewport) return;
    // The frame narrowing and the row growing (a label changing, a font
    // arriving late) change whether it overflows; neither is a render.
    const observer = new ResizeObserver(measure);
    observer.observe(root);
    if (viewport.firstElementChild) observer.observe(viewport.firstElementChild);
    viewport.addEventListener('scroll', measure, { passive: true });
    measure();
    return () => {
      observer.disconnect();
      viewport.removeEventListener('scroll', measure);
    };
  }, [measure, children]);

  const scroll = (direction: number) => {
    const viewport = viewportRef.current;
    if (!viewport) return;
    viewport.scrollBy({
      left: direction * viewport.clientWidth * SCROLL_STEP,
      behavior: reduce ? 'instant' : 'smooth'
    });
  };

  return (
    <div
      ref={rootRef}
      data-slot="overflow-row"
      className={cn('relative isolate flex max-w-full min-w-0 items-center', className)}
    >
      {edges.overflow ? (
        <ScrollArrow
          direction="left"
          label={label}
          disabled={!edges.left}
          controls={viewportId}
          onClick={() => scroll(-1)}
        />
      ) : null}
      <div
        ref={viewportRef}
        id={viewportId}
        className="scrollbar-none max-w-full min-w-0 overflow-x-auto [&::-webkit-scrollbar]:hidden"
        style={
          edges.overflow
            ? {
                maskImage: `linear-gradient(to right, ${
                  edges.left ? `transparent, black ${EDGE}px` : 'black, black 0px'
                }, ${edges.right ? `black calc(100% - ${EDGE}px), transparent` : 'black 100%'})`
              }
            : undefined
        }
        // An action reached with the keyboard is brought into view.
        onFocusCapture={(event) => {
          if (event.target instanceof HTMLElement && event.target !== viewportRef.current) {
            event.target.scrollIntoView({
              block: 'nearest',
              inline: 'nearest',
              behavior: reduce ? 'instant' : 'smooth'
            });
          }
        }}
      >
        {/* The padding keeps focus rings in. */}
        <div className={cn('flex w-max items-center gap-2 p-1 *:shrink-0', rowClassName)}>
          {children}
        </div>
      </div>
      {edges.overflow ? (
        <ScrollArrow
          direction="right"
          label={label}
          disabled={!edges.right}
          controls={viewportId}
          onClick={() => scroll(1)}
        />
      ) : null}
    </div>
  );
}

/** One end of the row: faded out rather than removed at its own end, so the
 *  other does not move under a reader aiming at it. */
function ScrollArrow({
  direction,
  label,
  disabled,
  controls,
  onClick
}: {
  direction: 'left' | 'right';
  label: string;
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
      aria-label={`Scroll ${label} ${direction}`}
      aria-controls={controls}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'absolute top-1/2 z-20 inline-flex -translate-y-1/2 items-center justify-center',
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
