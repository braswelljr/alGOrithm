import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

/**
 * The height and radius every control in a filter row takes.
 *
 * # Why this is CSS and not a `size` prop
 *
 * Two scales meet in a filter row and they do not intersect. `Input` and
 * `SelectTrigger` run h-7/8.5/10; `Button`, which is what a facet's trigger is,
 * runs h-6/7/8/9/10. There is no `size` a facet can be given that matches an
 * input: the nearest are 8 and 9 against 8.5. So no call site can line them up,
 * however carefully it passes props, and every filter row in the app had a
 * search box half a step taller than the facets beside it.
 *
 * 8.5 because it is where the two controls that already agree sit. The facets
 * and the buttons come to the search box rather than the other way round.
 *
 * Exported because two kinds of filter row exist: the registers, which go
 * through `DataTableToolbar`, and the public register, which is a hand-built row
 * with a label over each control. One definition, so they cannot drift.
 *
 * A descendant selector, not a child one: a facet's trigger is a button inside a
 * popover root rather than a direct child. Its option rows are portalled out, so
 * they keep their own size.
 */
/**
 * These are descendant selectors, so they beat a control's own class on
 * specificity rather than on order: `.toolbar button` is (0,1,1) against
 * `rounded-full` at (0,1,0), and no amount of `cn` merging changes that,
 * because tailwind-merge only ever sees one of the two class lists.
 *
 * That is worth knowing because it is how the whole toolbar stayed square
 * after `Button`, `Select` and `InputGroup` each became a pill: the components
 * were right, and this line was quietly overruling all three.
 */
export const TOOLBAR_CONTROL_SIZING =
  '[&_button]:h-8.5 [&_button]:rounded-full ' +
  '[&_[data-slot=select-trigger]]:h-8.5 [&_[data-slot=select-trigger]]:rounded-full ' +
  '[&_[data-slot=input-group]]:h-8.5 [&_[data-slot=input-group]]:rounded-full ' +
  // An icon-only button has to be square, and setting the height alone does
  // not make it one. `size="icon"` is `size-8` and `size="icon-sm"` is
  // `size-7`, which set width and height together; the `h-8.5` above outranks
  // only the height half, so every refresh, export and clear control in the
  // bar came out 32x34 or 28x34 - a shade taller than the search box beside
  // it and visibly not a circle now that it is rounded.
  //
  // Keyed on the attribute `Button` sets from its own size rather than on the
  // shape of the markup. Inferring it with `:has(> svg:only-child)` looked
  // right and was not: `:only-child` counts elements and ignores text, so a
  // button reading `<svg/>Reset` matched and lost its label.
  '[&_button[data-icon-only]]:w-8.5 [&_button[data-icon-only]]:p-0';

/**
 * The bar above a register: what narrows it on the left, what acts on it on the
 * right.
 *
 * # Why this is a component and not a class name people remember
 *
 * Every register here grew its own bar: `justify-between gap-5`, `gap-3`,
 * `gap-2`, one with no `justify-between` at all, one with the search nested
 * inside the group of controls, and one built from `SearchBox` rather than
 * `InputGroup`. Twenty near-identical bars read as twenty, and the differences
 * are invisible until two are open side by side, which is exactly when a reader
 * notices the filters have moved.
 *
 * Normalising the classes fixed what was on screen and enforced nothing: the
 * twenty-second register is written by copying the twenty-first, and the rule
 * lives in whichever one gets copied. This is the rule.
 *
 * # Why `ms-auto` and not `justify-between`
 *
 * `justify-between` spreads *every* child, so a bar that grows a third
 * top-level element stops being left-and-right and becomes three things strung
 * across the width with the middle one floating somewhere nobody chose. Giving
 * the search its own slot and pushing the rest into one group makes that
 * impossible to express, and `ms-auto` pins the group right whether or not
 * there is a search beside it.
 *
 * # What goes where
 *
 * `search` is the search box and nothing else: it is the control a reader
 * reaches for first and its position should never move. Everything else -
 * facets, the date range, export, refresh, clear - is a child, and they sit
 * together on the right in the order they are given.
 *
 * A register with no search passes nothing and its controls still sit right,
 * because the alternative is a group that jumps to the left on one page in
 * twenty.
 *
 * # Why the search slot grows
 *
 * The slot carried `min-w-0` and nothing else, so it was a flex item at
 * `width: auto`: it shrink-wrapped whatever was put in it. `InputGroup` is
 * `w-full`, and a percentage width against a parent that is sizing itself to
 * its content resolves to the content - so every search box in every register
 * rendered at the bare intrinsic width of a text input, and the `max-w-lg` the
 * call site asked for was never reached because the box never got near it.
 *
 * `grow` is what gives the slot a real width to be a percentage of. `basis-64`
 * is what keeps it honest when the row is narrow: at `flex-1`'s `basis: 0%` an
 * item always fits, so the controls would stay on the line and squeeze the
 * search to nothing rather than wrapping. With a 16rem basis the line cannot
 * hold both, the controls wrap beneath, and the search keeps its width.
 *
 * # Not `DataTableToolbarSection`
 *
 * That one is vendored with the table library, exported from nowhere and used by
 * nothing here. It is a bare flex container with no opinion about what sits
 * where: its own documentation shows hand-rolling `justify-between` around two
 * divs, which is the pattern this replaces. Left alone as vendored code; this is
 * the app's bar.
 *
 * # No `role="toolbar"`
 *
 * The role promises arrow-key navigation between the controls, and nothing here
 * implements it. A promise the keyboard does not keep is worse than no role: a
 * reader who takes it at its word finds the arrows do nothing. The controls are
 * ordinary focusable elements and tab through in the order they are written.
 */
export function DataTableToolbar({
  search,
  children,
  className
}: {
  /** The search input. Right, always, and on its own. */
  search?: ReactNode;
  /** Facets, exports, refresh, clear: everything that is not the search. */
  children?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-wrap items-center gap-3', className)}>
      {/* Facets on the left from `md` up: they are what a reader reaches for
          to narrow the register, and they read as a row of what is currently
          applied.
          
          Below `md` the order goes back to search first. A narrow bar wraps
          into stacked rows, and a column of facet chips above the search box
          pushes it off the first screen - so on a phone the thing most people
          open a register to do would need a scroll to reach.
          
          See TOOLBAR_CONTROL_SIZING: the two control scales do not intersect,
          so the heights are settled here rather than at twenty-two call
          sites. */}
      <div
        className={cn(
          'order-2 flex flex-wrap items-center gap-2 md:order-1',
          TOOLBAR_CONTROL_SIZING
        )}
      >
        {children}
      </div>
      {/* `grow` still lets the box take the slack, so a bar with two facets
          gives it most of the row and a bar with eight gives it the end. */}
      {search ? (
        <div
          className={cn(
            // Full width on a phone, where it is the first thing in the bar
            // and there is nothing to share the row with. From `md` it stops
            // growing and sits hard against the right edge: with `grow` it
            // stretched to fill whatever the facets left, so a register with
            // two facets got a search box half the screen wide and one with
            // eight got a stub, and neither lined up with the next register.
            'order-1 min-w-0 grow basis-64 md:order-2 md:ml-auto md:w-72 md:grow-0 md:basis-auto',
            TOOLBAR_CONTROL_SIZING
          )}
        >
          {search}
        </div>
      ) : null}
    </div>
  );
}
