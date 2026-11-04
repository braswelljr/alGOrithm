'use client';

import { useEffect, useId, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Button } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import type { TourStep } from '@/lib/tour/define';
import {
  inView,
  placeCard,
  spotlightBox,
  type Box,
  type CardPosition,
  type Size
} from '@/lib/tour/position';
import { cn } from '@/lib/utils';

/**
 * Whether movement should be cut: the in-app choice on `[data-preferences]`
 * (Settings, Accessibility) wins, and `system` follows the device.
 */
export function reducedMotion(): boolean {
  if (typeof window === 'undefined') return true;
  const choice = document.querySelector('[data-preferences]')?.getAttribute('data-motion');
  if (choice === 'reduce') return true;
  if (choice === 'full') return false;
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;
}

const viewportSize = (): Size => ({ width: window.innerWidth, height: window.innerHeight });

const boxOf = (el: Element): Box => {
  const r = el.getBoundingClientRect();
  return { top: r.top, left: r.left, width: r.width, height: r.height };
};

type Props = {
  target: Element;
  step: TourStep;
  index: number;
  total: number;
  onNext: () => void;
  onBack: () => void;
  onSkip: () => void;
  /** Told where focus went when something outside took it, such as a closing menu handing it back. */
  onFocusOutside?: (el: HTMLElement) => void;
};

/**
 * One step on screen: the page dimmed with a cut-out around the part being
 * explained, and a card beside it with the words and the buttons. Follows the
 * target as the page scrolls or resizes, one measurement a frame at most. The
 * cut-out is decoration (`aria-hidden`); the card is the dialog.
 */
export function TourOverlay({
  target,
  step,
  index,
  total,
  onNext,
  onBack,
  onSkip,
  onFocusOutside
}: Props) {
  const titleId = useId();
  const bodyId = useId();
  const card = useRef<HTMLDivElement>(null);
  const [spot, setSpot] = useState<Box | null>(null);
  const [place, setPlace] = useState<CardPosition | null>(null);
  // Moving between steps glides; following a scroll does not, or it would lag.
  const [gliding, setGliding] = useState(false);
  const still = reducedMotion();
  const last = index >= total - 1;

  // Bring the target on to the screen, once per step.
  useLayoutEffect(() => {
    if (!inView(boxOf(target), viewportSize())) {
      target.scrollIntoView?.({
        block: 'center',
        inline: 'nearest',
        behavior: still ? 'auto' : 'smooth'
      });
    }
    if (still) return;
    setGliding(true);
    const t = window.setTimeout(() => setGliding(false), 260);
    return () => window.clearTimeout(t);
  }, [target, still]);

  // Measure now and whenever anything could have moved it.
  useLayoutEffect(() => {
    let frame = 0;
    const measure = () => {
      frame = 0;
      const viewport = viewportSize();
      const box = boxOf(target);
      const size = card.current
        ? { width: card.current.offsetWidth, height: card.current.offsetHeight }
        : { width: 352, height: 200 };
      setSpot(spotlightBox(box, viewport));
      setPlace(placeCard(box, size, viewport, step.placement));
    };
    const schedule = () => {
      if (!frame) frame = window.requestAnimationFrame(measure);
    };
    measure();
    window.addEventListener('scroll', schedule, { capture: true, passive: true });
    window.addEventListener('resize', schedule, { passive: true });
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(schedule);
    observer?.observe(target);
    if (card.current) observer?.observe(card.current);
    return () => {
      if (frame) window.cancelAnimationFrame(frame);
      window.removeEventListener('scroll', schedule, { capture: true });
      window.removeEventListener('resize', schedule);
      observer?.disconnect();
    };
  }, [target, step]);

  // Each step is announced: focus moves to the card, which names itself.
  useEffect(() => {
    card.current?.focus({ preventScroll: true });
  }, [step]);

  // The card is modal: focus that lands elsewhere (a menu that has just
  // closed giving it back to its button) is noted, for when the tour ends,
  // and brought back.
  useEffect(() => {
    const onFocus = (event: FocusEvent) => {
      const el = event.target;
      if (!(el instanceof HTMLElement) || card.current?.contains(el)) return;
      onFocusOutside?.(el);
      card.current?.focus({ preventScroll: true });
    };
    document.addEventListener('focusin', onFocus);
    return () => document.removeEventListener('focusin', onFocus);
  }, [onFocusOutside]);

  // Keys: Right or Enter go on, Left goes back, Escape leaves; Tab stays in the card.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.isComposing) return;
      const from = event.target instanceof Element ? event.target : null;
      const inCard = !!from && !!card.current?.contains(from);
      const onControl = inCard && !!from?.closest('button, a');
      if (event.key === 'Escape') {
        event.preventDefault();
        event.stopPropagation();
        onSkip();
      } else if (event.key === 'ArrowRight' || (event.key === 'Enter' && !onControl)) {
        event.preventDefault();
        event.stopPropagation();
        onNext();
      } else if (event.key === 'ArrowLeft') {
        event.preventDefault();
        event.stopPropagation();
        if (index > 0) onBack();
      } else if (event.key === 'Tab') {
        const focusable = card.current?.querySelectorAll<HTMLElement>('button:not(:disabled)');
        if (!focusable || focusable.length < 1) return;
        const first = focusable[0];
        const end = focusable[focusable.length - 1];
        const active = document.activeElement;
        if (
          !inCard ||
          (event.shiftKey && active === first) ||
          (!event.shiftKey && active === end)
        ) {
          event.preventDefault();
          (event.shiftKey ? end : first).focus();
        }
      }
    };
    window.addEventListener('keydown', onKey, { capture: true });
    return () => window.removeEventListener('keydown', onKey, { capture: true });
  }, [index, onNext, onBack, onSkip]);

  const glide = gliding && !still;
  const sheet = place?.mode === 'sheet';

  return createPortal(
    <div
      data-tour-overlay=""
      className="fixed inset-0 z-[100]"
    >
      {/* Catches clicks so the page is not used by accident mid-tour. */}
      <div
        aria-hidden
        className="absolute inset-0"
      />
      {spot ? (
        <div
          aria-hidden
          data-tour-spotlight=""
          className={cn(
            'ring-primary pointer-events-none absolute top-0 left-0 rounded-lg shadow-[0_0_0_9999px_rgb(9_9_11/0.55)] ring-2 dark:shadow-[0_0_0_9999px_rgb(0_0_0/0.7)]',
            glide &&
              'transition-[transform,width,height] duration-[240ms] ease-[cubic-bezier(0.23,1,0.32,1)]'
          )}
          style={{
            width: spot.width,
            height: spot.height,
            transform: `translate3d(${spot.left}px, ${spot.top}px, 0)`
          }}
        />
      ) : null}
      <div
        ref={card}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={bodyId}
        tabIndex={-1}
        data-tour-card=""
        className={cn(
          // card-ok: the tour's floating step, placed over the page by hand.
          'border-border bg-card text-fg absolute w-[22rem] max-w-[calc(100vw-2rem)] rounded-xl border p-4 shadow-xl outline-none',
          sheet ? 'inset-x-4 w-auto max-w-none' : 'top-0 left-0',
          sheet && place.edge === 'bottom' && 'bottom-4',
          sheet && place.edge === 'top' && 'top-4',
          !place && 'invisible',
          glide &&
            !sheet &&
            'transition-transform duration-[240ms] ease-[cubic-bezier(0.23,1,0.32,1)]'
        )}
        style={
          place?.mode === 'anchored'
            ? { transform: `translate3d(${place.left}px, ${place.top}px, 0)` }
            : undefined
        }
      >
        <p className="text-muted text-xs font-semibold tracking-wide uppercase">
          Step {index + 1} of {total}
        </p>
        <Heading
          level={2}
          id={titleId}
          className="font-heading mt-1 text-base font-semibold"
        >
          {step.title}
        </Heading>
        <p
          id={bodyId}
          className="text-muted mt-1.5 text-sm"
        >
          {step.body}
        </p>
        <div className="mt-4 flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="bg-transparent"
            onClick={onSkip}
          >
            Skip tour
          </Button>
          <span className="flex-1" />
          {index > 0 ? (
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={onBack}
            >
              Back
            </Button>
          ) : null}
          <Button
            type="button"
            variant="primary"
            size="sm"
            onClick={onNext}
          >
            {last ? 'Done' : 'Next'}
          </Button>
        </div>
      </div>
    </div>,
    document.body
  );
}
