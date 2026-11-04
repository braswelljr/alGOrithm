'use client';

import type { ReactNode } from 'react';
import type { AlgorithmApi, Catalog } from '@braswelljr/algorithm-visualizer/core';
import { useSiteAlgorithm } from '@/hooks/use-site-algorithm';

/**
 * Shows its children once the wasm build is loaded, with the API and catalog.
 * loading replaces the default message while it loads; a page prerendered at
 * build time passes its content there, so it can be read before the wasm runs.
 */
export function WasmGate({
  children,
  loading
}: {
  children: (api: AlgorithmApi, catalog: Catalog) => ReactNode;
  loading?: ReactNode;
}) {
  const state = useSiteAlgorithm();
  if (state.status === 'loading') {
    if (loading) return <>{loading}</>;
    return (
      <p
        role="status"
        className="py-16 text-center text-muted"
      >
        Loading the algorithms...
      </p>
    );
  }
  if (state.status === 'error') {
    return (
      <div
        role="alert"
        className="rounded-lg border border-danger/40 bg-danger/5 p-4"
      >
        <p className="font-medium text-danger">The algorithms could not be loaded.</p>
        <p className="mt-1 text-sm text-muted">{state.error}</p>
      </div>
    );
  }
  return <>{children(state.api, state.catalog)}</>;
}
