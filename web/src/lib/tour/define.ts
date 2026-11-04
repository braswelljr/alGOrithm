/**
 * A guided tour: a few steps, each pointing at one part of a page and saying
 * in plain words what it is and why it is there (after gsl-certificates).
 *
 * Tours are data. Each lives in its own file under `lib/tour/tours/`, is
 * listed in `tours/index.ts`, and points at the page by stable
 * `data-tour="..."` attributes rather than by classes or text, so a restyled
 * page keeps its tour (docs/rules/a-page-with-a-tour-marks-its-parts-with-data-tour.md).
 */

/** Where the step's card sits beside its target; `auto` picks the side with room. */
export type TourPlacement = 'top' | 'bottom' | 'left' | 'right' | 'auto';

export type TourStep = {
  /** A CSS selector, always `[data-tour="..."]`. A step whose target is missing is skipped. */
  target: string;
  title: string;
  body: string;
  placement?: TourPlacement;
};

export type Tour = {
  /** Stable: whether a visitor has seen the tour is kept under it. */
  id: string;
  /** What the menus call it, such as "Tour of the visualizer". */
  title: string;
  steps: TourStep[];
  /** The page the tour belongs to, as a pattern (`/docs/:id`); none belongs nowhere in particular. */
  route?: string;
};

/** Declares a tour. Typed so a tour file reads as data. */
export function defineTour(t: Tour): Tour {
  return t;
}

/** A path without its trailing slash; the root stays "/". */
const trim = (p: string) => p.replace(/\/+$/, '') || '/';

/** Whether `pathname` is the page `pattern` names: `:x` is one segment, a final `*` the rest. */
export function matchRoute(pattern: string, pathname: string): boolean {
  const want = trim(pattern).split('/');
  const have = trim(pathname).split('/');
  for (let i = 0; i < want.length; i++) {
    if (want[i] === '*') return true;
    if (i >= have.length) return false;
    if (want[i].startsWith(':')) {
      if (!have[i]) return false;
      continue;
    }
    if (want[i] !== have[i]) return false;
  }
  return want.length === have.length;
}

/** Whether the tour belongs to the page at `pathname`. */
export function tourMatches(tour: Tour, pathname: string): boolean {
  return !!tour.route && matchRoute(tour.route, pathname);
}

/** Whether the tour's page can be opened as it is, with nothing to fill in. */
export function tourRouteIsConcrete(tour: Tour): boolean {
  return !!tour.route && !tour.route.includes(':') && !tour.route.includes('*');
}
