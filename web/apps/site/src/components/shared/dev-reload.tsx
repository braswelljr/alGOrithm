'use client';

import { useDevReload } from '@/hooks/use-dev-reload';

/** Reloads the page after each wasm rebuild, in development only. */
export function DevReload() {
  useDevReload();
  return null;
}
