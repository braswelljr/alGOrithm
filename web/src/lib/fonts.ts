import { JetBrains_Mono, Roboto_Serif, Source_Sans_3 } from 'next/font/google';

/**
 * The site's three faces, as in gsl-certificates and gsl-special-courses, all
 * through next/font: self-hosted at build time (a visitor's browser never asks
 * Google for anything), variable, latin only, and with the system fallback
 * sized to their metrics (adjustFontFallback) so swapping them in shifts
 * nothing.
 *
 *   --font-sans    Source Sans 3 (formerly Source Sans Pro): body and UI text
 *   --font-serif   Roboto Serif: headings and display text (--font-heading)
 *   --font-code    JetBrains Mono: code, counts and ids (first in --font-mono)
 *
 * These variables are unlayered, so they win over the theme's layered
 * defaults in src/styles/theme.css. The OG images cannot read next/font's
 * files, so src/lib/og.tsx loads the same families from @fontsource.
 */
export const sans = Source_Sans_3({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-sans',
  adjustFontFallback: true,
  preload: true
});

export const serif = Roboto_Serif({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-serif',
  adjustFontFallback: true,
  preload: true
});

export const codeFont = JetBrains_Mono({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-code',
  preload: false
});

/** The classes for <html>: they define --font-sans, --font-serif and --font-code. */
export const fontVariables = `${sans.variable} ${serif.variable} ${codeFont.variable}`;
