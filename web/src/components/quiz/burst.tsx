'use client';

import { useEffect, useRef } from 'react';

/** Particles in one burst, and how long it lasts. */
const COUNT = 36;
const LIFE_MS = 900;

/**
 * A short burst of particles from the middle of its box, drawn on a canvas,
 * each time `fire` changes (a streak milestone). Decoration only: hidden from
 * assistive technology, never blocking a click, and skipped with reduced
 * motion.
 */
export function Burst({ fire, reduced }: { fire: number; reduced: boolean }) {
  const canvas = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const el = canvas.current;
    if (!fire || reduced || !el) return;
    const ctx = el.getContext('2d');
    if (!ctx) return;
    const dpr = window.devicePixelRatio || 1;
    const { width, height } = el.getBoundingClientRect();
    el.width = width * dpr;
    el.height = height * dpr;
    ctx.scale(dpr, dpr);
    const css = getComputedStyle(el);
    const colors = ['--accent', '--success', '--fg'].map(
      (v) => css.getPropertyValue(v).trim() || 'currentColor'
    );
    const parts = Array.from({ length: COUNT }, (_, i) => {
      const angle = (i / COUNT) * Math.PI * 2 + Math.random() * 0.3;
      const speed = 2.5 + Math.random() * 3.5;
      return {
        x: width / 2,
        y: height / 2,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed - 1.5,
        size: 2 + Math.random() * 3,
        color: colors[i % colors.length]
      };
    });
    const start = performance.now();
    let frame = 0;
    const tick = (now: number) => {
      const t = (now - start) / LIFE_MS;
      ctx.clearRect(0, 0, width, height);
      if (t >= 1) return;
      ctx.globalAlpha = 1 - t;
      for (const p of parts) {
        p.x += p.vx;
        p.y += p.vy;
        p.vy += 0.12; // gravity
        p.vx *= 0.985;
        ctx.fillStyle = p.color;
        ctx.fillRect(p.x, p.y, p.size, p.size);
      }
      frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => {
      cancelAnimationFrame(frame);
      ctx.clearRect(0, 0, width, height);
    };
  }, [fire, reduced]);
  return (
    <canvas
      ref={canvas}
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 size-full"
    />
  );
}
