import { useEffect, useRef, useState } from 'react';

/**
 * The current step in words, in a polite live region. While playing fast it
 * updates at most twice a second, so a screen reader is not flooded.
 */
export function StepDescription({ text, busy }: { text: string; busy: boolean }) {
  const [shown, setShown] = useState(text);
  const last = useRef(0);
  useEffect(() => {
    const now = Date.now();
    const wait = busy ? Math.max(0, 500 - (now - last.current)) : 0;
    const t = setTimeout(() => {
      last.current = Date.now();
      setShown(text);
    }, wait);
    return () => clearTimeout(t);
  }, [text, busy]);
  return (
    <p
      className="av-live"
      aria-live="polite"
      aria-atomic="true"
    >
      {shown}
    </p>
  );
}
