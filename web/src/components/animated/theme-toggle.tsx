'use client';

// From beui.dev/components/motion/theme-toggle (MIT), by way of
// gsl-special-courses.
import { useCallback, useState, type ComponentPropsWithoutRef } from 'react';
import { Moon, Sun } from 'lucide-react';
import { useReducedMotionConfig } from 'motion/react';
import { useTheme } from 'next-themes';
import { ActionSwapIcon } from '@/components/animated/action-swap';
import { Button } from '@/components/ui/button';
import { useIsomorphicLayoutEffect } from '@/hooks/use-isomorphic-layout-effect';
import { EASE_OUT_CSS } from '@/lib/ease';
import { cn } from '@/lib/utils';

export type ThemeVariant = 'rectangle' | 'circle' | 'circle-blur' | 'blinds' | 'diagonal';

export type RectStart =
  'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'center' | 'bottom-up';

export interface ThemeToggleProps extends Omit<
  ComponentPropsWithoutRef<typeof Button>,
  'children' | 'onClick' | 'variant' | 'size'
> {
  /** Animation variant. Default: "rectangle". */
  variant?: ThemeVariant;
  /** Origin direction for the reveal. Default: "bottom-up". */
  start?: RectStart;
  iconClassName?: string;
}

const VT_STYLE_ID = 'beui-theme-toggle-vt';

// View transitions animate in CSS, not motion springs, so easing here is
// either EASE_OUT_CSS or a keyword. The circle variants keep the Material
// standard curve because their reveal expands symmetrically rather than
// decelerating. Durations differ per variant to match native OS mode switches.
const VT_CSS = `
html[data-beui-vt="rect"]::view-transition-old(root) {
  animation: none;
  mix-blend-mode: normal;
}
html[data-beui-vt="rect"]::view-transition-new(root) {
  mix-blend-mode: normal;
  animation: beui-rect-reveal 400ms ease-out;
}
html[data-beui-vt="circle"]::view-transition-old(root),
html[data-beui-vt="circle-blur"]::view-transition-old(root) {
  animation: none;
  mix-blend-mode: normal;
}
html[data-beui-vt="circle"]::view-transition-new(root) {
  mix-blend-mode: normal;
  animation: beui-circle-reveal 700ms cubic-bezier(0.4, 0, 0.2, 1);
}
html[data-beui-vt="circle-blur"]::view-transition-new(root) {
  mix-blend-mode: normal;
  animation: beui-circle-blur-reveal 700ms cubic-bezier(0.4, 0, 0.2, 1);
}
html[data-beui-vt="blinds"]::view-transition-old(root) {
  animation: none;
  mix-blend-mode: normal;
}
/* Slats: a masked band widens inside every 72px tile, so the new theme opens
   across the page like a shutter. The band edge has to be a registered custom
   property - mask-image itself is not animatable, but it re-resolves every
   frame the property ticks. mask-size fixes the tile at 72px rather than
   letting a repeating gradient's last stop define it, which is what keeps the
   20px soft edge from dragging the tile wider than the slat and leaving a
   feathered gap that never closes; it also means both ends land clean, fully
   transparent at -20px and fully opaque at 72px. Falling back to no mask
   (unregistered property, so the var is invalid) reveals the page in one
   step. */
@property --beui-vt-slat {
  syntax: "<length>";
  inherits: false;
  initial-value: 72px;
}
html[data-beui-vt="blinds"]::view-transition-new(root) {
  mix-blend-mode: normal;
  mask-image: linear-gradient(
    90deg,
    #000 0 var(--beui-vt-slat),
    transparent calc(var(--beui-vt-slat) + 20px)
  );
  mask-size: 72px 100%;
  mask-repeat: repeat;
  animation: beui-blinds-reveal 700ms ${EASE_OUT_CSS};
}
/* Diagonal: the new theme sweeps in from the top-left corner along the line the
   Settings previews split on, so choosing one plays out the picture of it. */
html[data-beui-vt="diagonal"]::view-transition-old(root) {
  animation: none;
  mix-blend-mode: normal;
}
html[data-beui-vt="diagonal"]::view-transition-new(root) {
  mix-blend-mode: normal;
  animation: beui-diagonal-reveal 650ms ${EASE_OUT_CSS};
}
@keyframes beui-diagonal-reveal {
  from { clip-path: polygon(0 0, 0 0, 0 0); }
  to   { clip-path: polygon(0 0, 200% 0, 0 200%); }
}
@keyframes beui-rect-reveal {
  from { clip-path: var(--beui-vt-from, inset(100% 0 0 0)); }
  to   { clip-path: inset(0 0 0 0); }
}
@keyframes beui-circle-reveal {
  from { clip-path: circle(0% at var(--beui-vt-origin, 50% 100%)); }
  to   { clip-path: circle(150% at var(--beui-vt-origin, 50% 100%)); }
}
@keyframes beui-circle-blur-reveal {
  from { clip-path: circle(0% at var(--beui-vt-origin, 50% 100%)); filter: blur(8px); }
  to   { clip-path: circle(150% at var(--beui-vt-origin, 50% 100%)); filter: blur(0px); }
}
@keyframes beui-blinds-reveal {
  from { --beui-vt-slat: -20px; }
  to   { --beui-vt-slat: 72px; }
}
`;

const RECT_FROM: Record<RectStart, string> = {
  'top-left': 'inset(0 100% 100% 0)',
  'top-right': 'inset(0 0 100% 100%)',
  'bottom-left': 'inset(100% 100% 0 0)',
  'bottom-right': 'inset(100% 0 0 100%)',
  center: 'inset(50% 50% 50% 50%)',
  'bottom-up': 'inset(100% 0 0 0)'
};

const CIRCLE_ORIGIN: Record<RectStart, string> = {
  'top-left': '0% 0%',
  'top-right': '100% 0%',
  'bottom-left': '0% 100%',
  'bottom-right': '100% 100%',
  center: '50% 50%',
  'bottom-up': '50% 100%'
};

export function useThemeToggle({
  variant = 'rectangle',
  start = 'bottom-up'
}: { variant?: ThemeVariant; start?: RectStart } = {}) {
  const { setTheme, resolvedTheme } = useTheme();
  // The config rather than the device alone, so the reduced motion a person
  // chose in Settings stops this the same way the operating system's does.
  const reduce = useReducedMotionConfig() ?? false;
  const [mounted, setMounted] = useState(false);
  // Both before paint: the icon is drawn for the resolved theme from the first
  // frame rather than swapping in a frame later, and the transition's
  // stylesheet is in the document before anybody can press the toggle.
  useIsomorphicLayoutEffect(() => setMounted(true), []);
  useIsomorphicLayoutEffect(() => {
    if (document.getElementById(VT_STYLE_ID)) return;
    const el = document.createElement('style');
    el.id = VT_STYLE_ID;
    el.textContent = VT_CSS;
    document.head.appendChild(el);
  }, []);
  const isDark = mounted && resolvedTheme === 'dark';

  // Stable across renders, so a list of rows built from it (the command
  // palette's) is rebuilt only when what it shows changes.
  const switchTo = useCallback(
    (next: 'light' | 'dark' | 'system') => {
      if (reduce || !('startViewTransition' in document)) {
        setTheme(next);
        return;
      }

      const root = document.documentElement;

      if (variant === 'rectangle') {
        root.style.setProperty('--beui-vt-from', RECT_FROM[start]);
        root.dataset.beuiVt = 'rect';
      } else if (variant === 'blinds' || variant === 'diagonal') {
        // Both sweep the whole viewport; there is no origin point to set.
        root.dataset.beuiVt = variant;
      } else {
        root.style.setProperty('--beui-vt-origin', CIRCLE_ORIGIN[start]);
        root.dataset.beuiVt = variant;
      }

      const vt = (
        document as Document & {
          startViewTransition(cb: () => void): { finished: Promise<void> };
        }
      ).startViewTransition(() => setTheme(next));

      vt.finished.finally(() => {
        delete root.dataset.beuiVt;
      });
    },
    [reduce, variant, start, setTheme]
  );

  const toggle = () => switchTo(isDark ? 'light' : 'dark');

  return { isDark, mounted, toggle, switchTo };
}

export function ThemeToggle({
  variant = 'rectangle',
  start = 'bottom-up',
  className,
  iconClassName,
  ...rest
}: ThemeToggleProps) {
  const { isDark, mounted, toggle } = useThemeToggle({ variant, start });

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      aria-label={mounted && isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      onClick={toggle}
      className={cn('flex items-center justify-center', className)}
      {...rest}
    >
      {mounted ? (
        <ActionSwapIcon
          value={isDark ? 'dark' : 'light'}
          animation="blur"
          className={iconClassName}
        >
          {isDark ? <Sun className={iconClassName} /> : <Moon className={iconClassName} />}
        </ActionSwapIcon>
      ) : (
        <span
          className={iconClassName}
          aria-hidden="true"
        />
      )}
    </Button>
  );
}
