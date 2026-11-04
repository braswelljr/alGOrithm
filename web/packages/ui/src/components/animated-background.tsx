'use client';

import { useRef } from 'react';
import { useInView } from '@braswelljr/algorithm-ui/hooks/use-in-view';
import { cn } from '@braswelljr/algorithm-ui/lib/utils';

type AnimatedBackgroundProps = {
  /** `hero` is the full aurora; `subtle` is a quieter wash for inner pages and sections. */
  variant?: 'hero' | 'subtle';
  className?: string;
};

/** One keycap outline on a 48px pitch: the tile the backdrop's grid repeats. */
const KEY_TILE = `url("data:image/svg+xml,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><rect x="3.5" y="3.5" width="41" height="41" rx="8" fill="none" stroke="black" stroke-width="1"/></svg>'
)}")`;

/**
 * The ambient layer behind a section: slow aurora glows over a faint grid of
 * keycap outlines that fades toward the edges, as if the page were a plate
 * with the keys lifted off. Fill the positioned parent with it. It is
 * decorative, pauses off screen, and holds still under reduced motion.
 */
export function AnimatedBackground({ variant = 'hero', className }: AnimatedBackgroundProps) {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref);
  const hero = variant === 'hero';

  return (
    <div
      ref={ref}
      aria-hidden="true"
      data-paused={!inView}
      className={cn('pointer-events-none absolute inset-0 -z-10 overflow-hidden', className)}
    >
      <div
        className={cn(
          'absolute -top-1/3 left-1/2 aspect-square w-[60rem] max-w-[160%] -translate-x-1/2',
          hero ? 'opacity-45 dark:opacity-35' : 'opacity-25 dark:opacity-20'
        )}
      >
        <div className="aurora-glow-cyan absolute top-[4%] left-[2%] size-3/5 animate-aurora opacity-80" />
        <div className="aurora-glow-violet absolute top-[14%] right-[0%] size-3/5 animate-aurora [animation-delay:-7s]" />
        {hero && (
          <div className="aurora-glow-amber absolute top-[34%] left-[26%] size-1/2 animate-aurora opacity-70 [animation-delay:-14s]" />
        )}
      </div>
      <div
        className={cn(
          'absolute inset-0 bg-fd-foreground',
          hero ? 'opacity-[0.07] dark:opacity-[0.09]' : 'opacity-[0.05]'
        )}
        style={{
          maskImage: `${KEY_TILE}, radial-gradient(ellipse 70% 60% at 50% 30%, black, transparent)`,
          maskComposite: 'intersect',
          WebkitMaskImage: `${KEY_TILE}, radial-gradient(ellipse 70% 60% at 50% 30%, black, transparent)`,
          WebkitMaskComposite: 'source-in'
        }}
      />
    </div>
  );
}

/**
 * An aurora fixed to the viewport, behind every section: the hero's glows
 * spread across the screen at about half their strength, so the whole page
 * sits in the same light as its top. The three glows orbit the middle of the
 * screen, a turn every 40 seconds, each still drifting on its own. It is
 * decorative, ignores the pointer and holds still under reduced motion.
 */
export function ViewportGlow({ className }: { className?: string }) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        'pointer-events-none fixed inset-0 -z-10 overflow-hidden opacity-45 dark:opacity-35',
        className
      )}
    >
      <div className="absolute top-1/2 left-1/2 size-[110vmin] -translate-x-1/2 -translate-y-1/2 animate-orbit">
        {/* The orbit is sized to the screen's short side, so the glows stay in view as they
            turn; three of them, a third of a turn apart, sweep round the middle. */}
        <div className="aurora-glow-cyan absolute top-[-6%] left-[22%] aspect-square w-[56%] animate-aurora [animation-duration:38s]" />
        <div className="aurora-glow-violet absolute right-[-8%] bottom-[2%] aspect-square w-[56%] animate-aurora [animation-delay:-12s] [animation-duration:44s]" />
        <div className="aurora-glow-amber absolute bottom-[2%] left-[-8%] aspect-square w-[56%] animate-aurora opacity-80 [animation-delay:-24s] [animation-duration:50s]" />
      </div>
    </div>
  );
}
