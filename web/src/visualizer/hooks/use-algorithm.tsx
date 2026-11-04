import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import type { AlgorithmApi } from '../core/api';
import { loadAlgorithm, type LoadOptions } from '../core/loader';
import type { Catalog } from '../core/schema';

export type AlgorithmState =
  | { status: 'loading' }
  | { status: 'ready'; api: AlgorithmApi; catalog: Catalog }
  | { status: 'error'; error: string };

const Context = createContext<LoadOptions | null>(null);

/** Says where the wasm build is served for the components inside it. */
export function AlgorithmProvider({ children, ...options }: LoadOptions & { children: ReactNode }) {
  return <Context.Provider value={options}>{children}</Context.Provider>;
}

/** Loads the algorithms (once for the whole page) and their catalog. */
export function useAlgorithm(options?: LoadOptions): AlgorithmState {
  const fromContext = useContext(Context);
  const opts = options ?? fromContext ?? {};
  const [state, setState] = useState<AlgorithmState>({ status: 'loading' });
  useEffect(() => {
    let live = true;
    loadAlgorithm(opts).then(
      (api) => {
        if (!live) return;
        try {
          setState({ status: 'ready', api, catalog: api.catalog() });
        } catch (err) {
          setState({ status: 'error', error: err instanceof Error ? err.message : String(err) });
        }
      },
      (err: unknown) =>
        live &&
        setState({ status: 'error', error: err instanceof Error ? err.message : String(err) })
    );
    return () => {
      live = false;
    };
  }, [opts.wasmUrl, opts.execUrl]);
  return state;
}
