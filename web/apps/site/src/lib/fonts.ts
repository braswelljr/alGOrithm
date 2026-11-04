import { JetBrains_Mono, Ysabeau_Office } from 'next/font/google';

/**
 * The site's fonts, all through next/font (self-hosted at build time, so a
 * visitor's browser never asks Google for anything), the same on every
 * platform so the layout is identical everywhere.
 *
 * Text: Ysabeau Office (OFL), a humanist sans with calligraphic stroke
 * modulation, gently flared terminals, open round bowls and a generous
 * x-height that keeps small UI text readable. Chosen over Signika (lower in
 * contrast, narrower) and Mukta (drawn for Devanagari first, more geometric).
 * next/font sizes the system fallback to its metrics (adjustFontFallback), so
 * swapping it in shifts nothing.
 *
 * Code: JetBrains Mono, as --font-code (theme.css puts it first in --font-mono).
 *
 * --font-sans set here is unlayered, so it wins over Tailwind's layered theme
 * default in packages/ui/src/styles/theme.css. The OG images cannot read these
 * files (next/font does not expose them to ImageResponse), so the OG renderer
 * loads Ysabeau Office's OFL .ttf on its own.
 */
export const sans = Ysabeau_Office({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-sans',
  adjustFontFallback: true,
  preload: true
});

export const codeFont = JetBrains_Mono({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-code',
  preload: false
});

/** The classes for <html>: they define --font-sans and --font-code. */
export const fontVariables = `${sans.variable} ${codeFont.variable}`;
