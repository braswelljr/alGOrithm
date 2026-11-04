'use client';

import * as React from 'react';
import { Kbd, KbdGroup } from '@/components/ui/kbd';
import { detectPlatform, keyLabels, parseKeys, speakKeys, type Platform } from '@/lib/shortcuts';
import { cn } from '@/lib/utils';

/** The visitor's platform, after mount (the server renders the PC keys). */
export function usePlatformKeys(): Platform {
  const [p, setP] = React.useState<Platform>('pc');
  React.useEffect(() => setP(detectPlatform()), []);
  return p;
}

/**
 * A key binding drawn as keycaps: ⌘ K, G then L, 1 / 2 / 3 / 4. Screen
 * readers hear the spoken form ("Command or Control plus K").
 *
 *   <KeyCombo keys="mod+k" />
 */
export function KeyCombo({ keys, className }: { keys: string; className?: string }) {
  const platform = usePlatformKeys();
  const combos = parseKeys(keys);
  return (
    <span
      className={cn('inline-flex flex-wrap items-center gap-1', className)}
      aria-label={speakKeys(keys)}
      role="img"
    >
      {combos.map((combo, i) => (
        <React.Fragment key={i}>
          {i > 0 ? (
            <span
              aria-hidden
              className="text-muted text-xs"
            >
              /
            </span>
          ) : null}
          {combo.map((chord, j) => (
            <React.Fragment key={j}>
              {j > 0 ? (
                <span
                  aria-hidden
                  className="text-muted text-xs"
                >
                  then
                </span>
              ) : null}
              <KbdGroup aria-hidden>
                {keyLabels(chord, platform).map((k) => (
                  <Kbd key={k}>{k}</Kbd>
                ))}
              </KbdGroup>
            </React.Fragment>
          ))}
        </React.Fragment>
      ))}
    </span>
  );
}
