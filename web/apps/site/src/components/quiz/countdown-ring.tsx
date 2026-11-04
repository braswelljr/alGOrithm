'use client';

import { useEffect, useRef } from 'react';
import { useGSAP } from '@gsap/react';
import gsap from 'gsap';

gsap.registerPlugin(useGSAP);

const R = 22;
const C = 2 * Math.PI * R;

/**
 * A ring that empties as time runs out, driven by one GSAP tween whose end is
 * the timer itself, so what is drawn and what counts can never disagree. With
 * reduced motion it still counts, but the ring moves in whole seconds instead
 * of sweeping. A new resetKey starts it again from full.
 */
export function CountdownRing({
  durationMs,
  running,
  resetKey,
  reduced,
  label,
  onDone
}: {
  durationMs: number;
  running: boolean;
  resetKey: string | number;
  reduced: boolean;
  /** What the timer is for: "Time for this question". */
  label: string;
  onDone(): void;
}) {
  const ring = useRef<SVGCircleElement>(null);
  const text = useRef<HTMLSpanElement>(null);
  const root = useRef<HTMLDivElement>(null);
  const tween = useRef<gsap.core.Tween | null>(null);
  const done = useRef(onDone);
  useEffect(() => {
    done.current = onDone;
  }, [onDone]);

  useGSAP(
    () => {
      const clock = { left: durationMs };
      const draw = () => {
        const shown = reduced ? Math.ceil(clock.left / 1000) * 1000 : clock.left;
        ring.current?.style.setProperty('stroke-dashoffset', String(C * (1 - shown / durationMs)));
        if (text.current) text.current.textContent = String(Math.ceil(clock.left / 1000));
        root.current?.toggleAttribute('data-low', clock.left <= durationMs * 0.25);
      };
      draw();
      tween.current = gsap.to(clock, {
        left: 0,
        duration: durationMs / 1000,
        ease: 'none',
        paused: true,
        onUpdate: draw,
        onComplete: () => done.current()
      });
    },
    { dependencies: [resetKey, durationMs, reduced], revertOnUpdate: true }
  );

  useEffect(() => {
    if (running) tween.current?.play();
    else tween.current?.pause();
  }, [running, resetKey, reduced]);

  return (
    <div
      ref={root}
      role="timer"
      aria-label={label}
      className="group relative grid size-13 shrink-0 place-items-center font-semibold tabular-nums"
    >
      <svg
        viewBox="0 0 52 52"
        aria-hidden="true"
        className="absolute inset-0 -rotate-90"
      >
        <circle
          className="fill-none stroke-subtle stroke-4"
          cx="26"
          cy="26"
          r={R}
        />
        <circle
          ref={ring}
          className="fill-none stroke-accent stroke-4 [stroke-linecap:round] group-data-low:stroke-danger"
          cx="26"
          cy="26"
          r={R}
          strokeDasharray={C}
        />
      </svg>
      <span>
        <span ref={text} />
        <span className="sr-only"> seconds left</span>
      </span>
    </div>
  );
}
