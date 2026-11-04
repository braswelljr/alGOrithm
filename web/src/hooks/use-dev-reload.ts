'use client';

import { useEffect } from 'react';
import { z } from 'zod';
import { basePath } from '@/config/site';

const buildSchema = z.object({ build: z.number() });

/**
 * In development only: polls /dev-build.json, which the repository's dev tool
 * rewrites after each wasm build, and reloads the page when it changes, so a
 * change to the Go code shows up without a manual refresh.
 */
export function useDevReload() {
  useEffect(() => {
    if (process.env.NODE_ENV !== 'development') return;
    let last: number | null = null;
    let live = true;
    const poll = async () => {
      try {
        const res = await fetch(`${basePath}/dev-build.json`, { cache: 'no-store' });
        if (!res.ok) return;
        const parsed = buildSchema.safeParse(await res.json());
        if (!parsed.success) return;
        if (last !== null && parsed.data.build !== last && live) location.reload();
        last = parsed.data.build;
      } catch {
        // No dev tool running: nothing to watch.
      }
    };
    void poll();
    const t = setInterval(poll, 1000);
    return () => {
      live = false;
      clearInterval(t);
    };
  }, []);
}
