'use client';

import type { ReactNode } from 'react';
import { useSiteAlgorithm } from '@/hooks/use-site-algorithm';
import type { AlgorithmApi, Catalog } from '@/visualizer/core';

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
        className="text-muted py-16 text-center"
      >
        Loading the algorithms...
      </p>
    );
  }
  if (state.status === 'error') {
    return (
      <div
        role="alert"
        className="border-danger/40 bg-danger/5 rounded-lg border p-4"
      >
        <p className="text-danger font-medium">The algorithms could not be loaded.</p>
        <p className="text-muted mt-1 text-sm">{state.error}</p>
      </div>
    );
  }
  return <>{children(state.api, state.catalog)}</>;
}
