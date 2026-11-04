/**
 * The motion tokens for JavaScript (Motion and GSAP), the same curves and
 * durations as the CSS variables in styles/theme.css, so everything moves
 * alike. Animate transform and opacity only.
 */
export const ease = {
  outStrong: [0.23, 1, 0.32, 1] as const,
  inOutStrong: [0.77, 0, 0.175, 1] as const,
  outExpo: [0.16, 1, 0.3, 1] as const
};

/** GSAP's names for the same curves (CustomEase is not needed for these). */
export const gsapEase = { out: 'expo.out', inOut: 'power3.inOut' } as const;

/** Seconds, for Motion and GSAP. */
export const duration = { fast: 0.12, base: 0.2, slow: 0.36 } as const;

/** The spring for things that move to a place: tab indicators, cards. */
export const spring = { type: 'spring', stiffness: 500, damping: 38, mass: 0.8 } as const;

/** A softer spring for layout changes and dialogs. */
export const softSpring = { type: 'spring', stiffness: 260, damping: 30 } as const;
