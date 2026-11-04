'use client';

import { useEffect, useState, type RefObject } from 'react';

/**
 * Whether the element is on screen, so decorative animation can stop while
 * nobody can see it. It starts false, so nothing animates before the first
 * observation says the element is on screen.
 */
export function useInView(ref: RefObject<Element | null>, rootMargin = '0px'): boolean {
  const [inView, setInView] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(([entry]) => setInView(entry.isIntersecting), {
      rootMargin
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [ref, rootMargin]);

  return inView;
}
