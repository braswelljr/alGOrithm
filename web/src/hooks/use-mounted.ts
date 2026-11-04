'use client';

import { useSyncExternalStore } from 'react';

const subscribe = () => () => {};

/**
 * True once rendering on the client, false on the server and in the first,
 * hydrating render, so markup that depends on the browser (the stored theme)
 * matches on both sides and only then changes.
 */
export function useMounted(): boolean {
  return useSyncExternalStore(
    subscribe,
    () => true,
    () => false
  );
}
