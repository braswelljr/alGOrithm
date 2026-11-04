'use client';

import { useEffect, useRef } from 'react';
import { animate } from 'motion/react';

/** A number that counts up to its new value, or jumps there with reduced motion. */
export function CountUp({ value, reduced }: { value: number; reduced: boolean }) {
  const el = useRef<HTMLSpanElement>(null);
  const shown = useRef(value);
  useEffect(() => {
    const node = el.current;
    if (!node) return;
    const from = shown.current;
    shown.current = value;
    if (reduced || from === value) {
      node.textContent = value.toLocaleString();
      return;
    }
    const controls = animate(from, value, {
      duration: 0.5,
      ease: [0.23, 1, 0.32, 1],
      onUpdate: (v) => {
        node.textContent = Math.round(v).toLocaleString();
      }
    });
    return () => controls.stop();
  }, [value, reduced]);
  return <span ref={el}>{value.toLocaleString()}</span>;
}
