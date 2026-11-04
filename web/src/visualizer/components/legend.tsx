import type { ReactNode } from 'react';

/** The symbols the array views use, and a key to them. */

export type MarkName =
  'compare' | 'swap' | 'write' | 'sorted' | 'pivot' | 'min' | 'range' | 'found' | 'visit';

const labels: Record<MarkName, string> = {
  compare: 'compared',
  swap: 'swapped',
  write: 'written',
  sorted: 'in final place',
  pivot: 'pivot',
  min: 'smallest so far',
  range: 'still searched',
  found: 'found',
  visit: 'read'
};

/** Patterns the symbols refer to. Render once inside each SVG that uses them. */
export function Patterns() {
  return (
    <defs>
      <pattern
        id="av-hatch-swap"
        width="4"
        height="4"
        patternUnits="userSpaceOnUse"
        patternTransform="rotate(45)"
      >
        <rect
          width="4"
          height="4"
          className="av-hatch-a"
        />
        <rect
          width="2"
          height="4"
          className="av-fill-swap"
        />
      </pattern>
      <pattern
        id="av-dots-visit"
        width="4"
        height="4"
        patternUnits="userSpaceOnUse"
      >
        <rect
          width="4"
          height="4"
          className="av-dots-a"
        />
        <circle
          cx="2"
          cy="2"
          r="1"
          className="av-fill-compare"
        />
      </pattern>
    </defs>
  );
}

export function MarkSymbol({ mark }: { mark: MarkName }) {
  switch (mark) {
    case 'compare':
      return (
        <path
          d="M2 2 L12 2 L7 10 Z"
          className="av-fill-compare"
        />
      );
    case 'swap':
      return (
        <rect
          x="1"
          y="1"
          width="12"
          height="10"
          fill="url(#av-hatch-swap)"
          className="av-stroke-swap"
        />
      );
    case 'write':
      return (
        <circle
          cx="7"
          cy="6"
          r="4.5"
          className="av-fill-write"
        />
      );
    case 'sorted':
      return (
        <>
          <rect
            x="1"
            y="1"
            width="12"
            height="7"
            className="av-fill-sorted"
          />
          <rect
            x="1"
            y="9"
            width="12"
            height="2"
            className="av-fill-fg"
          />
        </>
      );
    case 'pivot':
      return (
        <path
          d="M7 1 L12 6 L7 11 L2 6 Z"
          className="av-fill-pivot"
        />
      );
    case 'min':
      return (
        <rect
          x="2"
          y="1.5"
          width="10"
          height="9"
          fill="none"
          strokeWidth="1.5"
          strokeDasharray="2 1.5"
          className="av-stroke-pivot"
        />
      );
    case 'range':
      return (
        <path
          d="M3 1 L1 1 L1 11 L3 11 M11 1 L13 1 L13 11 L11 11"
          fill="none"
          strokeWidth="1.5"
          className="av-stroke-fg"
        />
      );
    case 'found':
      return (
        <path
          d="M2 6 L5.5 9.5 L12 2.5"
          fill="none"
          strokeWidth="2"
          className="av-stroke-found"
        />
      );
    case 'visit':
      return (
        <rect
          x="1"
          y="1"
          width="12"
          height="10"
          fill="url(#av-dots-visit)"
          className="av-stroke-compare"
        />
      );
  }
}

/** A key to the symbols a view shows. */
export function Legend({ marks }: { marks: MarkName[] }) {
  return (
    <ul
      aria-label="Key"
      className="av-legend"
    >
      {marks.map((m) => (
        <li key={m}>
          <svg
            viewBox="0 0 14 12"
            aria-hidden
          >
            <Patterns />
            <MarkSymbol mark={m} />
          </svg>
          {labels[m]}
        </li>
      ))}
    </ul>
  );
}

/** A key for views that are not SVG: a swatch drawn by a CSS class. */
export function SwatchLegend({ items }: { items: { label: string; swatch: ReactNode }[] }) {
  return (
    <ul
      aria-label="Key"
      className="av-legend"
    >
      {items.map((it) => (
        <li key={it.label}>
          {it.swatch}
          {it.label}
        </li>
      ))}
    </ul>
  );
}
