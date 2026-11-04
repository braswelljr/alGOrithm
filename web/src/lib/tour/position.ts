import type { TourPlacement } from './define';

export type Box = { top: number; left: number; width: number; height: number };
export type Size = { width: number; height: number };

/** Narrower than this, the step's card is a sheet across the bottom (or top) of the screen. */
export const SHEET_BELOW = 640;
/** The page's side gutter, kept clear on every edge. */
export const GUTTER = 16;
/** Between the target and its card. */
export const GAP = 12;
/** Around the target, inside the cut-out. */
export const SPOTLIGHT_PADDING = 6;

export type CardPosition =
  | { mode: 'anchored'; side: Exclude<TourPlacement, 'auto'>; top: number; left: number }
  | { mode: 'sheet'; edge: 'top' | 'bottom' };

/** The cut-out around a target: padded, and kept on the screen. */
export function spotlightBox(target: Box, viewport: Size): Box {
  const top = Math.max(target.top - SPOTLIGHT_PADDING, 0);
  const left = Math.max(target.left - SPOTLIGHT_PADDING, 0);
  const bottom = Math.min(target.top + target.height + SPOTLIGHT_PADDING, viewport.height);
  const right = Math.min(target.left + target.width + SPOTLIGHT_PADDING, viewport.width);
  return { top, left, width: Math.max(right - left, 0), height: Math.max(bottom - top, 0) };
}

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), Math.max(min, max));

/**
 * Where a step's card goes. On a phone it is a sheet on the edge away from
 * the target. Otherwise it sits on the side asked for, or for `auto` the
 * first side with room (below, above, right, left), else the roomiest; and it
 * slides along that side to stay on the screen.
 */
export function placeCard(
  target: Box,
  card: Size,
  viewport: Size,
  placement: TourPlacement = 'auto'
): CardPosition {
  if (viewport.width < SHEET_BELOW) {
    const centre = target.top + target.height / 2;
    return { mode: 'sheet', edge: centre > viewport.height * 0.55 ? 'top' : 'bottom' };
  }

  const room = {
    bottom: viewport.height - (target.top + target.height) - GAP - GUTTER,
    top: target.top - GAP - GUTTER,
    right: viewport.width - (target.left + target.width) - GAP - GUTTER,
    left: target.left - GAP - GUTTER
  };
  const fits = {
    bottom: room.bottom >= card.height,
    top: room.top >= card.height,
    right: room.right >= card.width,
    left: room.left >= card.width
  };
  const order: Array<keyof typeof room> = ['bottom', 'top', 'right', 'left'];
  let side: keyof typeof room;
  if (placement !== 'auto' && fits[placement]) side = placement;
  else {
    side =
      order.find((s) => fits[s]) ??
      order.reduce((best, s) => {
        const spare = (k: keyof typeof room) =>
          k === 'top' || k === 'bottom' ? room[k] - card.height : room[k] - card.width;
        return spare(s) > spare(best) ? s : best;
      }, order[0]);
  }

  const maxLeft = viewport.width - card.width - GUTTER;
  const maxTop = viewport.height - card.height - GUTTER;
  if (side === 'bottom' || side === 'top') {
    const left = clamp(target.left + target.width / 2 - card.width / 2, GUTTER, maxLeft);
    const top =
      side === 'bottom'
        ? clamp(target.top + target.height + GAP, GUTTER, maxTop)
        : clamp(target.top - GAP - card.height, GUTTER, maxTop);
    return { mode: 'anchored', side, top, left };
  }
  const top = clamp(target.top + target.height / 2 - card.height / 2, GUTTER, maxTop);
  const left =
    side === 'right'
      ? clamp(target.left + target.width + GAP, GUTTER, maxLeft)
      : clamp(target.left - GAP - card.width, GUTTER, maxLeft);
  return { mode: 'anchored', side, top, left };
}

/** Whether the box is wholly on the screen, so it need not be scrolled to. */
export function inView(target: Box, viewport: Size): boolean {
  return (
    target.top >= 0 &&
    target.left >= 0 &&
    target.top + target.height <= viewport.height &&
    target.left + target.width <= viewport.width
  );
}
