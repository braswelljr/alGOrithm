import type { ComponentProps } from 'react';
import { LOGO } from '@/lib/brand';
import { basePath } from '@/config/site';
import { cn } from '@/utils/utils';
import { ACTIVE_CONCEPT, CONCEPTS, markShapes, type ConceptId, type Palette } from './marks.ts';

/**
 * The site's logo, icon only. It draws ACTIVE_CONCEPT from marks.ts, so
 * switching the logo everywhere is that one line, then
 * `pnpm --dir web/apps/site icons` for the files in public/.
 */
export { ACTIVE_CONCEPT, CONCEPTS, type ConceptId };

/** The site's tokens, so the full-color mark themes with light and dark mode. */
const TOKENS: Palette = {
  primary: 'var(--accent)',
  secondary: 'var(--success)',
  muted: 'color-mix(in oklch, var(--accent) 55%, transparent)'
};

/** The one-color mark file, public/logos/mark-mono.svg, as a CSS mask. */
const MASK = `url("${basePath}${LOGO.mono.href}")`;

type LogoProps = Omit<ComponentProps<'span'>, 'children'> & {
  /** The accessible name; without one the mark is decoration and is hidden. */
  title?: string;
};

/**
 * The mark in one color, `currentColor`: public/logos/mark-mono.svg painted
 * through a CSS mask, so it takes the text color and the file stays the source.
 */
export function Logo({ title, className, style, ...props }: LogoProps) {
  const mask = MASK;
  return (
    <span
      role={title ? 'img' : undefined}
      aria-hidden={title ? undefined : true}
      aria-label={title}
      className={cn('inline-block size-6 shrink-0 bg-current', className)}
      style={{
        maskImage: mask,
        maskSize: '100% 100%',
        WebkitMaskImage: mask,
        WebkitMaskSize: '100% 100%',
        ...style
      }}
      {...props}
    />
  );
}

/**
 * The mark in full color from the site's accent tokens, inline, so it follows
 * the site's theme (which a visitor can switch) rather than only the system's,
 * as an <img> of public/logos/mark.svg would. No ids inside, so it can repeat.
 */
export function LogoColor({
  title,
  concept = ACTIVE_CONCEPT,
  className
}: {
  title?: string;
  concept?: ConceptId;
  className?: string;
}) {
  return (
    <svg
      viewBox="0 0 32 32"
      className={cn('size-6 shrink-0', className)}
      role={title ? 'img' : undefined}
      aria-hidden={title ? undefined : true}
      aria-label={title}
      dangerouslySetInnerHTML={{ __html: markShapes(concept, TOKENS) }}
    />
  );
}
