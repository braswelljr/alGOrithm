'use client';

import { useSyncExternalStore } from 'react';

/** The family of computer the visitor is on, as far as the keyboard in front of them goes. */
export type Platform = 'apple' | 'pc';

const subscribe = () => () => {};

const APPLE = /Macintosh|Mac OS X|iPhone|iPad|iPod/;

/**
 * Whether the visitor is on an Apple device or anything else. The server
 * cannot know, so it renders as `apple` and the first client render corrects
 * it; nothing that changes the page's layout should hang on the answer.
 */
export function usePlatform(): Platform {
  return useSyncExternalStore(
    subscribe,
    () => (APPLE.test(navigator.userAgent) ? 'apple' : 'pc'),
    () => 'apple'
  );
}
