'use client';

import type { ReactNode } from 'react';
import { motion, useReducedMotionConfig, type Variants } from 'motion/react';
import { cn } from '@/lib/utils';
import type {
  Contrast,
  Density,
  Focus,
  MotionPreference,
  TextSize,
  TextSpacing,
  Theme
} from '@/store/preferences';

// Pictures of what each display preference does, drawn inside the card that
// offers it.
//
// Where the stylesheet can show the real thing, a preview is a small piece of
// the dashboard scoped to that one option with the same attributes the
// dashboard root carries: the density preview is laid out at that density, the
// text-size preview set at that size. The stylesheet states every standard value
// as well as every changed one, so a preview of the standard keeps looking
// standard while the page around it is set otherwise. Where it cannot (a theme
// is a class on the document, and motion does not hold still), the preview is a
// drawing in fixed colours.

/** A piece of the dashboard laid out under one set of preferences. */
function Scope({
  children,
  className,
  ...attributes
}: {
  children: ReactNode;
  className?: string;
  'data-density'?: Density;
  'data-text-size'?: TextSize;
  'data-contrast'?: Contrast;
  'data-text-spacing'?: TextSpacing;
}) {
  return (
    <div
      data-preferences=""
      data-density="standard"
      data-text-size="standard"
      data-contrast="standard"
      data-text-spacing="standard"
      {...attributes}
      className={cn('absolute inset-0', className)}
    >
      {children}
    </div>
  );
}

// ── Theme ───────────────────────────────────────────────────────────────────

/**
 * A miniature dashboard in one theme: the sidebar, the top bar, and the cards
 * and rows of a page, as the shapes a skeleton draws.
 *
 * Fixed colours on purpose, with no `dark:` step: the picture of the light
 * theme has to stay light while the dashboard around it is dark, or the choice
 * would show what is already on the screen instead of what it changes to.
 */
function SkeletonDashboard({ tone, build }: { tone: 'light' | 'dark'; build: boolean }) {
  const light = tone === 'light';
  const ground = light ? 'bg-stone-100' : 'bg-stone-950';
  const card = light ? 'bg-white border-stone-200' : 'bg-stone-900 border-stone-800';
  const line = light ? 'bg-stone-200' : 'bg-stone-700';
  const strong = light ? 'bg-stone-300' : 'bg-stone-600';
  // Built piece by piece when chosen, as a page loads: the sidebar, the
  // heading, the cards, the rows. Shown whole otherwise.
  const piece = build ? { variants: PIECE } : {};
  return (
    <motion.div
      className={cn('absolute inset-0 flex', ground)}
      initial={build ? 'hidden' : false}
      animate="shown"
      variants={build ? BUILD : undefined}
    >
      {/* The sidebar, navy in both themes, as the real one is. */}
      <div className="bg-secondary flex w-1/4 flex-col gap-1.5 p-2">
        <motion.div
          {...piece}
          className="mb-1 size-3 rounded-sm bg-white/70"
        />
        {['w-4/5 bg-white/40', 'w-3/5 bg-white/25', 'w-4/5 bg-white/25', 'w-2/3 bg-white/25'].map(
          (shape) => (
            <motion.div
              key={shape}
              {...piece}
              className={cn('h-1.5 rounded-full', shape)}
            />
          )
        )}
      </div>
      <div className="flex flex-1 flex-col gap-2 p-2">
        <motion.div
          {...piece}
          className={cn('h-2 w-1/2 rounded-full', strong)}
        />
        <div className="grid grid-cols-3 gap-1.5">
          {[0, 1, 2].map((index) => (
            <motion.div
              key={index}
              {...piece}
              className={cn('flex h-7 flex-col justify-center gap-1 rounded border px-1.5', card)}
            >
              <div className={cn('h-1 w-2/3 rounded-full', line)} />
              <div className={cn('h-1.5 w-1/3 rounded-full', strong)} />
            </motion.div>
          ))}
        </div>
        <motion.div
          {...piece}
          className={cn('flex flex-1 flex-col gap-1.5 rounded border p-1.5', card)}
        >
          {[0, 1, 2, 3].map((index) => (
            <div
              key={index}
              className={cn('h-1 rounded-full', line, index % 2 === 0 ? 'w-11/12' : 'w-3/4')}
            />
          ))}
        </motion.div>
      </div>
    </motion.div>
  );
}

/** The skeleton's pieces arriving in turn, quickly enough to read as one act. */
const BUILD: Variants = { hidden: {}, shown: { transition: { staggerChildren: 0.035 } } };
const PIECE: Variants = {
  hidden: { opacity: 0, y: 3 },
  shown: { opacity: 1, y: 0, transition: { duration: 0.25, ease: 'easeOut' } }
};

/**
 * The dark half of following the device. Its edge sways slowly about the
 * diagonal, as the light does over a day; held on the diagonal when motion is
 * reduced. Five points with the far ones well outside the box, so the polygon
 * stays one shape however far the edge leans.
 */
const SPLIT_AT = (top: number, left: number) =>
  `polygon(${top}% 0, 300% 0, 300% 300%, 0 300%, 0 ${left}%)`;

/**
 * Each theme. Following the device is the two halves of the same dashboard,
 * split on the diagonal: it is light or dark depending on where it is opened.
 */
export function ThemePreview({ theme, selected }: { theme: Theme; selected: boolean }) {
  const reduce = useReducedMotionConfig() ?? false;
  const build = selected && !reduce;
  return (
    // Keyed on the choice, so choosing it again plays the build again.
    <div
      key={selected ? 'chosen' : 'offered'}
      className="absolute inset-0"
    >
      <SkeletonDashboard
        tone={theme === 'dark' ? 'dark' : 'light'}
        build={build}
      />
      {theme === 'system' ? (
        <motion.div
          className="absolute inset-0"
          initial={{ clipPath: SPLIT_AT(100, 100) }}
          animate={
            reduce
              ? { clipPath: SPLIT_AT(100, 100) }
              : {
                  clipPath: [
                    SPLIT_AT(100, 100),
                    SPLIT_AT(118, 82),
                    SPLIT_AT(82, 118),
                    SPLIT_AT(100, 100)
                  ]
                }
          }
          transition={
            reduce ? { duration: 0 } : { duration: 9, ease: 'easeInOut', repeat: Infinity }
          }
        >
          <SkeletonDashboard
            tone="dark"
            build={build}
          />
        </motion.div>
      ) : null}
    </div>
  );
}

// ── Density ─────────────────────────────────────────────────────────────────

/** A register's rows, laid out at the density offered. */
export function DensityPreview({ density }: { density: Density }) {
  return (
    <Scope
      data-density={density}
      className="flex flex-col p-2"
    >
      <div className="flex flex-col overflow-hidden rounded-md border border-stone-200 dark:border-stone-800">
        <div className="flex items-center gap-2 border-b border-stone-200 bg-stone-50 px-2 py-1.5 dark:border-stone-800 dark:bg-stone-950">
          <div className="h-1.5 w-1/4 rounded-full bg-stone-300 dark:bg-stone-700" />
          <div className="ms-auto h-1.5 w-1/6 rounded-full bg-stone-300 dark:bg-stone-700" />
        </div>
        {[0, 1, 2, 3, 4, 5].map((index) => (
          <div
            key={index}
            className="flex items-center gap-2 border-b border-stone-100 px-2 py-1.5 last:border-b-0 dark:border-stone-900"
          >
            <div className="size-3 shrink-0 rounded-full bg-stone-200 dark:bg-stone-800" />
            <div
              className={cn(
                'h-1.5 rounded-full bg-stone-200 dark:bg-stone-800',
                index % 2 === 0 ? 'w-1/2' : 'w-2/5'
              )}
            />
            <div className="bg-primary/60 ms-auto h-2 w-6 rounded-full" />
          </div>
        ))}
      </div>
    </Scope>
  );
}

// ── Text ────────────────────────────────────────────────────────────────────

/** A heading and a line of a page, set at the size offered. */
export function TextSizePreview({ size }: { size: TextSize }) {
  return (
    <Scope
      data-text-size={size}
      className="flex flex-col gap-1 p-3"
    >
      <span className="text-lg font-semibold tracking-tight text-stone-900 dark:text-stone-100">
        Applications
      </span>
      <span className="text-sm text-stone-600 dark:text-stone-400">
        Twelve cases are waiting for an officer to take them.
      </span>
    </Scope>
  );
}

/** A short paragraph, at the spacing offered. */
export function TextSpacingPreview({ spacing }: { spacing: TextSpacing }) {
  return (
    <Scope
      data-text-spacing={spacing}
      className="p-3"
    >
      <p className="text-sm text-stone-700 dark:text-stone-300">
        The panel visited the library and the moot court, and met the staff who teach the first
        year.
      </p>
    </Scope>
  );
}

// ── Contrast ────────────────────────────────────────────────────────────────

/** A card with a title, the muted line under it and its hairlines. */
export function ContrastPreview({ contrast }: { contrast: Contrast }) {
  return (
    <Scope
      data-contrast={contrast}
      className="flex flex-col gap-2 p-3"
    >
      <div className="flex flex-col gap-1 rounded-md border border-stone-200 p-2 dark:border-stone-800">
        <span className="text-sm font-medium text-stone-900 dark:text-stone-100">Site visit</span>
        <span className="text-xs text-stone-600 dark:text-stone-400">
          Scheduled, and published to the panel.
        </span>
      </div>
      <div className="h-px bg-stone-200 dark:bg-stone-800" />
      <div className="flex gap-2">
        <div className="h-1.5 w-1/3 rounded-full bg-stone-300 dark:bg-stone-700" />
        <div className="h-1.5 w-1/4 rounded-full bg-stone-300 dark:bg-stone-700" />
      </div>
    </Scope>
  );
}

// ── Focus ───────────────────────────────────────────────────────────────────

/**
 * Two fields and a button, with the keyboard on the button. Drawn rather than
 * focused, since a card that took the focus to show it would move the reader's
 * own place on the page.
 */
export function FocusPreview({ focus }: { focus: Focus }) {
  return (
    <div className="absolute inset-0 flex flex-col justify-center gap-2 p-3">
      <div className="h-5 rounded-md border border-stone-200 dark:border-stone-800" />
      <div className="h-5 rounded-md border border-stone-200 dark:border-stone-800" />
      <div className="flex justify-end pt-1">
        <span
          className={cn(
            'bg-primary rounded-md px-3 py-1 text-xs font-medium text-stone-950',
            focus === 'strong'
              ? 'shadow-[0_0_0_2px_var(--color-white),0_0_0_5px_var(--color-stone-950)] dark:shadow-[0_0_0_2px_var(--color-stone-950),0_0_0_5px_var(--color-white)]'
              : 'ring-ring ring-2 ring-offset-2 ring-offset-white dark:ring-offset-stone-950'
          )}
        >
          Continue
        </span>
      </div>
    </div>
  );
}

// ── Motion ──────────────────────────────────────────────────────────────────

/** A panel arriving: with its path drawn behind it, or simply there. */
function MotionDrawing({ moving }: { moving: boolean }) {
  const reduce = useReducedMotionConfig() ?? false;
  return (
    <div className="absolute inset-0 flex items-center justify-center bg-white dark:bg-stone-950">
      <div className="relative h-3/5 w-1/2">
        {moving
          ? [0.15, 0.3].map((opacity, index) => (
              <div
                key={opacity}
                style={{ opacity, transform: `translateX(-${(2 - index) * 22}%)` }}
                className="absolute inset-0 rounded-md border border-stone-300 bg-stone-100 dark:border-stone-700 dark:bg-stone-900"
              />
            ))
          : null}
        {/* With motion, the panel actually arrives, every few seconds. */}
        <motion.div
          initial={false}
          animate={moving && !reduce ? { x: ['-45%', '0%', '0%'], opacity: [0, 1, 1] } : { x: 0 }}
          transition={
            moving && !reduce
              ? { duration: 3, times: [0, 0.18, 1], ease: 'easeOut', repeat: Infinity }
              : { duration: 0 }
          }
          className="absolute inset-0 flex flex-col gap-1.5 rounded-md border border-stone-300 bg-white p-2 dark:border-stone-700 dark:bg-stone-900"
        >
          <div className="h-1.5 w-2/3 rounded-full bg-stone-300 dark:bg-stone-700" />
          <div className="h-1 w-1/2 rounded-full bg-stone-200 dark:bg-stone-800" />
        </motion.div>
      </div>
    </div>
  );
}

/** Following the device is both, split on the diagonal, as the theme is. */
export function MotionPreview({ motion }: { motion: MotionPreference }) {
  return (
    <>
      <MotionDrawing moving={motion !== 'reduce'} />
      {motion === 'system' ? (
        <div className="absolute inset-0 [clip-path:polygon(100%_0,100%_100%,0_100%)]">
          <MotionDrawing moving={false} />
        </div>
      ) : null}
    </>
  );
}
