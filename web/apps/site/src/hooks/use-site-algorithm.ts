'use client';

import { useEffect } from 'react';
import { useAlgorithm, type AlgorithmState } from '@braswelljr/algorithm-visualizer/react';
import { basePath, wasmOptions } from '@/config/site';

/** The algorithms, loaded from this site's copy of the wasm build. */
export function useSiteAlgorithm(): AlgorithmState {
  useWasmCache();
  return useAlgorithm(wasmOptions);
}

/**
 * Registers public/sw.js, which keeps the versioned wasm build in the Cache
 * API, the first time a page runs anything; production only, so development
 * always loads the fresh build. Best effort: a browser without service
 * workers just downloads the wasm as before.
 */
function useWasmCache() {
  useEffect(() => {
    if (process.env.NODE_ENV !== 'production' || !('serviceWorker' in navigator)) return;
    navigator.serviceWorker
      .register(`${basePath}/sw.js`, { scope: `${basePath}/` })
      .catch(() => {});
  }, []);
}
