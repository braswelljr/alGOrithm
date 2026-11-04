import * as React from 'react';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';
import { cva, type VariantProps } from 'class-variance-authority';

const badgeVariants = cva(
  'inline-flex items-center rounded-full border border-border font-semibold whitespace-nowrap transition-colors focus-visible:ring-2 focus-visible:ring-fg focus-visible:ring-offset-2 focus-visible:outline-none',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-fg text-bg',
        secondary: 'border-transparent bg-subtle text-fg',
        destructive: 'border-transparent bg-red-600! text-white!',
        outline: 'border-border bg-card text-muted',
        // blue-600, not 500: white on blue-500 is 3.7:1, under AA for text this size.
        info: 'border-transparent bg-blue-600! text-white!',
        success: 'border-transparent bg-emerald-500! text-zinc-950!',
        warning: 'border-transparent bg-yellow-500! text-zinc-950!',

        // Yellow, amber and orange all take dark text. They are bright
        // mid-tones: white on `amber-500` measures about 2.1:1 and on
        // `orange-500` about 2.6:1, both well under WCAG AA, while
        // `zinc-950` on them is above 8:1. The instinct is that a saturated
        // colour wants light text, and on these three it is the wrong one -
        // the same reason `bg-primary` takes `text-fg`.
        // ── Status tones ────────────────────────────────────────────────────
        // The solid palette every status badge on the platform draws from, and
        // what the server's `meta.tones` maps onto.
        //
        // # Why these are here rather than in three cva blocks
        //
        // They were in three: `tone-badge.tsx`, `application-status-badge.tsx`
        // and `case-status-badge.tsx`, each with its own copy of the same six
        // colours - and they had already drifted. `neutral` was `bg-muted`
        // in two of them and `bg-fg` in the third, so the same state
        // rendered a different grey depending on which screen you were on.
        //
        // # Why solid rather than tinted
        //
        // A status is the thing a reader scans a table for. A tinted badge -
        // pale background, coloured text - reads as decoration at a glance and
        // makes the row it sits in no easier to find; a solid one is a marker.
        // The four above are already solid, which is what these match.
        //
        // `!` on every declaration because a caller's `className` lands after
        // these in the class list and would otherwise win by source order
        // rather than by intent. The existing status variants do the same.
        neutral: 'border-transparent bg-muted! text-bg!',
        // Something is happening, and somebody else is doing it.
        progress: 'border-transparent bg-amber-500! text-zinc-950!',
        // Attention, but not failure. Distinct from `progress`: orange is
        // "look at this", amber is "this is moving".
        caution: 'border-transparent bg-orange-500! text-zinc-950!',
        danger: 'border-transparent bg-red-600! text-white!',

        // Shape only, no colours: for a badge whose palette is the caller's,
        // which is most status badges here. Selected automatically, see Badge.
        unstyled: 'border-transparent',
        /** The learning levels: a tint, an icon and the word (LevelBadge). */
        beginner: 'border-success/40 bg-success/10 text-success',
        intermediate: 'border-warning/40 bg-warning/10 text-warning',
        advanced: 'border-danger/40 bg-danger/10 text-danger',
        /** A cost in Big O, set in mono (ComplexityBadge). */
        cost: 'border-border bg-subtle font-mono text-fg'
      },
      /**
       * The height, by the name `Button` uses for the same step, so a badge
       * set beside a control can be sized to it. `default` is the small pill
       * a table cell wants; `md` and up are for a screen where the badge is
       * the content rather than a note on it.
       */
      size: {
        xs: 'h-4 px-1.5 text-xs',
        sm: 'h-5 px-1.5 text-xs',
        default: 'h-6 px-2.5 py-0.5 text-xs',
        md: 'h-7 gap-1.5 px-3 text-sm',
        lg: 'h-8 gap-1.5 px-3.5 text-sm',
        xl: 'h-9 gap-2 px-4 text-base'
      },
      /**
       * The corner.
       *
       * A variant rather than a `className`, for the reason `Button` has one:
       * the family had drifted into disagreeing with itself. `Badge`'s base is
       * a pill, `ToneBadge` overrode it to `rounded-md`, and
       * `ApplicationStatusBadge` overrode that back to `rounded-full` - so an
       * application's status was a different shape from every other status on
       * the platform, decided by which of three files last had its say.
       */
      /**
       * How the label is cased.
       *
       * Upper by default, which is the platform's convention: a badge is a
       * marker to be spotted rather than a phrase to be read, and the shouting
       * is what separates it from the text around it. `ToneBadge` had this and
       * ordinary badges did not, so a status was uppercase and the programme
       * chip beside it was not.
       *
       * `normal` is for the few that carry prose - a programme's full name, a
       * person's title - where upper-casing a sentence reads as a fault.
       */
      case: {
        upper: 'uppercase',
        normal: 'normal-case'
      },
      rounded: {
        none: 'rounded-none',
        sm: 'rounded-sm',
        md: 'rounded-md',
        lg: 'rounded-lg',
        full: 'rounded-full'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
      case: 'upper',
      rounded: 'full'
    }
  }
);
export type BadgeProps = React.HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>;

/**
 * Whether a class list paints a background of its own, unconditionally.
 *
 * Only an unmodified `bg-*` counts. A caller carrying nothing but `hover:bg-*`
 * or `dark:bg-*` is decorating the variant's colours rather than replacing
 * them, and should keep them.
 */
const PAINTS_OWN_BACKGROUND = /(^|\s)!?bg-/;

/**
 * A badge.
 *
 * A caller that paints its own background takes the whole palette, both themes,
 * and the variant contributes shape alone. Without this the variant's *dark*
 * background survives: tailwind-merge replaces `bg-fg` with the
 * caller's colour, but `` is a different key and stays, so a
 * green pill with white text turned into a white pill with white text the
 * moment the page was in dark mode. Every status badge on the platform is
 * written that way, so this is resolved here rather than in each of them.
 */
function Badge({ className, variant, size, case: casing, rounded, ...props }: BadgeProps) {
  const resolved = variant ?? (PAINTS_OWN_BACKGROUND.test(className ?? '') ? 'unstyled' : variant);
  // A span, so a badge may sit inside a sentence, a heading or a table cell's
  // line without breaking the markup around it.
  return (
    <span
      data-slot="badge"
      data-type-fixed=""
      className={cn(badgeVariants({ case: casing, rounded, size, variant: resolved }), className)}
      {...props}
    />
  );
}
export { Badge, badgeVariants };
