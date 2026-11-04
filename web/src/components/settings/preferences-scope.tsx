'use client';

import { MotionConfig } from 'motion/react';
import { usePreferencesStore } from '@/store/preferences';

/**
 * Applies a reader's display preferences to everything under it (ported from
 * gsl-special-courses' dashboard layout).
 *
 * Each preference is an attribute on one wrapper, which src/styles/preferences.css
 * keys the type scale, the density, the contrast and the underlined links off;
 * nothing changes a class. `contents`, so the wrapper takes no part in the
 * layout of what it holds. Reduced motion also goes to Motion, which otherwise
 * reads only the operating system's setting.
 */
export function PreferencesScope({ children }: { children: React.ReactNode }) {
  const { textSize, density, contrast, underlineLinks, motion, textSpacing, focus, boldText } =
    usePreferencesStore();
  return (
    <MotionConfig
      reducedMotion={motion === 'reduce' ? 'always' : motion === 'full' ? 'never' : 'user'}
    >
      <div
        data-preferences=""
        data-text-size={textSize}
        data-density={density}
        data-contrast={contrast}
        data-underline-links={underlineLinks ? 'true' : 'false'}
        data-motion={motion}
        data-text-spacing={textSpacing}
        data-focus={focus}
        data-bold-text={boldText ? 'true' : 'false'}
        className="contents"
      >
        {children}
      </div>
    </MotionConfig>
  );
}
