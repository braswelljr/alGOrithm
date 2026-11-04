/** The few icons the controls need, inline so the package has no icon dependency. */
import type { SVGProps } from 'react';

const base: SVGProps<SVGSVGElement> = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
  'aria-hidden': true,
  focusable: false
};

export const PlayIcon = () => (
  <svg {...base}>
    <path d="M6 4l14 8-14 8z" />
  </svg>
);
export const PauseIcon = () => (
  <svg {...base}>
    <path d="M7 4v16M17 4v16" />
  </svg>
);
export const StepBackIcon = () => (
  <svg {...base}>
    <path d="M18 5l-10 7 10 7zM6 5v14" />
  </svg>
);
export const StepForwardIcon = () => (
  <svg {...base}>
    <path d="M6 5l10 7-10 7zM18 5v14" />
  </svg>
);
export const SkipBackIcon = () => (
  <svg {...base}>
    <path d="M19 5l-7 7 7 7M12 5l-7 7 7 7" />
  </svg>
);
export const SkipForwardIcon = () => (
  <svg {...base}>
    <path d="M5 5l7 7-7 7M12 5l7 7-7 7" />
  </svg>
);
export const ReplayIcon = () => (
  <svg {...base}>
    <path d="M3 12a9 9 0 1 0 3-6.7L3 8M3 3v5h5" />
  </svg>
);
export const CopyIcon = () => (
  <svg {...base}>
    <rect
      x="9"
      y="9"
      width="12"
      height="12"
      rx="2"
    />
    <path d="M5 15V5a2 2 0 0 1 2-2h10" />
  </svg>
);
