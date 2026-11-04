'use client';

/**
 * The arena's frame: a screen on the site's own surface that holds the
 * select screen, a game's menu and the game itself. Everything inside is
 * driven by the keyboard as a game is (arrows to choose, Enter to take,
 * Escape to go back), and by pointer and touch as well. Its backdrop is
 * WebGL (ArenaBackdrop); the games draw on canvas and run the wasm.
 */
import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import { MotionConfig } from 'motion/react';
import { LuMaximize2, LuMinimize2 } from 'react-icons/lu';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { ArenaBackdrop } from './arena-backdrop';

/** Whether the arena is full screen: the boards take bigger cells there. */
const FullContext = createContext(false);
export const useArenaFull = () => useContext(FullContext);

/** The largest cell a board draws, inline and full screen, in pixels. */
export const cellCap = (full: boolean) => (full ? 40 : 24);

export function ArenaScreen({
  className,
  children,
  label
}: {
  className?: string;
  children: ReactNode;
  /** What the screen holds now, for a screen reader. */
  label: string;
}) {
  const host = useRef<HTMLElement>(null);
  const [full, setFull] = useState(false);

  // Leaving the browser's full screen (Escape, F11) leaves ours.
  useEffect(() => {
    const sync = () => {
      if (!document.fullscreenElement) setFull(false);
    };
    document.addEventListener('fullscreenchange', sync);
    return () => document.removeEventListener('fullscreenchange', sync);
  }, []);

  const toggle = async () => {
    if (full) {
      if (document.fullscreenElement) await document.exitFullscreen().catch(() => {});
      setFull(false);
      return;
    }
    // The browser's full screen where it is allowed; else the arena covers the window.
    await host.current?.requestFullscreen?.().catch(() => {});
    setFull(true);
  };

  return (
    <section
      ref={host}
      aria-label={label}
      data-arena=""
      className={cn(
        // The site's own surface and tokens, so the theme and every display
        // preference (text size, contrast, spacing, bold, focus) reach the
        // games; only reduced motion is set aside (data-arena), since a game
        // that does not move is not one.
        'bg-card text-fg relative isolate flex w-full flex-1 flex-col overflow-hidden',
        full
          ? 'fixed inset-0 z-50 h-dvh rounded-none'
          : // card-ok: the arena's console, a game screen rather than a content card.
            'border-border min-h-[max(30rem,calc(100svh-var(--header-height,3rem)-3rem))] rounded-2xl border',
        className
      )}
    >
      <ArenaBackdrop />
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={() => void toggle()}
        aria-pressed={full}
        aria-label={full ? 'Leave full screen' : 'Play full screen'}
        title={full ? 'Leave full screen (Esc)' : 'Full screen'}
        className="absolute end-3 top-3 z-30"
      >
        {full ? <LuMinimize2 aria-hidden /> : <LuMaximize2 aria-hidden />}
      </Button>
      <MotionConfig reducedMotion="never">
        <FullContext.Provider value={full}>{children}</FullContext.Provider>
      </MotionConfig>
    </section>
  );
}

/** The key hints along a screen's foot: "← → choose · Enter select". */
export function KeyHints({ hints }: { hints: [string, string][] }) {
  return (
    <p className="text-muted flex flex-wrap items-center gap-x-4 gap-y-1 text-xs">
      {hints.map(([keys, what]) => (
        <span
          key={keys}
          className="flex items-center gap-1.5"
        >
          <kbd className="keycap keycap-mod rounded-md px-1.5 py-0.5 font-mono text-[11px] [--cap-depth:2px]">
            {keys}
          </kbd>
          {what}
        </span>
      ))}
    </p>
  );
}
