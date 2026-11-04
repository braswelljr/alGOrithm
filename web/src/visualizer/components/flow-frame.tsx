'use client';

/**
 * A React Flow view with a full-screen toggle. The browser's Fullscreen API
 * where there is one; elsewhere (iPhone Safari) the view covers the window
 * instead. Escape leaves either. The view is drawn again on the way in and
 * out, so React Flow fits the graph to the new size.
 */
import { useEffect, useRef, useState, type ReactNode } from 'react';
import { LuMaximize2, LuMinimize2 } from 'react-icons/lu';
import { cn } from '@/lib/utils';

export function FlowFrame({ children }: { children: ReactNode }) {
  const host = useRef<HTMLDivElement>(null);
  const [full, setFull] = useState(false);

  useEffect(() => {
    const sync = () => {
      if (!document.fullscreenElement && full && host.current?.dataset.native === '1') {
        delete host.current.dataset.native;
        setFull(false);
      }
    };
    const esc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && full && !document.fullscreenElement) setFull(false);
    };
    document.addEventListener('fullscreenchange', sync);
    document.addEventListener('keydown', esc);
    return () => {
      document.removeEventListener('fullscreenchange', sync);
      document.removeEventListener('keydown', esc);
    };
  }, [full]);

  const toggle = async () => {
    const el = host.current;
    if (!el) return;
    if (full) {
      if (document.fullscreenElement) await document.exitFullscreen().catch(() => {});
      delete el.dataset.native;
      setFull(false);
      return;
    }
    if (el.requestFullscreen) {
      try {
        await el.requestFullscreen();
        el.dataset.native = '1';
      } catch {
        // Refused (an iframe without permission): cover the window instead.
      }
    }
    setFull(true);
  };

  return (
    <div
      ref={host}
      data-full={full || undefined}
      className={cn(
        'group/flow relative w-full',
        full && 'bg-bg fixed inset-0 z-50 flex flex-col p-4 sm:p-6'
      )}
    >
      <button
        type="button"
        onClick={() => void toggle()}
        aria-pressed={full}
        aria-label={full ? 'Leave full screen' : 'Show the view full screen'}
        title={full ? 'Leave full screen (Esc)' : 'Full screen'}
        className="bg-card text-fg hover:bg-subtle focus-visible:ring-accent absolute end-2 top-2 z-10 grid size-8 place-items-center rounded-md shadow-sm outline-none group-data-full/flow:end-6 group-data-full/flow:top-6 focus-visible:ring-2"
      >
        {full ? <LuMinimize2 aria-hidden /> : <LuMaximize2 aria-hidden />}
      </button>
      {/* In full screen every flow box takes the screen's height. */}
      <div
        key={full ? 'full' : 'inline'}
        className={cn(
          'w-full',
          full &&
            'flex min-h-0 flex-1 flex-col *:min-h-0 *:flex-1 [&_[data-flow-box]]:h-auto! [&_[data-flow-box]]:min-h-0 [&_[data-flow-box]]:flex-1 [&>figure]:flex [&>figure]:flex-col'
        )}
      >
        {children}
      </div>
    </div>
  );
}
